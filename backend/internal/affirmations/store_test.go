package affirmations_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/affirmations"
	"github.com/dlddu/pocket-aide/backend/internal/db"
)

func TestStoreCRUDPerUserIsolation(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "s.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO users (oidc_sub) VALUES ('alice'), ('bob')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	store := affirmations.NewStore(conn)
	ctx := context.Background()

	aliceItem, err := store.Create(ctx, 1, "alice line", affirmations.PriorityHigh)
	if err != nil {
		t.Fatalf("create alice: %v", err)
	}
	if aliceItem.Text != "alice line" || aliceItem.Priority != affirmations.PriorityHigh {
		t.Fatalf("alice item mismatch: %+v", aliceItem)
	}
	if aliceItem.CreatedAt == 0 || aliceItem.UpdatedAt == 0 {
		t.Fatalf("timestamps not set: %+v", aliceItem)
	}

	if _, err := store.Create(ctx, 2, "bob line", affirmations.PriorityLow); err != nil {
		t.Fatalf("create bob: %v", err)
	}

	aliceList, err := store.List(ctx, 1)
	if err != nil {
		t.Fatalf("list alice: %v", err)
	}
	if len(aliceList) != 1 || aliceList[0].Text != "alice line" {
		t.Fatalf("alice list wrong: %+v", aliceList)
	}

	// Bob's row must be invisible to Alice via Get/Update/Delete.
	if _, err := store.Get(ctx, 1, aliceItem.ID+999); !errors.Is(err, affirmations.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for nonexistent id, got %v", err)
	}

	bobList, err := store.List(ctx, 2)
	if err != nil {
		t.Fatalf("list bob: %v", err)
	}
	if len(bobList) != 1 {
		t.Fatalf("bob list wrong: %+v", bobList)
	}
	bobID := bobList[0].ID

	if _, err := store.Update(ctx, 1, bobID, "alice hijack", affirmations.PriorityHigh); !errors.Is(err, affirmations.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on cross-user update, got %v", err)
	}
	if err := store.Delete(ctx, 1, bobID); !errors.Is(err, affirmations.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on cross-user delete, got %v", err)
	}

	updated, err := store.Update(ctx, 1, aliceItem.ID, "alice v2", affirmations.PriorityNormal)
	if err != nil {
		t.Fatalf("update alice: %v", err)
	}
	if updated.Text != "alice v2" || updated.Priority != affirmations.PriorityNormal {
		t.Fatalf("update did not apply: %+v", updated)
	}

	if err := store.Delete(ctx, 1, aliceItem.ID); err != nil {
		t.Fatalf("delete alice: %v", err)
	}
	finalList, err := store.List(ctx, 1)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(finalList) != 0 {
		t.Fatalf("expected empty list, got %+v", finalList)
	}
}

func TestPriorityValid(t *testing.T) {
	for _, p := range []affirmations.Priority{
		affirmations.PriorityHigh, affirmations.PriorityNormal, affirmations.PriorityLow,
	} {
		if !p.Valid() {
			t.Errorf("expected %q valid", p)
		}
	}
	for _, bad := range []affirmations.Priority{"", "HIGH", "medium", "urgent"} {
		if bad.Valid() {
			t.Errorf("expected %q invalid", bad)
		}
	}
}
