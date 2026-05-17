//go:build integration

// Package-level integration tests that drive the real SQS path. Skipped by
// default (`//go:build integration`); CI activates them with the matching
// `-tags=integration` flag against a LocalStack container.
//
// Running locally:
//
//	docker run --rm -d -p 4566:4566 -e SERVICES=sqs --name ls localstack/localstack:3.8.1
//	AWS_ENDPOINT_URL=http://localhost:4566 AWS_REGION=us-east-1 \
//	  AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test \
//	  go test -tags=integration -race ./internal/githubwebhook/...
package githubwebhook

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func requireLocalStack(t *testing.T) {
	t.Helper()
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("AWS_ENDPOINT_URL not set; skipping LocalStack integration test")
	}
}

// createEphemeralQueue spins up a fresh queue per test so cases don't share
// state. The test is responsible for deleting it via t.Cleanup.
func createEphemeralQueue(t *testing.T, ctx context.Context) (*sqs.Client, string) {
	t.Helper()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("aws config: %v", err)
	}
	client := sqs.NewFromConfig(cfg)
	name := fmt.Sprintf("ghwebhook-it-%d", time.Now().UnixNano())
	create, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{QueueName: aws.String(name)})
	if err != nil {
		t.Fatalf("create queue: %v", err)
	}
	t.Cleanup(func() {
		_, _ = client.DeleteQueue(context.Background(), &sqs.DeleteQueueInput{QueueUrl: create.QueueUrl})
	})
	return client, aws.ToString(create.QueueUrl)
}

func sendEnvelope(t *testing.T, ctx context.Context, client *sqs.Client, queueURL string, body []byte) {
	t.Helper()
	_, err := client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
}

func TestIntegration_ConsumerDeliversValidMessage(t *testing.T) {
	requireLocalStack(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, queueURL := createEphemeralQueue(t, ctx)

	dispatched := make(chan WorkflowJobEvent, 1)
	consumer, err := New(ctx, queueURL, "", func(_ context.Context, evt WorkflowJobEvent) error {
		dispatched <- evt
		return nil
	})
	if err != nil {
		t.Fatalf("new consumer: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer.Run(ctx)
	}()

	sendEnvelope(t, ctx, client, queueURL, makeEventBridgeMessage(t, "workflow_job", completedWorkflowJob()))

	select {
	case evt := <-dispatched:
		if evt.Repo != "dlddu/pocket-aide" || evt.Conclusion != "failure" || evt.JobName != "lint" {
			t.Errorf("unexpected event: %+v", evt)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("dispatch was not called within 30s")
	}

	cancel()
	wg.Wait()
}

func TestIntegration_ConsumerDropsNonWorkflowJobMessages(t *testing.T) {
	requireLocalStack(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, queueURL := createEphemeralQueue(t, ctx)

	dispatched := make(chan WorkflowJobEvent, 1)
	consumer, err := New(ctx, queueURL, "", func(_ context.Context, evt WorkflowJobEvent) error {
		dispatched <- evt
		return nil
	})
	if err != nil {
		t.Fatalf("new consumer: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer.Run(ctx)
	}()

	// A detail-type we don't act on — must be silently dropped.
	sendEnvelope(t, ctx, client, queueURL, makeEventBridgeMessage(t, "push", map[string]any{"ref": "refs/heads/main"}))

	// A direct, non-EventBridge envelope — also silently dropped.
	bogus, _ := json.Marshal(map[string]any{"foo": "bar"})
	sendEnvelope(t, ctx, client, queueURL, bogus)

	select {
	case evt := <-dispatched:
		t.Fatalf("dispatch should not fire, got %+v", evt)
	case <-time.After(15 * time.Second):
		// Expected: nothing arrives.
	}

	cancel()
	wg.Wait()

	// Belt and suspenders: confirm the queue actually drained, so we know
	// the consumer processed (and dropped) the messages rather than them
	// sitting on the queue.
	out, err := client.GetQueueAttributes(context.Background(), &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(queueURL),
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameApproximateNumberOfMessages,
			types.QueueAttributeNameApproximateNumberOfMessagesNotVisible,
		},
	})
	if err != nil {
		t.Fatalf("get queue attrs: %v", err)
	}
	if visible := out.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessages)]; visible != "0" {
		t.Errorf("queue should be drained, visible=%s", visible)
	}
}
