// Package repository_test contains unit tests for the RoutineRepository.
// These tests use an in-memory SQLite database via testutil.NewTestDB and
// verify that RoutineRepository correctly reads and writes routine records.
//
// TDD Red Phase: the RoutineRepository type and its methods do not yet exist.
// All tests are expected to fail until repository/routine_repository.go is implemented.
package repository_test

import (
	"errors"
	"testing"
	"time"

	"github.com/dlddu/pocket-aide/repository"
	"github.com/dlddu/pocket-aide/testutil"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// seedUser inserts a user row and returns its auto-assigned ID.
func seedUser(t *testing.T, tdb *testutil.TestDB, email string) int64 {
	t.Helper()
	result, err := tdb.DB.Exec(
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		email, "hashed_password",
	)
	if err != nil {
		t.Fatalf("seedUser: failed to insert user %q: %v", email, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("seedUser: failed to get last insert ID: %v", err)
	}
	return id
}

// today returns today's date in YYYY-MM-DD format, matching the format used
// by the Complete method.
func today() string {
	return time.Now().UTC().Format("2006-01-02")
}

// addDays adds n days to a YYYY-MM-DD date string and returns the result in
// the same format. The test fails immediately on parse error.
func addDays(t *testing.T, dateStr string, n int) string {
	t.Helper()
	d, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		t.Fatalf("addDays: failed to parse date %q: %v", dateStr, err)
	}
	return d.AddDate(0, 0, n).Format("2006-01-02")
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

// TestRoutineRepository_Create_ReturnsNewRoutine verifies that Create inserts
// a routine record and returns the persisted routine with a non-zero ID, the
// correct name, interval_days, last_done_at, and a calculated next_due_date.
//
// Scenario:
//
//	Seed: user id=1.
//	Create(userID=1, name="샤워", intervalDays=1, lastDoneAt="2026-03-13")
//	→ Routine{ID: >0, UserID: 1, Name: "샤워", IntervalDays: 1,
//	          LastDoneAt: "2026-03-13", NextDueDate: "2026-03-14"}
func TestRoutineRepository_Create_ReturnsNewRoutine(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "create-routine@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	// Act
	routine, err := repo.Create(userID, "샤워", 1, "2026-03-13")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if routine == nil {
		t.Fatal("expected non-nil routine, got nil")
	}
	if routine.ID == 0 {
		t.Error("expected routine ID to be non-zero after creation")
	}
	if routine.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, routine.UserID)
	}
	if routine.Name != "샤워" {
		t.Errorf("expected Name '샤워', got '%s'", routine.Name)
	}
	if routine.IntervalDays != 1 {
		t.Errorf("expected IntervalDays 1, got %d", routine.IntervalDays)
	}
	if routine.LastDoneAt != "2026-03-13" {
		t.Errorf("expected LastDoneAt '2026-03-13', got '%s'", routine.LastDoneAt)
	}
	if routine.NextDueDate != "2026-03-14" {
		t.Errorf("expected NextDueDate '2026-03-14' (last_done_at + interval_days), got '%s'", routine.NextDueDate)
	}
}

// TestRoutineRepository_Create_PersistsToDatabase verifies that Create
// actually writes the row so that a direct DB query finds it.
//
// Scenario:
//
//	Seed: user id=1.
//	Create(userID=1, name="세수", intervalDays=1, lastDoneAt="2026-03-13")
//	→ row exists in routines table
func TestRoutineRepository_Create_PersistsToDatabase(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "create-persist@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	// Act
	created, err := repo.Create(userID, "세수", 1, "2026-03-13")
	if err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	// Assert: row must be visible via direct SQL
	var count int
	row := tdb.DB.QueryRow(
		"SELECT COUNT(*) FROM routines WHERE id = ? AND user_id = ? AND name = ?",
		created.ID, userID, "세수",
	)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to query routines table: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row in routines table, got %d", count)
	}
}

// TestRoutineRepository_Create_NextDueDateCalculation_MultipleIntervals
// verifies that next_due_date is correctly calculated for various interval
// lengths to confirm the date arithmetic is not hardcoded.
//
// Scenario:
//
//	Create(userID=1, name="세탁", intervalDays=7, lastDoneAt="2026-03-07")
//	→ NextDueDate == "2026-03-14"
//
//	Create(userID=1, name="청소", intervalDays=30, lastDoneAt="2026-02-01")
//	→ NextDueDate == "2026-03-03"
func TestRoutineRepository_Create_NextDueDateCalculation_MultipleIntervals(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "dday-intervals@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	cases := []struct {
		name         string
		intervalDays int
		lastDoneAt   string
		wantNextDue  string
	}{
		{"세탁", 7, "2026-03-07", "2026-03-14"},
		{"청소", 30, "2026-02-01", "2026-03-03"},
		{"운동", 3, "2026-03-11", "2026-03-14"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Act
			routine, err := repo.Create(userID, tc.name, tc.intervalDays, tc.lastDoneAt)

			// Assert
			if err != nil {
				t.Fatalf("Create(%q, %d, %q): unexpected error: %v", tc.name, tc.intervalDays, tc.lastDoneAt, err)
			}
			if routine.NextDueDate != tc.wantNextDue {
				t.Errorf("Create(%q, %d, %q): expected NextDueDate %q, got %q",
					tc.name, tc.intervalDays, tc.lastDoneAt, tc.wantNextDue, routine.NextDueDate)
			}
		})
	}
}

// TestRoutineRepository_Create_EmptyName_ReturnsError verifies that Create
// returns an error when an empty name is supplied.
//
// Scenario:
//
//	Create(userID=1, name="", intervalDays=1, lastDoneAt="2026-03-13")
//	→ non-nil error
func TestRoutineRepository_Create_EmptyName_ReturnsError(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "empty-name@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	// Act
	routine, err := repo.Create(userID, "", 1, "2026-03-13")

	// Assert
	if err == nil {
		t.Error("expected error for empty name, got nil")
	}
	if routine != nil {
		t.Errorf("expected nil routine on error, got %+v", routine)
	}
}

// TestRoutineRepository_Create_ZeroIntervalDays_ReturnsError verifies that
// Create rejects an interval_days value of zero.
//
// Scenario:
//
//	Create(userID=1, name="스트레칭", intervalDays=0, lastDoneAt="2026-03-13")
//	→ non-nil error
func TestRoutineRepository_Create_ZeroIntervalDays_ReturnsError(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "zero-interval@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	// Act
	routine, err := repo.Create(userID, "스트레칭", 0, "2026-03-13")

	// Assert
	if err == nil {
		t.Error("expected error for zero interval_days, got nil")
	}
	if routine != nil {
		t.Errorf("expected nil routine on error, got %+v", routine)
	}
}

// ---------------------------------------------------------------------------
// ListByUserID
// ---------------------------------------------------------------------------

// TestRoutineRepository_ListByUserID_ReturnsRoutines verifies that
// ListByUserID returns all routines belonging to the given user.
//
// Scenario:
//
//	Seed: user id=1 with two routines.
//	ListByUserID(1)
//	→ []Routine with len == 2
func TestRoutineRepository_ListByUserID_ReturnsRoutines(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "list-routines@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	if _, err := repo.Create(userID, "양치", 1, "2026-03-13"); err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}
	if _, err := repo.Create(userID, "세수", 2, "2026-03-12"); err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act
	routines, err := repo.ListByUserID(userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(routines) != 2 {
		t.Fatalf("expected 2 routines, got %d", len(routines))
	}
}

// TestRoutineRepository_ListByUserID_EmptyList verifies that ListByUserID
// returns an empty (non-nil) slice when the user has no routines.
//
// Scenario:
//
//	Seed: user id=1 with no routines.
//	ListByUserID(1)
//	→ []Routine{} (non-nil, len==0)
func TestRoutineRepository_ListByUserID_EmptyList(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "list-empty@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	// Act
	routines, err := repo.ListByUserID(userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if routines == nil {
		t.Error("expected non-nil slice, got nil")
	}
	if len(routines) != 0 {
		t.Errorf("expected empty slice, got %d routines", len(routines))
	}
}

// TestRoutineRepository_ListByUserID_IncludesDDay verifies that each routine
// returned by ListByUserID contains a DDay field calculated as
// next_due_date - today.
//
// Scenario (today = addDays(lastDoneAt, 0), nextDue = lastDoneAt + 7):
//
//	Create(userID, "세탁", 7, lastDoneAt) where nextDue is 3 days from now.
//	ListByUserID(userID)
//	→ routines[0].DDay == 3
func TestRoutineRepository_ListByUserID_IncludesDDay(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "list-dday@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	// last_done_at = today - 4 days → next_due_date = today + 3 days → d_day = 3
	lastDoneAt := addDays(t, today(), -4)
	if _, err := repo.Create(userID, "세탁", 7, lastDoneAt); err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act
	routines, err := repo.ListByUserID(userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(routines) == 0 {
		t.Fatal("expected at least 1 routine")
	}
	if routines[0].DDay != 3 {
		t.Errorf("expected DDay 3, got %d", routines[0].DDay)
	}
}

// TestRoutineRepository_ListByUserID_DDay_NegativeWhenOverdue verifies that
// DDay is negative when next_due_date is in the past.
//
// Scenario:
//
//	Create(userID, "조깅", 1, 5 days ago) → next_due_date = 4 days ago → d_day = -4
//	ListByUserID(userID)
//	→ routines[0].DDay == -4
func TestRoutineRepository_ListByUserID_DDay_NegativeWhenOverdue(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "list-overdue@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	// last_done_at = today - 5 days, interval = 1 → next_due_date = today - 4 days → d_day = -4
	lastDoneAt := addDays(t, today(), -5)
	if _, err := repo.Create(userID, "조깅", 1, lastDoneAt); err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act
	routines, err := repo.ListByUserID(userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(routines) == 0 {
		t.Fatal("expected at least 1 routine")
	}
	if routines[0].DDay != -4 {
		t.Errorf("expected DDay -4 (overdue), got %d", routines[0].DDay)
	}
}

// TestRoutineRepository_ListByUserID_DDay_ZeroWhenDueToday verifies that
// DDay is 0 when next_due_date equals today.
//
// Scenario:
//
//	Create(userID, "명상", 3, today - 3 days) → next_due_date = today → d_day = 0
//	ListByUserID(userID)
//	→ routines[0].DDay == 0
func TestRoutineRepository_ListByUserID_DDay_ZeroWhenDueToday(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "list-dday-zero@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	// last_done_at = today - 3 days, interval = 3 → next_due_date = today → d_day = 0
	lastDoneAt := addDays(t, today(), -3)
	if _, err := repo.Create(userID, "명상", 3, lastDoneAt); err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act
	routines, err := repo.ListByUserID(userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(routines) == 0 {
		t.Fatal("expected at least 1 routine")
	}
	if routines[0].DDay != 0 {
		t.Errorf("expected DDay 0 (due today), got %d", routines[0].DDay)
	}
}

// ---------------------------------------------------------------------------
// FindByID
// ---------------------------------------------------------------------------

// TestRoutineRepository_FindByID_ReturnsRoutine verifies that FindByID
// returns the matching routine when the ID and userID are both correct.
//
// Scenario:
//
//	Seed: one routine for user id=1.
//	FindByID(routineID, userID=1)
//	→ Routine{Name: "스트레칭"}, nil
func TestRoutineRepository_FindByID_ReturnsRoutine(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "find-routine@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	created, err := repo.Create(userID, "스트레칭", 3, "2026-03-11")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act
	found, err := repo.FindByID(created.ID, userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if found == nil {
		t.Fatal("expected non-nil routine, got nil")
	}
	if found.Name != "스트레칭" {
		t.Errorf("expected Name '스트레칭', got '%s'", found.Name)
	}
	if found.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, found.ID)
	}
}

// TestRoutineRepository_FindByID_NotFound_ReturnsErrNotFound verifies that
// FindByID returns repository.ErrRoutineNotFound when no routine with the
// given ID exists.
//
// Scenario:
//
//	FindByID(id=99999, userID=1)
//	→ errors.Is(err, repository.ErrRoutineNotFound) == true
func TestRoutineRepository_FindByID_NotFound_ReturnsErrNotFound(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "find-notfound@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	// Act
	found, err := repo.FindByID(99999, userID)

	// Assert
	if err == nil {
		t.Error("expected error for non-existent ID, got nil")
	}
	if found != nil {
		t.Errorf("expected nil routine when not found, got %+v", found)
	}
	if !errors.Is(err, repository.ErrRoutineNotFound) {
		t.Errorf("expected ErrRoutineNotFound, got: %v", err)
	}
}

// TestRoutineRepository_FindByID_WrongUser_ReturnsErrNotFound verifies that
// FindByID does not return a routine when the userID does not match the
// owner, enforcing user isolation.
//
// Scenario:
//
//	Seed: routine owned by user id=1.
//	FindByID(routineID, userID=2)
//	→ errors.Is(err, repository.ErrRoutineNotFound) == true
func TestRoutineRepository_FindByID_WrongUser_ReturnsErrNotFound(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	ownerID := seedUser(t, tdb, "owner@example.com")
	otherID := seedUser(t, tdb, "other-findbyid@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	created, err := repo.Create(ownerID, "독서", 7, "2026-03-07")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act: attempt to access owner's routine as the other user
	found, err := repo.FindByID(created.ID, otherID)

	// Assert
	if err == nil {
		t.Error("expected error when accessing another user's routine, got nil")
	}
	if found != nil {
		t.Errorf("expected nil routine for wrong user, got %+v", found)
	}
	if !errors.Is(err, repository.ErrRoutineNotFound) {
		t.Errorf("expected ErrRoutineNotFound, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

// TestRoutineRepository_Update_Name_ReturnsUpdatedRoutine verifies that
// Update persists a new name and returns the updated routine.
//
// Scenario:
//
//	Seed: routine{name:"스트레칭", intervalDays:3}.
//	Update(id, userID, name="스트레칭 (강화)")
//	→ Routine{Name: "스트레칭 (강화)"}, nil
func TestRoutineRepository_Update_Name_ReturnsUpdatedRoutine(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "update-name@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	created, err := repo.Create(userID, "스트레칭", 3, "2026-03-11")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act
	updates := repository.RoutineUpdates{Name: "스트레칭 (강화)"}
	updated, err := repo.Update(created.ID, userID, updates)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if updated == nil {
		t.Fatal("expected non-nil routine, got nil")
	}
	if updated.Name != "스트레칭 (강화)" {
		t.Errorf("expected Name '스트레칭 (강화)', got '%s'", updated.Name)
	}
}

// TestRoutineRepository_Update_IntervalDays_RecalculatesNextDueDate verifies
// that changing interval_days causes next_due_date to be recalculated as
// last_done_at + new_interval_days.
//
// Scenario:
//
//	Seed: routine{intervalDays:7, lastDoneAt:"2026-03-07"} → nextDue:"2026-03-14"
//	Update(id, userID, intervalDays=3)
//	→ Routine{NextDueDate: "2026-03-10"} (2026-03-07 + 3)
func TestRoutineRepository_Update_IntervalDays_RecalculatesNextDueDate(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "update-interval@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	created, err := repo.Create(userID, "조깅", 7, "2026-03-07")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}
	if created.NextDueDate != "2026-03-14" {
		t.Fatalf("precondition: expected initial NextDueDate '2026-03-14', got '%s'", created.NextDueDate)
	}

	// Act
	updates := repository.RoutineUpdates{IntervalDays: 3}
	updated, err := repo.Update(created.ID, userID, updates)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if updated.NextDueDate != "2026-03-10" {
		t.Errorf("expected NextDueDate '2026-03-10' after interval change (2026-03-07 + 3), got '%s'", updated.NextDueDate)
	}
	if updated.IntervalDays != 3 {
		t.Errorf("expected IntervalDays 3, got %d", updated.IntervalDays)
	}
}

// TestRoutineRepository_Update_Note_PersistsNote verifies that Update can
// set the optional note field.
//
// Scenario:
//
//	Seed: routine with empty note.
//	Update(id, userID, note="저녁 식사 후")
//	→ Routine{Note: "저녁 식사 후"}, nil
func TestRoutineRepository_Update_Note_PersistsNote(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "update-note@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	created, err := repo.Create(userID, "양치", 1, "2026-03-13")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act
	updates := repository.RoutineUpdates{Note: "저녁 식사 후"}
	updated, err := repo.Update(created.ID, userID, updates)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if updated.Note != "저녁 식사 후" {
		t.Errorf("expected Note '저녁 식사 후', got '%s'", updated.Note)
	}
}

// TestRoutineRepository_Update_NotFound_ReturnsErrNotFound verifies that
// Update returns ErrRoutineNotFound when no routine with the given ID exists.
//
// Scenario:
//
//	Update(id=99999, userID=1, updates={name:"없는 루틴"})
//	→ errors.Is(err, repository.ErrRoutineNotFound) == true
func TestRoutineRepository_Update_NotFound_ReturnsErrNotFound(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "update-notfound@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	// Act
	updates := repository.RoutineUpdates{Name: "없는 루틴"}
	updated, err := repo.Update(99999, userID, updates)

	// Assert
	if err == nil {
		t.Error("expected error for non-existent ID, got nil")
	}
	if updated != nil {
		t.Errorf("expected nil routine on not-found, got %+v", updated)
	}
	if !errors.Is(err, repository.ErrRoutineNotFound) {
		t.Errorf("expected ErrRoutineNotFound, got: %v", err)
	}
}

// TestRoutineRepository_Update_WrongUser_ReturnsErrNotFound verifies that
// Update refuses to modify a routine that belongs to another user.
//
// Scenario:
//
//	Seed: routine owned by user id=1.
//	Update(routineID, userID=2, updates={name:"탈취"})
//	→ errors.Is(err, repository.ErrRoutineNotFound) == true
func TestRoutineRepository_Update_WrongUser_ReturnsErrNotFound(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	ownerID := seedUser(t, tdb, "owner-update@example.com")
	otherID := seedUser(t, tdb, "other-update@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	created, err := repo.Create(ownerID, "독서", 7, "2026-03-07")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act: other user attempts to update owner's routine
	updates := repository.RoutineUpdates{Name: "탈취"}
	updated, err := repo.Update(created.ID, otherID, updates)

	// Assert
	if err == nil {
		t.Error("expected error when updating another user's routine, got nil")
	}
	if updated != nil {
		t.Errorf("expected nil routine for wrong user, got %+v", updated)
	}
	if !errors.Is(err, repository.ErrRoutineNotFound) {
		t.Errorf("expected ErrRoutineNotFound, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

// TestRoutineRepository_Delete_RemovesRoutine verifies that Delete removes the
// row from the database so that a subsequent FindByID returns ErrRoutineNotFound.
//
// Scenario:
//
//	Seed: one routine.
//	Delete(routineID, userID)   → nil
//	FindByID(routineID, userID) → ErrRoutineNotFound
func TestRoutineRepository_Delete_RemovesRoutine(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "delete-routine@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	created, err := repo.Create(userID, "삭제할 루틴", 7, "2026-03-07")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act
	err = repo.Delete(created.ID, userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error on Delete, got: %v", err)
	}

	_, findErr := repo.FindByID(created.ID, userID)
	if !errors.Is(findErr, repository.ErrRoutineNotFound) {
		t.Errorf("expected ErrRoutineNotFound after Delete, got: %v", findErr)
	}
}

// TestRoutineRepository_Delete_NotFound_ReturnsErrNotFound verifies that
// Delete returns ErrRoutineNotFound when the ID does not exist.
//
// Scenario:
//
//	Delete(id=99999, userID=1)
//	→ errors.Is(err, repository.ErrRoutineNotFound) == true
func TestRoutineRepository_Delete_NotFound_ReturnsErrNotFound(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "delete-notfound@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	// Act
	err := repo.Delete(99999, userID)

	// Assert
	if err == nil {
		t.Error("expected error for non-existent ID, got nil")
	}
	if !errors.Is(err, repository.ErrRoutineNotFound) {
		t.Errorf("expected ErrRoutineNotFound, got: %v", err)
	}
}

// TestRoutineRepository_Delete_WrongUser_ReturnsErrNotFound verifies that
// Delete refuses to remove a routine that belongs to another user.
//
// Scenario:
//
//	Seed: routine owned by user id=1.
//	Delete(routineID, userID=2)
//	→ errors.Is(err, repository.ErrRoutineNotFound) == true
//	→ routine is still present in the database
func TestRoutineRepository_Delete_WrongUser_ReturnsErrNotFound(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	ownerID := seedUser(t, tdb, "owner-delete@example.com")
	otherID := seedUser(t, tdb, "other-delete@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	created, err := repo.Create(ownerID, "독서", 7, "2026-03-07")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act: other user attempts to delete owner's routine
	err = repo.Delete(created.ID, otherID)

	// Assert: must fail
	if err == nil {
		t.Error("expected error when deleting another user's routine, got nil")
	}
	if !errors.Is(err, repository.ErrRoutineNotFound) {
		t.Errorf("expected ErrRoutineNotFound, got: %v", err)
	}

	// The original routine must still exist
	found, findErr := repo.FindByID(created.ID, ownerID)
	if findErr != nil {
		t.Errorf("expected owner's routine to still exist after failed delete, got error: %v", findErr)
	}
	if found == nil {
		t.Error("expected owner's routine to still exist after failed delete, got nil")
	}
}

// ---------------------------------------------------------------------------
// Complete
// ---------------------------------------------------------------------------

// TestRoutineRepository_Complete_UpdatesLastDoneAt verifies that Complete
// sets last_done_at to today.
//
// Scenario:
//
//	Seed: routine with last_done_at = "2026-03-01".
//	Complete(routineID, userID)
//	→ Routine{LastDoneAt: today()}
func TestRoutineRepository_Complete_UpdatesLastDoneAt(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "complete-lastdone@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	created, err := repo.Create(userID, "독서", 7, "2026-03-01")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	expectedToday := today()

	// Act
	completed, err := repo.Complete(created.ID, userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if completed == nil {
		t.Fatal("expected non-nil routine, got nil")
	}
	if completed.LastDoneAt != expectedToday {
		t.Errorf("expected LastDoneAt %q (today), got '%s'", expectedToday, completed.LastDoneAt)
	}
}

// TestRoutineRepository_Complete_RecalculatesNextDueDate verifies that
// Complete sets next_due_date to today + interval_days.
//
// Scenario:
//
//	Seed: routine{intervalDays:7, lastDoneAt:"2026-03-01"}.
//	Complete(routineID, userID)
//	→ Routine{NextDueDate: today() + 7 days}
func TestRoutineRepository_Complete_RecalculatesNextDueDate(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "complete-nextdue@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	created, err := repo.Create(userID, "운동", 7, "2026-03-01")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	expectedNextDue := addDays(t, today(), 7)

	// Act
	completed, err := repo.Complete(created.ID, userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if completed.NextDueDate != expectedNextDue {
		t.Errorf("expected NextDueDate %q (today + interval_days), got '%s'", expectedNextDue, completed.NextDueDate)
	}
}

// TestRoutineRepository_Complete_PersistsToDatabase verifies that Complete
// writes the updated last_done_at and next_due_date to the database so that a
// subsequent FindByID reflects the changes.
//
// Scenario:
//
//	Complete(routineID, userID)
//	FindByID(routineID, userID)
//	→ LastDoneAt == today(), NextDueDate == today() + interval_days
func TestRoutineRepository_Complete_PersistsToDatabase(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "complete-persist@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	created, err := repo.Create(userID, "명상", 3, "2026-03-01")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	expectedToday := today()
	expectedNextDue := addDays(t, expectedToday, 3)

	// Act
	if _, err := repo.Complete(created.ID, userID); err != nil {
		t.Fatalf("Complete returned unexpected error: %v", err)
	}

	// Assert via FindByID (reads fresh from DB)
	found, err := repo.FindByID(created.ID, userID)
	if err != nil {
		t.Fatalf("expected no error on FindByID after Complete, got: %v", err)
	}
	if found.LastDoneAt != expectedToday {
		t.Errorf("expected persisted LastDoneAt %q, got '%s'", expectedToday, found.LastDoneAt)
	}
	if found.NextDueDate != expectedNextDue {
		t.Errorf("expected persisted NextDueDate %q, got '%s'", expectedNextDue, found.NextDueDate)
	}
}

// TestRoutineRepository_Complete_NotFound_ReturnsErrNotFound verifies that
// Complete returns ErrRoutineNotFound when the ID does not exist.
//
// Scenario:
//
//	Complete(id=99999, userID=1)
//	→ errors.Is(err, repository.ErrRoutineNotFound) == true
func TestRoutineRepository_Complete_NotFound_ReturnsErrNotFound(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "complete-notfound@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	// Act
	completed, err := repo.Complete(99999, userID)

	// Assert
	if err == nil {
		t.Error("expected error for non-existent ID, got nil")
	}
	if completed != nil {
		t.Errorf("expected nil routine on not-found, got %+v", completed)
	}
	if !errors.Is(err, repository.ErrRoutineNotFound) {
		t.Errorf("expected ErrRoutineNotFound, got: %v", err)
	}
}

// TestRoutineRepository_Complete_WrongUser_ReturnsErrNotFound verifies that
// Complete refuses to update a routine belonging to another user.
//
// Scenario:
//
//	Seed: routine owned by user id=1.
//	Complete(routineID, userID=2)
//	→ errors.Is(err, repository.ErrRoutineNotFound) == true
func TestRoutineRepository_Complete_WrongUser_ReturnsErrNotFound(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	ownerID := seedUser(t, tdb, "owner-complete@example.com")
	otherID := seedUser(t, tdb, "other-complete@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	created, err := repo.Create(ownerID, "수영", 7, "2026-03-01")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act: other user attempts to complete owner's routine
	completed, err := repo.Complete(created.ID, otherID)

	// Assert
	if err == nil {
		t.Error("expected error when completing another user's routine, got nil")
	}
	if completed != nil {
		t.Errorf("expected nil routine for wrong user, got %+v", completed)
	}
	if !errors.Is(err, repository.ErrRoutineNotFound) {
		t.Errorf("expected ErrRoutineNotFound, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// User Isolation
// ---------------------------------------------------------------------------

// TestRoutineRepository_UserIsolation_ListByUserID_ExcludesOtherUsersRoutines
// verifies that ListByUserID never returns routines owned by a different user.
//
// Scenario:
//
//	Seed: user1 has 2 routines, user2 has 1 routine.
//	ListByUserID(user1)
//	→ all 2 routines belong to user1; user2's routine is absent
func TestRoutineRepository_UserIsolation_ListByUserID_ExcludesOtherUsersRoutines(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	user1ID := seedUser(t, tdb, "user1@example.com")
	user2ID := seedUser(t, tdb, "user2@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	if _, err := repo.Create(user1ID, "양치", 1, "2026-03-13"); err != nil {
		t.Fatalf("setup: Create user1 routine 1 failed: %v", err)
	}
	if _, err := repo.Create(user1ID, "세수", 1, "2026-03-13"); err != nil {
		t.Fatalf("setup: Create user1 routine 2 failed: %v", err)
	}
	if _, err := repo.Create(user2ID, "다른 사용자 루틴", 1, "2026-03-13"); err != nil {
		t.Fatalf("setup: Create user2 routine failed: %v", err)
	}

	// Act
	routines, err := repo.ListByUserID(user1ID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(routines) != 2 {
		t.Fatalf("expected 2 routines for user1, got %d", len(routines))
	}
	for i, r := range routines {
		if r.UserID != user1ID {
			t.Errorf("routine[%d] has UserID %d, expected %d", i, r.UserID, user1ID)
		}
	}
}

// TestRoutineRepository_UserIsolation_MultipleUsersIndependent verifies that
// two different users each see only their own routines without interference.
//
// Scenario:
//
//	user1 creates 1 routine, user2 creates 1 routine.
//	ListByUserID(user1) → len==1, all UserID==user1
//	ListByUserID(user2) → len==1, all UserID==user2
func TestRoutineRepository_UserIsolation_MultipleUsersIndependent(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	user1ID := seedUser(t, tdb, "iso-user1@example.com")
	user2ID := seedUser(t, tdb, "iso-user2@example.com")
	repo := repository.NewRoutineRepository(tdb.DB)

	if _, err := repo.Create(user1ID, "user1 루틴", 1, "2026-03-13"); err != nil {
		t.Fatalf("setup: Create user1 routine failed: %v", err)
	}
	if _, err := repo.Create(user2ID, "user2 루틴", 1, "2026-03-13"); err != nil {
		t.Fatalf("setup: Create user2 routine failed: %v", err)
	}

	// Act
	routines1, err1 := repo.ListByUserID(user1ID)
	routines2, err2 := repo.ListByUserID(user2ID)

	// Assert
	if err1 != nil {
		t.Fatalf("ListByUserID(user1): expected no error, got: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("ListByUserID(user2): expected no error, got: %v", err2)
	}
	if len(routines1) != 1 {
		t.Errorf("expected 1 routine for user1, got %d", len(routines1))
	}
	if len(routines2) != 1 {
		t.Errorf("expected 1 routine for user2, got %d", len(routines2))
	}
	if len(routines1) > 0 && routines1[0].UserID != user1ID {
		t.Errorf("user1 routine has wrong UserID: got %d", routines1[0].UserID)
	}
	if len(routines2) > 0 && routines2[0].UserID != user2ID {
		t.Errorf("user2 routine has wrong UserID: got %d", routines2[0].UserID)
	}
}
