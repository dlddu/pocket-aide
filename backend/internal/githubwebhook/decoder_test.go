package githubwebhook

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestDecodeAPIGatewayEnvelope_PlainBody(t *testing.T) {
	raw, _ := json.Marshal(apiGatewayEnvelope{
		Headers: map[string]string{
			"X-GitHub-Event":      "workflow_run",
			"X-Hub-Signature-256": "sha256=abc",
		},
		Body:            `{"hello":"world"}`,
		IsBase64Encoded: false,
	})
	headers, body, err := decodeAPIGatewayEnvelope(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if headers["x-github-event"] != "workflow_run" {
		t.Errorf("expected lowercased header lookup, got %#v", headers)
	}
	if string(body) != `{"hello":"world"}` {
		t.Errorf("body roundtrip mismatch: %s", body)
	}
}

func TestDecodeAPIGatewayEnvelope_Base64Body(t *testing.T) {
	original := []byte(`{"hello":"world"}`)
	raw, _ := json.Marshal(apiGatewayEnvelope{
		Headers:         map[string]string{"X-GitHub-Event": "workflow_run"},
		Body:            base64.StdEncoding.EncodeToString(original),
		IsBase64Encoded: true,
	})
	_, body, err := decodeAPIGatewayEnvelope(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != string(original) {
		t.Errorf("base64 body mismatch: got=%s want=%s", body, original)
	}
}

func TestDecodeAPIGatewayEnvelope_BadBase64(t *testing.T) {
	raw, _ := json.Marshal(apiGatewayEnvelope{
		Body:            "not_base64!@#",
		IsBase64Encoded: true,
	})
	if _, _, err := decodeAPIGatewayEnvelope(raw); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDecodeAPIGatewayEnvelope_MalformedJSON(t *testing.T) {
	if _, _, err := decodeAPIGatewayEnvelope([]byte(`{not json`)); err == nil {
		t.Fatal("expected error, got nil")
	}
}
