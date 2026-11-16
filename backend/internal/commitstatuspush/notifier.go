// Package commitstatuspush alerts the owner's devices to a new commit status request, with the same decide-from-the-notification shape as approval pushes (PRD-12 AC21).
package commitstatuspush

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/apns"
	"github.com/dlddu/pocket-aide/backend/internal/devicetokens"
	"github.com/dlddu/pocket-aide/backend/internal/push"
)

const Category = "COMMIT_STATUS_REQUEST"

const Kind = "commit_status_request"

type Sender interface {
	SendWithCategory(ctx context.Context, d devicetokens.Device, category, title, body string, data map[string]any) error
}

type Tokens interface {
	ListByUserID(ctx context.Context, userID int64) ([]devicetokens.Device, error)
	Delete(ctx context.Context, userID int64, token string) error
}

// Request is what one push says about a new request.
type Request struct {
	UserID    int64
	KeyName   string
	RequestID string
	Owner     string
	Repo      string
	SHA       string
	State     string
	Context   string
	Requester string
}

type Notifier struct {
	sender  Sender
	tokens  Tokens
	timeout time.Duration
}

func New(sender Sender, tokens Tokens) *Notifier {
	return &Notifier{sender: sender, tokens: tokens, timeout: 30 * time.Second}
}

func Title(keyName string) string {
	return "커밋 상태 요청 · " + keyName
}

func Body(req Request) string {
	sha := req.SHA
	if len(sha) > 7 {
		sha = sha[:7]
	}
	return req.Context + " → " + req.State + "\n" + req.Owner + "/" + req.Repo + "@" + sha + "\n" + req.Requester
}

func Data(requestID string) map[string]any {
	return map[string]any{"kind": Kind, "commit_status_request_id": requestID}
}

func (n *Notifier) Notify(req Request) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), n.timeout)
		defer cancel()
		n.Send(ctx, req)
	}()
}

func (n *Notifier) Send(ctx context.Context, req Request) {
	devices, err := n.tokens.ListByUserID(ctx, req.UserID)
	if err != nil {
		log.Printf("commit status push: list tokens for user=%d: %v", req.UserID, err)
		return
	}
	title := Title(req.KeyName)
	data := Data(req.RequestID)
	body := apns.FitAlertBody(Category, title, Body(req), data)
	for _, d := range devices {
		err := n.sender.SendWithCategory(ctx, d, Category, title, body, data)
		if err == nil || errors.Is(err, push.ErrPlatformDisabled) {
			continue
		}
		log.Printf("commit status push user=%d request=%s platform=%s: %v", req.UserID, req.RequestID, d.Platform, err)
		if push.IsTokenInvalid(err) {
			if err := n.tokens.Delete(ctx, req.UserID, d.Token); err != nil {
				log.Printf("commit status push: drop invalid token for user=%d: %v", req.UserID, err)
			}
		}
	}
}
