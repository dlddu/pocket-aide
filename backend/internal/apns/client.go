// Package apns is a thin wrapper around github.com/sideshow/apns2 for sending
// alert pushes to iOS devices. The .p8 auth key is parsed from a PEM byte
// slice (env-injected secret) so nothing touches disk — the deployment runs
// with readOnlyRootFilesystem.
package apns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

// Client wraps an apns2.Client plus the bundle ID we send as the APNs Topic.
type Client struct {
	cli      *apns2.Client
	bundleID string
}

// New parses the .p8 PEM in memory and returns a configured client.
// useProduction must agree with the iOS app's `aps-environment` entitlement —
// mismatched environments fail silently with no notification delivered.
func New(keyID, teamID, bundleID string, p8PEM []byte, useProduction bool) (*Client, error) {
	if keyID == "" || teamID == "" || bundleID == "" {
		return nil, errors.New("apns: keyID, teamID, bundleID are required")
	}
	if len(p8PEM) == 0 {
		return nil, errors.New("apns: p8 PEM is empty")
	}
	authKey, err := token.AuthKeyFromBytes(p8PEM)
	if err != nil {
		return nil, fmt.Errorf("apns: parse p8: %w", err)
	}
	tok := &token.Token{
		AuthKey: authKey,
		KeyID:   keyID,
		TeamID:  teamID,
	}
	cli := apns2.NewTokenClient(tok)
	if useProduction {
		cli = cli.Production()
	} else {
		cli = cli.Development()
	}
	return &Client{cli: cli, bundleID: bundleID}, nil
}

func (c *Client) WithHost(host string) *Client {
	c.cli.Host = strings.TrimRight(host, "/")
	if strings.HasPrefix(host, "http://") {
		c.cli.HTTPClient = &http.Client{Timeout: apns2.HTTPClientTimeout}
	}
	return c
}

// Send delivers a simple alert (title + body) to a single device token.
func (c *Client) Send(ctx context.Context, deviceToken, title, body string) error {
	return c.SendWithData(ctx, deviceToken, title, body, nil)
}

// SendWithData behaves like Send but also attaches custom top-level keys to
// the APNs payload. The iOS client reads these from
// `UNNotification.request.content.userInfo` to drive deep-link routing
// (PRD-10 AC7 — we send `event_id` so the receiver can highlight the right
// notification-history row).
func (c *Client) SendWithData(ctx context.Context, deviceToken, title, body string, data map[string]any) error {
	return c.SendWithCategory(ctx, deviceToken, "", title, body, data)
}

func (c *Client) SendWithCategory(ctx context.Context, deviceToken, category, title, body string, data map[string]any) error {
	n := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       c.bundleID,
		Payload:     alertPayload(category, title, body, data),
	}
	resp, err := c.cli.PushWithContext(ctx, n)
	if err != nil {
		return fmt.Errorf("apns push: %w", err)
	}
	if !resp.Sent() {
		return &RejectedError{StatusCode: resp.StatusCode, Reason: resp.Reason, ApnsID: resp.ApnsID}
	}
	return nil
}

const MaxPayloadBytes = 4096

type RejectedError struct {
	StatusCode int
	Reason     string
	ApnsID     string
}

func (e *RejectedError) Error() string {
	return fmt.Sprintf("apns rejected (status=%d reason=%q apns-id=%s)", e.StatusCode, e.Reason, e.ApnsID)
}

func (e *RejectedError) TokenInvalid() bool {
	switch e.Reason {
	case apns2.ReasonBadDeviceToken, apns2.ReasonUnregistered, apns2.ReasonDeviceTokenNotForTopic:
		return true
	}
	return e.StatusCode == http.StatusGone
}

func alertPayload(category, title, body string, data map[string]any) *payload.Payload {
	p := payload.NewPayload().AlertTitle(title).AlertBody(body).Sound("default")
	if category != "" {
		p = p.Category(category)
	}
	for k, v := range data {
		p = p.Custom(k, v)
	}
	return p
}

func PayloadSize(category, title, body string, data map[string]any) int {
	raw, err := json.Marshal(alertPayload(category, title, body, data))
	if err != nil {
		return 0
	}
	return len(raw)
}

func FitAlertBody(category, title, body string, data map[string]any) string {
	if PayloadSize(category, title, body, data) <= MaxPayloadBytes {
		return body
	}
	runes := []rune(body)
	lo, hi := 0, len(runes)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if PayloadSize(category, title, string(runes[:mid])+"…", data) <= MaxPayloadBytes {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return string(runes[:lo]) + "…"
}
