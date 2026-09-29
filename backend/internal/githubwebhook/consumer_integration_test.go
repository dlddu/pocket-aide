//go:build integration

package githubwebhook

import (
	"context"
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

func sendNativeMessage(t *testing.T, ctx context.Context, client *sqs.Client, queueURL, eventType string, payload any) {
	t.Helper()
	_, err := client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:          aws.String(queueURL),
		MessageBody:       aws.String(string(mustMarshal(t, payload))),
		MessageAttributes: eventAttrs(eventType),
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

	dispatched := make(chan WorkflowRunEvent, 1)
	consumer, err := New(ctx, queueURL, "", func(_ context.Context, evt WorkflowRunEvent) error {
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

	sendNativeMessage(t, ctx, client, queueURL, "workflow_run", completedWorkflowRun())

	select {
	case evt := <-dispatched:
		if evt.Repo != "dlddu/pocket-aide" || evt.Conclusion != "failure" || evt.WorkflowName != "CI" {
			t.Errorf("unexpected event: %+v", evt)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("dispatch was not called within 30s")
	}

	cancel()
	wg.Wait()
}

func TestIntegration_ConsumerDropsNonWorkflowRunMessages(t *testing.T) {
	requireLocalStack(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, queueURL := createEphemeralQueue(t, ctx)

	dispatched := make(chan WorkflowRunEvent, 1)
	consumer, err := New(ctx, queueURL, "", func(_ context.Context, evt WorkflowRunEvent) error {
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

	sendNativeMessage(t, ctx, client, queueURL, "push", map[string]any{"ref": "refs/heads/main"})

	sendNativeMessage(t, ctx, client, queueURL, "", map[string]any{"foo": "bar"})

	select {
	case evt := <-dispatched:
		t.Fatalf("dispatch should not fire, got %+v", evt)
	case <-time.After(15 * time.Second):
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
