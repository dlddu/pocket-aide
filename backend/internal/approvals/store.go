package approvals

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type Status string

const (
	StatusPending  Status = "PENDING"
	StatusApproved Status = "APPROVED"
	StatusRejected Status = "REJECTED"
	StatusExpired  Status = "EXPIRED"
)

type Decision string

const (
	DecisionApprove Decision = "approve"
	DecisionReject  Decision = "reject"
)

const (
	keyRawPrefix   = "pak_"
	keyPrefixChars = 12
)

var (
	ErrNotFound          = errors.New("not found")
	ErrDuplicateName     = errors.New("caller name already in use")
	ErrDuplicateExternal = errors.New("externalId already used by this caller key")
	ErrAlreadyRevoked    = errors.New("caller key already revoked")
	ErrInvalidKey        = errors.New("invalid caller key")
	ErrNotPending        = errors.New("request is not pending")
	ErrInvalidCursor     = errors.New("invalid history cursor")
	ErrInvalidMode       = errors.New("auto_response must be one of NONE, AUTO_APPROVE, AUTO_REJECT")
)

// AutoResponse is a user's row in user_approval_settings: how requests filed by caller keys are processed at creation; a user without a row gets AutoResponseNone.
type AutoResponse string

const (
	AutoResponseNone    AutoResponse = "NONE"
	AutoResponseApprove AutoResponse = "AUTO_APPROVE"
	AutoResponseReject  AutoResponse = "AUTO_REJECT"
)

func (m AutoResponse) Valid() bool {
	switch m {
	case AutoResponseNone, AutoResponseApprove, AutoResponseReject:
		return true
	}
	return false
}

// CallerKey is a row of approval_caller_keys: a named key an external caller presents to file approval requests for one user.
type CallerKey struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	KeyPrefix  string `json:"key_prefix"`
	CreatedAt  int64  `json:"created_at"`
	LastUsedAt *int64 `json:"last_used_at"`
	RevokedAt  *int64 `json:"revoked_at"`
}

type IssuedKey struct {
	CallerKey
	Key string `json:"key"`
}

// Request is a row of approval_requests: one approval asked through a caller key, and the user's decision on it.
type Request struct {
	ID            string
	UserID        int64
	CallerKeyID   int64
	CallerKeyName string
	KeyRevoked    bool
	ExternalID    string
	Context       string
	RequesterName string
	Status        Status
	DecisionMode  string
	CreatedAt     int64
	ExpiresAt     *int64
	ProcessedAt   *int64
}

type NewRequest struct {
	ExternalID     string
	Context        string
	RequesterName  string
	TimeoutSeconds *int64
}

type Store struct {
	db  *sql.DB
	now func() time.Time
}

func New(db *sql.DB) *Store {
	return &Store{db: db, now: time.Now}
}

func NewWithClock(db *sql.DB, now func() time.Time) *Store {
	return &Store{db: db, now: now}
}

func hashKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func (s *Store) IssueKey(ctx context.Context, userID int64, name string) (IssuedKey, error) {
	secret, err := randomToken(32)
	if err != nil {
		return IssuedKey{}, err
	}
	raw := keyRawPrefix + secret
	k := IssuedKey{
		CallerKey: CallerKey{Name: name, KeyPrefix: raw[:keyPrefixChars], CreatedAt: s.now().Unix()},
		Key:       raw,
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO approval_caller_keys (user_id, name, key_hash, key_prefix, created_at) VALUES (?, ?, ?, ?, ?)`,
		userID, name, hashKey(raw), k.KeyPrefix, k.CreatedAt)
	if isUniqueViolation(err) {
		return IssuedKey{}, ErrDuplicateName
	}
	if err != nil {
		return IssuedKey{}, fmt.Errorf("insert caller key: %w", err)
	}
	if k.ID, err = res.LastInsertId(); err != nil {
		return IssuedKey{}, fmt.Errorf("caller key id: %w", err)
	}
	return k, nil
}

func (s *Store) ListKeys(ctx context.Context, userID int64) ([]CallerKey, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, key_prefix, created_at, last_used_at, revoked_at FROM approval_caller_keys WHERE user_id = ? ORDER BY created_at DESC, id DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("list caller keys: %w", err)
	}
	defer func() { _ = rows.Close() }()
	keys := []CallerKey{}
	for rows.Next() {
		var k CallerKey
		if err := rows.Scan(&k.ID, &k.Name, &k.KeyPrefix, &k.CreatedAt, &k.LastUsedAt, &k.RevokedAt); err != nil {
			return nil, fmt.Errorf("scan caller key: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (s *Store) RevokeKey(ctx context.Context, userID, keyID int64) (CallerKey, error) {
	now := s.now().Unix()
	res, err := s.db.ExecContext(ctx,
		`UPDATE approval_caller_keys SET revoked_at = ? WHERE id = ? AND user_id = ? AND revoked_at IS NULL`,
		now, keyID, userID)
	if err != nil {
		return CallerKey{}, fmt.Errorf("revoke caller key: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return CallerKey{}, fmt.Errorf("revoke caller key: %w", err)
	}
	k, err := s.getKey(ctx, userID, keyID)
	if err != nil {
		return CallerKey{}, err
	}
	if n == 0 {
		return k, ErrAlreadyRevoked
	}
	return k, nil
}

func (s *Store) getKey(ctx context.Context, userID, keyID int64) (CallerKey, error) {
	var k CallerKey
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, key_prefix, created_at, last_used_at, revoked_at FROM approval_caller_keys WHERE id = ? AND user_id = ?`,
		keyID, userID).Scan(&k.ID, &k.Name, &k.KeyPrefix, &k.CreatedAt, &k.LastUsedAt, &k.RevokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return CallerKey{}, ErrNotFound
	}
	if err != nil {
		return CallerKey{}, fmt.Errorf("select caller key: %w", err)
	}
	return k, nil
}

type Caller struct {
	KeyID  int64
	UserID int64
	Name   string
}

func (s *Store) Authenticate(ctx context.Context, raw string) (Caller, error) {
	if !strings.HasPrefix(raw, keyRawPrefix) {
		return Caller{}, ErrInvalidKey
	}
	var c Caller
	var revokedAt *int64
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, revoked_at FROM approval_caller_keys WHERE key_hash = ?`,
		hashKey(raw)).Scan(&c.KeyID, &c.UserID, &c.Name, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Caller{}, ErrInvalidKey
	}
	if err != nil {
		return Caller{}, fmt.Errorf("select caller key: %w", err)
	}
	if revokedAt != nil {
		return Caller{}, ErrInvalidKey
	}
	if _, err := s.db.ExecContext(ctx,
		`UPDATE approval_caller_keys SET last_used_at = ? WHERE id = ?`, s.now().Unix(), c.KeyID); err != nil {
		return Caller{}, fmt.Errorf("touch caller key: %w", err)
	}
	return c, nil
}

const requestColumns = `r.id, r.user_id, r.caller_key_id, k.name, k.revoked_at, r.external_id, r.context, r.requester_name, r.status, r.decision_mode, r.created_at, r.expires_at, r.processed_at`

func (s *Store) scanRequest(sc interface{ Scan(...any) error }) (Request, error) {
	var (
		r         Request
		revokedAt *int64
		mode      sql.NullString
	)
	if err := sc.Scan(&r.ID, &r.UserID, &r.CallerKeyID, &r.CallerKeyName, &revokedAt, &r.ExternalID, &r.Context,
		&r.RequesterName, &r.Status, &mode, &r.CreatedAt, &r.ExpiresAt, &r.ProcessedAt); err != nil {
		return Request{}, err
	}
	r.KeyRevoked = revokedAt != nil
	r.DecisionMode = mode.String
	if r.Status == StatusPending && r.ExpiresAt != nil && *r.ExpiresAt <= s.now().Unix() {
		r.Status = StatusExpired
		r.ProcessedAt = r.ExpiresAt
		r.DecisionMode = "expired"
	}
	return r, nil
}

func (s *Store) AutoResponse(ctx context.Context, userID int64) (AutoResponse, error) {
	var mode string
	err := s.db.QueryRowContext(ctx,
		`SELECT auto_response FROM user_approval_settings WHERE user_id = ?`, userID).Scan(&mode)
	if errors.Is(err, sql.ErrNoRows) {
		return AutoResponseNone, nil
	}
	if err != nil {
		return "", fmt.Errorf("select approval settings: %w", err)
	}
	return AutoResponse(mode), nil
}

func (s *Store) SetAutoResponse(ctx context.Context, userID int64, mode AutoResponse) (AutoResponse, error) {
	if !mode.Valid() {
		return "", ErrInvalidMode
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO user_approval_settings (user_id, auto_response) VALUES (?, ?) ON CONFLICT(user_id) DO UPDATE SET auto_response = excluded.auto_response, updated_at = unixepoch()`,
		userID, string(mode)); err != nil {
		return "", fmt.Errorf("upsert approval settings: %w", err)
	}
	return mode, nil
}

func (s *Store) Create(ctx context.Context, caller Caller, in NewRequest) (Request, error) {
	id, err := randomID()
	if err != nil {
		return Request{}, err
	}
	mode, err := s.AutoResponse(ctx, caller.UserID)
	if err != nil {
		return Request{}, err
	}
	created := s.now().Unix()
	var expiresAt *int64
	if in.TimeoutSeconds != nil {
		e := created + *in.TimeoutSeconds
		expiresAt = &e
	}
	status := StatusPending
	var decisionMode *string
	var processedAt *int64
	if mode != AutoResponseNone {
		status = StatusApproved
		if mode == AutoResponseReject {
			status = StatusRejected
		}
		auto := "auto"
		decisionMode, processedAt = &auto, &created
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO approval_requests (id, user_id, caller_key_id, external_id, context, requester_name, status, decision_mode, created_at, expires_at, processed_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, caller.UserID, caller.KeyID, in.ExternalID, in.Context, in.RequesterName, status, decisionMode, created, expiresAt, processedAt)
	if isUniqueViolation(err) {
		return Request{}, ErrDuplicateExternal
	}
	if err != nil {
		return Request{}, fmt.Errorf("insert approval request: %w", err)
	}
	return s.GetForCaller(ctx, caller, id)
}

func (s *Store) GetForCaller(ctx context.Context, caller Caller, id string) (Request, error) {
	r, err := s.scanRequest(s.db.QueryRowContext(ctx,
		`SELECT `+requestColumns+` FROM approval_requests r JOIN approval_caller_keys k ON k.id = r.caller_key_id WHERE r.id = ? AND r.caller_key_id = ?`,
		id, caller.KeyID))
	if errors.Is(err, sql.ErrNoRows) {
		return Request{}, ErrNotFound
	}
	if err != nil {
		return Request{}, fmt.Errorf("select approval request: %w", err)
	}
	return r, nil
}

func (s *Store) Get(ctx context.Context, userID int64, id string) (Request, error) {
	r, err := s.scanRequest(s.db.QueryRowContext(ctx,
		`SELECT `+requestColumns+` FROM approval_requests r JOIN approval_caller_keys k ON k.id = r.caller_key_id WHERE r.id = ? AND r.user_id = ?`,
		id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Request{}, ErrNotFound
	}
	if err != nil {
		return Request{}, fmt.Errorf("select approval request: %w", err)
	}
	return r, nil
}

func (s *Store) ListPending(ctx context.Context, userID int64) ([]Request, error) {
	now := s.now().Unix()
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+requestColumns+` FROM approval_requests r JOIN approval_caller_keys k ON k.id = r.caller_key_id WHERE r.user_id = ? AND r.status = 'PENDING' AND (r.expires_at IS NULL OR r.expires_at > ?) ORDER BY r.created_at DESC, r.id DESC`,
		userID, now)
	if err != nil {
		return nil, fmt.Errorf("list approval requests: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := []Request{}
	for rows.Next() {
		r, err := s.scanRequest(rows)
		if err != nil {
			return nil, fmt.Errorf("scan approval request: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) Decide(ctx context.Context, userID int64, id string, d Decision) (Request, error) {
	status := StatusApproved
	if d == DecisionReject {
		status = StatusRejected
	}
	now := s.now().Unix()
	res, err := s.db.ExecContext(ctx,
		`UPDATE approval_requests SET status = ?, decision_mode = 'manual', processed_at = ? WHERE id = ? AND user_id = ? AND status = 'PENDING' AND (expires_at IS NULL OR expires_at > ?)`,
		status, now, id, userID, now)
	if err != nil {
		return Request{}, fmt.Errorf("decide approval request: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Request{}, fmt.Errorf("decide approval request: %w", err)
	}
	r, err := s.Get(ctx, userID, id)
	if err != nil {
		return Request{}, err
	}
	if n == 0 {
		return r, ErrNotPending
	}
	return r, nil
}

const HistoryPageSize = 20

type HistoryPage struct {
	Items      []Request
	NextCursor string
}

func historyCursor(r Request) string {
	return strconv.FormatInt(*r.ProcessedAt, 10) + "." + r.ID
}

func parseHistoryCursor(c string) (int64, string, error) {
	ts, id, ok := strings.Cut(c, ".")
	if !ok || id == "" {
		return 0, "", ErrInvalidCursor
	}
	n, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return 0, "", ErrInvalidCursor
	}
	return n, id, nil
}

func (s *Store) ListHistory(ctx context.Context, userID int64, cursor string, limit int) (HistoryPage, error) {
	beforeTS, beforeID := int64(math.MaxInt64), ""
	if cursor != "" {
		ts, id, err := parseHistoryCursor(cursor)
		if err != nil {
			return HistoryPage{}, err
		}
		beforeTS, beforeID = ts, id
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+requestColumns+` FROM approval_requests r JOIN approval_caller_keys k ON k.id = r.caller_key_id WHERE r.user_id = ? AND (r.status IN ('APPROVED', 'REJECTED') OR (r.status = 'PENDING' AND r.expires_at <= ?)) AND (COALESCE(r.processed_at, r.expires_at) < ? OR (COALESCE(r.processed_at, r.expires_at) = ? AND r.id < ?)) ORDER BY COALESCE(r.processed_at, r.expires_at) DESC, r.id DESC LIMIT ?`,
		userID, s.now().Unix(), beforeTS, beforeTS, beforeID, limit+1)
	if err != nil {
		return HistoryPage{}, fmt.Errorf("list approval history: %w", err)
	}
	defer func() { _ = rows.Close() }()
	page := HistoryPage{Items: []Request{}}
	for rows.Next() {
		r, err := s.scanRequest(rows)
		if err != nil {
			return HistoryPage{}, fmt.Errorf("scan approval request: %w", err)
		}
		page.Items = append(page.Items, r)
	}
	if err := rows.Err(); err != nil {
		return HistoryPage{}, err
	}
	if len(page.Items) > limit {
		page.Items = page.Items[:limit]
		page.NextCursor = historyCursor(page.Items[limit-1])
	}
	return page, nil
}
