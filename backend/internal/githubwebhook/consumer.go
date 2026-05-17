// Package githubwebhook consumes GitHub webhook deliveries that have been
// forwarded to an SQS queue (typically via API Gateway → SQS integration).
// It verifies the X-Hub-Signature-256 HMAC, filters to workflow_run completed
// events, and hands the parsed event to a caller-supplied dispatch func.
//
// Idempotency, retries, and dedupe are intentionally out of scope — the draft
// only validates HMAC and forwards once.
package githubwebhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// WorkflowRunEvent is the subset of fields the dispatcher cares about.
type WorkflowRunEvent struct {
	Repo         string // e.g. "dlddu/pocket-aide"
	WorkflowName string
	HeadBranch   string
	Conclusion   string // success | failure | cancelled | ...
	HTMLURL      string
}

// DispatchFunc is invoked once per accepted (HMAC-verified, completed)
// workflow_run event. Its error is logged but never blocks message deletion —
// the draft favours throughput over guaranteed delivery.
type DispatchFunc func(ctx context.Context, evt WorkflowRunEvent) error

// Consumer long-polls a single SQS queue.
type Consumer struct {
	client   *sqs.Client
	queueURL string
	secret   []byte
	dispatch DispatchFunc
}

// New loads AWS config from the default credential chain (instance profile,
// AWS_* env vars, etc.) and returns a Consumer ready for Run. If roleARN is
// non-empty, the SQS client uses credentials obtained by assuming that role
// via STS — the default chain provides the base credentials that sign the
// AssumeRole call.
func New(ctx context.Context, queueURL, secret, roleARN string, dispatch DispatchFunc) (*Consumer, error) {
	if queueURL == "" {
		return nil, errors.New("queueURL is empty")
	}
	if secret == "" {
		return nil, errors.New("secret is empty")
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
		client:   sqs.NewFromConfig(awsCfg),
		queueURL: queueURL,
		secret:   []byte(secret),
		dispatch: dispatch,
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
		// Log and still delete — failures here are unrecoverable (bad
		// signature, malformed payload). Leaving the message would just
		// stall the queue.
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
	headers, payload, err := decodeAPIGatewayEnvelope(raw)
	if err != nil {
		return fmt.Errorf("envelope: %w", err)
	}
	if event := headers["x-github-event"]; event != "workflow_run" {
		// Not a workflow_run — the webhook is subscribed to events we don't
		// act on. Log once per message so operators can see what's filling
		// the queue without having to sample messages directly.
		log.Printf("githubwebhook: skipping event=%q (not workflow_run)", event)
		return nil
	}
	if err := verifyHMAC(c.secret, payload, headers["x-hub-signature-256"]); err != nil {
		return fmt.Errorf("hmac: %w", err)
	}
	var parsed workflowRunPayload
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return fmt.Errorf("parse workflow_run: %w", err)
	}
	if parsed.Action != "completed" {
		// GitHub sends requested/in_progress/completed for every run;
		// we only push on completed. Log so the requested/in_progress
		// volume is observable.
		log.Printf("githubwebhook: skipping workflow_run action=%q (not completed)", parsed.Action)
		return nil
	}
	evt := WorkflowRunEvent{
		Repo:         parsed.Repository.FullName,
		WorkflowName: parsed.WorkflowRun.Name,
		HeadBranch:   parsed.WorkflowRun.HeadBranch,
		Conclusion:   parsed.WorkflowRun.Conclusion,
		HTMLURL:      parsed.WorkflowRun.HTMLURL,
	}
	if err := c.dispatch(ctx, evt); err != nil {
		return fmt.Errorf("dispatch: %w", err)
	}
	log.Printf("githubwebhook: dispatched repo=%s branch=%s conclusion=%s", evt.Repo, evt.HeadBranch, evt.Conclusion)
	return nil
}

type workflowRunPayload struct {
	Action     string `json:"action"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	WorkflowRun struct {
		Name       string `json:"name"`
		HeadBranch string `json:"head_branch"`
		Conclusion string `json:"conclusion"`
		HTMLURL    string `json:"html_url"`
	} `json:"workflow_run"`
}
