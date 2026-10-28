package approvalpush

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/apns"
)

type Sender interface {
	SendWithCategory(ctx context.Context, deviceToken, category, title, body string, data map[string]any) error
}

const Category = "APPROVAL_REQUEST"

type Tokens interface {
	ListByUserID(ctx context.Context, userID int64) ([]string, error)
	Delete(ctx context.Context, userID int64, token string) error
}

type Request struct {
	UserID     int64
	KeyName    string
	ApprovalID string
	Context    string
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
	return "승인 요청 · " + keyName
}

func Data(approvalID string) map[string]any {
	return map[string]any{"approval_id": approvalID}
}

func (n *Notifier) Notify(req Request) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), n.timeout)
		defer cancel()
		n.Send(ctx, req)
	}()
}

func (n *Notifier) Send(ctx context.Context, req Request) {
	tokens, err := n.tokens.ListByUserID(ctx, req.UserID)
	if err != nil {
		log.Printf("approval push: list tokens for user=%d: %v", req.UserID, err)
		return
	}
	title := Title(req.KeyName)
	data := Data(req.ApprovalID)
	body := apns.FitAlertBody(Category, title, req.Context, data)
	for _, t := range tokens {
		err := n.sender.SendWithCategory(ctx, t, Category, title, body, data)
		if err == nil {
			continue
		}
		log.Printf("approval push user=%d approval=%s: %v", req.UserID, req.ApprovalID, err)
		var rejected *apns.RejectedError
		if errors.As(err, &rejected) && rejected.TokenInvalid() {
			if err := n.tokens.Delete(ctx, req.UserID, t); err != nil {
				log.Printf("approval push: drop invalid token for user=%d: %v", req.UserID, err)
			}
		}
	}
}
