// Package push routes an alert to the push service that matches each device's platform.
package push

import (
	"context"
	"errors"
	"fmt"

	"github.com/dlddu/pocket-aide/backend/internal/apns"
	"github.com/dlddu/pocket-aide/backend/internal/devicetokens"
	"github.com/dlddu/pocket-aide/backend/internal/fcm"
)

// ErrPlatformDisabled means the device's platform has no configured push client, so nothing was sent.
var ErrPlatformDisabled = errors.New("push: platform has no configured client")

// TokenInvalid is satisfied by push-service rejections that tell whether the device token is dead.
type TokenInvalid interface {
	TokenInvalid() bool
}

var (
	_ TokenInvalid = (*apns.RejectedError)(nil)
	_ TokenInvalid = (*fcm.RejectedError)(nil)
)

// IsTokenInvalid reports whether err says the push service will never accept that device token again.
func IsTokenInvalid(err error) bool {
	var ti TokenInvalid
	return errors.As(err, &ti) && ti.TokenInvalid()
}

type sender interface {
	SendWithCategory(ctx context.Context, token, category, title, body string, data map[string]any) error
}

// Router holds one optional client per platform.
type Router struct {
	apns sender
	fcm  sender
}

// New builds a Router; a nil client disables its platform.
func New(apnsClient *apns.Client, fcmClient *fcm.Client) *Router {
	r := &Router{}
	if apnsClient != nil {
		r.apns = apnsClient
	}
	if fcmClient != nil {
		r.fcm = fcmClient
	}
	return r
}

// SendWithData delivers an alert with custom data and no category.
func (r *Router) SendWithData(ctx context.Context, d devicetokens.Device, title, body string, data map[string]any) error {
	return r.SendWithCategory(ctx, d, "", title, body, data)
}

// SendWithCategory delivers an alert to d through its platform's client.
func (r *Router) SendWithCategory(ctx context.Context, d devicetokens.Device, category, title, body string, data map[string]any) error {
	var s sender
	switch d.Platform {
	case devicetokens.PlatformIOS:
		s = r.apns
	case devicetokens.PlatformAndroid:
		s = r.fcm
	default:
		return fmt.Errorf("push: unknown platform %q", d.Platform)
	}
	if s == nil {
		return ErrPlatformDisabled
	}
	return s.SendWithCategory(ctx, d.Token, category, title, body, data)
}
