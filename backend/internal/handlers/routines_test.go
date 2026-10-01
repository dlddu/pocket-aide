package handlers_test

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
	"github.com/dlddu/pocket-aide/backend/internal/routines"
)

func newRoutinesRouter(t *testing.T) http.Handler {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "routines.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	store := routines.New(conn)
	r := chi.NewRouter()
	r.Get("/api/routines", handlers.ListRoutines(store))
	r.Post("/api/routines", handlers.CreateRoutine(store))
	r.Delete("/api/routines/{id}", handlers.DeleteRoutine(store))
	r.Post("/api/routines/{id}/steps", handlers.AddRoutineStep(store))
	r.Delete("/api/routines/{id}/steps/{stepID}", handlers.DeleteRoutineStep(store))
	r.Get("/api/routines/days/{day}", handlers.ListRoutineDay(store))
	r.Patch("/api/routines/{id}/days/{day}/steps/{stepID}", handlers.SetRoutineStepCheck(store))
	r.Get("/api/routines/{id}/history/{day}", handlers.RoutineHistory(store))
	return r
}

func routinePath(id int64, rest string) string {
	return "/api/routines/" + strconv.FormatInt(id, 10) + rest
}

func decodeInto[T any](t *testing.T, rec interface{ Result() *http.Response }, want int) T {
	t.Helper()
	res := rec.Result()
	defer func() { _ = res.Body.Close() }()
	var out T
	if res.StatusCode != want {
		t.Fatalf("status: got %d want %d", res.StatusCode, want)
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return out
}

func createRoutine(t *testing.T, router http.Handler, body map[string]any, userID int64) routines.Routine {
	t.Helper()
	rec := doTodo(t, router, http.MethodPost, "/api/routines", body, userID)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create routine: got %d (body=%s)", rec.Code, rec.Body.String())
	}
	return decodeInto[routines.Routine](t, rec, http.StatusCreated)
}

type dayResponse struct {
	Day   string                `json:"day"`
	Items []routines.DayRoutine `json:"items"`
}

func checkStep(t *testing.T, router http.Handler, id, stepID int64, day string, checked bool) routines.DayRoutine {
	t.Helper()
	rec := doTodo(t, router, http.MethodPatch, routinePath(id, "/days/"+day+"/steps/"+strconv.FormatInt(stepID, 10)), map[string]any{"checked": checked}, 1)
	if rec.Code != http.StatusOK {
		t.Fatalf("check step: got %d (body=%s)", rec.Code, rec.Body.String())
	}
	return decodeInto[routines.DayRoutine](t, rec, http.StatusOK)
}

func TestRoutineCreateKeepsStructure(t *testing.T) {
	router := newRoutinesRouter(t)
	created := createRoutine(t, router, map[string]any{
		"name":      " 아침 루틴 ",
		"cadence":   "daily",
		"start_day": "2026-09-01",
		"steps":     []string{"물 한 컵", "스트레칭 5분", "  ", "아침 산책", "일기 한 줄", "다짐 듣기"},
	}, 1)
	if created.Name != "아침 루틴" || created.Cadence != routines.CadenceDaily || len(created.Steps) != 5 {
		t.Fatalf("created: %+v", created)
	}

	added := doTodo(t, router, http.MethodPost, routinePath(created.ID, "/steps"), map[string]any{"title": "샤워"}, 1)
	withStep := decodeInto[routines.Routine](t, added, http.StatusCreated)
	if len(withStep.Steps) != 6 || withStep.Steps[5].Title != "샤워" {
		t.Fatalf("after add step: %+v", withStep.Steps)
	}
	del := doTodo(t, router, http.MethodDelete, routinePath(created.ID, "/steps/"+strconv.FormatInt(withStep.Steps[0].ID, 10)), nil, 1)
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete step: got %d", del.Code)
	}

	list := decodeInto[struct {
		Items []routines.Routine `json:"items"`
	}](t, doTodo(t, router, http.MethodGet, "/api/routines", nil, 1), http.StatusOK)
	if len(list.Items) != 1 {
		t.Fatalf("list: %+v", list.Items)
	}
	titles := []string{}
	for _, s := range list.Items[0].Steps {
		titles = append(titles, s.Title)
	}
	want := []string{"스트레칭 5분", "아침 산책", "일기 한 줄", "다짐 듣기", "샤워"}
	if len(titles) != len(want) {
		t.Fatalf("steps after re-entry: %v", titles)
	}
	for i := range want {
		if titles[i] != want[i] {
			t.Fatalf("steps after re-entry: %v, want %v", titles, want)
		}
	}
}

func TestRoutineValidation(t *testing.T) {
	router := newRoutinesRouter(t)
	cases := []map[string]any{
		{"name": "  ", "cadence": "daily"},
		{"name": "x", "cadence": "hourly"},
		{"name": "x", "cadence": "weekdays"},
		{"name": "x", "cadence": "weekly", "weekdays": 3},
		{"name": "x", "cadence": "monthly", "month_day": 0},
		{"name": "x", "cadence": "daily", "start_day": "10/01"},
		{"name": "x", "cadence": "daily", "color": "green"},
	}
	for _, body := range cases {
		if rec := doTodo(t, router, http.MethodPost, "/api/routines", body, 1); rec.Code != http.StatusBadRequest {
			t.Errorf("POST %v: got %d want 400", body, rec.Code)
		}
	}
	if rec := doTodo(t, router, http.MethodGet, "/api/routines/days/2026-13-01", nil, 1); rec.Code != http.StatusBadRequest {
		t.Errorf("bad day: got %d want 400", rec.Code)
	}
}

func TestRoutineDayShowsOnlyScheduledRoutines(t *testing.T) {
	router := newRoutinesRouter(t)
	createRoutine(t, router, map[string]any{"name": "아침", "cadence": "daily", "start_day": "2026-09-01", "steps": []string{"a"}}, 1)
	createRoutine(t, router, map[string]any{"name": "주간 회고", "cadence": "weekly", "weekdays": routines.WeekdayBit(0), "start_day": "2026-09-01", "steps": []string{"a"}}, 1)
	createRoutine(t, router, map[string]any{"name": "월말 정산", "cadence": "monthly", "month_day": 31, "start_day": "2026-09-01", "steps": []string{"a"}}, 1)
	createRoutine(t, router, map[string]any{"name": "내일부터", "cadence": "daily", "start_day": "2026-10-05", "steps": []string{"a"}}, 1)

	names := func(day string) []string {
		resp := decodeInto[dayResponse](t, doTodo(t, router, http.MethodGet, "/api/routines/days/"+day, nil, 1), http.StatusOK)
		out := []string{}
		for _, r := range resp.Items {
			out = append(out, r.Name)
		}
		return out
	}
	if got := names("2026-09-30"); len(got) != 2 || got[0] != "아침" || got[1] != "월말 정산" {
		t.Errorf("wednesday 09-30: %v", got)
	}
	if got := names("2026-10-04"); len(got) != 2 || got[0] != "아침" || got[1] != "주간 회고" {
		t.Errorf("sunday 10-04: %v", got)
	}
	if got := names("2026-10-05"); len(got) != 2 || got[1] != "내일부터" {
		t.Errorf("monday 10-05: %v", got)
	}
}

func TestRoutineChecksCompleteTheDay(t *testing.T) {
	router := newRoutinesRouter(t)
	r := createRoutine(t, router, map[string]any{"name": "저녁 정리", "cadence": "daily", "start_day": "2026-09-01", "steps": []string{"책상", "옷", "책"}}, 1)
	day := "2026-10-01"

	v := checkStep(t, router, r.ID, r.Steps[0].ID, day, true)
	if v.Done != 1 || v.Total != 3 || v.Completed {
		t.Fatalf("after one check: %+v", v)
	}
	checkStep(t, router, r.ID, r.Steps[0].ID, day, true)
	checkStep(t, router, r.ID, r.Steps[1].ID, day, true)
	v = checkStep(t, router, r.ID, r.Steps[2].ID, day, true)
	if v.Done != 3 || !v.Completed {
		t.Fatalf("all steps checked: %+v", v)
	}
	v = checkStep(t, router, r.ID, r.Steps[1].ID, day, false)
	if v.Done != 2 || v.Completed || v.Steps[1].Checked {
		t.Fatalf("after uncheck: %+v", v)
	}

	resp := decodeInto[dayResponse](t, doTodo(t, router, http.MethodGet, "/api/routines/days/2026-10-02", nil, 1), http.StatusOK)
	if len(resp.Items) != 1 || resp.Items[0].Done != 0 {
		t.Fatalf("checks must be per day: %+v", resp.Items)
	}
}

func TestRoutineCheckRejectsUnscheduledDay(t *testing.T) {
	router := newRoutinesRouter(t)
	r := createRoutine(t, router, map[string]any{"name": "주간 회고", "cadence": "weekly", "weekdays": routines.WeekdayBit(0), "start_day": "2026-09-01", "steps": []string{"a"}}, 1)
	rec := doTodo(t, router, http.MethodPatch, routinePath(r.ID, "/days/2026-10-01/steps/"+strconv.FormatInt(r.Steps[0].ID, 10)), map[string]any{"checked": true}, 1)
	if rec.Code != http.StatusConflict {
		t.Fatalf("unscheduled day: got %d want 409", rec.Code)
	}
	rec = doTodo(t, router, http.MethodPatch, routinePath(r.ID, "/days/2026-10-04/steps/"+strconv.FormatInt(r.Steps[0].ID, 10)), map[string]any{}, 1)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing checked: got %d want 400", rec.Code)
	}
}

func TestRoutineHistoryByDay(t *testing.T) {
	router := newRoutinesRouter(t)
	r := createRoutine(t, router, map[string]any{"name": "아침", "cadence": "daily", "start_day": "2026-09-25", "steps": []string{"a", "b"}}, 1)
	checkStep(t, router, r.ID, r.Steps[0].ID, "2026-09-29", true)
	checkStep(t, router, r.ID, r.Steps[1].ID, "2026-09-29", true)
	checkStep(t, router, r.ID, r.Steps[0].ID, "2026-09-30", true)

	hist := decodeInto[struct {
		Items []routines.HistoryDay `json:"items"`
	}](t, doTodo(t, router, http.MethodGet, routinePath(r.ID, "/history/2026-10-01"), nil, 1), http.StatusOK)
	if len(hist.Items) != 30 || hist.Items[0].Day != "2026-09-02" || hist.Items[29].Day != "2026-10-01" {
		t.Fatalf("history window: %d items, first=%v", len(hist.Items), hist.Items[0])
	}
	byDay := map[string]routines.HistoryDay{}
	for _, h := range hist.Items {
		byDay[h.Day] = h
	}
	if h := byDay["2026-09-24"]; h.Scheduled {
		t.Errorf("before start day must not be scheduled: %+v", h)
	}
	if h := byDay["2026-09-29"]; !h.Scheduled || !h.Completed || h.Done != 2 {
		t.Errorf("09-29: %+v", h)
	}
	if h := byDay["2026-09-30"]; !h.Scheduled || h.Completed || h.Done != 1 {
		t.Errorf("09-30: %+v", h)
	}
	if h := byDay["2026-10-01"]; !h.Scheduled || h.Completed || h.Done != 0 || h.Total != 2 {
		t.Errorf("10-01: %+v", h)
	}
}

func TestRoutineUserScoping(t *testing.T) {
	router := newRoutinesRouter(t)
	r := createRoutine(t, router, map[string]any{"name": "아침", "cadence": "daily", "start_day": "2026-09-01", "steps": []string{"a"}}, 1)
	stepPath := routinePath(r.ID, "/days/2026-10-01/steps/"+strconv.FormatInt(r.Steps[0].ID, 10))
	if rec := doTodo(t, router, http.MethodPatch, stepPath, map[string]any{"checked": true}, 2); rec.Code != http.StatusNotFound {
		t.Errorf("other user check: got %d want 404", rec.Code)
	}
	if rec := doTodo(t, router, http.MethodPost, routinePath(r.ID, "/steps"), map[string]any{"title": "x"}, 2); rec.Code != http.StatusNotFound {
		t.Errorf("other user add step: got %d want 404", rec.Code)
	}
	if rec := doTodo(t, router, http.MethodGet, routinePath(r.ID, "/history/2026-10-01"), nil, 2); rec.Code != http.StatusNotFound {
		t.Errorf("other user history: got %d want 404", rec.Code)
	}
	if rec := doTodo(t, router, http.MethodDelete, routinePath(r.ID, ""), nil, 2); rec.Code != http.StatusNotFound {
		t.Errorf("other user delete: got %d want 404", rec.Code)
	}
	resp := decodeInto[dayResponse](t, doTodo(t, router, http.MethodGet, "/api/routines/days/2026-10-01", nil, 2), http.StatusOK)
	if len(resp.Items) != 0 {
		t.Errorf("other user day view: %+v", resp.Items)
	}
	if rec := doTodo(t, router, http.MethodDelete, routinePath(r.ID, ""), nil, 1); rec.Code != http.StatusNoContent {
		t.Errorf("owner delete: got %d want 204", rec.Code)
	}
}
