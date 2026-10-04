package notificationsettings_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/notificationsettings"
)

func newStore(t *testing.T) *notificationsettings.Store {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "ns.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	return notificationsettings.New(conn)
}

func boolPtr(b bool) *bool { return &b }

func outcomesPtr(o notificationsettings.Outcomes) *notificationsettings.Outcomes { return &o }

func TestGet_DefaultsWhenNoRow(t *testing.T) {
	store := newStore(t)
	got, err := store.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != notificationsettings.Default() {
		t.Errorf("got %+v want %+v", got, notificationsettings.Default())
	}
	if !got.Enabled || got.Outcomes != notificationsettings.OutcomesBoth {
		t.Errorf("default must be enabled + both, got %+v", got)
	}
}

func TestUpdate_PartialPatchKeepsOtherField(t *testing.T) {
	store := newStore(t)
	ctx := context.Background()

	got, err := store.Update(ctx, 1, notificationsettings.Patch{Outcomes: outcomesPtr(notificationsettings.OutcomesFailure)})
	if err != nil {
		t.Fatalf("update outcomes: %v", err)
	}
	if !got.Enabled || got.Outcomes != notificationsettings.OutcomesFailure {
		t.Errorf("after outcomes patch: %+v", got)
	}

	got, err = store.Update(ctx, 1, notificationsettings.Patch{Enabled: boolPtr(false)})
	if err != nil {
		t.Fatalf("update enabled: %v", err)
	}
	if got.Enabled || got.Outcomes != notificationsettings.OutcomesFailure {
		t.Errorf("after enabled patch: %+v", got)
	}

	reloaded, err := store.Get(ctx, 1)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if reloaded != got {
		t.Errorf("reloaded %+v want %+v", reloaded, got)
	}

	other, err := store.Get(ctx, 2)
	if err != nil {
		t.Fatalf("get other: %v", err)
	}
	if other != notificationsettings.Default() {
		t.Errorf("user 2 must keep defaults, got %+v", other)
	}
}

func TestUpdate_RejectsUnknownOutcomes(t *testing.T) {
	store := newStore(t)
	_, err := store.Update(context.Background(), 1, notificationsettings.Patch{Outcomes: outcomesPtr("cancelled")})
	if !errors.Is(err, notificationsettings.ErrInvalidOutcomes) {
		t.Errorf("expected ErrInvalidOutcomes, got %v", err)
	}
}

func TestAllowsPush(t *testing.T) {
	cases := []struct {
		name       string
		settings   notificationsettings.Settings
		conclusion string
		want       bool
	}{
		{"both success", notificationsettings.Settings{Enabled: true, Outcomes: notificationsettings.OutcomesBoth}, "success", true},
		{"both failure", notificationsettings.Settings{Enabled: true, Outcomes: notificationsettings.OutcomesBoth}, "failure", true},
		{"both cancelled", notificationsettings.Settings{Enabled: true, Outcomes: notificationsettings.OutcomesBoth}, "cancelled", true},
		{"success-only success", notificationsettings.Settings{Enabled: true, Outcomes: notificationsettings.OutcomesSuccess}, "success", true},
		{"success-only failure", notificationsettings.Settings{Enabled: true, Outcomes: notificationsettings.OutcomesSuccess}, "failure", false},
		{"success-only cancelled", notificationsettings.Settings{Enabled: true, Outcomes: notificationsettings.OutcomesSuccess}, "cancelled", false},
		{"failure-only failure", notificationsettings.Settings{Enabled: true, Outcomes: notificationsettings.OutcomesFailure}, "failure", true},
		{"failure-only timed_out", notificationsettings.Settings{Enabled: true, Outcomes: notificationsettings.OutcomesFailure}, "timed_out", true},
		{"failure-only startup_failure", notificationsettings.Settings{Enabled: true, Outcomes: notificationsettings.OutcomesFailure}, "startup_failure", true},
		{"failure-only success", notificationsettings.Settings{Enabled: true, Outcomes: notificationsettings.OutcomesFailure}, "success", false},
		{"failure-only cancelled", notificationsettings.Settings{Enabled: true, Outcomes: notificationsettings.OutcomesFailure}, "cancelled", false},
		{"disabled success", notificationsettings.Settings{Enabled: false, Outcomes: notificationsettings.OutcomesBoth}, "success", false},
		{"disabled failure", notificationsettings.Settings{Enabled: false, Outcomes: notificationsettings.OutcomesFailure}, "failure", false},
	}
	for _, tc := range cases {
		if got := tc.settings.AllowsPush(tc.conclusion); got != tc.want {
			t.Errorf("%s: AllowsPush(%q) = %v want %v", tc.name, tc.conclusion, got, tc.want)
		}
	}
}
