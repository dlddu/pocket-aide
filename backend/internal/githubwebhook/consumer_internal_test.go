package githubwebhook

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// mustMarshal JSON-encodes v or fails the test. Shared with the integration
// test, which builds the same native message bodies.
func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

// eventAttrs builds the SQS message-attribute map the API Gateway → SQS
// integration produces. An empty eventType yields nil (no attributes),
// simulating a producer other than the integration.
func eventAttrs(eventType string) map[string]types.MessageAttributeValue {
	if eventType == "" {
		return nil
	}
	return map[string]types.MessageAttributeValue{
		attrGitHubEvent: {
			DataType:    aws.String("String"),
			StringValue: aws.String(eventType),
		},
	}
}

// makeNativeMessage builds the SQS message the API Gateway → SQS integration
// delivers: the raw GitHub event JSON as the body, with the event type carried
// in the x-github-event message attribute.
func makeNativeMessage(t *testing.T, eventType string, payload any) types.Message {
	t.Helper()
	body := mustMarshal(t, payload)
	return types.Message{
		Body:              aws.String(string(body)),
		MessageAttributes: eventAttrs(eventType),
	}
}

func completedWorkflowRun() map[string]any {
	return map[string]any{
		"action": "completed",
		"repository": map[string]any{
			"full_name": "dlddu/pocket-aide",
			"html_url":  "https://github.com/dlddu/pocket-aide",
		},
		"workflow_run": map[string]any{
			"name":        "CI",
			"head_branch": "main",
			"head_sha":    "a3f9c27deadbeef",
			"conclusion":  "failure",
			"html_url":    "https://github.com/dlddu/pocket-aide/actions/runs/1",
		},
	}
}

func completedWorkflowRunWithPR(number int) map[string]any {
	p := completedWorkflowRun()
	wr := p["workflow_run"].(map[string]any)
	wr["pull_requests"] = []map[string]any{
		{"number": number, "title": "feat(x): hello"},
	}
	return p
}

// requestedWorkflowRun models a "CI 시작" event: action=requested, a null
// conclusion, and a non-terminal status.
func requestedWorkflowRun() map[string]any {
	p := completedWorkflowRun()
	p["action"] = "requested"
	wr := p["workflow_run"].(map[string]any)
	wr["status"] = "queued"
	delete(wr, "conclusion")
	return p
}

// recorderConsumer wires a Consumer with a dispatch func that records every
// invocation, so each test case can assert how many times — and with what
// payload — dispatch fired.
func recorderConsumer(t *testing.T) (*Consumer, *[]WorkflowRunEvent) {
	t.Helper()
	var got []WorkflowRunEvent
	c := &Consumer{
		dispatch: func(_ context.Context, evt WorkflowRunEvent) error {
			got = append(got, evt)
			return nil
		},
	}
	return c, &got
}

func TestProcess_HappyPath(t *testing.T) {
	c, got := recorderConsumer(t)
	msg := makeNativeMessage(t, "workflow_run", completedWorkflowRun())

	if err := c.process(context.Background(), msg); err != nil {
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
	if evt.CommitURL != "https://github.com/dlddu/pocket-aide/commit/a3f9c27deadbeef" {
		t.Errorf("commit url: got %q", evt.CommitURL)
	}
	if evt.HeadSHA != "a3f9c27deadbeef" {
		t.Errorf("head sha: got %q want %q", evt.HeadSHA, "a3f9c27deadbeef")
	}
	if evt.PRNumber != 0 || evt.PRTitle != "" || evt.PRURL != "" {
		t.Errorf("expected zero PR fields when pull_requests missing, got n=%d t=%q u=%q",
			evt.PRNumber, evt.PRTitle, evt.PRURL)
	}
}

func TestProcess_WithPullRequest(t *testing.T) {
	c, got := recorderConsumer(t)
	msg := makeNativeMessage(t, "workflow_run", completedWorkflowRunWithPR(42))

	if err := c.process(context.Background(), msg); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(*got) != 1 {
		t.Fatalf("dispatch invocations: got %d want 1", len(*got))
	}
	evt := (*got)[0]
	if evt.PRNumber != 42 {
		t.Errorf("pr number: got %d want 42", evt.PRNumber)
	}
	if evt.PRTitle != "feat(x): hello" {
		t.Errorf("pr title: got %q", evt.PRTitle)
	}
	if evt.PRURL != "https://github.com/dlddu/pocket-aide/pull/42" {
		t.Errorf("pr url: got %q", evt.PRURL)
	}
}

func TestProcess_NonWorkflowRunEventSilentlyDropped(t *testing.T) {
	c, got := recorderConsumer(t)
	// Some other GitHub event the same webhook forwards.
	msg := makeNativeMessage(t, "push", map[string]any{"ref": "refs/heads/main"})

	if err := c.process(context.Background(), msg); err != nil {
		t.Fatalf("expected nil for non-workflow_run, got %v", err)
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on non-workflow_run event")
	}
}

func TestProcess_MissingEventTypeSilentlyDropped(t *testing.T) {
	c, got := recorderConsumer(t)
	// No x-github-event attribute — e.g. a producer other than the API Gateway
	// integration published directly to the queue.
	msg := makeNativeMessage(t, "", map[string]any{"hello": "world"})

	if err := c.process(context.Background(), msg); err != nil {
		t.Fatalf("expected nil for missing event type, got %v", err)
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on missing event type")
	}
}

func TestProcess_RequestedStartEventDispatched(t *testing.T) {
	c, got := recorderConsumer(t)
	msg := makeNativeMessage(t, "workflow_run", requestedWorkflowRun())

	if err := c.process(context.Background(), msg); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(*got) != 1 {
		t.Fatalf("dispatch invocations: got %d want 1", len(*got))
	}
	// A requested run has a null conclusion; we normalize it to the run
	// status so the history row carries a non-empty in-progress marker.
	if evt := (*got)[0]; evt.Conclusion != "queued" {
		t.Errorf("start event conclusion: got %q want %q", evt.Conclusion, "queued")
	}
}

func TestProcess_RequestedWithoutStatusFallsBackToInProgress(t *testing.T) {
	c, got := recorderConsumer(t)
	payload := requestedWorkflowRun()
	delete(payload["workflow_run"].(map[string]any), "status")
	msg := makeNativeMessage(t, "workflow_run", payload)

	if err := c.process(context.Background(), msg); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(*got) != 1 {
		t.Fatalf("dispatch invocations: got %d want 1", len(*got))
	}
	if evt := (*got)[0]; evt.Conclusion != "in_progress" {
		t.Errorf("start event conclusion fallback: got %q want %q", evt.Conclusion, "in_progress")
	}
}

func TestProcess_InProgressActionDropped(t *testing.T) {
	c, got := recorderConsumer(t)
	payload := completedWorkflowRun()
	payload["action"] = "in_progress"
	msg := makeNativeMessage(t, "workflow_run", payload)

	if err := c.process(context.Background(), msg); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on action=in_progress")
	}
}

func TestProcess_MalformedBody(t *testing.T) {
	c, got := recorderConsumer(t)
	// x-github-event says workflow_run, but the body isn't valid JSON.
	msg := types.Message{
		Body:              aws.String("not json"),
		MessageAttributes: eventAttrs("workflow_run"),
	}
	if err := c.process(context.Background(), msg); err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on malformed body")
	}
}

func TestProcess_BodyNotObject(t *testing.T) {
	c, got := recorderConsumer(t)
	// Valid JSON, but a workflow_run body must be an object; an array fails to
	// unmarshal into workflowRunPayload.
	msg := types.Message{
		Body:              aws.String(`[1, 2, 3]`),
		MessageAttributes: eventAttrs("workflow_run"),
	}
	if err := c.process(context.Background(), msg); err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if len(*got) != 0 {
		t.Errorf("dispatch fired on non-object body")
	}
}

func TestProcess_DispatchErrorPropagates(t *testing.T) {
	c := &Consumer{
		dispatch: func(_ context.Context, _ WorkflowRunEvent) error {
			return errBoom
		},
	}
	msg := makeNativeMessage(t, "workflow_run", completedWorkflowRun())
	if err := c.process(context.Background(), msg); err == nil {
		t.Fatal("expected dispatch error to propagate, got nil")
	}
}

var errBoom = &boomError{}

type boomError struct{}

func (*boomError) Error() string { return "boom" }
