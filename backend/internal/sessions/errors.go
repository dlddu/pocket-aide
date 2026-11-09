package sessions

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ErrorCodeAuthPoolExhausted is the code the app sees when every login slot
// of the session's type is held: session-platform answers 503 for that and
// for a disabled feature alike, and only its body tells them apart.
const ErrorCodeAuthPoolExhausted = "auth_pool_exhausted"

// AuthPoolExhausted reports whether a platform answer is session-platform's
// "auth pool exhausted" (agent-auth-pool AC-N2), which a session create or a
// restore can meet.
func AuthPoolExhausted(status int, body []byte) bool {
	if status != http.StatusServiceUnavailable {
		return false
	}
	var resp struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &resp) != nil {
		return false
	}
	return strings.HasPrefix(resp.Error, "auth pool exhausted")
}
