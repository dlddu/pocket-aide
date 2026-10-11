package handlers_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type historyBody struct {
	Items      []appApprovalBody `json:"items"`
	NextCursor *string           `json:"next_cursor"`
}

func (e *approvalEnv) history(sub, cursor string) historyBody {
	e.t.Helper()
	target := "/api/approvals/history"
	if cursor != "" {
		target += "?cursor=" + url.QueryEscape(cursor)
	}
	rec := e.do(http.MethodGet, target, "oidc-"+sub, nil)
	expectCode(e.t, rec, http.StatusOK, "list history")
	return decodeApproval[historyBody](e.t, rec)
}

func (e *approvalEnv) decide(sub, id, verb string) {
	e.t.Helper()
	expectCode(e.t, e.do(http.MethodPost, "/api/approvals/"+id+"/"+verb, "oidc-"+sub, nil), http.StatusOK, verb+" "+id)
}

func TestApprovalHistoryShowsProcessingModeNewestFirst(t *testing.T) {
	env := newApprovalEnv(t)
	k := env.issue("a", "infra-ci")

	manual := env.create(k.Key, map[string]any{"externalId": "m", "context": "수동"})
	env.advance(time.Second)
	env.decide("a", manual.ID, "approve")

	env.advance(time.Second)
	auto := env.create(k.Key, map[string]any{"externalId": "auto", "context": "자동"})
	if _, err := env.conn.Exec(`UPDATE approval_requests SET status = 'REJECTED', decision_mode = 'auto', processed_at = created_at WHERE id = ?`, auto.ID); err != nil {
		t.Fatalf("mark auto: %v", err)
	}

	env.advance(time.Second)
	expiring := env.create(k.Key, map[string]any{"externalId": "e", "context": "만료", "timeoutSeconds": 5})
	waiting := env.create(k.Key, map[string]any{"externalId": "w", "context": "대기"})
	env.advance(6 * time.Second)

	got := env.history("a", "")
	if got.NextCursor != nil {
		t.Fatalf("history fits one page but has next_cursor %q", *got.NextCursor)
	}
	want := []struct{ id, status, mode string }{
		{expiring.ID, "EXPIRED", "expired"},
		{auto.ID, "REJECTED", "auto"},
		{manual.ID, "APPROVED", "manual"},
	}
	if len(got.Items) != len(want) {
		t.Fatalf("history: got %+v, want %d processed items", got.Items, len(want))
	}
	for i, w := range want {
		it := got.Items[i]
		if it.ID != w.id || it.Status != w.status || it.DecisionMode != w.mode || it.KeyName != "infra-ci" || it.ProcessedAt == nil {
			t.Fatalf("history[%d]: got %+v, want id=%s status=%s mode=%s", i, it, w.id, w.status, w.mode)
		}
	}
	if *got.Items[0].ProcessedAt != *expiring.ExpiresAt {
		t.Fatalf("expired item processed_at %d, want expires_at %d", *got.Items[0].ProcessedAt, *expiring.ExpiresAt)
	}
	for _, it := range got.Items {
		if it.ID == waiting.ID {
			t.Fatalf("pending request %s is in the history", waiting.ID)
		}
	}
	if other := env.history("b", ""); len(other.Items) != 0 {
		t.Fatalf("user b sees a's history: %+v", other.Items)
	}
}

func TestApprovalHistoryCursorHasNoGapsOrDuplicates(t *testing.T) {
	env := newApprovalEnv(t)
	k := env.issue("a", "infra-ci")
	const total = 47
	var ids []string
	for i := 0; i < total; i++ {
		c := env.create(k.Key, map[string]any{"externalId": fmt.Sprintf("r-%d", i), "context": "c"})
		ids = append(ids, c.ID)
	}
	for i, id := range ids {
		if i%5 == 0 {
			env.advance(time.Second)
		}
		verb := "approve"
		if i%2 == 1 {
			verb = "reject"
		}
		env.decide("a", id, verb)
	}

	seen := map[string]bool{}
	var order []string
	cursor := ""
	pages := 0
	for {
		page := env.history("a", cursor)
		pages++
		for _, it := range page.Items {
			if seen[it.ID] {
				t.Fatalf("duplicate %s on page %d", it.ID, pages)
			}
			seen[it.ID] = true
			order = append(order, it.ID)
		}
		if page.NextCursor == nil {
			break
		}
		cursor = *page.NextCursor
		if pages > total {
			t.Fatalf("cursor does not terminate")
		}
	}
	if len(order) != total {
		t.Fatalf("history returned %d items over %d pages, want %d", len(order), pages, total)
	}
	if pages != 3 {
		t.Fatalf("pages = %d, want 3 for %d items", pages, total)
	}
	var prev int64 = 1 << 62
	prevID := "~"
	for _, id := range order {
		var ts int64
		if err := env.conn.QueryRow(`SELECT processed_at FROM approval_requests WHERE id = ?`, id).Scan(&ts); err != nil {
			t.Fatalf("select processed_at: %v", err)
		}
		if ts > prev || (ts == prev && id >= prevID) {
			t.Fatalf("history not newest first at %s (%d after %d)", id, ts, prev)
		}
		prev, prevID = ts, id
	}

	expectCode(t, env.do(http.MethodGet, "/api/approvals/history?cursor=bogus", "oidc-a", nil), http.StatusBadRequest, "bad cursor")
}
