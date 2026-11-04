package mail

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	stdmail "net/mail"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

// ErrAuth means the IMAP server rejected the login (typically a revoked or
// mistyped app password). Anything else wrong with a connection is a plain
// error — only ErrAuth makes an account "needs reconnect".
var ErrAuth = errors.New("mail: IMAP authentication failed")

const (
	previewRunes = 120
	// peekBytes is how much of each message is fetched to build the preview:
	// headers plus the first bytes of the body.
	peekBytes = 16 * 1024
)

// Message is one unseen inbox message as shown in the mail list (PRD-8 AC11).
type Message struct {
	UID        uint32    `json:"uid"`
	From       string    `json:"from"`
	Subject    string    `json:"subject"`
	ReceivedAt time.Time `json:"received_at"`
	Preview    string    `json:"preview"`
}

// IMAP is the read-only IMAP access the service needs.
type IMAP interface {
	// Probe logs in and out; it is the connection test run when an account is
	// saved. A rejected login is ErrAuth.
	Probe(ctx context.Context, cred Credentials) error
	// Unseen returns up to limit of the newest unseen INBOX messages without
	// changing any message state.
	Unseen(ctx context.Context, cred Credentials, limit int) ([]Message, error)
}

// Client is the go-imap backed IMAP implementation. Plaintext (Insecure) is
// for tests against an in-memory server only.
type Client struct {
	Timeout   time.Duration
	TLSConfig *tls.Config
	Insecure  bool
}

func (c *Client) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return 20 * time.Second
}

func (c *Client) open(ctx context.Context, cred Credentials) (*imapclient.Client, func(), error) {
	addr := net.JoinHostPort(cred.Host, strconv.Itoa(cred.Port))
	opts := &imapclient.Options{
		TLSConfig: c.TLSConfig,
		Dialer:    &net.Dialer{Timeout: c.timeout()},
	}
	var cl *imapclient.Client
	var err error
	if c.Insecure {
		cl, err = imapclient.DialInsecure(addr, opts)
	} else {
		cl, err = imapclient.DialTLS(addr, opts)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("mail: connect %s: %w", addr, err)
	}
	// go-imap has no per-call context; closing the connection unblocks any
	// pending command when the caller gives up.
	stop := context.AfterFunc(ctx, func() { _ = cl.Close() })
	closeAll := func() {
		stop()
		_ = cl.Logout().Wait()
		_ = cl.Close()
	}
	if err := cl.Login(cred.Email, cred.Password).Wait(); err != nil {
		closeAll()
		var ie *imap.Error
		if errors.As(err, &ie) {
			return nil, nil, ErrAuth
		}
		return nil, nil, fmt.Errorf("mail: login: %w", err)
	}
	return cl, closeAll, nil
}

// Probe implements IMAP.
func (c *Client) Probe(ctx context.Context, cred Credentials) error {
	_, done, err := c.open(ctx, cred)
	if err != nil {
		return err
	}
	done()
	return nil
}

// Unseen implements IMAP. The mailbox is selected read-only and bodies are
// fetched with BODY.PEEK, so no message is ever marked read.
func (c *Client) Unseen(ctx context.Context, cred Credentials, limit int) ([]Message, error) {
	cl, done, err := c.open(ctx, cred)
	if err != nil {
		return nil, err
	}
	defer done()

	if _, err := cl.Select("INBOX", &imap.SelectOptions{ReadOnly: true}).Wait(); err != nil {
		return nil, fmt.Errorf("mail: select INBOX: %w", err)
	}
	found, err := cl.UIDSearch(&imap.SearchCriteria{NotFlag: []imap.Flag{imap.FlagSeen}}, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("mail: search unseen: %w", err)
	}
	uids := found.AllUIDs()
	if len(uids) == 0 {
		return []Message{}, nil
	}
	sort.Slice(uids, func(i, j int) bool { return uids[i] > uids[j] })
	if len(uids) > limit {
		uids = uids[:limit]
	}
	section := &imap.FetchItemBodySection{Peek: true, Partial: &imap.SectionPartial{Offset: 0, Size: peekBytes}}
	bufs, err := cl.Fetch(imap.UIDSetNum(uids...), &imap.FetchOptions{
		UID:          true,
		Envelope:     true,
		InternalDate: true,
		BodySection:  []*imap.FetchItemBodySection{section},
	}).Collect()
	if err != nil {
		return nil, fmt.Errorf("mail: fetch: %w", err)
	}
	out := make([]Message, 0, len(bufs))
	for _, b := range bufs {
		m := Message{UID: uint32(b.UID), ReceivedAt: b.InternalDate}
		if b.Envelope != nil {
			m.Subject = b.Envelope.Subject
			m.From = formatFrom(b.Envelope.From)
		}
		if len(b.BodySection) > 0 {
			m.Preview = previewOf(b.BodySection[0].Bytes)
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ReceivedAt.After(out[j].ReceivedAt) })
	return out, nil
}

func formatFrom(from []imap.Address) string {
	if len(from) == 0 {
		return ""
	}
	a := from[0]
	if a.Name != "" {
		return a.Name
	}
	return a.Addr()
}

var (
	tagRe   = regexp.MustCompile(`(?s)<(script|style)[^>]*>.*?</(script|style)>|<[^>]*>`)
	spaceRe = regexp.MustCompile(`\s+`)
)

// previewOf returns a one-line preview from the leading bytes of a raw
// message: the first text/plain part (or tag-stripped text/html), decoded as
// UTF-8 and collapsed to one line. It returns "" when nothing readable is
// found — the list shows the row without a preview rather than failing.
func previewOf(raw []byte) string {
	msg, err := stdmail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return ""
	}
	text := textOf(msg.Header.Get("Content-Type"), msg.Header.Get("Content-Transfer-Encoding"), msg.Body, 0)
	text = spaceRe.ReplaceAllString(strings.TrimSpace(text), " ")
	text = strings.ToValidUTF8(text, "")
	if utf8.RuneCountInString(text) > previewRunes {
		text = string([]rune(text)[:previewRunes])
	}
	return text
}

func textOf(contentType, cte string, body io.Reader, depth int) string {
	mt, params, err := mime.ParseMediaType(contentType)
	if err != nil || contentType == "" {
		mt = "text/plain"
	}
	switch {
	case strings.HasPrefix(mt, "multipart/") && depth < 3:
		mr := multipart.NewReader(body, params["boundary"])
		var html string
		for {
			part, err := mr.NextPart()
			if err != nil {
				break
			}
			t := textOf(part.Header.Get("Content-Type"), part.Header.Get("Content-Transfer-Encoding"), part, depth+1)
			pmt, _, _ := mime.ParseMediaType(part.Header.Get("Content-Type"))
			if t == "" {
				continue
			}
			if pmt == "" || pmt == "text/plain" || strings.HasPrefix(pmt, "multipart/") {
				return t
			}
			if pmt == "text/html" && html == "" {
				html = t
			}
		}
		return html
	case mt == "text/plain":
		return decodeBody(cte, body)
	case mt == "text/html":
		return tagRe.ReplaceAllString(decodeBody(cte, body), " ")
	}
	return ""
}

// decodeBody reads as much as it can: the fetch is truncated, so a decoding
// error near the end is expected and the text read so far is kept.
func decodeBody(cte string, body io.Reader) string {
	r := body
	switch strings.ToLower(strings.TrimSpace(cte)) {
	case "quoted-printable":
		r = quotedprintable.NewReader(body)
	case "base64":
		r = base64.NewDecoder(base64.StdEncoding, &stripNewlines{r: body})
	}
	b, _ := io.ReadAll(io.LimitReader(r, peekBytes))
	return string(b)
}

type stripNewlines struct{ r io.Reader }

func (s *stripNewlines) Read(p []byte) (int, error) {
	n, err := s.r.Read(p)
	w := 0
	for i := 0; i < n; i++ {
		if p[i] != '\r' && p[i] != '\n' {
			p[w] = p[i]
			w++
		}
	}
	return w, err
}
