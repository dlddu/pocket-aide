// Package githubwebhook consumes GitHub webhook deliveries that have been
// forwarded into an SQS queue via the EventBridge → SQS target integration.
// EventBridge wraps each delivery in its standard envelope; this package
// unwraps the envelope, filters to workflow_run requested/completed events,
// and hands the parsed event to a caller-supplied dispatch func.
//
// Message authenticity is enforced by the IAM/queue-policy boundary around
// the SQS queue, not by an HMAC. The EventBridge bus is the only producer
// expected to hold sqs:SendMessage on this queue; the GitHub-side webhook
// signature is consumed and discarded by EventBridge before forwarding.
//
// Idempotency, retries, and dedupe are intentionally out of scope — the draft
// validates the message shape and forwards once.
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
// Sourced from EventBridge `detail.workflow_run` / `detail.repository`.
type WorkflowRunEvent struct {
	Repo         string // e.g. "dlddu/pocket-aide"
	WorkflowName string // workflow name, e.g. "CI"
	HeadBranch   string
	Action       string // "requested" | "completed"
	Conclusion   string // success | failure | cancelled | skipped | ... ; empty for requested
	HTMLURL      string
}

// DispatchFunc is invoked once per accepted (workflow_run requested/completed) event.
// Its error is logged but never blocks message deletion — the draft favours
// throughput over guaranteed delivery.
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
		// Log and still delete — failures here are unrecoverable (malformed
		// envelope, malformed payload). Leaving the message would just stall
		// the queue.
		log.Printf("githubwebhook: dropping message: %v", err)
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
	// GitHub sends requested/in_progress/completed for every run. We push on
	// requested (run kicked off) and completed (final result); in_progress
	// fires per-job and would just spam.
	if parsed.Action != "requested" && parsed.Action != "completed" {
		log.Printf("githubwebhook: skipping workflow_run action=%q (not requested/completed)", parsed.Action)
		c.debugLogEnvelope(raw)
		return nil
	}
	evt := WorkflowRunEvent{
		Repo:         parsed.Repository.FullName,
		WorkflowName: parsed.WorkflowRun.Name,
		HeadBranch:   parsed.WorkflowRun.HeadBranch,
		Action:       parsed.Action,
		Conclusion:   parsed.WorkflowRun.Conclusion,
		HTMLURL:      parsed.WorkflowRun.HTMLURL,
	}
	if err := c.dispatch(ctx, evt); err != nil {
		return fmt.Errorf("dispatch: %w", err)
	}
	log.Printf("githubwebhook: dispatched repo=%s workflow=%s branch=%s action=%s conclusion=%s",
		evt.Repo, evt.WorkflowName, evt.HeadBranch, evt.Action, evt.Conclusion)
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
	} `json:"repository"`
	WorkflowRun struct {
		Name       string `json:"name"` // workflow name, e.g. "CI"
		HeadBranch string `json:"head_branch"`
		Conclusion string `json:"conclusion"`
		HTMLURL    string `json:"html_url"`
	} `json:"workflow_run"`
}
