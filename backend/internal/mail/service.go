package mail

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
)

// MaxMessages is the list cap (PRD-8 AC11: 최근 50통).
const MaxMessages = 50

const defaultPort = 993

// ErrInvalidAccount is returned for a malformed address or password.
var ErrInvalidAccount = errors.New("email and app password are required")

// ErrHostRequired is returned when the address's domain has no known IMAP
// server and none was supplied.
var ErrHostRequired = errors.New("imap_host is required for this mail service")

// presets auto-fill the IMAP server for the major services (AC11).
var presets = map[string]string{
	"gmail.com":      "imap.gmail.com",
	"googlemail.com": "imap.gmail.com",
	"outlook.com":    "outlook.office365.com",
	"hotmail.com":    "outlook.office365.com",
	"live.com":       "outlook.office365.com",
	"icloud.com":     "imap.mail.me.com",
	"me.com":         "imap.mail.me.com",
	"yahoo.com":      "imap.mail.yahoo.com",
	"naver.com":      "imap.naver.com",
	"daum.net":       "imap.daum.net",
	"hanmail.net":    "imap.daum.net",
	"kakao.com":      "imap.kakao.com",
}

// Account health in the message list response.
const (
	StatusOK       = "ok"
	StatusReauth   = "reauth_required"
	StatusUnstable = "error"
)

// AccountStatus reports how one account fared in a list fetch.
type AccountStatus struct {
	AccountID int64  `json:"account_id"`
	Status    string `json:"status"`
}

// ListedMessage is a Message tagged with its account (AC11: 계정 라벨).
type ListedMessage struct {
	AccountID    int64  `json:"account_id"`
	AccountEmail string `json:"account_email"`
	Message
}

// Service ties the account store to IMAP access.
type Service struct {
	Store *Store
	IMAP  IMAP
}

// Connect tests the login and, only if it succeeds, stores the account.
func (s *Service) Connect(ctx context.Context, userID int64, email, password, host string) (Account, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 || password == "" {
		return Account{}, ErrInvalidAccount
	}
	host = strings.TrimSpace(host)
	if host == "" {
		host = presets[email[at+1:]]
	}
	if host == "" {
		return Account{}, ErrHostRequired
	}
	cred := Credentials{Email: email, Password: password, Host: host, Port: defaultPort}
	if err := s.IMAP.Probe(ctx, cred); err != nil {
		return Account{}, err
	}
	return s.Store.Add(ctx, userID, cred)
}

// Reconnect replaces the app password of an existing account after a
// successful login test and clears its "needs reconnect" flag.
func (s *Service) Reconnect(ctx context.Context, userID, id int64, password string) error {
	if password == "" {
		return ErrInvalidAccount
	}
	acct, err := s.Store.Account(ctx, userID, id)
	if err != nil {
		return err
	}
	// The stored password is not read: a lost key must not block a reconnect,
	// and the new password replaces it anyway.
	cred := Credentials{Email: acct.Email, Host: acct.IMAPHost, Port: acct.IMAPPort}
	cred.Password = password
	if err := s.IMAP.Probe(ctx, cred); err != nil {
		return err
	}
	return s.Store.ReplacePassword(ctx, userID, id, password)
}

// Messages merges the unseen mail of the user's accounts (or just accountID
// when non-zero), newest first, capped at MaxMessages. A failing account does
// not fail the list: it is reported in the statuses, and a rejected login also
// flags the account "needs reconnect".
func (s *Service) Messages(ctx context.Context, userID, accountID int64) ([]ListedMessage, []AccountStatus, error) {
	accounts, err := s.Store.List(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if accountID != 0 {
		filtered := accounts[:0:0]
		for _, a := range accounts {
			if a.ID == accountID {
				filtered = append(filtered, a)
			}
		}
		if len(filtered) == 0 {
			return nil, nil, ErrNotFound
		}
		accounts = filtered
	}

	type result struct {
		status AccountStatus
		msgs   []ListedMessage
	}
	results := make([]result, len(accounts))
	var wg sync.WaitGroup
	for i, a := range accounts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = result{status: AccountStatus{AccountID: a.ID, Status: StatusOK}}
			cred, err := s.Store.Credentials(ctx, userID, a.ID)
			var msgs []Message
			if err == nil {
				msgs, err = s.IMAP.Unseen(ctx, cred, MaxMessages)
			}
			switch {
			case err == nil:
				if a.NeedsReconnect {
					_ = s.Store.SetNeedsReconnect(ctx, userID, a.ID, false)
				}
				for _, m := range msgs {
					results[i].msgs = append(results[i].msgs, ListedMessage{AccountID: a.ID, AccountEmail: a.Email, Message: m})
				}
			case errors.Is(err, ErrAuth):
				_ = s.Store.SetNeedsReconnect(ctx, userID, a.ID, true)
				results[i].status.Status = StatusReauth
			default:
				results[i].status.Status = StatusUnstable
			}
		}()
	}
	wg.Wait()

	all := make([]ListedMessage, 0)
	statuses := make([]AccountStatus, 0, len(results))
	for _, r := range results {
		statuses = append(statuses, r.status)
		all = append(all, r.msgs...)
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].ReceivedAt.After(all[j].ReceivedAt) })
	if len(all) > MaxMessages {
		all = all[:MaxMessages]
	}
	return all, statuses, nil
}
