package sessions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultPlatformURL = "http://control-plane.session-platform.svc.cluster.local"
	apiPrefix          = "/api/v1"
	maxResponseBytes   = 17 << 20
)

type Client struct {
	baseURL string
	http    *http.Client
	tokens  TokenSource
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/") + apiPrefix,
		http:    &http.Client{Timeout: 120 * time.Second},
	}
}

// WithTokens makes every call carry a bearer token from tokens, so the
// platform answers as that token's tenant. Without it the call has no
// Authorization header and the platform treats it as the operator.
func (c *Client) WithTokens(tokens TokenSource) *Client {
	c.tokens = tokens
	return c
}

func (c *Client) Do(ctx context.Context, method, path string, body []byte) (int, []byte, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, nil, fmt.Errorf("new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.tokens != nil {
		token, err := c.tokens.Token(ctx)
		if err != nil {
			return 0, nil, fmt.Errorf("platform token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("do: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return 0, nil, fmt.Errorf("read body: %w", err)
	}
	return resp.StatusCode, raw, nil
}

func (c *Client) SessionName(ctx context.Context, id string) (string, error) {
	status, raw, err := c.Do(ctx, http.MethodGet, "/sessions/"+url.PathEscape(id), nil)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("get session: status %d", status)
	}
	var s Session
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", fmt.Errorf("decode session: %w", err)
	}
	return s.Name, nil
}
