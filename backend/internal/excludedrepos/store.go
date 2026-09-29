// Package excludedrepos is the storage layer for the PR-monitor blacklist
// (PRD-10 AC6).
package excludedrepos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var repoFormat = regexp.MustCompile(`^[\w.\-]+/[\w.\-]+$`)

// ErrInvalidRepo is returned by Add when the supplied name doesn't look like
// `owner/repo`.
var ErrInvalidRepo = errors.New("repo must match owner/repo format")

// ErrNotFound is returned when a query targets a row that does not exist or
// is not owned by the caller.
var ErrNotFound = errors.New("excluded repo not found")

// ExcludedRepo is one row in the user_excluded_repos table.
type ExcludedRepo struct {
	ID           int64  `json:"id"`
	RepoFullName string `json:"repo_full_name"`
	CreatedAt    int64  `json:"created_at"`
}

// Store is the user-scoped facade.
type Store struct {
	db *sql.DB
}

// New wires a Store onto an open *sql.DB.
func New(db *sql.DB) *Store { return &Store{db: db} }

// List returns every excluded-repo row for userID, newest first.
func (s *Store) List(ctx context.Context, userID int64) ([]ExcludedRepo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, repo_full_name, created_at
		FROM user_excluded_repos
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list excluded repos: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]ExcludedRepo, 0)
	for rows.Next() {
		var r ExcludedRepo
		if err := rows.Scan(&r.ID, &r.RepoFullName, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan excluded repo: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

// Add inserts a new exclusion.
func (s *Store) Add(ctx context.Context, userID int64, repo string) (ExcludedRepo, error) {
	if !repoFormat.MatchString(repo) {
		return ExcludedRepo{}, ErrInvalidRepo
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO user_excluded_repos (user_id, repo_full_name)
		VALUES (?, ?)
	`, userID, repo)
	if err != nil {
		if isUniqueViolation(err) {
			return ExcludedRepo{}, ErrAlreadyExcluded
		}
		return ExcludedRepo{}, fmt.Errorf("insert excluded repo: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return ExcludedRepo{}, fmt.Errorf("last insert id: %w", err)
	}
	var r ExcludedRepo
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, repo_full_name, created_at
		FROM user_excluded_repos
		WHERE id = ? AND user_id = ?
	`, id, userID).Scan(&r.ID, &r.RepoFullName, &r.CreatedAt); err != nil {
		return ExcludedRepo{}, fmt.Errorf("reload excluded repo: %w", err)
	}
	return r, nil
}

// ErrAlreadyExcluded is returned when Add hits the (user, repo) unique index.
var ErrAlreadyExcluded = errors.New("repo already excluded")

// Delete removes an exclusion owned by userID.
func (s *Store) Delete(ctx context.Context, userID, id int64) error {
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM user_excluded_repos WHERE id = ? AND user_id = ?
	`, id, userID)
	if err != nil {
		return fmt.Errorf("delete excluded repo: %w", err)
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

// ListUserIDsExcluding returns every user_id that has NOT excluded `repo`.
// A user with no exclusions at all matches every repo (blacklist default =
// empty), so the query starts from all provisioned users rather than joining
// on the exclusions table.
func (s *Store) ListUserIDsExcluding(ctx context.Context, repo string) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id FROM users
		WHERE id NOT IN (
			SELECT user_id FROM user_excluded_repos WHERE repo_full_name = ?
		)
		ORDER BY id
	`, repo)
	if err != nil {
		return nil, fmt.Errorf("list matched users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan user id: %w", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE") && strings.Contains(msg, "constraint")
}
