// Package externalnotify sends a plain alert push from an external caller to every device of the caller's owner.
package externalnotify

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/dlddu/pocket-aide/backend/internal/apns"
)

type Sender interface {
	SendWithData(ctx context.Context, deviceToken, title, body string, data map[string]any) error
}

type Tokens interface {
	ListByUserID(ctx context.Context, userID int64) ([]string, error)
	Delete(ctx context.Context, userID int64, token string) error
}

// Result counts the owner's registered devices and how many APNs accepted.
type Result struct {
	Devices int `json:"devices"`
	Sent    int `json:"sent"`
}

type Notifier struct {
	sender Sender
	tokens Tokens
}

func New(sender Sender, tokens Tokens) *Notifier {
	return &Notifier{sender: sender, tokens: tokens}
}

// Send delivers synchronously and returns an error only when the devices cannot be listed.
func (n *Notifier) Send(ctx context.Context, userID int64, title, body string) (Result, error) {
	tokens, err := n.tokens.ListByUserID(ctx, userID)
	if err != nil {
		return Result{}, fmt.Errorf("list device tokens: %w", err)
	}
	res := Result{Devices: len(tokens)}
	body = apns.FitAlertBody(title, body, nil)
	for _, t := range tokens {
		err := n.sender.SendWithData(ctx, t, title, body, nil)
		if err == nil {
			res.Sent++
			continue
		}
		log.Printf("external notification user=%d: %v", userID, err)
		var rejected *apns.RejectedError
		if errors.As(err, &rejected) && rejected.TokenInvalid() {
			if err := n.tokens.Delete(ctx, userID, t); err != nil {
				log.Printf("external notification: drop invalid token for user=%d: %v", userID, err)
			}
		}
	}
	return res, nil
}
