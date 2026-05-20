// Package githubwebhook consumes GitHub webhook deliveries that have been
// forwarded into an SQS queue via the EventBridge → SQS target integration.
// EventBridge wraps each delivery in its standard envelope; this package
// unwraps the envelope, filters to workflow_run completed events, and hands
// the parsed event to a caller-supplied dispatch func.
//
// Message authenticity is enforced by the IAM/queue-policy boundary around
// the SQS queue, not by an HMAC. The EventBridge bus is the only producer
// expected to hold sqs:SendMessage on this queue; the GitHub-side webhook
// signature is consumed and discarded by EventBridge before forwarding.
//
// Retry semantics: a dispatch func that returns a non-nil error keeps the
// SQS message on the queue (SkipDeleteMessage on the receive cycle) so it
// will be re-delivered after the queue's VisibilityTimeout (30s — see PRD
// operational notes). A DLQ with maxReceiveCount=5 backstops malformed
// messages so the queue does not stall.
package githubwebhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// WorkflowRunEvent is the subset of fields the dispatcher cares about.
// Sourced from EventBridge `detail.workflow_run` / `detail.repository` /
// `detail.workflow_run.pull_requests[]`.
//
// PR fields (PRNumber/PRTitle/PRURL) are zero-valued when the workflow run
// is not associated with a pull request (e.g. a push to main triggered the
// workflow directly). Notifications still fire in that case — the iOS card
// falls back to "repo — conclusion · workflow_name".
type WorkflowRunEvent struct {
	Repo         string // e.g. "dlddu/pocket-aide"
	WorkflowName string // workflow name, e.g. "CI"
	HeadBranch   string
	HeadSHA      string // commit SHA — used by the iOS client to group PR-less rows (AC13)
	Conclusion   string // success | failure | cancelled | skipped | ...
	HTMLURL      string // run URL
	CommitURL    string // head commit URL on GitHub
	PRNumber     int    // 0 when no PR linked
	PRTitle      string // "" when no PR linked
	PRURL        string // "" when no PR linked
}

// DispatchFunc is invoked once per accepted (workflow_run completed) event.
// A non-nil error causes handleMessage to SKIP DeleteMessage, so the SQS
// message is re-delivered after VisibilityTimeout. Use that to signal
// "retry this event later" (e.g. the notification-history write failed).
// PRD-10 AC11 requires this — history must persist before pushes go out.
type DispatchFunc func(ctx context.Context, evt WorkflowRunEvent) error

// Consumer long-polls a single SQS queue.
//
// debugLogHeaders / debugLogBody are opt-in toggles read from
// DEBUG_LOG_ENVELOPE_HEADERS / DEBUG_LOG_ENVELOPE_BODY at construction time.
// When set, they cause silent-drop paths in process() to dump additional
// envelope detail — useful for diagnosing unexpected messages on the queue
// (test sends from the AWS console, non-EventBridge producers, etc.).
type Consumer struct {
	client          *sqs.Client
	queueURL        string
	dispatch        DispatchFunc
	debugLogHeaders bool
	debugLogBody    bool
}

// New loads AWS config from the default credential chain (instance profile,
// AWS_* env vars, etc.) and returns a Consumer ready for Run. If roleARN is
// non-empty, the SQS client uses credentials obtained by assuming that role
// via STS — the default chain provides the base credentials that sign the
// AssumeRole call.
func New(ctx context.Context, queueURL, roleARN string, dispatch DispatchFunc) (*Consumer, error) {
	if queueURL == "" {
		return nil, errors.New("queueURL is empty")
	}
	if dispatch == nil {
		return nil, errors.New("dispatch is nil")
	}
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}
	if roleARN != "" {
		provider := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(awsCfg), roleARN, func(o *stscreds.AssumeRoleOptions) {
			o.RoleSessionName = "pocket-aide-sqs"
		})
		awsCfg.Credentials = aws.NewCredentialsCache(provider)
	}
	return &Consumer{
		client:          sqs.NewFromConfig(awsCfg),
		queueURL:        queueURL,
		dispatch:        dispatch,
		debugLogHeaders: os.Getenv("DEBUG_LOG_ENVELOPE_HEADERS") == "1",
		debugLogBody:    os.Getenv("DEBUG_LOG_ENVELOPE_BODY") == "1",
	}, nil
}

// Run loops until ctx is cancelled, draining the queue with long polling.
func (c *Consumer) Run(ctx context.Context) {
	log.Printf("githubwebhook: SQS consumer starting (queue=%s)", c.queueURL)
	for {
		if ctx.Err() != nil {
			log.Printf("githubwebhook: consumer stopping: %v", ctx.Err())
			return
		}
		out, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20,
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("githubwebhook: receive error: %v", err)
			continue
		}
		for _, msg := range out.Messages {
			c.handleMessage(ctx, msg)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg types.Message) {
	body := []byte(aws.ToString(msg.Body))
	if err := c.process(ctx, body); err != nil {
		// Two kinds of failure land here:
		//   1) Malformed envelope/payload — replays would never succeed.
		//   2) Transient dispatch error (e.g. SQLite busy, history write
		//      failed) — should be retried.
		// We don't reliably distinguish them in code, so we take the safer
		// option for PRD-10 AC11: SKIP DeleteMessage on any process error.
		// SQS re-delivers after VisibilityTimeout (configured to 30s — see
		// PRD operational notes). Malformed messages will eventually land on
		// the DLQ after maxReceiveCount attempts, so they don't stall the
		// queue indefinitely.
		log.Printf("githubwebhook: process error, leaving message for retry: %v", err)
		return
	}
	if _, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	}); err != nil {
		log.Printf("githubwebhook: delete error: %v", err)
	}
}

func (c *Consumer) process(ctx context.Context, raw []byte) error {
	detailType, source, detail, err := decodeEventBridgeEnvelope(raw)
	if err != nil {
		return fmt.Errorf("envelope: %w", err)
	}
	if detailType != "workflow_run" {
		// Not a workflow_run — could be any other GitHub event forwarded
		// by the same bus, an AWS test message, etc. Log once so operators
		// can see what's filling the queue without sampling messages
		// directly. Source is included for the same reason.
		log.Printf("githubwebhook: skipping detail_type=%q source=%q (not workflow_run)", detailType, source)
		c.debugLogEnvelope(raw)
		return nil
	}
	var parsed workflowRunPayload
	if err := json.Unmarshal(detail, &parsed); err != nil {
		return fmt.Errorf("parse workflow_run: %w", err)
	}
	if parsed.Action != "completed" {
		// GitHub sends requested/in_progress/completed for every run;
		// we only push on completed. Log so the non-completed volume is
		// observable.
		log.Printf("githubwebhook: skipping workflow_run action=%q (not completed)", parsed.Action)
		c.debugLogEnvelope(raw)
		return nil
	}
	evt := WorkflowRunEvent{
		Repo:         parsed.Repository.FullName,
		WorkflowName: parsed.WorkflowRun.Name,
		HeadBranch:   parsed.WorkflowRun.HeadBranch,
		HeadSHA:      parsed.WorkflowRun.HeadSHA,
		Conclusion:   parsed.WorkflowRun.Conclusion,
		HTMLURL:      parsed.WorkflowRun.HTMLURL,
	}
	if parsed.WorkflowRun.HeadSHA != "" && parsed.Repository.HTMLURL != "" {
		evt.CommitURL = parsed.Repository.HTMLURL + "/commit/" + parsed.WorkflowRun.HeadSHA
	}
	if len(parsed.WorkflowRun.PullRequests) > 0 {
		pr := parsed.WorkflowRun.PullRequests[0]
		evt.PRNumber = pr.Number
		evt.PRTitle = pr.Title
		// GitHub doesn't include the PR HTML URL inside the workflow_run's
		// pull_requests[] entries (only the API url). Synthesize it from the
		// repo + number — stable and matches what a user would expect.
		if parsed.Repository.HTMLURL != "" && pr.Number > 0 {
			evt.PRURL = fmt.Sprintf("%s/pull/%d", parsed.Repository.HTMLURL, pr.Number)
		}
	}
	if err := c.dispatch(ctx, evt); err != nil {
		return fmt.Errorf("dispatch: %w", err)
	}
	log.Printf("githubwebhook: dispatched repo=%s workflow=%s branch=%s conclusion=%s",
		evt.Repo, evt.WorkflowName, evt.HeadBranch, evt.Conclusion)
	return nil
}

// debugLogEnvelope dumps extra envelope detail when the corresponding
// opt-in env var is set. Called from the silent-drop paths only — happy-path
// messages already log a structured "dispatched ..." line.
//
// For EventBridge envelopes the "header keys" we surface are the top-level
// fields of the envelope JSON itself (detail-type, source, detail, etc.).
// This is the analogue of header keys in the old API-Gateway-shaped envelope
// and is sufficient to identify whether a message came from EventBridge or
// from another producer.
//
// Body output is capped at maxBodyPrefix bytes and printed via %q so any
// non-UTF-8 bytes are safely escaped.
func (c *Consumer) debugLogEnvelope(raw []byte) {
	if c.debugLogHeaders {
		var top map[string]json.RawMessage
		if err := json.Unmarshal(raw, &top); err == nil {
			keys := make([]string, 0, len(top))
			for k := range top {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			log.Printf("githubwebhook: debug envelope_keys=%v", keys)
		} else {
			log.Printf("githubwebhook: debug envelope_keys=<unparseable: %v>", err)
		}
	}
	if c.debugLogBody {
		const maxBodyPrefix = 256
		prefix := raw
		truncated := false
		if len(prefix) > maxBodyPrefix {
			prefix = prefix[:maxBodyPrefix]
			truncated = true
		}
		log.Printf("githubwebhook: debug body_prefix=%q total_len=%d truncated=%t",
			prefix, len(raw), truncated)
	}
}

type workflowRunPayload struct {
	Action     string `json:"action"`
	Repository struct {
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
	WorkflowRun struct {
		Name         string                   `json:"name"` // workflow name, e.g. "CI"
		HeadBranch   string                   `json:"head_branch"`
		HeadSHA      string                   `json:"head_sha"`
		Conclusion   string                   `json:"conclusion"`
		HTMLURL      string                   `json:"html_url"`
		PullRequests []workflowRunPullRequest `json:"pull_requests"`
	} `json:"workflow_run"`
}

type workflowRunPullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"` // API url, not the html_url — we synthesize the html_url separately
}
