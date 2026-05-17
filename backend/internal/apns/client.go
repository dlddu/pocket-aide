// Package apns is a thin wrapper around github.com/sideshow/apns2 for sending
// alert pushes to iOS devices. The .p8 auth key is parsed from a PEM byte
// slice (env-injected secret) so nothing touches disk — the deployment runs
// with readOnlyRootFilesystem.
package apns

import (
	"context"
	"errors"
	"fmt"

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
// useProduction selects api.push.apple.com (true) vs api.sandbox.push.apple.com
// (false). It must agree with the iOS app's `aps-environment` entitlement —
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

// Send delivers a simple alert (title + body) to a single device token. APNs
// non-2xx responses become Go errors with the reason string included so the
// caller can log them per-token.
func (c *Client) Send(ctx context.Context, deviceToken, title, body string) error {
	n := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       c.bundleID,
		Payload:     payload.NewPayload().AlertTitle(title).AlertBody(body).Sound("default"),
	}
	resp, err := c.cli.PushWithContext(ctx, n)
	if err != nil {
		return fmt.Errorf("apns push: %w", err)
	}
	if !resp.Sent() {
		return fmt.Errorf("apns rejected (status=%d reason=%q apns-id=%s)", resp.StatusCode, resp.Reason, resp.ApnsID)
	}
	return nil
}
