// Package commitstatus keeps the commit statuses external callers ask to write, until the owner approves one and it is written to GitHub (PRD-12 AC18–AC21).
package commitstatus

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/approvals"
)

type Status string

const (
	StatusPending  Status = "PENDING"
	StatusApproved Status = "APPROVED"
	StatusRejected Status = "REJECTED"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrNotPending = errors.New("request is not pending")
)

// Request is one commit status an external caller asked to write, and the owner's decision on it.
type Request struct {
	ID             string `json:"id"`
	Status         Status `json:"status"`
	Owner          string `json:"owner"`
	Repo           string `json:"repo"`
	SHA            string `json:"sha"`
	State          string `json:"state"`
	Context        string `json:"context"`
	Description    string `json:"description"`
	TargetURL      string `json:"target_url"`
	Requester      string `json:"requester"`
	Reason         string `json:"reason"`
	SessionID      string `json:"session_id"`
	SessionName    string `json:"session_name"`
	KeyName        string `json:"key_name"`
	KeyRevoked     bool   `json:"key_revoked"`
	GitHubStatusID *int64 `json:"github_status_id"`
	LastError      string `json:"last_error"`
	CreatedAt      int64  `json:"created_at"`
	ProcessedAt    *int64 `json:"processed_at"`
}

// NewRequest is what a caller files; every field a status write needs is here.
type NewRequest struct {
	Owner       string
	Repo        string
	SHA         string
	State       string
	Context     string
	Description string
	TargetURL   string
	Requester   string
	Reason      string
	SessionID   string
	SessionName string
}

// Store persists commit status requests per user.
type Store struct {
	db  *sql.DB
	now func() time.Time
}

// New wires a Store onto an open *sql.DB.
func New(db *sql.DB) *Store { return &Store{db: db, now: time.Now} }

// NewWithClock is New with an injectable clock.
func NewWithClock(db *sql.DB, now func() time.Time) *Store { return &Store{db: db, now: now} }

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random: %w", err)
	}
	return hex.EncodeToString(b), nil
}

const requestColumns = `r.id, k.name, k.revoked_at, r.owner, r.repo, r.sha, r.state, r.context, r.description, r.target_url, r.requester, r.reason, r.session_id, r.session_name, r.status, r.github_status_id, r.last_error, r.created_at, r.processed_at`

func scanRequest(sc interface{ Scan(...any) error }) (Request, error) {
	var (
		r         Request
		revokedAt *int64
	)
	if err := sc.Scan(&r.ID, &r.KeyName, &revokedAt, &r.Owner, &r.Repo, &r.SHA, &r.State, &r.Context,
		&r.Description, &r.TargetURL, &r.Requester, &r.Reason, &r.SessionID, &r.SessionName, &r.Status, &r.GitHubStatusID,
		&r.LastError, &r.CreatedAt, &r.ProcessedAt); err != nil {
		return Request{}, err
	}
	r.KeyRevoked = revokedAt != nil
	return r, nil
}

// Create stores a pending request for the user who registered the caller.
func (s *Store) Create(ctx context.Context, caller approvals.Caller, in NewRequest) (Request, error) {
	id, err := randomID()
	if err != nil {
		return Request{}, err
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO commit_status_requests (id, user_id, caller_key_id, owner, repo, sha, state, context, description, target_url, requester, reason, session_id, session_name, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, caller.UserID, caller.KeyID, in.Owner, in.Repo, strings.ToLower(in.SHA), in.State, in.Context, in.Description, in.TargetURL,
		in.Requester, in.Reason, in.SessionID, in.SessionName, s.now().Unix()); err != nil {
		return Request{}, fmt.Errorf("insert commit status request: %w", err)
	}
	return s.Get(ctx, caller.UserID, id)
}

// Get returns one of the user's requests.
func (s *Store) Get(ctx context.Context, userID int64, id string) (Request, error) {
	r, err := scanRequest(s.db.QueryRowContext(ctx,
		`SELECT `+requestColumns+` FROM commit_status_requests r JOIN approval_caller_keys k ON k.id = r.caller_key_id WHERE r.id = ? AND r.user_id = ?`,
		id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Request{}, ErrNotFound
	}
	if err != nil {
		return Request{}, fmt.Errorf("select commit status request: %w", err)
	}
	return r, nil
}

// ListPending returns the user's undecided requests, newest first.
func (s *Store) ListPending(ctx context.Context, userID int64) ([]Request, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+requestColumns+` FROM commit_status_requests r JOIN approval_caller_keys k ON k.id = r.caller_key_id WHERE r.user_id = ? AND r.status = 'PENDING' ORDER BY r.created_at DESC, r.id DESC`,
		userID)
	return collect(rows, err)
}

// ListHistory returns the user's most recently decided requests, newest decision first.
func (s *Store) ListHistory(ctx context.Context, userID int64, limit int) ([]Request, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+requestColumns+` FROM commit_status_requests r JOIN approval_caller_keys k ON k.id = r.caller_key_id WHERE r.user_id = ? AND r.status IN ('APPROVED', 'REJECTED') ORDER BY r.processed_at DESC, r.id DESC LIMIT ?`,
		userID, limit)
	return collect(rows, err)
}

func collect(rows *sql.Rows, err error) ([]Request, error) {
	if err != nil {
		return nil, fmt.Errorf("list commit status requests: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := []Request{}
	for rows.Next() {
		r, err := scanRequest(rows)
		if err != nil {
			return nil, fmt.Errorf("scan commit status request: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) markApproved(ctx context.Context, userID int64, id string, githubStatusID int64) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE commit_status_requests SET status = 'APPROVED', github_status_id = ?, last_error = '', processed_at = ? WHERE id = ? AND user_id = ? AND status = 'PENDING'`,
		githubStatusID, s.now().Unix(), id, userID)
	return expectOne(res, err, "approve commit status request")
}

func (s *Store) markRejected(ctx context.Context, userID int64, id string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE commit_status_requests SET status = 'REJECTED', processed_at = ? WHERE id = ? AND user_id = ? AND status = 'PENDING'`,
		s.now().Unix(), id, userID)
	return expectOne(res, err, "reject commit status request")
}

func (s *Store) setLastError(ctx context.Context, userID int64, id, msg string) error {
	if _, err := s.db.ExecContext(ctx,
		`UPDATE commit_status_requests SET last_error = ? WHERE id = ? AND user_id = ? AND status = 'PENDING'`,
		msg, id, userID); err != nil {
		return fmt.Errorf("record commit status write error: %w", err)
	}
	return nil
}

func expectOne(res sql.Result, err error, what string) error {
	if err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	if n == 0 {
		return ErrNotPending
	}
	return nil
}
