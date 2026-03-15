// Package repository_test contains unit tests for the SentenceRepository.
// These tests use an in-memory SQLite database via testutil.NewTestDB and
// verify that SentenceRepository correctly reads and writes sentences records.
//
// TDD Red Phase: SentenceRepository and its methods do not yet exist.
// All tests are expected to fail until repository/sentence_repository.go
// is implemented and the following DB migrations are applied:
//
// Migration 000007_sentence_categories.up.sql (prerequisite):
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
// Migration 000008_sentences.up.sql:
//
//	CREATE TABLE IF NOT EXISTS sentences (
//	    id          INTEGER PRIMARY KEY AUTOINCREMENT,
//	    user_id     INTEGER NOT NULL,
//	    category_id INTEGER NOT NULL,
//	    content     TEXT    NOT NULL,
//	    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
//	    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
//	    FOREIGN KEY (user_id)     REFERENCES users(id),
//	    FOREIGN KEY (category_id) REFERENCES sentence_categories(id)
//	);
//
// Migration 000008_sentences.down.sql:
//
//	DROP TABLE IF EXISTS sentences;
package repository_test

import (
	"errors"
	"testing"

	"github.com/dlddu/pocket-aide/repository"
	"github.com/dlddu/pocket-aide/testutil"
)

// seedSentenceCategory inserts a sentence_categories row and returns its
// auto-assigned ID. It is used by sentence tests to satisfy the foreign key
// constraint on sentences.category_id.
func seedSentenceCategory(t *testing.T, tdb *testutil.TestDB, userID int64, name string) int64 {
	t.Helper()
	result, err := tdb.DB.Exec(
		"INSERT INTO sentence_categories (user_id, name) VALUES (?, ?)",
		userID, name,
	)
	if err != nil {
		t.Fatalf("seedSentenceCategory: failed to insert category %q: %v", name, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("seedSentenceCategory: failed to get last insert ID: %v", err)
	}
	return id
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

// TestSentenceRepository_Create_ReturnsSentence verifies that Create inserts a
// sentences record and returns the persisted sentence with a non-zero ID, the
// correct user_id, content, and category_id.
//
// Scenario:
//
//	Seed: user + category "인사말".
//	Create(userID, categoryID, "안녕하세요")
//	→ Sentence{ID: >0, UserID: userID, Content: "안녕하세요", CategoryID: categoryID}, nil
func TestSentenceRepository_Create_ReturnsSentence(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "s-create@example.com")
	categoryID := seedSentenceCategory(t, tdb, userID, "인사말")
	repo := repository.NewSentenceRepository(tdb.DB)

	// Act
	sentence, err := repo.Create(userID, categoryID, "안녕하세요")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if sentence == nil {
		t.Fatal("expected non-nil sentence, got nil")
	}
	if sentence.ID == 0 {
		t.Error("expected sentence ID to be non-zero after creation")
	}
	if sentence.UserID != userID {
		t.Errorf("expected UserID %d, got %d", userID, sentence.UserID)
	}
	if sentence.Content != "안녕하세요" {
		t.Errorf("expected Content '안녕하세요', got '%s'", sentence.Content)
	}
	if sentence.CategoryID != categoryID {
		t.Errorf("expected CategoryID %d, got %d", categoryID, sentence.CategoryID)
	}
}

// TestSentenceRepository_Create_EmptyContent_ReturnsError verifies that Create
// returns a non-nil error when an empty content string is supplied, without
// inserting any record.
//
// Scenario:
//
//	Create(userID, categoryID, "")
//	→ non-nil error, nil sentence
func TestSentenceRepository_Create_EmptyContent_ReturnsError(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "s-empty-content@example.com")
	categoryID := seedSentenceCategory(t, tdb, userID, "인사말")
	repo := repository.NewSentenceRepository(tdb.DB)

	// Act
	sentence, err := repo.Create(userID, categoryID, "")

	// Assert
	if err == nil {
		t.Error("expected error for empty content, got nil")
	}
	if sentence != nil {
		t.Errorf("expected nil sentence on error, got %+v", sentence)
	}
}

// ---------------------------------------------------------------------------
// ListByUserID
// ---------------------------------------------------------------------------

// TestSentenceRepository_ListByUserID_ReturnsList verifies that ListByUserID
// returns all sentences belonging to the given user.
//
// Scenario:
//
//	Seed: user with two sentences in the same category.
//	ListByUserID(userID)
//	→ []*Sentence with len == 2
func TestSentenceRepository_ListByUserID_ReturnsList(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "s-list@example.com")
	categoryID := seedSentenceCategory(t, tdb, userID, "인사말")
	repo := repository.NewSentenceRepository(tdb.DB)

	if _, err := repo.Create(userID, categoryID, "안녕하세요"); err != nil {
		t.Fatalf("setup: Create '안녕하세요' failed: %v", err)
	}
	if _, err := repo.Create(userID, categoryID, "반갑습니다"); err != nil {
		t.Fatalf("setup: Create '반갑습니다' failed: %v", err)
	}

	// Act
	sentences, err := repo.ListByUserID(userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(sentences) != 2 {
		t.Fatalf("expected 2 sentences, got %d", len(sentences))
	}
	for i, s := range sentences {
		if s.UserID != userID {
			t.Errorf("sentences[%d]: expected UserID %d, got %d", i, userID, s.UserID)
		}
	}
}

// TestSentenceRepository_ListByUserID_Empty_ReturnsEmptySlice verifies that
// ListByUserID returns a non-nil empty slice (not nil) when the user has no
// sentences.
//
// Scenario:
//
//	Seed: user with no sentences.
//	ListByUserID(userID)
//	→ non-nil empty slice, nil error
func TestSentenceRepository_ListByUserID_Empty_ReturnsEmptySlice(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "s-list-empty@example.com")
	repo := repository.NewSentenceRepository(tdb.DB)

	// Act
	sentences, err := repo.ListByUserID(userID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if sentences == nil {
		t.Error("expected non-nil slice, got nil")
	}
	if len(sentences) != 0 {
		t.Errorf("expected empty slice, got %d sentences", len(sentences))
	}
}

// ---------------------------------------------------------------------------
// ListByCategoryID
// ---------------------------------------------------------------------------

// TestSentenceRepository_ListByCategoryID_ReturnsFiltered verifies that
// ListByCategoryID returns only the sentences belonging to the specified
// category, excluding sentences from other categories owned by the same user.
//
// Scenario:
//
//	Seed: user with two categories; one sentence in each.
//	ListByCategoryID(userID, categoryID_1)
//	→ []*Sentence with len == 1, all CategoryID == categoryID_1
func TestSentenceRepository_ListByCategoryID_ReturnsFiltered(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "s-list-category@example.com")
	cat1ID := seedSentenceCategory(t, tdb, userID, "인사말")
	cat2ID := seedSentenceCategory(t, tdb, userID, "감사 표현")
	repo := repository.NewSentenceRepository(tdb.DB)

	if _, err := repo.Create(userID, cat1ID, "안녕하세요"); err != nil {
		t.Fatalf("setup: Create sentence in cat1 failed: %v", err)
	}
	if _, err := repo.Create(userID, cat2ID, "감사합니다"); err != nil {
		t.Fatalf("setup: Create sentence in cat2 failed: %v", err)
	}

	// Act
	sentences, err := repo.ListByCategoryID(userID, cat1ID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(sentences) != 1 {
		t.Fatalf("expected 1 sentence for category_id=%d, got %d", cat1ID, len(sentences))
	}
	if sentences[0].CategoryID != cat1ID {
		t.Errorf("expected CategoryID %d, got %d", cat1ID, sentences[0].CategoryID)
	}
	if sentences[0].Content != "안녕하세요" {
		t.Errorf("expected Content '안녕하세요', got '%s'", sentences[0].Content)
	}
}

// ---------------------------------------------------------------------------
// FindByID
// ---------------------------------------------------------------------------

// TestSentenceRepository_FindByID_ReturnsSentence verifies that FindByID
// returns the matching sentence when the ID and userID are both correct.
//
// Scenario:
//
//	Seed: one sentence for user.
//	FindByID(sentenceID, userID)
//	→ Sentence{Content: "안녕하세요"}, nil
func TestSentenceRepository_FindByID_ReturnsSentence(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "s-find@example.com")
	categoryID := seedSentenceCategory(t, tdb, userID, "인사말")
	repo := repository.NewSentenceRepository(tdb.DB)

	created, err := repo.Create(userID, categoryID, "안녕하세요")
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
		t.Fatal("expected non-nil sentence, got nil")
	}
	if found.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, found.ID)
	}
	if found.Content != "안녕하세요" {
		t.Errorf("expected Content '안녕하세요', got '%s'", found.Content)
	}
}

// TestSentenceRepository_FindByID_NotFound_ReturnsError verifies that
// FindByID returns ErrSentenceNotFound when no sentence with the given ID
// exists (or it belongs to a different user).
//
// Scenario:
//
//	FindByID(99999, userID)
//	→ errors.Is(err, repository.ErrSentenceNotFound) == true
func TestSentenceRepository_FindByID_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "s-find-notfound@example.com")
	repo := repository.NewSentenceRepository(tdb.DB)

	// Act
	found, err := repo.FindByID(99999, userID)

	// Assert
	if err == nil {
		t.Error("expected error for non-existent ID, got nil")
	}
	if found != nil {
		t.Errorf("expected nil sentence when not found, got %+v", found)
	}
	if !errors.Is(err, repository.ErrSentenceNotFound) {
		t.Errorf("expected ErrSentenceNotFound, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

// TestSentenceRepository_Update_ReturnsUpdated verifies that Update persists
// the new content and returns the updated sentence.
//
// Scenario:
//
//	Seed: sentence{content:"안녕하세요"}.
//	Update(sentenceID, userID, "안녕히 가세요")
//	→ Sentence{Content: "안녕히 가세요"}, nil
func TestSentenceRepository_Update_ReturnsUpdated(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "s-update@example.com")
	categoryID := seedSentenceCategory(t, tdb, userID, "인사말")
	repo := repository.NewSentenceRepository(tdb.DB)

	created, err := repo.Create(userID, categoryID, "안녕하세요")
	if err != nil {
		t.Fatalf("setup: Create failed: %v", err)
	}

	// Act
	updated, err := repo.Update(created.ID, userID, "안녕히 가세요")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if updated == nil {
		t.Fatal("expected non-nil sentence, got nil")
	}
	if updated.Content != "안녕히 가세요" {
		t.Errorf("expected Content '안녕히 가세요', got '%s'", updated.Content)
	}
	if updated.ID != created.ID {
		t.Errorf("expected same ID %d, got %d", created.ID, updated.ID)
	}
}

// TestSentenceRepository_Update_NotFound_ReturnsError verifies that Update
// returns ErrSentenceNotFound when no sentence with the given ID exists.
//
// Scenario:
//
//	Update(99999, userID, "없는 문장")
//	→ errors.Is(err, repository.ErrSentenceNotFound) == true
func TestSentenceRepository_Update_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "s-update-notfound@example.com")
	repo := repository.NewSentenceRepository(tdb.DB)

	// Act
	updated, err := repo.Update(99999, userID, "없는 문장")

	// Assert
	if err == nil {
		t.Error("expected error for non-existent ID, got nil")
	}
	if updated != nil {
		t.Errorf("expected nil sentence on not-found, got %+v", updated)
	}
	if !errors.Is(err, repository.ErrSentenceNotFound) {
		t.Errorf("expected ErrSentenceNotFound, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

// TestSentenceRepository_Delete_RemovesRecord verifies that Delete removes the
// row so that a subsequent FindByID returns ErrSentenceNotFound.
//
// Scenario:
//
//	Seed: one sentence.
//	Delete(sentenceID, userID)   → nil
//	FindByID(sentenceID, userID) → ErrSentenceNotFound
func TestSentenceRepository_Delete_RemovesRecord(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "s-delete@example.com")
	categoryID := seedSentenceCategory(t, tdb, userID, "인사말")
	repo := repository.NewSentenceRepository(tdb.DB)

	created, err := repo.Create(userID, categoryID, "삭제할 문장")
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
	if !errors.Is(findErr, repository.ErrSentenceNotFound) {
		t.Errorf("expected ErrSentenceNotFound after Delete, got: %v", findErr)
	}
}

// TestSentenceRepository_Delete_NotFound_ReturnsError verifies that Delete
// returns ErrSentenceNotFound when the ID does not exist.
//
// Scenario:
//
//	Delete(99999, userID)
//	→ errors.Is(err, repository.ErrSentenceNotFound) == true
func TestSentenceRepository_Delete_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	userID := seedUser(t, tdb, "s-delete-notfound@example.com")
	repo := repository.NewSentenceRepository(tdb.DB)

	// Act
	err := repo.Delete(99999, userID)

	// Assert
	if err == nil {
		t.Error("expected error for non-existent ID, got nil")
	}
	if !errors.Is(err, repository.ErrSentenceNotFound) {
		t.Errorf("expected ErrSentenceNotFound, got: %v", err)
	}
}
