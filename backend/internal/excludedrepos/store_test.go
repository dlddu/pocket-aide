package excludedrepos_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/excludedrepos"
)

func newStore(t *testing.T) (*excludedrepos.Store, []int64) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "ex.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	// Two users — exclusion behaviour only makes sense across multiple subjects.
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	return excludedrepos.New(conn), []int64{1, 2}
}

func TestAdd_ValidatesShape(t *testing.T) {
	store, _ := newStore(t)
	ctx := context.Background()

	for _, bad := range []string{
		"",
		"justrepo",
		"too/many/slashes",
		"with spaces/repo",
		"dlddu/repo!",
	} {
		_, err := store.Add(ctx, 1, bad)
		if !errors.Is(err, excludedrepos.ErrInvalidRepo) {
			t.Errorf("Add(%q): expected ErrInvalidRepo, got %v", bad, err)
		}
	}

	if _, err := store.Add(ctx, 1, "dlddu/pocket-aide"); err != nil {
		t.Errorf("Add valid name: %v", err)
	}
}

func TestAdd_DuplicateReturnsAlreadyExcluded(t *testing.T) {
	store, _ := newStore(t)
	ctx := context.Background()

	if _, err := store.Add(ctx, 1, "dlddu/repo"); err != nil {
		t.Fatalf("first add: %v", err)
	}
	if _, err := store.Add(ctx, 1, "dlddu/repo"); !errors.Is(err, excludedrepos.ErrAlreadyExcluded) {
		t.Errorf("dup add: expected ErrAlreadyExcluded, got %v", err)
	}
}

func TestList_UserScoped(t *testing.T) {
	store, _ := newStore(t)
	ctx := context.Background()

	_, _ = store.Add(ctx, 1, "dlddu/a")
	_, _ = store.Add(ctx, 1, "dlddu/b")
	_, _ = store.Add(ctx, 2, "dlddu/c")

	u1, err := store.List(ctx, 1)
	if err != nil {
		t.Fatalf("list u1: %v", err)
	}
	if len(u1) != 2 {
		t.Errorf("u1 list size: got %d want 2", len(u1))
	}
	for _, r := range u1 {
		if r.RepoFullName == "dlddu/c" {
			t.Errorf("u1 list leaked u2's repo: %+v", r)
		}
	}
}

func TestDelete_OtherUserNotFound(t *testing.T) {
	store, _ := newStore(t)
	ctx := context.Background()

	added, err := store.Add(ctx, 1, "dlddu/x")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := store.Delete(ctx, 2, added.ID); !errors.Is(err, excludedrepos.ErrNotFound) {
		t.Errorf("cross-user delete: expected ErrNotFound, got %v", err)
	}
	if err := store.Delete(ctx, 1, added.ID); err != nil {
		t.Errorf("owner delete: %v", err)
	}
}

func TestListUserIDsExcluding(t *testing.T) {
	store, users := newStore(t)
	ctx := context.Background()

	// u1 excludes pocket-aide; u2 does not.
	if _, err := store.Add(ctx, 1, "dlddu/pocket-aide"); err != nil {
		t.Fatalf("seed: %v", err)
	}

	got, err := store.ListUserIDsExcluding(ctx, "dlddu/pocket-aide")
	if err != nil {
		t.Fatalf("list matched: %v", err)
	}
	if len(got) != 1 || got[0] != 2 {
		t.Errorf("matched for pocket-aide: got %v want [2]", got)
	}

	got, err = store.ListUserIDsExcluding(ctx, "dlddu/some-other-repo")
	if err != nil {
		t.Fatalf("list matched (no exclusion): %v", err)
	}
	if len(got) != len(users) {
		t.Errorf("matched for some-other-repo: got %v want all users", got)
	}
}
