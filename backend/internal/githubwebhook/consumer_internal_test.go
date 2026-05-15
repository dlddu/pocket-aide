package githubwebhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

// makeEnvelope builds a faithful API Gateway → SQS message body with the
// given GitHub event header and JSON payload. signWith is the HMAC secret
// the producer uses; pass a different value than the consumer's secret to
// exercise the rejection path.
func makeEnvelope(t *testing.T, event string, payload any, signWith []byte) []byte {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	mac := hmac.New(sha256.New, signWith)
	mac.Write(body)
	env := apiGatewayEnvelope{
		Headers: map[string]string{
			"X-GitHub-Event":      event,
			"X-Hub-Signature-256": "sha256=" + hex.EncodeToString(mac.Sum(nil)),
		},
		Body:            string(body),
		IsBase64Encoded: false,
	}
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	return raw
}

func completedWorkflowRun() map[string]any {
	return map[string]any{
		"action": "completed",
		"repository": map[string]any{
			"full_name": "dlddu/pocket-aide",
		},
		"workflow_run": map[string]any{
			"name":        "CI",
			"head_branch": "main",
			"conclusion":  "failure",
			"html_url":    "https://github.com/dlddu/pocket-aide/actions/runs/1",
		},
	}
}

// recorderConsumer wires a Consumer with a dispatch func that records every
// invocation, so each test case can assert how many times — and with what
// payload — dispatch fired.
func recorderConsumer(t *testing.T, secret string) (*Consumer, *[]WorkflowRunEvent) {
	t.Helper()
	var got []WorkflowRunEvent
	c := &Consumer{
		secret: []byte(secret),
		dispatch: func(_ context.Context, evt WorkflowRunEvent) error {
			got = append(got, evt)
			return nil
		},
	}
	return c, &got
}

func TestProcess_HappyPath(t *testing.T) {
	const secret = "shared-secret"
	c, got := recorderConsumer(t, secret)
	raw := makeEnvelope(t, "workflow_run", completedWorkflowRun(), []byte(secret))

	if err := c.process(context.Background(), raw); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(*got) != 1 {
		t.Fatalf("dispatch invocations: got %d want 1", len(*got))
	}
	evt := (*got)[0]
	if evt.Repo != "dlddu/pocket-aide" || evt.WorkflowName != "CI" ||
		evt.HeadBranch != "main" || evt.Conclusion != "failure" {
		t.Errorf("unexpected event: %+v", evt)
	}
}

func TestProcess_WrongSignatureRejected(t *testing.T) {
	c, got := recorderConsumer(t, "shared-secret")
	raw := makeEnvelope(t, "workflow_run", completedWorkflowRun(), []byte("other-secret"))

	if err := c.process(context.Background(), raw); err == nil {
		t.Fatal("expected hmac error, got nil")
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on bad signature: %d invocations", len(*got))
	}
}

func TestProcess_NonWorkflowRunSilentlyDropped(t *testing.T) {
	const secret = "shared-secret"
	c, got := recorderConsumer(t, secret)
	// "push" events arrive on the same webhook with a valid signature; we
	// just don't care about them. Should return nil so the message is
	// deleted from the queue.
	raw := makeEnvelope(t, "push", map[string]any{"ref": "refs/heads/main"}, []byte(secret))

	if err := c.process(context.Background(), raw); err != nil {
		t.Fatalf("expected nil for non-workflow_run, got %v", err)
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on non-workflow_run event")
	}
}

func TestProcess_ActionOtherThanCompletedDropped(t *testing.T) {
	const secret = "shared-secret"
	c, got := recorderConsumer(t, secret)
	payload := completedWorkflowRun()
	payload["action"] = "requested"
	raw := makeEnvelope(t, "workflow_run", payload, []byte(secret))

	if err := c.process(context.Background(), raw); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on action=requested")
	}
}

func TestProcess_MalformedEnvelope(t *testing.T) {
	c, got := recorderConsumer(t, "shared-secret")
	if err := c.process(context.Background(), []byte("not json")); err == nil {
		t.Fatal("expected envelope error, got nil")
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on malformed envelope")
	}
}

func TestProcess_MalformedInnerPayload(t *testing.T) {
	const secret = "shared-secret"
	c, got := recorderConsumer(t, secret)
	// Valid envelope, valid HMAC, but the body isn't a workflow_run JSON
	// object — it's a JSON array. parse will fail.
	rawBody := []byte(`[1, 2, 3]`)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	env, _ := json.Marshal(apiGatewayEnvelope{
		Headers: map[string]string{
			"X-GitHub-Event":      "workflow_run",
			"X-Hub-Signature-256": "sha256=" + hex.EncodeToString(mac.Sum(nil)),
		},
		Body: string(rawBody),
	})
	if err := c.process(context.Background(), env); err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on malformed payload")
	}
}

func TestProcess_DispatchErrorPropagates(t *testing.T) {
	const secret = "shared-secret"
	c := &Consumer{
		secret: []byte(secret),
		dispatch: func(_ context.Context, _ WorkflowRunEvent) error {
			return errBoom
		},
	}
	raw := makeEnvelope(t, "workflow_run", completedWorkflowRun(), []byte(secret))
	if err := c.process(context.Background(), raw); err == nil {
		t.Fatal("expected dispatch error to propagate, got nil")
	}
}

var errBoom = &boomError{}

type boomError struct{}

func (*boomError) Error() string { return "boom" }
