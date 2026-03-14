// Package repository provides data access layer implementations.
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrTodoNotFound is returned when no todo with the given criteria exists.
var ErrTodoNotFound = errors.New("todo not found")

// Todo represents a todo record from the database.
type Todo struct {
	ID          int64
	UserID      int64
	Title       string
	Type        string
	Note        string
	Priority    string
	DueDate     *string
	CompletedAt *string
}

// TodoUpdates holds the optional fields that can be updated on a Todo.
// Zero-value fields ("") are treated as "no change".
type TodoUpdates struct {
	Title    string
	Note     string
	Priority string
	DueDate  *string
}

// TodoRepository provides database access for todo records.
type TodoRepository struct {
	db *sql.DB
}

// NewTodoRepository creates a new TodoRepository backed by the given database.
func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// Create inserts a new todo for the given user and returns the persisted record.
// Returns an error if title is empty. The priority defaults to "medium" via the
// DB column default.
func (r *TodoRepository) Create(userID int64, title string, todoType string) (*Todo, error) {
	if title == "" {
		return nil, fmt.Errorf("title must not be empty")
	}
	if todoType == "" {
		todoType = "personal"
	}

	result, err := r.db.Exec(
		`INSERT INTO todos (user_id, title, type) VALUES (?, ?, ?)`,
		userID, title, todoType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert todo: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return r.FindByID(id, userID)
}

// CreateWithPriority inserts a new todo with an explicit priority value.
// If priority is empty, "medium" is used as the default.
// Returns an error if title is empty.
func (r *TodoRepository) CreateWithPriority(userID int64, title string, todoType string, priority string) (*Todo, error) {
	if title == "" {
		return nil, fmt.Errorf("title must not be empty")
	}
	if todoType == "" {
		todoType = "personal"
	}
	if priority == "" {
		priority = "medium"
	}

	result, err := r.db.Exec(
		`INSERT INTO todos (user_id, title, type, priority) VALUES (?, ?, ?, ?)`,
		userID, title, todoType, priority,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert todo: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return r.FindByID(id, userID)
}

// ListByUserIDAndType returns all todos belonging to the given user filtered
// by type. If todoType is empty, all types are returned.
// When todoType is "work", results are sorted by priority (high → medium → low),
// then by id ascending. For other types the order is by id ascending.
// Returns a non-nil empty slice when the user has no matching todos.
func (r *TodoRepository) ListByUserIDAndType(userID int64, todoType string) ([]*Todo, error) {
	var rows *sql.Rows
	var err error

	if todoType == "work" {
		rows, err = r.db.Query(
			`SELECT id, user_id, title, type, note, priority, due_date, completed_at
			 FROM todos WHERE user_id = ? AND type = ?
			 ORDER BY CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 ELSE 4 END ASC, id ASC`,
			userID, todoType,
		)
	} else if todoType != "" {
		rows, err = r.db.Query(
			`SELECT id, user_id, title, type, note, priority, due_date, completed_at
			 FROM todos WHERE user_id = ? AND type = ? ORDER BY id ASC`,
			userID, todoType,
		)
	} else {
		rows, err = r.db.Query(
			`SELECT id, user_id, title, type, note, priority, due_date, completed_at
			 FROM todos WHERE user_id = ? ORDER BY id ASC`,
			userID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query todos: %w", err)
	}
	defer rows.Close()

	todos := make([]*Todo, 0)
	for rows.Next() {
		td, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, td)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return todos, nil
}

// FindByID returns the todo with the given ID owned by userID.
// Returns ErrTodoNotFound when the ID does not exist or belongs to another user.
func (r *TodoRepository) FindByID(id int64, userID int64) (*Todo, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, title, type, note, priority, due_date, completed_at
		 FROM todos WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	td, err := scanTodoRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w", ErrTodoNotFound)
		}
		return nil, fmt.Errorf("failed to query todo: %w", err)
	}
	return td, nil
}

// Update applies the non-zero fields of updates to the todo with the given ID
// owned by userID.
// Returns ErrTodoNotFound when the ID does not exist or belongs to another user.
func (r *TodoRepository) Update(id int64, userID int64, updates TodoUpdates) (*Todo, error) {
	current, err := r.FindByID(id, userID)
	if err != nil {
		return nil, err
	}

	newTitle := current.Title
	if updates.Title != "" {
		newTitle = updates.Title
	}

	newNote := current.Note
	if updates.Note != "" {
		newNote = updates.Note
	}

	newPriority := current.Priority
	if updates.Priority != "" {
		newPriority = updates.Priority
	}

	_, err = r.db.Exec(
		`UPDATE todos SET title = ?, note = ?, priority = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND user_id = ?`,
		newTitle, newNote, newPriority, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	return &Todo{
		ID:          current.ID,
		UserID:      current.UserID,
		Title:       newTitle,
		Type:        current.Type,
		Note:        newNote,
		Priority:    newPriority,
		DueDate:     current.DueDate,
		CompletedAt: current.CompletedAt,
	}, nil
}

// Delete removes the todo with the given ID owned by userID.
// Returns ErrTodoNotFound when the ID does not exist or belongs to another user.
func (r *TodoRepository) Delete(id int64, userID int64) error {
	result, err := r.db.Exec(
		`DELETE FROM todos WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w", ErrTodoNotFound)
	}
	return nil
}

// Toggle switches completed_at between NULL and current timestamp.
// If currently NULL (pending), sets it to now; if non-NULL (completed),
// sets it back to NULL.
// Returns ErrTodoNotFound when the ID does not exist or belongs to another user.
func (r *TodoRepository) Toggle(id int64, userID int64) (*Todo, error) {
	current, err := r.FindByID(id, userID)
	if err != nil {
		return nil, err
	}

	var newCompletedAt *string
	if current.CompletedAt == nil {
		// pending → completed
		now := time.Now().UTC().Format(time.RFC3339)
		newCompletedAt = &now
		_, err = r.db.Exec(
			`UPDATE todos SET completed_at = ?, updated_at = CURRENT_TIMESTAMP
			 WHERE id = ? AND user_id = ?`,
			now, id, userID,
		)
	} else {
		// completed → pending
		newCompletedAt = nil
		_, err = r.db.Exec(
			`UPDATE todos SET completed_at = NULL, updated_at = CURRENT_TIMESTAMP
			 WHERE id = ? AND user_id = ?`,
			id, userID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to toggle todo: %w", err)
	}

	return &Todo{
		ID:          current.ID,
		UserID:      current.UserID,
		Title:       current.Title,
		Type:        current.Type,
		Note:        current.Note,
		Priority:    current.Priority,
		DueDate:     current.DueDate,
		CompletedAt: newCompletedAt,
	}, nil
}

// scanTodo scans a row from *sql.Rows into a Todo struct.
func scanTodo(rows *sql.Rows) (*Todo, error) {
	var td Todo
	var note sql.NullString
	var priority sql.NullString
	var dueDate sql.NullString
	var completedAt sql.NullString

	if err := rows.Scan(&td.ID, &td.UserID, &td.Title, &td.Type, &note, &priority, &dueDate, &completedAt); err != nil {
		return nil, fmt.Errorf("failed to scan todo: %w", err)
	}
	if note.Valid {
		td.Note = note.String
	}
	if priority.Valid {
		td.Priority = priority.String
	}
	if dueDate.Valid {
		td.DueDate = &dueDate.String
	}
	if completedAt.Valid {
		td.CompletedAt = &completedAt.String
	}
	return &td, nil
}

// scanTodoRow scans a *sql.Row into a Todo struct.
func scanTodoRow(row *sql.Row) (*Todo, error) {
	var td Todo
	var note sql.NullString
	var priority sql.NullString
	var dueDate sql.NullString
	var completedAt sql.NullString

	if err := row.Scan(&td.ID, &td.UserID, &td.Title, &td.Type, &note, &priority, &dueDate, &completedAt); err != nil {
		return nil, err
	}
	if note.Valid {
		td.Note = note.String
	}
	if priority.Valid {
		td.Priority = priority.String
	}
	if dueDate.Valid {
		td.DueDate = &dueDate.String
	}
	if completedAt.Valid {
		td.CompletedAt = &completedAt.String
	}
	return &td, nil
}
