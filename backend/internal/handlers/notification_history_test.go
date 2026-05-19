package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
	"github.com/dlddu/pocket-aide/backend/internal/notificationhistory"
)

func newHistoryStore(t *testing.T) *notificationhistory.Store {
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

func chiRouterForHistory(store *notificationhistory.Store) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/notification-history", handlers.ListNotificationHistory(store))
	r.Post("/api/notification-history/{id}/ack", handlers.AcknowledgeNotification(store))
	return r
}

func TestNotificationHistory_ListIsUserScoped(t *testing.T) {
	store := newHistoryStore(t)
	router := chiRouterForHistory(store)

	if _, err := store.InsertBatchTx(context.Background(), []int64{1}, notificationhistory.Event{
		RepoFullName: "dlddu/pocket-aide", Conclusion: "success",
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodGet, "/api/notification-history", nil, 2))
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}
	var got struct {
		Items []notificationhistory.Item `json:"items"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if len(got.Items) != 0 {
		t.Errorf("u2 should see 0 items, got %d", len(got.Items))
	}
}

func TestNotificationHistory_AckOwnRow(t *testing.T) {
	store := newHistoryStore(t)
	router := chiRouterForHistory(store)

	ids, err := store.InsertBatchTx(context.Background(), []int64{1}, notificationhistory.Event{
		RepoFullName: "dlddu/pocket-aide", Conclusion: "failure",
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(
		http.MethodPost,
		"/api/notification-history/"+strconv.FormatInt(ids[0], 10)+"/ack",
		nil, 1,
	))
	if rec.Code != http.StatusNoContent {
		t.Errorf("ack: got %d want 204 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestNotificationHistory_AckOtherUserReturns404(t *testing.T) {
	store := newHistoryStore(t)
	router := chiRouterForHistory(store)

	ids, err := store.InsertBatchTx(context.Background(), []int64{1}, notificationhistory.Event{
		RepoFullName: "dlddu/pocket-aide", Conclusion: "failure",
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(
		http.MethodPost,
		"/api/notification-history/"+strconv.FormatInt(ids[0], 10)+"/ack",
		nil, 2,
	))
	if rec.Code != http.StatusNotFound {
		t.Errorf("cross-user ack: got %d want 404", rec.Code)
	}
}

func TestNotificationHistory_ListSupportsBeforeAndLimit(t *testing.T) {
	store := newHistoryStore(t)
	router := chiRouterForHistory(store)

	for i := 0; i < 5; i++ {
		if _, err := store.InsertBatchTx(context.Background(), []int64{1}, notificationhistory.Event{
			RepoFullName: "dlddu/pocket-aide", Conclusion: "success",
		}); err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodGet, "/api/notification-history?limit=2", nil, 1))
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	var page1 struct {
		Items []notificationhistory.Item `json:"items"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&page1)
	if len(page1.Items) != 2 {
		t.Fatalf("page1 size: got %d want 2", len(page1.Items))
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(
		http.MethodGet,
		"/api/notification-history?limit=2&before="+strconv.FormatInt(page1.Items[1].ID, 10),
		nil, 1,
	))
	var page2 struct {
		Items []notificationhistory.Item `json:"items"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&page2)
	if len(page2.Items) != 2 {
		t.Errorf("page2 size: got %d want 2", len(page2.Items))
	}
	if len(page2.Items) > 0 && page2.Items[0].ID >= page1.Items[1].ID {
		t.Errorf("page2 first id %d should be < page1 last id %d",
			page2.Items[0].ID, page1.Items[1].ID)
	}
}
