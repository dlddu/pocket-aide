package githubwebhook

import (
	"context"
	"encoding/json"
	"testing"
)

// makeEventBridgeMessage builds a faithful EventBridge → SQS message body
// with the given detail-type and detail (GitHub event) payload.
func makeEventBridgeMessage(t *testing.T, detailType string, detail any) []byte {
	t.Helper()
	detailRaw, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("marshal detail: %v", err)
	}
	raw, err := json.Marshal(eventBridgeEnvelope{
		DetailType: detailType,
		Source:     "github.webhooks",
		Detail:     detailRaw,
	})
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	return raw
}

func completedWorkflowJob() map[string]any {
	return map[string]any{
		"action": "completed",
		"repository": map[string]any{
			"full_name": "dlddu/pocket-aide",
		},
		"workflow_job": map[string]any{
			"name":          "lint",
			"workflow_name": "CI",
			"head_branch":   "main",
			"conclusion":    "failure",
			"html_url":      "https://github.com/dlddu/pocket-aide/actions/runs/1/job/1",
		},
	}
}

// recorderConsumer wires a Consumer with a dispatch func that records every
// invocation, so each test case can assert how many times — and with what
// payload — dispatch fired.
func recorderConsumer(t *testing.T) (*Consumer, *[]WorkflowJobEvent) {
	t.Helper()
	var got []WorkflowJobEvent
	c := &Consumer{
		dispatch: func(_ context.Context, evt WorkflowJobEvent) error {
			got = append(got, evt)
			return nil
		},
	}
	return c, &got
}

func TestProcess_HappyPath(t *testing.T) {
	c, got := recorderConsumer(t)
	raw := makeEventBridgeMessage(t, "workflow_job", completedWorkflowJob())

	if err := c.process(context.Background(), raw); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(*got) != 1 {
		t.Fatalf("dispatch invocations: got %d want 1", len(*got))
	}
	evt := (*got)[0]
	if evt.Repo != "dlddu/pocket-aide" || evt.WorkflowName != "CI" ||
		evt.JobName != "lint" || evt.HeadBranch != "main" ||
		evt.Conclusion != "failure" {
		t.Errorf("unexpected event: %+v", evt)
	}
}

func TestProcess_NonWorkflowJobDetailTypeSilentlyDropped(t *testing.T) {
	c, got := recorderConsumer(t)
	// Some other GitHub event forwarded by the same bus.
	raw := makeEventBridgeMessage(t, "push", map[string]any{"ref": "refs/heads/main"})

	if err := c.process(context.Background(), raw); err != nil {
		t.Fatalf("expected nil for non-workflow_job, got %v", err)
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on non-workflow_job detail-type")
	}
}

func TestProcess_EmptyDetailTypeSilentlyDropped(t *testing.T) {
	c, got := recorderConsumer(t)
	// Envelope decodes but has no detail-type — e.g. a non-EventBridge
	// producer published directly to the queue.
	raw, _ := json.Marshal(map[string]any{"hello": "world"})

	if err := c.process(context.Background(), raw); err != nil {
		t.Fatalf("expected nil for empty detail-type, got %v", err)
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on empty detail-type")
	}
}

func TestProcess_ActionOtherThanCompletedDropped(t *testing.T) {
	c, got := recorderConsumer(t)
	payload := completedWorkflowJob()
	payload["action"] = "queued"
	raw := makeEventBridgeMessage(t, "workflow_job", payload)

	if err := c.process(context.Background(), raw); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on action=queued")
	}
}

func TestProcess_MalformedEnvelope(t *testing.T) {
	c, got := recorderConsumer(t)
	if err := c.process(context.Background(), []byte("not json")); err == nil {
		t.Fatal("expected envelope error, got nil")
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on malformed envelope")
	}
}

func TestProcess_MalformedDetail(t *testing.T) {
	c, got := recorderConsumer(t)
	// Valid envelope, but detail isn't a workflow_job JSON object — it's a
	// JSON array. Unmarshal into workflowJobPayload will fail.
	raw, _ := json.Marshal(eventBridgeEnvelope{
		DetailType: "workflow_job",
		Source:     "github.webhooks",
		Detail:     json.RawMessage(`[1, 2, 3]`),
	})
	if err := c.process(context.Background(), raw); err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on malformed detail")
	}
}

func TestProcess_DispatchErrorPropagates(t *testing.T) {
	c := &Consumer{
		dispatch: func(_ context.Context, _ WorkflowJobEvent) error {
			return errBoom
		},
	}
	raw := makeEventBridgeMessage(t, "workflow_job", completedWorkflowJob())
	if err := c.process(context.Background(), raw); err == nil {
		t.Fatal("expected dispatch error to propagate, got nil")
	}
}

var errBoom = &boomError{}

type boomError struct{}

func (*boomError) Error() string { return "boom" }
