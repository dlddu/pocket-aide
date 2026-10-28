// Package externalnotify sends and records plain alert pushes from an external caller to every device of the caller's owner.
package externalnotify

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/dlddu/pocket-aide/backend/internal/apns"
)

// Data is the payload the app reads to open that notification's detail on tap; without the
// kind marker the app treats a push with no deep-link id as a CI push.
func Data(notificationID int64) map[string]any {
	return map[string]any{"kind": "external_notification", "notification_id": notificationID}
}

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
func (n *Notifier) Send(ctx context.Context, userID, notificationID int64, title, body string) (Result, error) {
	tokens, err := n.tokens.ListByUserID(ctx, userID)
	if err != nil {
		return Result{}, fmt.Errorf("list device tokens: %w", err)
	}
	res := Result{Devices: len(tokens)}
	data := Data(notificationID)
	body = apns.FitAlertBody("", title, body, data)
	for _, t := range tokens {
		err := n.sender.SendWithData(ctx, t, title, body, data)
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
