package githubwebhook

import (
	"encoding/json"
	"testing"
)

func TestDecodeEventBridgeEnvelope_WellFormed(t *testing.T) {
	raw, _ := json.Marshal(eventBridgeEnvelope{
		DetailType: "workflow_job",
		Source:     "github.webhooks",
		Detail:     json.RawMessage(`{"action":"completed","workflow_job":{"name":"lint"}}`),
	})
	dt, src, detail, err := decodeEventBridgeEnvelope(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dt != "workflow_job" {
		t.Errorf("detail-type: got %q want %q", dt, "workflow_job")
	}
	if src != "github.webhooks" {
		t.Errorf("source: got %q want %q", src, "github.webhooks")
	}
	if string(detail) != `{"action":"completed","workflow_job":{"name":"lint"}}` {
		t.Errorf("detail roundtrip mismatch: %s", detail)
	}
}

func TestDecodeEventBridgeEnvelope_MalformedJSON(t *testing.T) {
	if _, _, _, err := decodeEventBridgeEnvelope([]byte(`{not json`)); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDecodeEventBridgeEnvelope_EmptyFieldsAreNotErrors(t *testing.T) {
	// An envelope that decodes but is missing detail-type / detail should
	// NOT return an error here — the caller decides whether to drop it.
	raw, _ := json.Marshal(map[string]any{"some": "other-shape"})
	dt, src, detail, err := decodeEventBridgeEnvelope(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dt != "" || src != "" || len(detail) != 0 {
		t.Errorf("expected zero values, got dt=%q src=%q detail=%s", dt, src, detail)
	}
}
