// Package notificationhistory persists CI-completion events per user so the
// PR-monitor screen can show a history (PRD-10 AC11) and the user can
// explicitly acknowledge each item (AC12).
//
// One workflow_run event becomes N rows — one per matched user — so each
// user's `acknowledged_at` is independent. The dispatch pipeline uses
// InsertBatchTx to write every row inside a single transaction so a partial
// failure leaves no orphan rows and the SQS message can be retried cleanly.
package notificationhistory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrNotFound is returned when a query targets a row that does not exist or
// is not owned by the caller.
var ErrNotFound = errors.New("notification history item not found")

// Event is the workflow_run-completed shape supplied by the dispatch
// pipeline. PR fields are optional (zero values mean "no PR linked").
type Event struct {
	RepoFullName string
	PRNumber     int    // 0 when no PR linked
	PRTitle      string // "" when no PR linked
	PRURL        string // "" when no PR linked
	CommitURL    string
	RunURL       string
	WorkflowName string
	HeadBranch   string
	HeadSHA      string // commit SHA — grouping key fallback when no PR is linked (AC13)
	Conclusion   string
}

// Item is one persisted row. Pointer fields use sql.Null* under the hood
// during scan but are exposed as Go-native pointers so the JSON layer can
// emit nulls cleanly.
type Item struct {
	ID             int64   `json:"id"`
	RepoFullName   string  `json:"repo_full_name"`
	PRNumber       *int    `json:"pr_number,omitempty"`
	PRTitle        *string `json:"pr_title,omitempty"`
	PRURL          *string `json:"pr_url,omitempty"`
	CommitURL      *string `json:"commit_url,omitempty"`
	RunURL         *string `json:"run_url,omitempty"`
	WorkflowName   string  `json:"workflow_name"`
	HeadBranch     string  `json:"head_branch"`
	HeadSHA        string  `json:"head_sha"`
	Conclusion     string  `json:"conclusion"`
	AcknowledgedAt *int64  `json:"acknowledged_at,omitempty"`
	CreatedAt      int64   `json:"created_at"`
}

// Store is the user-scoped facade.
type Store struct {
	db *sql.DB
}

// New wires a Store onto an open *sql.DB.
func New(db *sql.DB) *Store { return &Store{db: db} }

// List returns up to `limit` history items belonging to userID, newest
// first. When beforeID > 0 the result is keyset-paginated: only rows with
// id < beforeID are returned. Limit is clamped to [1, 100].
func (s *Store) List(ctx context.Context, userID int64, limit int, beforeID int64) ([]Item, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	var (
		rows *sql.Rows
		err  error
	)
	if beforeID > 0 {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, repo_full_name, pr_number, pr_title, pr_url,
			       commit_url, run_url, workflow_name, head_branch,
			       head_sha, conclusion, acknowledged_at, created_at
			FROM notification_history
			WHERE user_id = ? AND id < ?
			ORDER BY id DESC
			LIMIT ?
		`, userID, beforeID, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, repo_full_name, pr_number, pr_title, pr_url,
			       commit_url, run_url, workflow_name, head_branch,
			       head_sha, conclusion, acknowledged_at, created_at
			FROM notification_history
			WHERE user_id = ?
			ORDER BY id DESC
			LIMIT ?
		`, userID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("list notification history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]Item, 0)
	for rows.Next() {
		var (
			it      Item
			prNum   sql.NullInt64
			prTitle sql.NullString
			prURL   sql.NullString
			commit  sql.NullString
			run     sql.NullString
			ackedAt sql.NullInt64
		)
		if err := rows.Scan(
			&it.ID, &it.RepoFullName, &prNum, &prTitle, &prURL,
			&commit, &run, &it.WorkflowName, &it.HeadBranch,
			&it.HeadSHA, &it.Conclusion, &ackedAt, &it.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification history: %w", err)
		}
		if prNum.Valid {
			n := int(prNum.Int64)
			it.PRNumber = &n
		}
		if prTitle.Valid {
			s := prTitle.String
			it.PRTitle = &s
		}
		if prURL.Valid {
			s := prURL.String
			it.PRURL = &s
		}
		if commit.Valid {
			s := commit.String
			it.CommitURL = &s
		}
		if run.Valid {
			s := run.String
			it.RunURL = &s
		}
		if ackedAt.Valid {
			t := ackedAt.Int64
			it.AcknowledgedAt = &t
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

// Acknowledge stamps the supplied row with the current unix-epoch time. The
// (user_id, id) filter means callers cannot ack rows that belong to other
// users — those return ErrNotFound, never 200 — which is what AC12 needs.
//
// Idempotent: re-acknowledging an already-acked row leaves the original
// timestamp untouched. The UPDATE filter `acknowledged_at IS NULL` would
// detect the no-op, but we prefer ErrNotFound semantics only for "not your
// row," so we just let the update succeed silently.
func (s *Store) Acknowledge(ctx context.Context, userID, id int64) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE notification_history
		SET acknowledged_at = CASE
			WHEN acknowledged_at IS NULL THEN unixepoch()
			ELSE acknowledged_at
		END
		WHERE id = ? AND user_id = ?
	`, id, userID)
	if err != nil {
		return fmt.Errorf("ack notification history: %w", err)
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

// InsertBatchTx inserts one row per userID inside a single transaction.
// Returns the inserted row IDs in the same order as userIDs. Roll back on
// the first failure so partial state never reaches the table — the SQS
// message is then retried on visibility timeout expiry.
func (s *Store) InsertBatchTx(ctx context.Context, userIDs []int64, evt Event) ([]int64, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO notification_history (
			user_id, repo_full_name, pr_number, pr_title, pr_url,
			commit_url, run_url, workflow_name, head_branch, head_sha, conclusion
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	ids := make([]int64, 0, len(userIDs))
	for _, uid := range userIDs {
		res, err := stmt.ExecContext(ctx, uid,
			evt.RepoFullName,
			nullableInt(evt.PRNumber),
			nullableString(evt.PRTitle),
			nullableString(evt.PRURL),
			nullableString(evt.CommitURL),
			nullableString(evt.RunURL),
			evt.WorkflowName,
			evt.HeadBranch,
			evt.HeadSHA,
			evt.Conclusion,
		)
		if err != nil {
			return nil, fmt.Errorf("insert history for user %d: %w", uid, err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("last insert id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return ids, nil
}

// Get returns a single row by id, scoped to userID. Used by tests; production
// reads always come through List or the dispatch path.
func (s *Store) Get(ctx context.Context, userID, id int64) (Item, error) {
	items, err := s.db.QueryContext(ctx, `
		SELECT id, repo_full_name, pr_number, pr_title, pr_url,
		       commit_url, run_url, workflow_name, head_branch,
		       head_sha, conclusion, acknowledged_at, created_at
		FROM notification_history
		WHERE id = ? AND user_id = ?
	`, id, userID)
	if err != nil {
		return Item{}, fmt.Errorf("get history: %w", err)
	}
	defer func() { _ = items.Close() }()
	if !items.Next() {
		if err := items.Err(); err != nil {
			return Item{}, fmt.Errorf("rows err: %w", err)
		}
		return Item{}, ErrNotFound
	}
	var (
		it      Item
		prNum   sql.NullInt64
		prTitle sql.NullString
		prURL   sql.NullString
		commit  sql.NullString
		run     sql.NullString
		ackedAt sql.NullInt64
	)
	if err := items.Scan(
		&it.ID, &it.RepoFullName, &prNum, &prTitle, &prURL,
		&commit, &run, &it.WorkflowName, &it.HeadBranch,
		&it.HeadSHA, &it.Conclusion, &ackedAt, &it.CreatedAt,
	); err != nil {
		return Item{}, fmt.Errorf("scan: %w", err)
	}
	if prNum.Valid {
		n := int(prNum.Int64)
		it.PRNumber = &n
	}
	if prTitle.Valid {
		s := prTitle.String
		it.PRTitle = &s
	}
	if prURL.Valid {
		s := prURL.String
		it.PRURL = &s
	}
	if commit.Valid {
		s := commit.String
		it.CommitURL = &s
	}
	if run.Valid {
		s := run.String
		it.RunURL = &s
	}
	if ackedAt.Valid {
		t := ackedAt.Int64
		it.AcknowledgedAt = &t
	}
	return it, nil
}

func nullableInt(v int) any {
	if v == 0 {
		return nil
	}
	return v
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
