package githubwebhook

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// apiGatewayEnvelope is the standard AWS API Gateway → SQS proxy shape:
// the headers map and the raw HTTP body (optionally base64-encoded).
type apiGatewayEnvelope struct {
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
}

// decodeAPIGatewayEnvelope unwraps the SQS message body and returns the
// (case-normalised) headers and the raw GitHub webhook payload.
func decodeAPIGatewayEnvelope(raw []byte) (map[string]string, []byte, error) {
	var env apiGatewayEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, nil, fmt.Errorf("decode envelope: %w", err)
	}
	body := []byte(env.Body)
	if env.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(env.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("base64 body: %w", err)
		}
		body = decoded
	}
	headers := normaliseHeaders(env.Headers)
	return headers, body, nil
}

// normaliseHeaders lower-cases every key so callers can lookup with a fixed
// case. API Gateway's casing of "X-Hub-Signature-256" varies.
func normaliseHeaders(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[lowercase(k)] = v
	}
	return out
}

func lowercase(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
