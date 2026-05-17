package githubwebhook

import (
	"encoding/json"
	"fmt"
)

// eventBridgeEnvelope is the wrapper EventBridge publishes to its SQS target
// when a GitHub webhook (forwarded via the EventBridge GitHub integration)
// arrives. The fields we care about are detail-type (which GitHub event this
// is) and detail (the raw event payload). Other fields (id, time, source,
// resources, account, region, version) are present but ignored.
//
// See AWS EventBridge → SQS target docs and the GitHub partner event source.
type eventBridgeEnvelope struct {
	DetailType string          `json:"detail-type"`
	Source     string          `json:"source"`
	Detail     json.RawMessage `json:"detail"`
}

// decodeEventBridgeEnvelope unwraps the SQS message body and returns the
// EventBridge detail-type and the raw GitHub event payload from `detail`.
// Source is returned too so callers can log it for diagnosis.
//
// Returns an error only when the JSON cannot be parsed at all. Whether the
// detail-type is one we act on is the caller's decision (so unexpected
// producers can be silent-dropped without churning the error path).
func decodeEventBridgeEnvelope(raw []byte) (detailType, source string, detail []byte, err error) {
	var env eventBridgeEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return "", "", nil, fmt.Errorf("decode envelope: %w", err)
	}
	return env.DetailType, env.Source, []byte(env.Detail), nil
}
