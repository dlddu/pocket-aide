// Package repository provides data access layer implementations.
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrRoutineNotFound is returned when no routine with the given criteria exists.
var ErrRoutineNotFound = errors.New("routine not found")

// Routine represents a routine record from the database.
type Routine struct {
	ID            int64
	UserID        int64
	Name          string
	IntervalDays  int
	LastDoneAt    string // YYYY-MM-DD
	NextDueDate   string // YYYY-MM-DD = LastDoneAt + IntervalDays
	DDay          int    // next_due_date - today (negative = overdue)
	Note          string
	NotifyEnabled bool
}

// RoutineUpdates holds the optional fields that can be updated on a Routine.
// Zero-value fields ("", 0) are treated as "no change".
// NotifyEnabled is a pointer so that nil means "no change".
type RoutineUpdates struct {
	Name          string
	IntervalDays  int
	Note          string
	NotifyEnabled *bool
}

// RoutineRepository provides database access for routine records.
type RoutineRepository struct {
	db *sql.DB
}

// NewRoutineRepository creates a new RoutineRepository backed by the given database.
func NewRoutineRepository(db *sql.DB) *RoutineRepository {
	return &RoutineRepository{db: db}
}

// calcNextDueDate calculates next_due_date = lastDoneAt + intervalDays.
// lastDoneAt must be in YYYY-MM-DD format.
func calcNextDueDate(lastDoneAt string, intervalDays int) (string, error) {
	d, err := time.Parse("2006-01-02", lastDoneAt)
	if err != nil {
		return "", fmt.Errorf("invalid lastDoneAt date %q: %w", lastDoneAt, err)
	}
	return d.AddDate(0, 0, intervalDays).Format("2006-01-02"), nil
}

// calcDDay calculates the D-day as next_due_date - today (UTC).
// A negative value means the routine is overdue.
func calcDDay(nextDueDate string) (int, error) {
	due, err := time.Parse("2006-01-02", nextDueDate)
	if err != nil {
		return 0, fmt.Errorf("invalid nextDueDate %q: %w", nextDueDate, err)
	}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	diff := due.Sub(today)
	return int(diff.Hours() / 24), nil
}

// Create inserts a new routine for the given user and returns the persisted record.
// Returns an error if name is empty or intervalDays is not positive.
func (r *RoutineRepository) Create(userID int64, name string, intervalDays int, lastDoneAt string) (*Routine, error) {
	if name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}
	if intervalDays <= 0 {
		return nil, fmt.Errorf("intervalDays must be positive, got %d", intervalDays)
	}

	nextDueDate, err := calcNextDueDate(lastDoneAt, intervalDays)
	if err != nil {
		return nil, err
	}

	result, err := r.db.Exec(
		`INSERT INTO routines (user_id, name, interval_days, last_done_at, next_due_date, note, notify_enabled)
		 VALUES (?, ?, ?, ?, ?, '', 0)`,
		userID, name, intervalDays, lastDoneAt, nextDueDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert routine: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	dDay, err := calcDDay(nextDueDate)
	if err != nil {
		return nil, err
	}

	return &Routine{
		ID:            id,
		UserID:        userID,
		Name:          name,
		IntervalDays:  intervalDays,
		LastDoneAt:    lastDoneAt,
		NextDueDate:   nextDueDate,
		DDay:          dDay,
		Note:          "",
		NotifyEnabled: false,
	}, nil
}

// ListByUserID returns all routines belonging to the given user.
// Returns a non-nil empty slice when the user has no routines.
func (r *RoutineRepository) ListByUserID(userID int64) ([]*Routine, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, interval_days, last_done_at, next_due_date, note, notify_enabled
		 FROM routines WHERE user_id = ? ORDER BY id ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query routines: %w", err)
	}
	defer rows.Close()

	routines := make([]*Routine, 0)
	for rows.Next() {
		rt, err := scanRoutine(rows)
		if err != nil {
			return nil, err
		}
		routines = append(routines, rt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return routines, nil
}

// FindByID returns the routine with the given ID owned by userID.
// Returns ErrRoutineNotFound when the ID does not exist or belongs to another user.
func (r *RoutineRepository) FindByID(id int64, userID int64) (*Routine, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, name, interval_days, last_done_at, next_due_date, note, notify_enabled
		 FROM routines WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	rt, err := scanRoutineRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w", ErrRoutineNotFound)
		}
		return nil, fmt.Errorf("failed to query routine: %w", err)
	}
	return rt, nil
}

// Update applies the non-zero fields of updates to the routine with the given ID
// owned by userID. IntervalDays changes trigger a next_due_date recalculation.
// Returns ErrRoutineNotFound when the ID does not exist or belongs to another user.
func (r *RoutineRepository) Update(id int64, userID int64, updates RoutineUpdates) (*Routine, error) {
	// Fetch the current record first (also validates ownership)
	current, err := r.FindByID(id, userID)
	if err != nil {
		return nil, err
	}

	// Apply updates (zero-values mean "no change")
	newName := current.Name
	if updates.Name != "" {
		newName = updates.Name
	}

	newNote := current.Note
	if updates.Note != "" {
		newNote = updates.Note
	}

	newIntervalDays := current.IntervalDays
	if updates.IntervalDays != 0 {
		newIntervalDays = updates.IntervalDays
	}

	newNotifyEnabled := current.NotifyEnabled
	if updates.NotifyEnabled != nil {
		newNotifyEnabled = *updates.NotifyEnabled
	}

	// Recalculate next_due_date if interval changed
	newNextDueDate := current.NextDueDate
	if updates.IntervalDays != 0 {
		newNextDueDate, err = calcNextDueDate(current.LastDoneAt, newIntervalDays)
		if err != nil {
			return nil, err
		}
	}

	notifyInt := 0
	if newNotifyEnabled {
		notifyInt = 1
	}

	_, err = r.db.Exec(
		`UPDATE routines SET name = ?, interval_days = ?, next_due_date = ?, note = ?, notify_enabled = ?,
		 updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`,
		newName, newIntervalDays, newNextDueDate, newNote, notifyInt, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update routine: %w", err)
	}

	dDay, err := calcDDay(newNextDueDate)
	if err != nil {
		return nil, err
	}

	return &Routine{
		ID:            current.ID,
		UserID:        current.UserID,
		Name:          newName,
		IntervalDays:  newIntervalDays,
		LastDoneAt:    current.LastDoneAt,
		NextDueDate:   newNextDueDate,
		DDay:          dDay,
		Note:          newNote,
		NotifyEnabled: newNotifyEnabled,
	}, nil
}

// Delete removes the routine with the given ID owned by userID.
// Returns ErrRoutineNotFound when the ID does not exist or belongs to another user.
func (r *RoutineRepository) Delete(id int64, userID int64) error {
	result, err := r.db.Exec(
		`DELETE FROM routines WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete routine: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w", ErrRoutineNotFound)
	}
	return nil
}

// Complete sets last_done_at to today (UTC) and recalculates next_due_date.
// Returns ErrRoutineNotFound when the ID does not exist or belongs to another user.
func (r *RoutineRepository) Complete(id int64, userID int64) (*Routine, error) {
	current, err := r.FindByID(id, userID)
	if err != nil {
		return nil, err
	}

	todayStr := time.Now().UTC().Format("2006-01-02")
	newNextDueDate, err := calcNextDueDate(todayStr, current.IntervalDays)
	if err != nil {
		return nil, err
	}

	_, err = r.db.Exec(
		`UPDATE routines SET last_done_at = ?, next_due_date = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND user_id = ?`,
		todayStr, newNextDueDate, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to complete routine: %w", err)
	}

	dDay, err := calcDDay(newNextDueDate)
	if err != nil {
		return nil, err
	}

	return &Routine{
		ID:            current.ID,
		UserID:        current.UserID,
		Name:          current.Name,
		IntervalDays:  current.IntervalDays,
		LastDoneAt:    todayStr,
		NextDueDate:   newNextDueDate,
		DDay:          dDay,
		Note:          current.Note,
		NotifyEnabled: current.NotifyEnabled,
	}, nil
}

// scanRoutine scans a row from *sql.Rows into a Routine struct.
func scanRoutine(rows *sql.Rows) (*Routine, error) {
	var rt Routine
	var notifyInt int
	if err := rows.Scan(&rt.ID, &rt.UserID, &rt.Name, &rt.IntervalDays, &rt.LastDoneAt, &rt.NextDueDate, &rt.Note, &notifyInt); err != nil {
		return nil, fmt.Errorf("failed to scan routine: %w", err)
	}
	rt.NotifyEnabled = notifyInt != 0

	dDay, err := calcDDay(rt.NextDueDate)
	if err != nil {
		return nil, err
	}
	rt.DDay = dDay
	return &rt, nil
}

// scanRoutineRow scans a *sql.Row into a Routine struct.
func scanRoutineRow(row *sql.Row) (*Routine, error) {
	var rt Routine
	var notifyInt int
	if err := row.Scan(&rt.ID, &rt.UserID, &rt.Name, &rt.IntervalDays, &rt.LastDoneAt, &rt.NextDueDate, &rt.Note, &notifyInt); err != nil {
		return nil, err
	}
	rt.NotifyEnabled = notifyInt != 0

	dDay, err := calcDDay(rt.NextDueDate)
	if err != nil {
		return nil, err
	}
	rt.DDay = dDay
	return &rt, nil
}
