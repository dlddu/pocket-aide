// Package fcm sends alert pushes to Android devices through the Firebase Cloud Messaging HTTP v1 API.
package fcm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"
)

const (
	DefaultEndpoint = "https://fcm.googleapis.com"
	Scope           = "https://www.googleapis.com/auth/firebase.messaging"

	defaultTokenURL = "https://oauth2.googleapis.com/token"
	fcmErrorType    = "type.googleapis.com/google.firebase.fcm.v1.FcmError"
)

// Android notification channel IDs; they must match the channels the Android app creates, or the push lands in the OS fallback channel.
const (
	ChannelApproval = "approval"
	ChannelDefault  = "default"
)

// Client posts messages for one Firebase project.
type Client struct {
	projectID string
	endpoint  string
	http      *http.Client
	channels  map[string]string
}

type serviceAccount struct {
	Type         string `json:"type"`
	ProjectID    string `json:"project_id"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
	TokenURI     string `json:"token_uri"`
}

// New parses a Firebase service account key JSON in memory and returns a client authorised for that key's project.
func New(serviceAccountJSON []byte) (*Client, error) {
	var sa serviceAccount
	if err := json.Unmarshal(serviceAccountJSON, &sa); err != nil {
		return nil, fmt.Errorf("fcm: parse service account json: %w", err)
	}
	if sa.Type != "" && sa.Type != "service_account" {
		return nil, fmt.Errorf("fcm: credentials type %q is not service_account", sa.Type)
	}
	if sa.ProjectID == "" || sa.ClientEmail == "" || sa.PrivateKey == "" {
		return nil, errors.New("fcm: service account json needs project_id, client_email and private_key")
	}
	cfg := &jwt.Config{
		Email:        sa.ClientEmail,
		PrivateKey:   []byte(sa.PrivateKey),
		PrivateKeyID: sa.PrivateKeyID,
		Scopes:       []string{Scope},
		TokenURL:     sa.TokenURI,
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = defaultTokenURL
	}
	return NewWithTokenSource(sa.ProjectID, cfg.TokenSource(context.Background()))
}

// NewWithTokenSource builds a client that authorises each request with ts.
func NewWithTokenSource(projectID string, ts oauth2.TokenSource) (*Client, error) {
	if projectID == "" {
		return nil, errors.New("fcm: project id is required")
	}
	if ts == nil {
		return nil, errors.New("fcm: token source is required")
	}
	hc := &http.Client{
		Timeout:   10 * time.Second,
		Transport: &oauth2.Transport{Source: oauth2.ReuseTokenSource(nil, ts)},
	}
	return &Client{projectID: projectID, endpoint: DefaultEndpoint, http: hc}, nil
}

// WithEndpoint points the client at another FCM base URL, such as a test server.
func (c *Client) WithEndpoint(baseURL string) *Client {
	c.endpoint = strings.TrimRight(baseURL, "/")
	return c
}

// WithChannels maps push categories to Android notification channels; unmapped categories use ChannelDefault.
func (c *Client) WithChannels(byCategory map[string]string) *Client {
	c.channels = byCategory
	return c
}

// ProjectID reports the Firebase project the client sends to.
func (c *Client) ProjectID() string { return c.projectID }

// SendWithData delivers an alert with custom data and no category.
func (c *Client) SendWithData(ctx context.Context, token, title, body string, data map[string]any) error {
	return c.SendWithCategory(ctx, token, "", title, body, data)
}

// SendWithCategory delivers an alert to one registration token; the category travels as the "category" data key.
func (c *Client) SendWithCategory(ctx context.Context, token, category, title, body string, data map[string]any) error {
	raw, err := json.Marshal(sendRequest{Message: c.message(token, category, title, body, data)})
	if err != nil {
		return fmt.Errorf("fcm: encode message: %w", err)
	}
	u := c.endpoint + "/v1/projects/" + url.PathEscape(c.projectID) + "/messages:send"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("fcm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("fcm push: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode/100 == 2 {
		return nil
	}
	return rejected(resp.StatusCode, respBody)
}

type sendRequest struct {
	Message message `json:"message"`
}

type message struct {
	Token        string            `json:"token"`
	Notification notification      `json:"notification"`
	Data         map[string]string `json:"data,omitempty"`
	Android      androidConfig     `json:"android"`
}

type notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type androidConfig struct {
	Priority     string              `json:"priority"`
	Notification androidNotification `json:"notification"`
}

type androidNotification struct {
	ChannelID string `json:"channel_id"`
}

func (c *Client) message(token, category, title, body string, data map[string]any) message {
	channel, ok := c.channels[category]
	if !ok {
		channel = ChannelDefault
	}
	return message{
		Token:        token,
		Notification: notification{Title: title, Body: body},
		Data:         stringData(category, data),
		Android: androidConfig{
			Priority:     "high",
			Notification: androidNotification{ChannelID: channel},
		},
	}
}

// FCM data values must be strings.
func stringData(category string, data map[string]any) map[string]string {
	if len(data) == 0 && category == "" {
		return nil
	}
	out := make(map[string]string, len(data)+1)
	for k, v := range data {
		if s, ok := v.(string); ok {
			out[k] = s
			continue
		}
		raw, err := json.Marshal(v)
		if err != nil {
			out[k] = fmt.Sprint(v)
			continue
		}
		out[k] = string(raw)
	}
	if category != "" {
		out["category"] = category
	}
	return out
}

// RejectedError is a non-2xx answer from the FCM send endpoint.
type RejectedError struct {
	StatusCode int
	Status     string
	ErrorCode  string
	Message    string
}

func (e *RejectedError) Error() string {
	return fmt.Sprintf("fcm rejected (status=%d %s errorCode=%q): %s", e.StatusCode, e.Status, e.ErrorCode, e.Message)
}

// TokenInvalid reports whether FCM will never accept this registration token again.
func (e *RejectedError) TokenInvalid() bool {
	switch e.ErrorCode {
	case "UNREGISTERED", "SENDER_ID_MISMATCH":
		return true
	}
	return e.Status == "NOT_FOUND" || e.StatusCode == http.StatusNotFound
}

type errorBody struct {
	Error struct {
		Message string `json:"message"`
		Status  string `json:"status"`
		Details []struct {
			Type      string `json:"@type"`
			ErrorCode string `json:"errorCode"`
		} `json:"details"`
	} `json:"error"`
}

func rejected(statusCode int, body []byte) *RejectedError {
	e := &RejectedError{StatusCode: statusCode}
	var eb errorBody
	if json.Unmarshal(body, &eb) != nil {
		return e
	}
	e.Status = eb.Error.Status
	e.Message = eb.Error.Message
	for _, d := range eb.Error.Details {
		if d.Type == fcmErrorType && d.ErrorCode != "" {
			e.ErrorCode = d.ErrorCode
			break
		}
	}
	return e
}
