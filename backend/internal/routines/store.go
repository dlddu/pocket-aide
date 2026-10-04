package routines

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound     = errors.New("routine not found")
	ErrNotScheduled = errors.New("routine is not scheduled on that day")
)

// Step is one row of the routine_steps table.
type Step struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Position int    `json:"position"`
}

// Routine is one row of the routines table, with its steps in position order.
type Routine struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Schedule
	StartDay  string `json:"start_day"`
	Steps     []Step `json:"steps"`
	CreatedAt int64  `json:"created_at"`
}

func (r Routine) ScheduledOn(day time.Time) bool {
	return day.Format(dayLayout) >= r.StartDay && r.Matches(day)
}

// DayStep is a step with whether a routine_step_checks row marks it done on that day.
type DayStep struct {
	Step
	Checked bool `json:"checked"`
}

type DayRoutine struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Schedule
	Day       string    `json:"day"`
	Steps     []DayStep `json:"steps"`
	Done      int       `json:"done"`
	Total     int       `json:"total"`
	Completed bool      `json:"completed"`
}

type HistoryDay struct {
	Day       string `json:"day"`
	Scheduled bool   `json:"scheduled"`
	Done      int    `json:"done"`
	Total     int    `json:"total"`
	Completed bool   `json:"completed"`
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store { return &Store{db: db} }

type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func load(ctx context.Context, q querier, userID int64, only int64) ([]Routine, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, name, cadence, weekdays, month_day, start_day, created_at
		FROM routines
		WHERE user_id = ? AND (? = 0 OR id = ?)
		ORDER BY id
	`, userID, only, only)
	if err != nil {
		return nil, fmt.Errorf("list routines: %w", err)
	}
	out := make([]Routine, 0)
	index := map[int64]int{}
	for rows.Next() {
		var r Routine
		if err := rows.Scan(&r.ID, &r.Name, &r.Cadence, &r.Weekdays, &r.MonthDay, &r.StartDay, &r.CreatedAt); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan routine: %w", err)
		}
		r.Steps = make([]Step, 0)
		index[r.ID] = len(out)
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("rows err: %w", err)
	}
	_ = rows.Close()

	stepRows, err := q.QueryContext(ctx, `
		SELECT s.routine_id, s.id, s.title, s.position
		FROM routine_steps s
		JOIN routines r ON r.id = s.routine_id
		WHERE r.user_id = ? AND (? = 0 OR r.id = ?)
		ORDER BY s.routine_id, s.position, s.id
	`, userID, only, only)
	if err != nil {
		return nil, fmt.Errorf("list routine steps: %w", err)
	}
	defer func() { _ = stepRows.Close() }()
	for stepRows.Next() {
		var routineID int64
		var s Step
		if err := stepRows.Scan(&routineID, &s.ID, &s.Title, &s.Position); err != nil {
			return nil, fmt.Errorf("scan routine step: %w", err)
		}
		if i, ok := index[routineID]; ok {
			out[i].Steps = append(out[i].Steps, s)
		}
	}
	if err := stepRows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

func (s *Store) List(ctx context.Context, userID int64) ([]Routine, error) {
	return load(ctx, s.db, userID, 0)
}

func (s *Store) Get(ctx context.Context, userID, id int64) (Routine, error) {
	if id <= 0 {
		return Routine{}, ErrNotFound
	}
	list, err := load(ctx, s.db, userID, id)
	if err != nil {
		return Routine{}, err
	}
	if len(list) == 0 {
		return Routine{}, ErrNotFound
	}
	return list[0], nil
}

func (s *Store) Create(ctx context.Context, userID int64, name string, schedule Schedule, startDay string, steps []string) (Routine, error) {
	schedule, err := schedule.Normalized()
	if err != nil {
		return Routine{}, err
	}
	if _, err := ParseDay(startDay); err != nil {
		return Routine{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Routine{}, fmt.Errorf("begin create routine: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `
		INSERT INTO routines (user_id, name, cadence, weekdays, month_day, start_day)
		VALUES (?, ?, ?, ?, ?, ?)
	`, userID, name, schedule.Cadence, schedule.Weekdays, schedule.MonthDay, startDay)
	if err != nil {
		return Routine{}, fmt.Errorf("insert routine: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Routine{}, fmt.Errorf("last insert id: %w", err)
	}
	for i, title := range steps {
		if _, err := tx.ExecContext(ctx, `INSERT INTO routine_steps (routine_id, position, title) VALUES (?, ?, ?)`, id, i, title); err != nil {
			return Routine{}, fmt.Errorf("insert routine step: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return Routine{}, fmt.Errorf("commit create routine: %w", err)
	}
	return s.Get(ctx, userID, id)
}

func (s *Store) Delete(ctx context.Context, userID, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM routines WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("delete routine: %w", err)
	}
	return expectOne(res)
}

func (s *Store) AddStep(ctx context.Context, userID, routineID int64, title string) (Routine, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO routine_steps (routine_id, position, title)
		SELECT r.id, COALESCE((SELECT MAX(position) + 1 FROM routine_steps WHERE routine_id = r.id), 0), ?
		FROM routines r
		WHERE r.id = ? AND r.user_id = ?
	`, title, routineID, userID)
	if err != nil {
		return Routine{}, fmt.Errorf("insert routine step: %w", err)
	}
	if err := expectOne(res); err != nil {
		return Routine{}, err
	}
	return s.Get(ctx, userID, routineID)
}

func (s *Store) DeleteStep(ctx context.Context, userID, routineID, stepID int64) error {
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM routine_steps
		WHERE id = ? AND routine_id IN (SELECT id FROM routines WHERE id = ? AND user_id = ?)
	`, stepID, routineID, userID)
	if err != nil {
		return fmt.Errorf("delete routine step: %w", err)
	}
	return expectOne(res)
}

func (s *Store) checkedSteps(ctx context.Context, userID int64, day string) (map[int64]bool, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.step_id
		FROM routine_step_checks c
		JOIN routine_steps s ON s.id = c.step_id
		JOIN routines r ON r.id = s.routine_id
		WHERE r.user_id = ? AND c.day = ?
	`, userID, day)
	if err != nil {
		return nil, fmt.Errorf("list routine checks: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan routine check: %w", err)
		}
		out[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

func dayView(r Routine, day string, checked map[int64]bool) DayRoutine {
	v := DayRoutine{ID: r.ID, Name: r.Name, Schedule: r.Schedule, Day: day, Steps: make([]DayStep, 0, len(r.Steps)), Total: len(r.Steps)}
	for _, st := range r.Steps {
		c := checked[st.ID]
		if c {
			v.Done++
		}
		v.Steps = append(v.Steps, DayStep{Step: st, Checked: c})
	}
	v.Completed = v.Total > 0 && v.Done == v.Total
	return v
}

func (s *Store) Day(ctx context.Context, userID int64, day string) ([]DayRoutine, error) {
	t, err := ParseDay(day)
	if err != nil {
		return nil, err
	}
	all, err := s.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	checked, err := s.checkedSteps(ctx, userID, day)
	if err != nil {
		return nil, err
	}
	out := make([]DayRoutine, 0, len(all))
	for _, r := range all {
		if r.ScheduledOn(t) {
			out = append(out, dayView(r, day, checked))
		}
	}
	return out, nil
}

func (s *Store) SetCheck(ctx context.Context, userID, routineID, stepID int64, day string, checked bool) (DayRoutine, error) {
	t, err := ParseDay(day)
	if err != nil {
		return DayRoutine{}, err
	}
	r, err := s.Get(ctx, userID, routineID)
	if err != nil {
		return DayRoutine{}, err
	}
	found := false
	for _, st := range r.Steps {
		if st.ID == stepID {
			found = true
			break
		}
	}
	if !found {
		return DayRoutine{}, ErrNotFound
	}
	if !r.ScheduledOn(t) {
		return DayRoutine{}, ErrNotScheduled
	}
	if checked {
		_, err = s.db.ExecContext(ctx, `INSERT OR IGNORE INTO routine_step_checks (step_id, day) VALUES (?, ?)`, stepID, day)
	} else {
		_, err = s.db.ExecContext(ctx, `DELETE FROM routine_step_checks WHERE step_id = ? AND day = ?`, stepID, day)
	}
	if err != nil {
		return DayRoutine{}, fmt.Errorf("set routine check: %w", err)
	}
	marks, err := s.checkedSteps(ctx, userID, day)
	if err != nil {
		return DayRoutine{}, err
	}
	return dayView(r, day, marks), nil
}

func (s *Store) History(ctx context.Context, userID, routineID int64, endDay string, days int) ([]HistoryDay, error) {
	end, err := ParseDay(endDay)
	if err != nil {
		return nil, err
	}
	r, err := s.Get(ctx, userID, routineID)
	if err != nil {
		return nil, err
	}
	start := end.AddDate(0, 0, -(days - 1))
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.day, COUNT(*)
		FROM routine_step_checks c
		JOIN routine_steps s ON s.id = c.step_id
		WHERE s.routine_id = ? AND c.day BETWEEN ? AND ?
		GROUP BY c.day
	`, routineID, start.Format(dayLayout), endDay)
	if err != nil {
		return nil, fmt.Errorf("routine history: %w", err)
	}
	defer func() { _ = rows.Close() }()
	done := map[string]int{}
	for rows.Next() {
		var day string
		var n int
		if err := rows.Scan(&day, &n); err != nil {
			return nil, fmt.Errorf("scan routine history: %w", err)
		}
		done[day] = n
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	out := make([]HistoryDay, 0, days)
	total := len(r.Steps)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format(dayLayout)
		h := HistoryDay{Day: key, Scheduled: r.ScheduledOn(d), Done: done[key], Total: total}
		h.Completed = h.Scheduled && total > 0 && h.Done == total
		out = append(out, h)
	}
	return out, nil
}

func expectOne(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
