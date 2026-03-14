// Package repository_test contains unit tests for the TodoRepository.
package repository_test

import (
	"errors"
	"testing"

	"github.com/dlddu/pocket-aide/repository"
	"github.com/dlddu/pocket-aide/testutil"
)

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestTodoRepository_Create_ReturnsNewTodo(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-create@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	todo, err := repo.Create(userID, "장보기", "personal")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if todo == nil {
		t.Fatal("expected non-nil todo, got nil")
	}
	if todo.ID == 0 {
		t.Error("expected todo ID to be non-zero after creation")
	}
	if todo.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, todo.UserID)
	}
	if todo.Title != "장보기" {
		t.Errorf("expected Title '장보기', got '%s'", todo.Title)
	}
	if todo.Type != "personal" {
		t.Errorf("expected Type 'personal', got '%s'", todo.Type)
	}
	if todo.CompletedAt != nil {
		t.Errorf("expected CompletedAt nil, got %v", todo.CompletedAt)
	}
}

func TestTodoRepository_Create_EmptyTitle_ReturnsError(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-empty-title@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	todo, err := repo.Create(userID, "", "personal")

	if err == nil {
		t.Error("expected error for empty title, got nil")
	}
	if todo != nil {
		t.Errorf("expected nil todo on error, got %+v", todo)
	}
}

func TestTodoRepository_Create_DefaultType(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-default-type@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	todo, err := repo.Create(userID, "기본 타입", "")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if todo.Type != "personal" {
		t.Errorf("expected default type 'personal', got '%s'", todo.Type)
	}
}

// ---------------------------------------------------------------------------
// ListByUserIDAndType
// ---------------------------------------------------------------------------

func TestTodoRepository_ListByUserIDAndType_ReturnsTodos(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-list@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	if _, err := repo.Create(userID, "장보기", "personal"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := repo.Create(userID, "독서", "personal"); err != nil {
		t.Fatalf("setup: %v", err)
	}

	todos, err := repo.ListByUserIDAndType(userID, "personal")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(todos))
	}
	for i, td := range todos {
		if td.Type != "personal" {
			t.Errorf("todos[%d]: expected type 'personal', got '%s'", i, td.Type)
		}
	}
}

func TestTodoRepository_ListByUserIDAndType_EmptyList(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-list-empty@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	todos, err := repo.ListByUserIDAndType(userID, "personal")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if todos == nil {
		t.Error("expected non-nil slice, got nil")
	}
	if len(todos) != 0 {
		t.Errorf("expected empty slice, got %d todos", len(todos))
	}
}

func TestTodoRepository_ListByUserIDAndType_UserIsolation(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	user1ID := seedUser(t, tdb, "todo-iso-user1@example.com")
	user2ID := seedUser(t, tdb, "todo-iso-user2@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	if _, err := repo.Create(user1ID, "user1 todo", "personal"); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := repo.Create(user2ID, "user2 todo", "personal"); err != nil {
		t.Fatalf("setup: %v", err)
	}

	todos, err := repo.ListByUserIDAndType(user1ID, "personal")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo for user1, got %d", len(todos))
	}
	if todos[0].UserID != user1ID {
		t.Errorf("expected UserID %d, got %d", user1ID, todos[0].UserID)
	}
}

// ---------------------------------------------------------------------------
// FindByID
// ---------------------------------------------------------------------------

func TestTodoRepository_FindByID_ReturnsTodo(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-find@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	created, err := repo.Create(userID, "운동하기", "personal")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	found, err := repo.FindByID(created.ID, userID)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if found.Title != "운동하기" {
		t.Errorf("expected Title '운동하기', got '%s'", found.Title)
	}
}

func TestTodoRepository_FindByID_NotFound(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-find-notfound@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	found, err := repo.FindByID(99999, userID)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if found != nil {
		t.Errorf("expected nil, got %+v", found)
	}
	if !errors.Is(err, repository.ErrTodoNotFound) {
		t.Errorf("expected ErrTodoNotFound, got: %v", err)
	}
}

func TestTodoRepository_FindByID_WrongUser(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	ownerID := seedUser(t, tdb, "todo-owner-find@example.com")
	otherID := seedUser(t, tdb, "todo-other-find@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	created, err := repo.Create(ownerID, "비밀 투두", "personal")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	found, err := repo.FindByID(created.ID, otherID)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if found != nil {
		t.Errorf("expected nil, got %+v", found)
	}
	if !errors.Is(err, repository.ErrTodoNotFound) {
		t.Errorf("expected ErrTodoNotFound, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestTodoRepository_Update_Title(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-update@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	created, err := repo.Create(userID, "운동하기", "personal")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	updated, err := repo.Update(created.ID, userID, repository.TodoUpdates{Title: "운동하기 (30분)"})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if updated.Title != "운동하기 (30분)" {
		t.Errorf("expected Title '운동하기 (30분)', got '%s'", updated.Title)
	}
}

func TestTodoRepository_Update_NotFound(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-update-notfound@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	updated, err := repo.Update(99999, userID, repository.TodoUpdates{Title: "없는 투두"})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if updated != nil {
		t.Errorf("expected nil, got %+v", updated)
	}
	if !errors.Is(err, repository.ErrTodoNotFound) {
		t.Errorf("expected ErrTodoNotFound, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestTodoRepository_Delete_RemovesTodo(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-delete@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	created, err := repo.Create(userID, "삭제할 투두", "personal")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	err = repo.Delete(created.ID, userID)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, findErr := repo.FindByID(created.ID, userID)
	if !errors.Is(findErr, repository.ErrTodoNotFound) {
		t.Errorf("expected ErrTodoNotFound after delete, got: %v", findErr)
	}
}

func TestTodoRepository_Delete_NotFound(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-delete-notfound@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	err := repo.Delete(99999, userID)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, repository.ErrTodoNotFound) {
		t.Errorf("expected ErrTodoNotFound, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Toggle
// ---------------------------------------------------------------------------

func TestTodoRepository_Toggle_CompletesTodo(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-toggle-complete@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	created, err := repo.Create(userID, "책 읽기", "personal")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if created.CompletedAt != nil {
		t.Fatalf("precondition: expected nil CompletedAt, got %v", created.CompletedAt)
	}

	toggled, err := repo.Toggle(created.ID, userID)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if toggled.CompletedAt == nil {
		t.Error("expected non-nil CompletedAt after toggle (pending → completed)")
	}
}

func TestTodoRepository_Toggle_UncompletesTodo(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-toggle-uncomplete@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	created, err := repo.Create(userID, "일기 쓰기", "personal")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// First toggle: pending → completed
	if _, err := repo.Toggle(created.ID, userID); err != nil {
		t.Fatalf("1st toggle: %v", err)
	}

	// Second toggle: completed → pending
	toggled, err := repo.Toggle(created.ID, userID)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if toggled.CompletedAt != nil {
		t.Errorf("expected nil CompletedAt after 2nd toggle (completed → pending), got %v", toggled.CompletedAt)
	}
}

func TestTodoRepository_Toggle_NotFound(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-toggle-notfound@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	toggled, err := repo.Toggle(99999, userID)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if toggled != nil {
		t.Errorf("expected nil, got %+v", toggled)
	}
	if !errors.Is(err, repository.ErrTodoNotFound) {
		t.Errorf("expected ErrTodoNotFound, got: %v", err)
	}
}

func TestTodoRepository_Toggle_PersistsToDatabase(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "todo-toggle-persist@example.com")
	repo := repository.NewTodoRepository(tdb.DB)

	created, err := repo.Create(userID, "영화 보기", "personal")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, err := repo.Toggle(created.ID, userID); err != nil {
		t.Fatalf("toggle: %v", err)
	}

	found, err := repo.FindByID(created.ID, userID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.CompletedAt == nil {
		t.Error("expected non-nil CompletedAt in DB after toggle")
	}
}
