// Package todos is the storage layer for PRD-3 (개인 투두 / 회사 투두) —
// user-scoped CRUD over two physically separate tables, one per area.
package todos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Area selects which collection a call operates on. Each area maps to its own
// table; there is deliberately no operation that reads or writes both, and no
// operation that moves an item between them (PRD-3 AC4).
type Area string

const (
	AreaPersonal Area = "personal"
	AreaWork     Area = "work"
)

// Valid reports whether a is one of the two areas.
func (a Area) Valid() bool {
	_, ok := tables[a]
	return ok
}

// tables is the only place an area turns into SQL. Table names cannot be bound
// as parameters, so they come from this fixed map and never from the request.
var tables = map[Area]string{
	AreaPersonal: "personal_todos",
	AreaWork:     "work_todos",
}

// Priority is optional on a todo. The DB CHECK constraint mirrors these values.
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

// Valid reports whether p is one of the three allowed values.
func (p Priority) Valid() bool {
	switch p {
	case PriorityHigh, PriorityNormal, PriorityLow:
		return true
	default:
		return false
	}
}

// Fields is the user-editable part of a todo. Update overwrites all of them.
type Fields struct {
	Title    string
	Memo     string
	DueDate  *string // YYYY-MM-DD, nil when the todo has no deadline
	Priority *Priority
	Done     bool
}

// Todo is one row of either area's table.
type Todo struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Memo        string    `json:"memo"`
	DueDate     *string   `json:"due_date"`
	Priority    *Priority `json:"priority"`
	CompletedAt *int64    `json:"completed_at"`
	CreatedAt   int64     `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
}

// ErrNotFound is returned when a query targets a row that does not exist in
// the given area or is not owned by the caller.
var ErrNotFound = errors.New("todo not found")

// ErrInvalidArea is returned for an area outside AreaPersonal/AreaWork.
var ErrInvalidArea = errors.New("invalid todo area")

// Store is the user-scoped CRUD facade.
type Store struct {
	db *sql.DB
}

// New wires a Store onto an open *sql.DB.
func New(db *sql.DB) *Store { return &Store{db: db} }

func table(area Area) (string, error) {
	t, ok := tables[area]
	if !ok {
		return "", ErrInvalidArea
	}
	return t, nil
}

const columns = `id, title, memo, due_date, priority, completed_at, created_at, updated_at`

type scanner interface{ Scan(dest ...any) error }

func scanTodo(s scanner) (Todo, error) {
	var t Todo
	var due, prio sql.NullString
	var completed sql.NullInt64
	if err := s.Scan(&t.ID, &t.Title, &t.Memo, &due, &prio, &completed, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return Todo{}, err
	}
	if due.Valid {
		t.DueDate = &due.String
	}
	if prio.Valid {
		p := Priority(prio.String)
		t.Priority = &p
	}
	if completed.Valid {
		t.CompletedAt = &completed.Int64
	}
	return t, nil
}

// List returns every todo in area that belongs to userID: open items first
// (earliest due date first, undated last), then completed items newest first.
func (s *Store) List(ctx context.Context, area Area, userID int64) ([]Todo, error) {
	tbl, err := table(area)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT `+columns+`
		FROM `+tbl+`
		WHERE user_id = ?
		ORDER BY completed_at IS NOT NULL, completed_at DESC,
		         due_date IS NULL, due_date, created_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]Todo, 0)
	for rows.Next() {
		t, err := scanTodo(rows)
		if err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

// Get returns a single todo in area that belongs to userID.
func (s *Store) Get(ctx context.Context, area Area, userID, id int64) (Todo, error) {
	tbl, err := table(area)
	if err != nil {
		return Todo{}, err
	}
	t, err := scanTodo(s.db.QueryRowContext(ctx, `
		SELECT `+columns+`
		FROM `+tbl+`
		WHERE id = ? AND user_id = ?
	`, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Todo{}, ErrNotFound
	}
	if err != nil {
		return Todo{}, fmt.Errorf("get todo: %w", err)
	}
	return t, nil
}

// Create inserts a new todo into area for userID and returns the row.
func (s *Store) Create(ctx context.Context, area Area, userID int64, f Fields) (Todo, error) {
	tbl, err := table(area)
	if err != nil {
		return Todo{}, err
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO `+tbl+` (user_id, title, memo, due_date, priority, completed_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, userID, f.Title, f.Memo, f.DueDate, f.Priority, completedAt(nil, f.Done))
	if err != nil {
		return Todo{}, fmt.Errorf("insert todo: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Todo{}, fmt.Errorf("last insert id: %w", err)
	}
	return s.Get(ctx, area, userID, id)
}

// Update overwrites every editable field of an existing todo. Marking an
// already-completed todo done again keeps its original completion time.
func (s *Store) Update(ctx context.Context, area Area, userID, id int64, f Fields) (Todo, error) {
	current, err := s.Get(ctx, area, userID, id)
	if err != nil {
		return Todo{}, err
	}
	tbl, _ := table(area)
	res, err := s.db.ExecContext(ctx, `
		UPDATE `+tbl+`
		SET title = ?, memo = ?, due_date = ?, priority = ?, completed_at = ?, updated_at = unixepoch()
		WHERE id = ? AND user_id = ?
	`, f.Title, f.Memo, f.DueDate, f.Priority, completedAt(current.CompletedAt, f.Done), id, userID)
	if err != nil {
		return Todo{}, fmt.Errorf("update todo: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Todo{}, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return Todo{}, ErrNotFound
	}
	return s.Get(ctx, area, userID, id)
}

// Delete removes a todo in area that belongs to userID.
func (s *Store) Delete(ctx context.Context, area Area, userID, id int64) error {
	tbl, err := table(area)
	if err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM `+tbl+` WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func completedAt(current *int64, done bool) any {
	if !done {
		return nil
	}
	if current != nil {
		return *current
	}
	return time.Now().Unix()
}
