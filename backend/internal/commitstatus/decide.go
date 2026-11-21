package commitstatus

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"
)

const (
	MaxContextRunes     = 255
	MaxDescriptionRunes = 140
	MaxTargetURLRunes   = 2000
	MaxRequesterRunes   = 200
	MaxReasonRunes      = 2000
	MaxSessionIDRunes   = 128
	maxWriteErrorRunes  = 500
)

var (
	ErrWriterUnavailable = errors.New("commit status writes are not configured")
	fullSHA              = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
	repoName             = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	states               = map[string]bool{"error": true, "failure": true, "pending": true, "success": true}
)

// WriteError is a GitHub failure while writing an approved request; the request stays pending with the message recorded.
type WriteError struct {
	Request Request
	Message string
}

func (e *WriteError) Error() string { return "write commit status: " + e.Message }

// Validate checks a request the way GitHub would, so a request that cannot be written is refused when it is filed rather than when it is approved.
func Validate(in NewRequest) error {
	switch {
	case !repoName.MatchString(in.Owner):
		return errors.New("owner must be a GitHub account name")
	case !repoName.MatchString(in.Repo):
		return errors.New("repo must be a repository name without its owner")
	case !fullSHA.MatchString(in.SHA):
		return errors.New("sha must be a 40-character hex commit sha")
	case !states[in.State]:
		return errors.New("state must be one of error, failure, pending, success")
	case strings.TrimSpace(in.Context) == "":
		return errors.New("context is required")
	case utf8.RuneCountInString(in.Context) > MaxContextRunes:
		return fmt.Errorf("context must be at most %d characters", MaxContextRunes)
	case utf8.RuneCountInString(in.Description) > MaxDescriptionRunes:
		return fmt.Errorf("description must be at most %d characters", MaxDescriptionRunes)
	case strings.TrimSpace(in.Requester) == "":
		return errors.New("requester is required")
	case utf8.RuneCountInString(in.Requester) > MaxRequesterRunes:
		return fmt.Errorf("requester must be at most %d characters", MaxRequesterRunes)
	case utf8.RuneCountInString(in.Reason) > MaxReasonRunes:
		return fmt.Errorf("reason must be at most %d characters", MaxReasonRunes)
	case utf8.RuneCountInString(in.SessionID) > MaxSessionIDRunes:
		return fmt.Errorf("sessionId must be at most %d characters", MaxSessionIDRunes)
	}
	if in.TargetURL != "" {
		u, err := url.Parse(in.TargetURL)
		if err != nil || !u.IsAbs() || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" ||
			utf8.RuneCountInString(in.TargetURL) > MaxTargetURLRunes {
			return errors.New("targetUrl must be an absolute http(s) URL")
		}
	}
	return nil
}

// Decider applies the owner's decision; one mutex makes sure a request is written to GitHub at most once even when the app and a notification action decide it together.
type Decider struct {
	store  *Store
	writer Writer
	mu     sync.Mutex
}

// NewDecider builds a Decider; a nil writer refuses approvals with ErrWriterUnavailable but still rejects.
func NewDecider(store *Store, writer Writer) *Decider {
	return &Decider{store: store, writer: writer}
}

// Approve writes the request's status to GitHub and marks it approved.
func (d *Decider) Approve(ctx context.Context, userID int64, id string) (Request, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	r, err := d.pending(ctx, userID, id)
	if err != nil {
		return r, err
	}
	if d.writer == nil {
		return r, ErrWriterUnavailable
	}
	statusID, werr := d.writer.WriteStatus(ctx, StatusWrite{
		Owner: r.Owner, Repo: r.Repo, SHA: r.SHA, State: r.State,
		Context: r.Context, Description: r.Description, TargetURL: r.TargetURL,
	})
	if werr != nil {
		msg := truncate(werr.Error(), maxWriteErrorRunes)
		if err := d.store.setLastError(context.WithoutCancel(ctx), userID, id, msg); err != nil {
			return r, err
		}
		latest, err := d.store.Get(context.WithoutCancel(ctx), userID, id)
		if err != nil {
			return r, err
		}
		return latest, &WriteError{Request: latest, Message: msg}
	}
	if err := d.store.markApproved(context.WithoutCancel(ctx), userID, id, statusID); err != nil {
		return r, err
	}
	return d.store.Get(context.WithoutCancel(ctx), userID, id)
}

// Reject marks the request rejected without touching GitHub.
func (d *Decider) Reject(ctx context.Context, userID int64, id string) (Request, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	r, err := d.pending(ctx, userID, id)
	if err != nil {
		return r, err
	}
	if err := d.store.markRejected(ctx, userID, id); err != nil {
		return r, err
	}
	return d.store.Get(ctx, userID, id)
}

func (d *Decider) pending(ctx context.Context, userID int64, id string) (Request, error) {
	r, err := d.store.Get(ctx, userID, id)
	if err != nil {
		return r, err
	}
	if r.Status != StatusPending {
		return r, ErrNotPending
	}
	return r, nil
}

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max]) + "…"
}
