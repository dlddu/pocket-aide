// Package chat is PRD-1 (AI 채팅 탭): each user's one conversation, its
// messages, and the codex session on session-platform that answers it.
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

	platformSessionID string
	transcriptOffset  int64
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

const sessionColumns = `id, title, created_at, updated_at, platform_session_id, transcript_offset`

func scanSession(row interface{ Scan(...any) error }) (Session, error) {
	var out Session
	err := row.Scan(&out.ID, &out.Title, &out.CreatedAt, &out.UpdatedAt, &out.platformSessionID, &out.transcriptOffset)
	return out, err
}

// Conversation returns the user's conversation — the most recently active
// one, from before the chat tab became a single conversation — creating it
// on first use.
func (s *Store) Conversation(ctx context.Context, userID int64) (Session, error) {
	out, err := scanSession(s.db.QueryRowContext(ctx,
		`SELECT `+sessionColumns+` FROM chat_sessions WHERE user_id = ? ORDER BY updated_at DESC, id DESC LIMIT 1`, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return s.CreateSession(ctx, userID)
	}
	if err != nil {
		return Session{}, fmt.Errorf("get chat conversation: %w", err)
	}
	return out, nil
}

// SetPlatform records the codex session answering a conversation and how far
// its transcript has been read.
func (s *Store) SetPlatform(ctx context.Context, userID, id int64, platformSessionID string, offset int64) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE chat_sessions SET platform_session_id = ?, transcript_offset = ? WHERE id = ? AND user_id = ?`,
		platformSessionID, offset, id, userID)
	if err != nil {
		return fmt.Errorf("update chat platform session: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateSession starts an empty conversation.
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
	out, err := scanSession(s.db.QueryRowContext(ctx,
		`SELECT `+sessionColumns+` FROM chat_sessions WHERE id = ? AND user_id = ?`, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("get chat session: %w", err)
	}
	return out, nil
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

// QueuedMessage is a message held for after the turn in progress — the
// store-backed twin of Engine's in-memory Pending, durable so a restart
// between the busy check and the write does not lose it.
type QueuedMessage struct {
	Content  string `json:"content"`
	QueuedAt int64  `json:"queued_at"`
}

// GetQueued returns the message queued behind the session's turn in
// progress, if any.
func (s *Store) GetQueued(ctx context.Context, userID, sessionID int64) (*QueuedMessage, error) {
	var content sql.NullString
	var queuedAt sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT queued_content, queued_at FROM chat_sessions WHERE id = ? AND user_id = ?`, sessionID, userID,
	).Scan(&content, &queuedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get queued message: %w", err)
	}
	if !content.Valid || content.String == "" {
		return nil, nil
	}
	return &QueuedMessage{Content: content.String, QueuedAt: queuedAt.Int64}, nil
}

// SetQueued holds content for after the session's turn in progress ends.
func (s *Store) SetQueued(ctx context.Context, userID, sessionID int64, content string, queuedAt int64) (QueuedMessage, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE chat_sessions SET queued_content = ?, queued_at = ? WHERE id = ? AND user_id = ?`,
		content, queuedAt, sessionID, userID)
	if err != nil {
		return QueuedMessage{}, fmt.Errorf("set queued message: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return QueuedMessage{}, ErrNotFound
	}
	return QueuedMessage{Content: content, QueuedAt: queuedAt}, nil
}

// ClearQueued cancels a queued message, reporting whether one was present.
func (s *Store) ClearQueued(ctx context.Context, userID, sessionID int64) (bool, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE chat_sessions SET queued_content = NULL, queued_at = NULL
		 WHERE id = ? AND user_id = ? AND queued_content IS NOT NULL`, sessionID, userID)
	if err != nil {
		return false, fmt.Errorf("clear queued message: %w", err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// TakeQueued atomically returns and clears the queued message, if any, so a
// concurrent cancel and drain can never both act on the same message.
func (s *Store) TakeQueued(ctx context.Context, userID, sessionID int64) (*QueuedMessage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var content sql.NullString
	var queuedAt sql.NullInt64
	err = tx.QueryRowContext(ctx,
		`SELECT queued_content, queued_at FROM chat_sessions WHERE id = ? AND user_id = ?`, sessionID, userID,
	).Scan(&content, &queuedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get queued message: %w", err)
	}
	if !content.Valid || content.String == "" {
		return nil, nil
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE chat_sessions SET queued_content = NULL, queued_at = NULL WHERE id = ? AND user_id = ?`,
		sessionID, userID); err != nil {
		return nil, fmt.Errorf("clear queued message: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &QueuedMessage{Content: content.String, QueuedAt: queuedAt.Int64}, nil
}

func deriveTitle(content string) string {
	t := strings.Join(strings.Fields(content), " ")
	r := []rune(t)
	if len(r) > titleRunes {
		return string(r[:titleRunes]) + "…"
	}
	return t
}
