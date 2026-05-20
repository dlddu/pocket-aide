package notificationhistory_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/notificationhistory"
)

func newStore(t *testing.T) *notificationhistory.Store {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "nh.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return notificationhistory.New(conn)
}

func TestInsertBatchTx_FansOutAcrossUsers(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	evt := notificationhistory.Event{
		RepoFullName: "dlddu/pocket-aide",
		PRNumber:     42,
		PRTitle:      "feat(widget): rotation",
		PRURL:        "https://github.com/dlddu/pocket-aide/pull/42",
		CommitURL:    "https://github.com/dlddu/pocket-aide/commit/abc",
		RunURL:       "https://github.com/dlddu/pocket-aide/actions/runs/1",
		WorkflowName: "CI",
		HeadBranch:   "feature/x",
		HeadSHA:      "abc123",
		Conclusion:   "failure",
	}
	ids, err := store.InsertBatchTx(ctx, []int64{1, 2}, evt)
	if err != nil {
		t.Fatalf("insert batch: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("got %d ids, want 2", len(ids))
	}

	// Each user sees only their own row.
	for _, uid := range []int64{1, 2} {
		items, err := store.List(ctx, uid, 100, 0)
		if err != nil {
			t.Fatalf("list uid=%d: %v", uid, err)
		}
		if len(items) != 1 {
			t.Errorf("uid=%d list size: got %d want 1", uid, len(items))
		}
		if items[0].PRNumber == nil || *items[0].PRNumber != 42 {
			t.Errorf("uid=%d PR number lost: %+v", uid, items[0].PRNumber)
		}
		if items[0].HeadSHA != "abc123" {
			t.Errorf("uid=%d head_sha: got %q want %q", uid, items[0].HeadSHA, "abc123")
		}
	}
}

func TestInsertBatchTx_PRFieldsNilWhenEmpty(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	evt := notificationhistory.Event{
		RepoFullName: "dlddu/pocket-aide",
		// PRNumber/PRTitle/PRURL deliberately zero — main-direct-push case.
		CommitURL:    "https://example/commit",
		RunURL:       "https://example/run",
		WorkflowName: "CI",
		HeadBranch:   "main",
		Conclusion:   "success",
	}
	if _, err := store.InsertBatchTx(ctx, []int64{1}, evt); err != nil {
		t.Fatalf("insert: %v", err)
	}
	items, err := store.List(ctx, 1, 100, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if items[0].PRNumber != nil || items[0].PRTitle != nil || items[0].PRURL != nil {
		t.Errorf("expected nil PR fields, got %+v %+v %+v",
			items[0].PRNumber, items[0].PRTitle, items[0].PRURL)
	}
}

func TestAcknowledge_OwnRowSetsTimestamp(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	ids, err := store.InsertBatchTx(ctx, []int64{1}, notificationhistory.Event{
		RepoFullName: "dlddu/pocket-aide",
		Conclusion:   "success",
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := store.Acknowledge(ctx, 1, ids[0]); err != nil {
		t.Fatalf("ack: %v", err)
	}

	it, err := store.Get(ctx, 1, ids[0])
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if it.AcknowledgedAt == nil {
		t.Errorf("acknowledged_at not set after Acknowledge")
	}
}

func TestAcknowledge_OtherUserReturnsNotFound(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	ids, err := store.InsertBatchTx(ctx, []int64{1}, notificationhistory.Event{
		RepoFullName: "dlddu/pocket-aide",
		Conclusion:   "success",
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := store.Acknowledge(ctx, 2, ids[0]); !errors.Is(err, notificationhistory.ErrNotFound) {
		t.Errorf("cross-user ack: expected ErrNotFound, got %v", err)
	}
}

func TestAcknowledge_IsIdempotent(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	ids, err := store.InsertBatchTx(ctx, []int64{1}, notificationhistory.Event{
		RepoFullName: "dlddu/pocket-aide",
		Conclusion:   "failure",
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := store.Acknowledge(ctx, 1, ids[0]); err != nil {
		t.Fatalf("first ack: %v", err)
	}
	first, _ := store.Get(ctx, 1, ids[0])
	if err := store.Acknowledge(ctx, 1, ids[0]); err != nil {
		t.Fatalf("second ack: %v", err)
	}
	second, _ := store.Get(ctx, 1, ids[0])
	if first.AcknowledgedAt == nil || second.AcknowledgedAt == nil ||
		*first.AcknowledgedAt != *second.AcknowledgedAt {
		t.Errorf("re-ack should not change timestamp: first=%v second=%v",
			first.AcknowledgedAt, second.AcknowledgedAt)
	}
}

func TestList_KeysetPagination(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if _, err := store.InsertBatchTx(ctx, []int64{1}, notificationhistory.Event{
			RepoFullName: "dlddu/pocket-aide",
			Conclusion:   "success",
		}); err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}

	page1, err := store.List(ctx, 1, 2, 0)
	if err != nil {
		t.Fatalf("page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("page1 size: got %d want 2", len(page1))
	}

	page2, err := store.List(ctx, 1, 2, page1[1].ID)
	if err != nil {
		t.Fatalf("page2: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("page2 size: got %d want 2", len(page2))
	}
	if page2[0].ID >= page1[1].ID {
		t.Errorf("page2 first id %d should be < page1 last id %d",
			page2[0].ID, page1[1].ID)
	}
}
