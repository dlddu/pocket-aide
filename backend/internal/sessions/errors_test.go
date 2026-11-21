package sessions

import (
	"net/http"
	"testing"
)

func TestAuthPoolExhausted(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   bool
	}{
		{http.StatusServiceUnavailable, `{"error":"auth pool exhausted"}`, true},
		{http.StatusServiceUnavailable, `{"error":"auth pool exhausted: all 2 slots are held"}`, true},
		{http.StatusServiceUnavailable, `{"error":"archive disabled"}`, false},
		{http.StatusServiceUnavailable, `not json`, false},
		{http.StatusConflict, `{"error":"auth pool exhausted"}`, false},
	} {
		if got := AuthPoolExhausted(tc.status, []byte(tc.body)); got != tc.want {
			t.Errorf("%d %s = %v, want %v", tc.status, tc.body, got, tc.want)
		}
	}
}
