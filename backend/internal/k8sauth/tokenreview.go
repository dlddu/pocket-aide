// Package k8sauth verifies Kubernetes ServiceAccount tokens with the cluster's TokenReview API.
package k8sauth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	inClusterTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	inClusterCAPath    = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	saUserPrefix       = "system:serviceaccount:"
	cacheTTL           = 60 * time.Second
	cacheMax           = 1024
)

var (
	ErrUnauthenticated = errors.New("token is not an authenticated service account token for this audience")
	ErrUnavailable     = errors.New("token review unavailable")
)

type Config struct {
	APIURL   string
	Audience string
	Issuer   string
	// Token returns this server's own credential for the TokenReview call; it is re-read per call because kubelet rotates it.
	Token      func() (string, error)
	HTTPClient *http.Client
}

type Reviewer struct {
	cfg Config
	now func() time.Time

	mu    sync.Mutex
	cache map[[32]byte]cached
}

type cached struct {
	serviceAccount string
	err            error
	expires        time.Time
}

func New(cfg Config) *Reviewer {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Reviewer{cfg: cfg, now: time.Now, cache: map[[32]byte]cached{}}
}

func InCluster(audience, issuer string) (*Reviewer, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" || port == "" {
		return nil, errors.New("KUBERNETES_SERVICE_HOST/PORT not set: not running in a cluster")
	}
	ca, err := os.ReadFile(inClusterCAPath)
	if err != nil {
		return nil, fmt.Errorf("read cluster CA: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(ca) {
		return nil, errors.New("cluster CA has no certificates")
	}
	return New(Config{
		APIURL:   "https://" + strings.Trim(host, "[]") + ":" + port,
		Audience: audience,
		Issuer:   issuer,
		Token: func() (string, error) {
			b, err := os.ReadFile(inClusterTokenPath)
			return strings.TrimSpace(string(b)), err
		},
		HTTPClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}},
		},
	}), nil
}

// Review returns the "<namespace>/<name>" of the ServiceAccount the token was issued to.
func (r *Reviewer) Review(ctx context.Context, token string) (string, error) {
	claims, ok := peekClaims(token)
	if !ok || (r.cfg.Issuer != "" && claims.Issuer != r.cfg.Issuer) || !claims.Audience.has(r.cfg.Audience) {
		return "", ErrUnauthenticated
	}
	key := sha256.Sum256([]byte(token))
	now := r.now()
	if c, ok := r.lookup(key, now); ok {
		return c.serviceAccount, c.err
	}
	sa, err := r.review(ctx, token)
	if err != nil && !errors.Is(err, ErrUnauthenticated) {
		return "", err
	}
	expires := now.Add(cacheTTL)
	if claims.Expiry > 0 {
		if exp := time.Unix(claims.Expiry, 0); exp.Before(expires) {
			expires = exp
		}
	}
	r.store(key, cached{serviceAccount: sa, err: err, expires: expires}, now)
	return sa, err
}

func (r *Reviewer) lookup(key [32]byte, now time.Time) (cached, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.cache[key]
	if !ok || !now.Before(c.expires) {
		return cached{}, false
	}
	return c, true
}

func (r *Reviewer) store(key [32]byte, c cached, now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.cache) >= cacheMax {
		for k, v := range r.cache {
			if !now.Before(v.expires) {
				delete(r.cache, k)
			}
		}
		if len(r.cache) >= cacheMax {
			r.cache = map[[32]byte]cached{}
		}
	}
	r.cache[key] = c
}

type tokenReview struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Spec       struct {
		Token     string   `json:"token"`
		Audiences []string `json:"audiences"`
	} `json:"spec"`
	Status struct {
		Authenticated bool `json:"authenticated"`
		User          struct {
			Username string `json:"username"`
		} `json:"user"`
		Audiences []string `json:"audiences"`
		Error     string   `json:"error"`
	} `json:"status"`
}

func (r *Reviewer) review(ctx context.Context, token string) (string, error) {
	own, err := r.cfg.Token()
	if err != nil {
		return "", fmt.Errorf("%w: read own token: %v", ErrUnavailable, err)
	}
	var in tokenReview
	in.APIVersion, in.Kind = "authentication.k8s.io/v1", "TokenReview"
	in.Spec.Token, in.Spec.Audiences = token, []string{r.cfg.Audience}
	body, err := json.Marshal(in)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.cfg.APIURL+"/apis/authentication.k8s.io/v1/tokenreviews", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+own)
	resp, err := r.cfg.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: apiserver returned %d", ErrUnavailable, resp.StatusCode)
	}
	var out tokenReview
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("%w: decode: %v", ErrUnavailable, err)
	}
	if !out.Status.Authenticated || !audiences(out.Status.Audiences).has(r.cfg.Audience) {
		return "", ErrUnauthenticated
	}
	ns, name, ok := strings.Cut(strings.TrimPrefix(out.Status.User.Username, saUserPrefix), ":")
	if !strings.HasPrefix(out.Status.User.Username, saUserPrefix) || !ok || ns == "" || name == "" {
		return "", ErrUnauthenticated
	}
	return ns + "/" + name, nil
}

type audiences []string

func (a audiences) has(want string) bool {
	for _, v := range a {
		if v == want {
			return true
		}
	}
	return false
}

// UnmarshalJSON accepts both forms RFC 7519 allows for "aud": a string or an array of strings.
func (a *audiences) UnmarshalJSON(b []byte) error {
	var one string
	if err := json.Unmarshal(b, &one); err == nil {
		*a = audiences{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(b, &many); err != nil {
		return err
	}
	*a = many
	return nil
}

type claims struct {
	Issuer   string    `json:"iss"`
	Audience audiences `json:"aud"`
	Expiry   int64     `json:"exp"`
}

// peekClaims reads the unverified payload only to skip TokenReview for tokens that cannot pass it; TokenReview is the verification.
func peekClaims(token string) (claims, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return claims{}, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return claims{}, false
	}
	var c claims
	if err := json.Unmarshal(raw, &c); err != nil {
		return claims{}, false
	}
	return c, true
}
