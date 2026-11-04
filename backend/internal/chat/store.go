// Package chat is the storage layer for PRD-1 (AI 채팅 탭) — conversation
// sessions and their messages, so a conversation survives re-entry (AC1, AC4).
package chat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Role is who wrote a message. The DB CHECK constraint mirrors these values.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// titleRunes caps the auto-generated session title (first user message).
const titleRunes = 40

// ErrNotFound is returned when a session does not exist or is not owned by
// the caller.
var ErrNotFound = errors.New("chat session not found")

// Session is one row of chat_sessions.
type Session struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// Message is one row of chat_messages.
type Message struct {
	ID        int64  `json:"id"`
	Role      Role   `json:"role"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

// Store is the user-scoped facade.
type Store struct {
	db *sql.DB
}

// New wires a Store onto an open *sql.DB.
func New(db *sql.DB) *Store { return &Store{db: db} }

const sessionColumns = `id, title, created_at, updated_at`

// CreateSession starts an empty conversation (PRD-1 AC4 「새 대화 시작」).
func (s *Store) CreateSession(ctx context.Context, userID int64) (Session, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO chat_sessions (user_id) VALUES (?)`, userID)
	if err != nil {
		return Session{}, fmt.Errorf("insert chat session: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Session{}, fmt.Errorf("last insert id: %w", err)
	}
	return s.GetSession(ctx, userID, id)
}

// GetSession returns one session owned by userID.
func (s *Store) GetSession(ctx context.Context, userID, id int64) (Session, error) {
	var out Session
	err := s.db.QueryRowContext(ctx,
		`SELECT `+sessionColumns+` FROM chat_sessions WHERE id = ? AND user_id = ?`, id, userID,
	).Scan(&out.ID, &out.Title, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("get chat session: %w", err)
	}
	return out, nil
}

// ListSessions returns the caller's sessions, most recently active first
// (PRD-1 AC4 「과거 대화 목록 조회」).
func (s *Store) ListSessions(ctx context.Context, userID int64) ([]Session, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+sessionColumns+` FROM chat_sessions WHERE user_id = ? ORDER BY updated_at DESC, id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list chat sessions: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := []Session{}
	for rows.Next() {
		var it Session
		if err := rows.Scan(&it.ID, &it.Title, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan chat session: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// Messages returns a session's messages oldest first. A limit > 0 keeps only
// the most recent limit messages (still oldest first).
func (s *Store) Messages(ctx context.Context, userID, sessionID int64, limit int) ([]Message, error) {
	if _, err := s.GetSession(ctx, userID, sessionID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, role, content, created_at FROM chat_messages WHERE session_id = ? ORDER BY id`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list chat messages: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := []Message{}
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chat message: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, nil
}

// AppendMessage stores a message in a session owned by userID and bumps the
// session's activity time. The first user message also becomes the title.
func (s *Store) AppendMessage(ctx context.Context, userID, sessionID int64, role Role, content string) (Message, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Message{}, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var title string
	err = tx.QueryRowContext(ctx,
		`SELECT title FROM chat_sessions WHERE id = ? AND user_id = ?`, sessionID, userID).Scan(&title)
	if errors.Is(err, sql.ErrNoRows) {
		return Message{}, ErrNotFound
	}
	if err != nil {
		return Message{}, fmt.Errorf("lookup chat session: %w", err)
	}
	res, err := tx.ExecContext(ctx,
		`INSERT INTO chat_messages (session_id, role, content) VALUES (?, ?, ?)`, sessionID, string(role), content)
	if err != nil {
		return Message{}, fmt.Errorf("insert chat message: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Message{}, fmt.Errorf("last insert id: %w", err)
	}
	if title == "" && role == RoleUser {
		title = deriveTitle(content)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE chat_sessions SET title = ?, updated_at = unixepoch() WHERE id = ?`, title, sessionID); err != nil {
		return Message{}, fmt.Errorf("touch chat session: %w", err)
	}
	var m Message
	if err := tx.QueryRowContext(ctx,
		`SELECT id, role, content, created_at FROM chat_messages WHERE id = ?`, id,
	).Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
		return Message{}, fmt.Errorf("reload chat message: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Message{}, fmt.Errorf("commit: %w", err)
	}
	return m, nil
}

func deriveTitle(content string) string {
	t := strings.Join(strings.Fields(content), " ")
	r := []rune(t)
	if len(r) > titleRunes {
		return string(r[:titleRunes]) + "…"
	}
	return t
}
