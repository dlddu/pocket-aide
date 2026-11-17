package commitstatus

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// DefaultAPIBase is GitHub's REST API root.
const DefaultAPIBase = "https://api.github.com"

const (
	githubAPIVersion = "2022-11-28"
	userAgent        = "pocket-aide"
	requestTimeout   = 20 * time.Second
)

// StatusWrite is the commit status an approved request writes.
type StatusWrite struct {
	Owner       string
	Repo        string
	SHA         string
	State       string
	Context     string
	Description string
	TargetURL   string
}

// Writer writes one commit status and returns GitHub's id for it.
type Writer interface {
	WriteStatus(ctx context.Context, w StatusWrite) (int64, error)
}

// App writes commit statuses as a GitHub App, minting an installation token scoped to one repository and statuses: write for each write.
type App struct {
	clientID string
	key      *rsa.PrivateKey
	apiBase  string
	http     *http.Client
	now      func() time.Time
}

// NewApp parses the App's PEM private key; apiBase empty means DefaultAPIBase.
func NewApp(clientID string, privateKeyPEM []byte, apiBase string) (*App, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return nil, errors.New("github app client id is empty")
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse github app private key: %w", err)
	}
	if apiBase == "" {
		apiBase = DefaultAPIBase
	}
	return &App{
		clientID: clientID,
		key:      key,
		apiBase:  strings.TrimRight(apiBase, "/"),
		http:     &http.Client{Timeout: requestTimeout},
		now:      time.Now,
	}, nil
}

func (a *App) appJWT() (string, error) {
	now := a.now()
	signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		Issuer:    a.clientID,
		IssuedAt:  jwt.NewNumericDate(now.Add(-time.Minute)),
		ExpiresAt: jwt.NewNumericDate(now.Add(9 * time.Minute)),
	}).SignedString(a.key)
	if err != nil {
		return "", fmt.Errorf("sign github app jwt: %w", err)
	}
	return signed, nil
}

// WriteStatus finds the App's installation on the repository, mints a token for that repository with statuses: write only, writes the status and revokes the token.
func (a *App) WriteStatus(ctx context.Context, w StatusWrite) (int64, error) {
	appToken, err := a.appJWT()
	if err != nil {
		return 0, err
	}
	repoPath := "/repos/" + url.PathEscape(w.Owner) + "/" + url.PathEscape(w.Repo)

	var installation struct {
		ID int64 `json:"id"`
	}
	if err := a.do(ctx, http.MethodGet, repoPath+"/installation", appToken, nil, &installation); err != nil {
		return 0, fmt.Errorf("find the app installation on %s/%s: %w", w.Owner, w.Repo, err)
	}

	var minted struct {
		Token string `json:"token"`
	}
	if err := a.do(ctx, http.MethodPost, fmt.Sprintf("/app/installations/%d/access_tokens", installation.ID), appToken, map[string]any{
		"repositories": []string{w.Repo},
		"permissions":  map[string]string{"statuses": "write"},
	}, &minted); err != nil {
		return 0, fmt.Errorf("mint an installation token: %w", err)
	}
	defer func() {
		_ = a.do(context.WithoutCancel(ctx), http.MethodDelete, "/installation/token", minted.Token, nil, nil)
	}()

	body := map[string]string{"state": w.State, "context": w.Context}
	if w.Description != "" {
		body["description"] = w.Description
	}
	if w.TargetURL != "" {
		body["target_url"] = w.TargetURL
	}
	var created struct {
		ID int64 `json:"id"`
	}
	if err := a.do(ctx, http.MethodPost, repoPath+"/statuses/"+url.PathEscape(w.SHA), minted.Token, body, &created); err != nil {
		return 0, fmt.Errorf("write the commit status: %w", err)
	}
	return created.ID, nil
}

func (a *App) do(ctx context.Context, method, path, token string, in, out any) error {
	var reader io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, a.apiBase+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", githubAPIVersion)
	req.Header.Set("User-Agent", userAgent)
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GitHub answered %d: %s", resp.StatusCode, githubMessage(raw))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("unreadable GitHub response: %w", err)
	}
	return nil
}

func githubMessage(raw []byte) string {
	var body struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(raw, &body) == nil && body.Message != "" {
		return body.Message
	}
	text := strings.TrimSpace(string(raw))
	if len(text) > 300 {
		text = text[:300]
	}
	return text
}
