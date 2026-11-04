package mail

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ErrNotFound is returned when a query targets an account that does not exist
// or is not owned by the caller.
var ErrNotFound = errors.New("mail account not found")

// ErrAlreadyConnected is returned by Add when the user already connected the
// same address.
var ErrAlreadyConnected = errors.New("mail account already connected")

// Account is one connected IMAP account. The app password never leaves the
// store except through Credentials.
type Account struct {
	ID             int64  `json:"id"`
	Email          string `json:"email"`
	IMAPHost       string `json:"imap_host"`
	IMAPPort       int    `json:"imap_port"`
	NeedsReconnect bool   `json:"needs_reconnect"`
	CreatedAt      int64  `json:"created_at"`
}

// Credentials is what an IMAP connection needs.
type Credentials struct {
	Email    string
	Password string
	Host     string
	Port     int
}

// Store is the user-scoped facade over mail_accounts.
type Store struct {
	db     *sql.DB
	cipher *Cipher
}

// NewStore wires a Store onto an open *sql.DB.
func NewStore(db *sql.DB, c *Cipher) *Store { return &Store{db: db, cipher: c} }

const accountColumns = `id, email, imap_host, imap_port, needs_reconnect, created_at`

func scanAccount(sc interface{ Scan(...any) error }) (Account, error) {
	var a Account
	var reconnect int
	if err := sc.Scan(&a.ID, &a.Email, &a.IMAPHost, &a.IMAPPort, &reconnect, &a.CreatedAt); err != nil {
		return Account{}, err
	}
	a.NeedsReconnect = reconnect != 0
	return a, nil
}

// List returns the user's accounts, oldest first.
func (s *Store) List(ctx context.Context, userID int64) ([]Account, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT `+accountColumns+`
		FROM mail_accounts
		WHERE user_id = ?
		ORDER BY created_at ASC, id ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list mail accounts: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := make([]Account, 0)
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, fmt.Errorf("scan mail account: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

// Add stores a new account with its app password sealed.
func (s *Store) Add(ctx context.Context, userID int64, cred Credentials) (Account, error) {
	sealed, err := s.cipher.Encrypt(cred.Password)
	if err != nil {
		return Account{}, err
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO mail_accounts (user_id, email, imap_host, imap_port, password_enc)
		VALUES (?, ?, ?, ?, ?)
	`, userID, cred.Email, cred.Host, cred.Port, sealed)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return Account{}, ErrAlreadyConnected
		}
		return Account{}, fmt.Errorf("insert mail account: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Account{}, fmt.Errorf("last insert id: %w", err)
	}
	return scanAccount(s.db.QueryRowContext(ctx, `
		SELECT `+accountColumns+` FROM mail_accounts WHERE id = ? AND user_id = ?
	`, id, userID))
}

// Account returns one account without touching its sealed password.
func (s *Store) Account(ctx context.Context, userID, id int64) (Account, error) {
	a, err := scanAccount(s.db.QueryRowContext(ctx, `
		SELECT `+accountColumns+` FROM mail_accounts WHERE id = ? AND user_id = ?
	`, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("load mail account: %w", err)
	}
	return a, nil
}

// Credentials opens the stored app password for one account.
func (s *Store) Credentials(ctx context.Context, userID, id int64) (Credentials, error) {
	var c Credentials
	var sealed string
	err := s.db.QueryRowContext(ctx, `
		SELECT email, imap_host, imap_port, password_enc
		FROM mail_accounts WHERE id = ? AND user_id = ?
	`, id, userID).Scan(&c.Email, &c.Host, &c.Port, &sealed)
	if errors.Is(err, sql.ErrNoRows) {
		return Credentials{}, ErrNotFound
	}
	if err != nil {
		return Credentials{}, fmt.Errorf("load mail credentials: %w", err)
	}
	c.Password, err = s.cipher.Decrypt(sealed)
	if err != nil {
		return Credentials{}, err
	}
	return c, nil
}

// ReplacePassword re-seals a new app password (reconnect) and clears the
// reconnect flag.
func (s *Store) ReplacePassword(ctx context.Context, userID, id int64, password string) error {
	sealed, err := s.cipher.Encrypt(password)
	if err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE mail_accounts SET password_enc = ?, needs_reconnect = 0
		WHERE id = ? AND user_id = ?
	`, sealed, id, userID)
	return affectedOrNotFound(res, err, "replace mail password")
}

// SetNeedsReconnect flags (or unflags) an account whose IMAP login failed. The
// stored password is kept: a revoked app password is repaired by reconnecting,
// not by deleting.
func (s *Store) SetNeedsReconnect(ctx context.Context, userID, id int64, needs bool) error {
	v := 0
	if needs {
		v = 1
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE mail_accounts SET needs_reconnect = ? WHERE id = ? AND user_id = ?
	`, v, id, userID)
	return affectedOrNotFound(res, err, "flag mail account")
}

// Delete disconnects an account; the sealed password goes with the row.
func (s *Store) Delete(ctx context.Context, userID, id int64) error {
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM mail_accounts WHERE id = ? AND user_id = ?
	`, id, userID)
	return affectedOrNotFound(res, err, "delete mail account")
}

func affectedOrNotFound(res sql.Result, err error, what string) error {
	if err != nil {
		return fmt.Errorf("%s: %w", what, err)
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
