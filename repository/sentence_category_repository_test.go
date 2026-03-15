// Package repository_test contains unit tests for the SentenceCategoryRepository.
// These tests use an in-memory SQLite database via testutil.NewTestDB and
// verify that SentenceCategoryRepository correctly reads and writes
// sentence_categories records.
//
// TDD Red Phase: SentenceCategoryRepository and its methods do not yet exist.
// All tests are expected to fail until repository/sentence_category_repository.go
// is implemented and the following DB migrations are applied:
//
// Migration 000007_sentence_categories.up.sql:
//
//	CREATE TABLE IF NOT EXISTS sentence_categories (
//	    id         INTEGER PRIMARY KEY AUTOINCREMENT,
//	    user_id    INTEGER NOT NULL,
//	    name       TEXT    NOT NULL,
//	    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//	    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//	    FOREIGN KEY (user_id) REFERENCES users(id)
//	);
//
// Migration 000007_sentence_categories.down.sql:
//
//	DROP TABLE IF EXISTS sentence_categories;
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

// TestSentenceCategoryRepository_Create_ReturnsCategory verifies that Create
// inserts a sentence_categories record and returns the persisted category with
// a non-zero ID, the correct user_id, and the correct name.
//
// Scenario:
//
//	Seed: user.
//	Create(userID, "인사말")
//	→ SentenceCategory{ID: >0, UserID: userID, Name: "인사말"}, nil
func TestSentenceCategoryRepository_Create_ReturnsCategory(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "sc-create@example.com")
	repo := repository.NewSentenceCategoryRepository(tdb.DB)

	// Act
	cat, err := repo.Create(userID, "인사말")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cat == nil {
		t.Fatal("expected non-nil category, got nil")
	}
	if cat.ID == 0 {
		t.Error("expected category ID to be non-zero after creation")
	}
	if cat.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, cat.UserID)
	}
	if cat.Name != "인사말" {
		t.Errorf("expected Name '인사말', got '%s'", cat.Name)
	}
}

// TestSentenceCategoryRepository_Create_EmptyName_ReturnsError verifies that
// Create returns a non-nil error when an empty name is supplied, without
// inserting any record.
//
// Scenario:
//
//	Create(userID, "")
//	→ non-nil error, nil category
func TestSentenceCategoryRepository_Create_EmptyName_ReturnsError(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "sc-empty-name@example.com")
	repo := repository.NewSentenceCategoryRepository(tdb.DB)

	// Act
	cat, err := repo.Create(userID, "")

	// Assert
	if err == nil {
		t.Error("expected error for empty name, got nil")
	}
	if cat != nil {
		t.Errorf("expected nil category on error, got %+v", cat)
	}
}

// ---------------------------------------------------------------------------
// ListByUserID
// ---------------------------------------------------------------------------

// TestSentenceCategoryRepository_ListByUserID_ReturnsList verifies that
// ListByUserID returns all categories belonging to the given user.
//
// Scenario:
//
//	Seed: user with two categories "인사말" and "감사 표현".
//	ListByUserID(userID)
//	→ []*SentenceCategory with len == 2
func TestSentenceCategoryRepository_ListByUserID_ReturnsList(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "sc-list@example.com")
	repo := repository.NewSentenceCategoryRepository(tdb.DB)

	if _, err := repo.Create(userID, "인사말"); err != nil {
		t.Fatalf("setup: Create '인사말' failed: %v", err)
	}
	if _, err := repo.Create(userID, "감사 표현"); err != nil {
		t.Fatalf("setup: Create '감사 표현' failed: %v", err)
	}

	// Act
	categories, err := repo.ListByUserID(userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(categories))
	}
	for i, c := range categories {
		if c.UserID != userID {
			t.Errorf("categories[%d]: expected UserID %d, got %d", i, userID, c.UserID)
		}
	}
}

// TestSentenceCategoryRepository_ListByUserID_Empty_ReturnsEmptySlice verifies
// that ListByUserID returns a non-nil empty slice (not nil) when the user has
// no categories.
//
// Scenario:
//
//	Seed: user with no categories.
//	ListByUserID(userID)
//	→ non-nil empty slice, nil error
func TestSentenceCategoryRepository_ListByUserID_Empty_ReturnsEmptySlice(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "sc-list-empty@example.com")
	repo := repository.NewSentenceCategoryRepository(tdb.DB)

	// Act
	categories, err := repo.ListByUserID(userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if categories == nil {
		t.Error("expected non-nil slice, got nil")
	}
	if len(categories) != 0 {
		t.Errorf("expected empty slice, got %d categories", len(categories))
	}
}

// ---------------------------------------------------------------------------
// FindByID
// ---------------------------------------------------------------------------

// TestSentenceCategoryRepository_FindByID_ReturnsCategory verifies that
// FindByID returns the matching category when the ID and userID are correct.
//
// Scenario:
//
//	Seed: one category for user.
//	FindByID(categoryID, userID)
//	→ SentenceCategory{Name: "인사말"}, nil
func TestSentenceCategoryRepository_FindByID_ReturnsCategory(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "sc-find@example.com")
	repo := repository.NewSentenceCategoryRepository(tdb.DB)

	created, err := repo.Create(userID, "인사말")
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
		t.Fatal("expected non-nil category, got nil")
	}
	if found.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, found.ID)
	}
	if found.Name != "인사말" {
		t.Errorf("expected Name '인사말', got '%s'", found.Name)
	}
}

// TestSentenceCategoryRepository_FindByID_NotFound_ReturnsError verifies that
// FindByID returns ErrSentenceCategoryNotFound when no category with the given
// ID exists (or it belongs to a different user).
//
// Scenario:
//
//	FindByID(99999, userID)
//	→ errors.Is(err, repository.ErrSentenceCategoryNotFound) == true
func TestSentenceCategoryRepository_FindByID_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "sc-find-notfound@example.com")
	repo := repository.NewSentenceCategoryRepository(tdb.DB)

	// Act
	found, err := repo.FindByID(99999, userID)

	// Assert
	if err == nil {
		t.Error("expected error for non-existent ID, got nil")
	}
	if found != nil {
		t.Errorf("expected nil category when not found, got %+v", found)
	}
	if !errors.Is(err, repository.ErrSentenceCategoryNotFound) {
		t.Errorf("expected ErrSentenceCategoryNotFound, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

// TestSentenceCategoryRepository_Update_ReturnsUpdated verifies that Update
// persists the new name and returns the updated category.
//
// Scenario:
//
//	Seed: category{name:"인사말"}.
//	Update(categoryID, userID, "감사 표현")
//	→ SentenceCategory{Name: "감사 표현"}, nil
func TestSentenceCategoryRepository_Update_ReturnsUpdated(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "sc-update@example.com")
	repo := repository.NewSentenceCategoryRepository(tdb.DB)

	created, err := repo.Create(userID, "인사말")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act
	updated, err := repo.Update(created.ID, userID, "감사 표현")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if updated == nil {
		t.Fatal("expected non-nil category, got nil")
	}
	if updated.Name != "감사 표현" {
		t.Errorf("expected Name '감사 표현', got '%s'", updated.Name)
	}
	if updated.ID != created.ID {
		t.Errorf("expected same ID %d, got %d", created.ID, updated.ID)
	}
}

// TestSentenceCategoryRepository_Update_NotFound_ReturnsError verifies that
// Update returns ErrSentenceCategoryNotFound when no category with the given
// ID exists.
//
// Scenario:
//
//	Update(99999, userID, "없는 카테고리")
//	→ errors.Is(err, repository.ErrSentenceCategoryNotFound) == true
func TestSentenceCategoryRepository_Update_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "sc-update-notfound@example.com")
	repo := repository.NewSentenceCategoryRepository(tdb.DB)

	// Act
	updated, err := repo.Update(99999, userID, "없는 카테고리")

	// Assert
	if err == nil {
		t.Error("expected error for non-existent ID, got nil")
	}
	if updated != nil {
		t.Errorf("expected nil category on not-found, got %+v", updated)
	}
	if !errors.Is(err, repository.ErrSentenceCategoryNotFound) {
		t.Errorf("expected ErrSentenceCategoryNotFound, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

// TestSentenceCategoryRepository_Delete_RemovesRecord verifies that Delete
// removes the row so that a subsequent FindByID returns ErrSentenceCategoryNotFound.
//
// Scenario:
//
//	Seed: one category.
//	Delete(categoryID, userID)   → nil
//	FindByID(categoryID, userID) → ErrSentenceCategoryNotFound
func TestSentenceCategoryRepository_Delete_RemovesRecord(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "sc-delete@example.com")
	repo := repository.NewSentenceCategoryRepository(tdb.DB)

	created, err := repo.Create(userID, "삭제할 카테고리")
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
	if !errors.Is(findErr, repository.ErrSentenceCategoryNotFound) {
		t.Errorf("expected ErrSentenceCategoryNotFound after Delete, got: %v", findErr)
	}
}

// TestSentenceCategoryRepository_Delete_NotFound_ReturnsError verifies that
// Delete returns ErrSentenceCategoryNotFound when the ID does not exist.
//
// Scenario:
//
//	Delete(99999, userID)
//	→ errors.Is(err, repository.ErrSentenceCategoryNotFound) == true
func TestSentenceCategoryRepository_Delete_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "sc-delete-notfound@example.com")
	repo := repository.NewSentenceCategoryRepository(tdb.DB)

	// Act
	err := repo.Delete(99999, userID)

	// Assert
	if err == nil {
		t.Error("expected error for non-existent ID, got nil")
	}
	if !errors.Is(err, repository.ErrSentenceCategoryNotFound) {
		t.Errorf("expected ErrSentenceCategoryNotFound, got: %v", err)
	}
}
