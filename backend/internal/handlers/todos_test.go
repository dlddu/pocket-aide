package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
	"github.com/dlddu/pocket-aide/backend/internal/todos"
)

func newTodoRouter(t *testing.T) http.Handler {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "todos.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	store := todos.New(conn)
	r := chi.NewRouter()
	r.Get("/api/todos/{area}", handlers.ListTodos(store))
	r.Post("/api/todos/{area}", handlers.CreateTodo(store))
	r.Patch("/api/todos/{area}/{id}", handlers.UpdateTodo(store))
	r.Delete("/api/todos/{area}/{id}", handlers.DeleteTodo(store))
	return r
}

func doTodo(t *testing.T, router http.Handler, method, target string, body any, userID int64) *httptest.ResponseRecorder {
	t.Helper()
	var raw []byte
	if body != nil {
		var err error
		if raw, err = json.Marshal(body); err != nil {
			t.Fatalf("marshal: %v", err)
		}
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(method, target, raw, userID))
	return rec
}

func createTodo(t *testing.T, router http.Handler, area string, body map[string]any, userID int64) todos.Todo {
	t.Helper()
	rec := doTodo(t, router, http.MethodPost, "/api/todos/"+area, body, userID)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %s: got %d want 201 (body=%s)", area, rec.Code, rec.Body.String())
	}
	var created todos.Todo
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return created
}

func listTodos(t *testing.T, router http.Handler, area string, userID int64) []todos.Todo {
	t.Helper()
	rec := doTodo(t, router, http.MethodGet, "/api/todos/"+area, nil, userID)
	if rec.Code != http.StatusOK {
		t.Fatalf("list %s: got %d want 200 (body=%s)", area, rec.Code, rec.Body.String())
	}
	var resp struct {
		Items []todos.Todo `json:"items"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp.Items
}

// PRD-3 AC1: an item added in one area never shows up in the other area's list.
func TestTodoAreasAreSeparateCollections(t *testing.T) {
	router := newTodoRouter(t)
	createTodo(t, router, "personal", map[string]any{"title": "치과 예약 잡기"}, 1)
	createTodo(t, router, "work", map[string]any{"title": "분기 리뷰 deck"}, 1)

	personal := listTodos(t, router, "personal", 1)
	if len(personal) != 1 || personal[0].Title != "치과 예약 잡기" {
		t.Errorf("personal list: %+v", personal)
	}
	work := listTodos(t, router, "work", 1)
	if len(work) != 1 || work[0].Title != "분기 리뷰 deck" {
		t.Errorf("work list: %+v", work)
	}
}

// PRD-3 AC2: add / edit / complete / delete with title, memo, due date and an
// optional priority, behaving identically in both areas.
func TestTodoCRUDInBothAreas(t *testing.T) {
	for _, area := range []string{"personal", "work"} {
		t.Run(area, func(t *testing.T) {
			router := newTodoRouter(t)
			created := createTodo(t, router, area, map[string]any{
				"title": "  retention fact-check  ", "memo": "메모", "due_date": "2026-10-02", "priority": "high",
			}, 1)
			if created.Title != "retention fact-check" || created.Memo != "메모" ||
				created.DueDate == nil || *created.DueDate != "2026-10-02" ||
				created.Priority == nil || *created.Priority != todos.PriorityHigh || created.CompletedAt != nil {
				t.Fatalf("unexpected created: %+v", created)
			}

			path := "/api/todos/" + area + "/" + strconv.FormatInt(created.ID, 10)
			rec := doTodo(t, router, http.MethodPatch, path, map[string]any{"title": "retention fact-check", "done": true}, 1)
			if rec.Code != http.StatusOK {
				t.Fatalf("complete: got %d (body=%s)", rec.Code, rec.Body.String())
			}
			var done todos.Todo
			_ = json.NewDecoder(rec.Body).Decode(&done)
			if done.CompletedAt == nil || done.DueDate != nil || done.Priority != nil || done.Memo != "" {
				t.Errorf("complete should set completed_at and overwrite the rest: %+v", done)
			}

			rec = doTodo(t, router, http.MethodPatch, path, map[string]any{"title": "retention fact-check", "done": true}, 1)
			var again todos.Todo
			_ = json.NewDecoder(rec.Body).Decode(&again)
			if again.CompletedAt == nil || *again.CompletedAt != *done.CompletedAt {
				t.Errorf("re-saving a done todo must keep completed_at: %v vs %v", again.CompletedAt, done.CompletedAt)
			}

			rec = doTodo(t, router, http.MethodPatch, path, map[string]any{"title": "reopened", "done": false}, 1)
			var reopened todos.Todo
			_ = json.NewDecoder(rec.Body).Decode(&reopened)
			if reopened.CompletedAt != nil || reopened.Title != "reopened" {
				t.Errorf("reopen: %+v", reopened)
			}

			rec = doTodo(t, router, http.MethodDelete, path, nil, 1)
			if rec.Code != http.StatusNoContent {
				t.Fatalf("delete: got %d", rec.Code)
			}
			if items := listTodos(t, router, area, 1); len(items) != 0 {
				t.Errorf("list after delete: %+v", items)
			}
		})
	}
}

// PRD-3 AC4: there is no way to move an item across areas — an id from one
// area does not resolve in the other, and the body cannot name an area.
func TestTodoCannotCrossAreas(t *testing.T) {
	router := newTodoRouter(t)
	created := createTodo(t, router, "personal", map[string]any{"title": "책장 정리"}, 1)
	workPath := "/api/todos/work/" + strconv.FormatInt(created.ID, 10)

	if rec := doTodo(t, router, http.MethodPatch, workPath, map[string]any{"title": "x"}, 1); rec.Code != http.StatusNotFound {
		t.Errorf("patch via other area: got %d want 404", rec.Code)
	}
	if rec := doTodo(t, router, http.MethodDelete, workPath, nil, 1); rec.Code != http.StatusNotFound {
		t.Errorf("delete via other area: got %d want 404", rec.Code)
	}
	personalPath := "/api/todos/personal/" + strconv.FormatInt(created.ID, 10)
	if rec := doTodo(t, router, http.MethodPatch, personalPath, map[string]any{"title": "x", "area": "work"}, 1); rec.Code != http.StatusBadRequest {
		t.Errorf("area in body: got %d want 400", rec.Code)
	}
	if items := listTodos(t, router, "personal", 1); len(items) != 1 || items[0].Title != "책장 정리" {
		t.Errorf("personal item must be untouched: %+v", items)
	}
}

func TestTodoUserScopingAndValidation(t *testing.T) {
	router := newTodoRouter(t)
	created := createTodo(t, router, "work", map[string]any{"title": "1:1 일정"}, 1)
	if items := listTodos(t, router, "work", 2); len(items) != 0 {
		t.Errorf("user 2 must not see user 1's todos: %+v", items)
	}
	path := "/api/todos/work/" + strconv.FormatInt(created.ID, 10)
	if rec := doTodo(t, router, http.MethodDelete, path, nil, 2); rec.Code != http.StatusNotFound {
		t.Errorf("cross-user delete: got %d want 404", rec.Code)
	}

	cases := []struct {
		name string
		path string
		body map[string]any
		want int
	}{
		{"unknown area", "/api/todos/shared", map[string]any{"title": "x"}, http.StatusNotFound},
		{"empty title", "/api/todos/work", map[string]any{"title": "   "}, http.StatusBadRequest},
		{"bad due date", "/api/todos/work", map[string]any{"title": "x", "due_date": "10/02"}, http.StatusBadRequest},
		{"bad priority", "/api/todos/work", map[string]any{"title": "x", "priority": "urgent"}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		if rec := doTodo(t, router, http.MethodPost, tc.path, tc.body, 1); rec.Code != tc.want {
			t.Errorf("%s: got %d want %d", tc.name, rec.Code, tc.want)
		}
	}
	if rec := doTodo(t, router, http.MethodGet, "/api/todos/shared", nil, 1); rec.Code != http.StatusNotFound {
		t.Errorf("list unknown area: got %d want 404", rec.Code)
	}
}

func TestTodoListOrdering(t *testing.T) {
	router := newTodoRouter(t)
	undated := createTodo(t, router, "personal", map[string]any{"title": "undated"}, 1)
	later := createTodo(t, router, "personal", map[string]any{"title": "later", "due_date": "2026-10-09"}, 1)
	sooner := createTodo(t, router, "personal", map[string]any{"title": "sooner", "due_date": "2026-10-01"}, 1)
	done := createTodo(t, router, "personal", map[string]any{"title": "done", "due_date": "2026-09-01", "done": true}, 1)

	got := listTodos(t, router, "personal", 1)
	want := []int64{sooner.ID, later.ID, undated.ID, done.ID}
	if len(got) != len(want) {
		t.Fatalf("len: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i] {
			t.Errorf("position %d: got %q", i, got[i].Title)
		}
	}
}
