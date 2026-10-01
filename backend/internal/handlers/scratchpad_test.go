package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/affirmations"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
	"github.com/dlddu/pocket-aide/backend/internal/routines"
	"github.com/dlddu/pocket-aide/backend/internal/scratchpad"
	"github.com/dlddu/pocket-aide/backend/internal/todos"
)

func newScratchpadRouter(t *testing.T) http.Handler {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "scratchpad.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	store := scratchpad.New(conn)
	todoStore := todos.New(conn)
	affStore := affirmations.New(conn)
	routineStore := routines.New(conn)
	r := chi.NewRouter()
	r.Get("/api/scratchpad", handlers.ListScratchpad(store))
	r.Post("/api/scratchpad", handlers.CreateScratchpadItem(store))
	r.Delete("/api/scratchpad/{id}", handlers.DeleteScratchpadItem(store))
	r.Post("/api/scratchpad/{id}/move", handlers.MoveScratchpadItem(store, todoStore, affStore, routineStore))
	r.Get("/api/todos/{area}", handlers.ListTodos(todoStore))
	r.Get("/api/affirmations", handlers.ListAffirmations(affStore))
	r.Get("/api/routines", handlers.ListRoutines(routineStore))
	return r
}

func createScratch(t *testing.T, router http.Handler, body map[string]any, userID int64) scratchpad.Item {
	t.Helper()
	rec := doTodo(t, router, http.MethodPost, "/api/scratchpad", body, userID)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d want 201 (body=%s)", rec.Code, rec.Body.String())
	}
	var created scratchpad.Item
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return created
}

func listScratch(t *testing.T, router http.Handler, userID int64) []scratchpad.Item {
	t.Helper()
	rec := doTodo(t, router, http.MethodGet, "/api/scratchpad", nil, userID)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var body struct {
		Items []scratchpad.Item `json:"items"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body.Items
}

func moveScratch(t *testing.T, router http.Handler, id int64, target string, userID int64) *httptest.ResponseRecorder {
	t.Helper()
	return doTodo(t, router, http.MethodPost, "/api/scratchpad/"+strconv.FormatInt(id, 10)+"/move", map[string]any{"target": target}, userID)
}

func TestScratchpadCreateRecordsMetadata(t *testing.T) {
	router := newScratchpadRouter(t)

	typed := createScratch(t, router, map[string]any{"text": "  엄마 생신 선물  "}, 1)
	if typed.Text != "엄마 생신 선물" || typed.Source != scratchpad.SourceText {
		t.Fatalf("typed item: got %+v, want trimmed text with source=text", typed)
	}
	if typed.CapturedAt == 0 {
		t.Fatalf("typed item: captured_at not set")
	}

	shortcut := createScratch(t, router, map[string]any{"text": "retention fact-check", "source": "shortcut", "captured_at": 1700000000}, 1)
	if shortcut.Source != scratchpad.SourceShortcut || shortcut.CapturedAt != 1700000000 {
		t.Fatalf("shortcut item: got %+v", shortcut)
	}

	items := listScratch(t, router, 1)
	if len(items) != 2 || items[0].ID != typed.ID || items[1].ID != shortcut.ID {
		t.Fatalf("list order: got %+v, want newest capture first", items)
	}
}

func TestScratchpadValidation(t *testing.T) {
	router := newScratchpadRouter(t)
	cases := []map[string]any{
		{"text": "   "},
		{"text": "x", "source": "email"},
		{"text": "x", "area": "personal"},
	}
	for _, body := range cases {
		rec := doTodo(t, router, http.MethodPost, "/api/scratchpad", body, 1)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("POST %v: got %d want 400", body, rec.Code)
		}
	}
	item := createScratch(t, router, map[string]any{"text": "x"}, 1)
	if rec := moveScratch(t, router, item.ID, "chat", 1); rec.Code != http.StatusBadRequest {
		t.Errorf("move to chat: got %d want 400", rec.Code)
	}
	if n := len(listScratch(t, router, 1)); n != 1 {
		t.Errorf("rejected move must keep the item: got %d items", n)
	}
}

func TestScratchpadMoveToTodoAreas(t *testing.T) {
	router := newScratchpadRouter(t)
	for _, area := range []string{"personal", "work"} {
		item := createScratch(t, router, map[string]any{"text": area + " 메모"}, 1)
		rec := moveScratch(t, router, item.ID, area, 1)
		if rec.Code != http.StatusOK {
			t.Fatalf("move to %s: got %d (body=%s)", area, rec.Code, rec.Body.String())
		}
		var resp struct {
			Target string      `json:"target"`
			Todo   *todos.Todo `json:"todo"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Target != area || resp.Todo == nil || resp.Todo.Title != area+" 메모" {
			t.Fatalf("move to %s: response %+v", area, resp)
		}
		got := listTodos(t, router, area, 1)
		if len(got) != 1 || got[0].ID != resp.Todo.ID {
			t.Fatalf("%s list after move: got %+v", area, got)
		}
	}
	if n := len(listScratch(t, router, 1)); n != 0 {
		t.Fatalf("moved items must leave the scratchpad: %d remain", n)
	}
}

func TestScratchpadMoveToAffirmation(t *testing.T) {
	router := newScratchpadRouter(t)
	item := createScratch(t, router, map[string]any{"text": "compounding은 매일 1%"}, 1)
	rec := moveScratch(t, router, item.ID, "affirmation", 1)
	if rec.Code != http.StatusOK {
		t.Fatalf("move: got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		Affirmation *affirmations.Affirmation `json:"affirmation"`
		Todo        *todos.Todo               `json:"todo"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Todo != nil || resp.Affirmation == nil || resp.Affirmation.Text != item.Text || resp.Affirmation.Priority != affirmations.PriorityNormal {
		t.Fatalf("response: %+v", resp)
	}
	list := doTodo(t, router, http.MethodGet, "/api/affirmations", nil, 1)
	var body struct {
		Items []affirmations.Affirmation `json:"items"`
	}
	if err := json.NewDecoder(list.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].ID != resp.Affirmation.ID {
		t.Fatalf("affirmations after move: %+v", body.Items)
	}
	if n := len(listScratch(t, router, 1)); n != 0 {
		t.Fatalf("moved item must leave the scratchpad: %d remain", n)
	}
}

func TestScratchpadMoveToRoutine(t *testing.T) {
	router := newScratchpadRouter(t)
	item := createScratch(t, router, map[string]any{"text": "저녁 정리"}, 1)
	rec := moveScratch(t, router, item.ID, "routine", 1)
	if rec.Code != http.StatusOK {
		t.Fatalf("move to routine: got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var resp struct {
		Target  string            `json:"target"`
		Routine *routines.Routine `json:"routine"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Target != "routine" || resp.Routine == nil || resp.Routine.Name != item.Text || resp.Routine.Cadence != routines.CadenceDaily || len(resp.Routine.Steps) != 0 {
		t.Fatalf("move response: %+v", resp)
	}
	if n := len(listScratch(t, router, 1)); n != 0 {
		t.Fatalf("moved item must leave the scratchpad: got %d items", n)
	}
	list := doTodo(t, router, http.MethodGet, "/api/routines", nil, 1)
	var body struct {
		Items []routines.Routine `json:"items"`
	}
	if err := json.NewDecoder(list.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].ID != resp.Routine.ID {
		t.Fatalf("routines after move: %+v", body.Items)
	}
}

func TestScratchpadUserScoping(t *testing.T) {
	router := newScratchpadRouter(t)
	item := createScratch(t, router, map[string]any{"text": "mine"}, 1)
	if n := len(listScratch(t, router, 2)); n != 0 {
		t.Fatalf("other user sees %d items", n)
	}
	if rec := moveScratch(t, router, item.ID, "personal", 2); rec.Code != http.StatusNotFound {
		t.Fatalf("other user move: got %d want 404", rec.Code)
	}
	if got := listTodos(t, router, "personal", 2); len(got) != 0 {
		t.Fatalf("failed move created a todo for user 2: %+v", got)
	}
	target := "/api/scratchpad/" + strconv.FormatInt(item.ID, 10)
	if rec := doTodo(t, router, http.MethodDelete, target, nil, 2); rec.Code != http.StatusNotFound {
		t.Fatalf("other user delete: got %d want 404", rec.Code)
	}
	if rec := doTodo(t, router, http.MethodDelete, target, nil, 1); rec.Code != http.StatusNoContent {
		t.Fatalf("owner delete: got %d want 204", rec.Code)
	}
	if n := len(listScratch(t, router, 1)); n != 0 {
		t.Fatalf("deleted item still listed")
	}
}
