package mail

import (
	"strings"
	"testing"
)

func TestPreviewOf(t *testing.T) {
	cases := map[string]struct{ raw, want string }{
		"plain":            {"Subject: s\r\n\r\nhello\r\n there", "hello there"},
		"quoted-printable": {"Content-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n=ED=95=9C=EA=B8=80 caf=C3=A9", "한글 café"},
		"base64":           {"Content-Type: text/plain\r\nContent-Transfer-Encoding: base64\r\n\r\n7JWI64WV7ZWY7IS47JqU\r\n", "안녕하세요"},
		"html only":        {"Content-Type: text/html\r\n\r\n<style>p{}</style><p>Hi <b>there</b></p>", "Hi there"},
		"multipart prefers plain": {
			"Content-Type: multipart/alternative; boundary=XX\r\n\r\n--XX\r\nContent-Type: text/html\r\n\r\n<p>html</p>\r\n--XX\r\nContent-Type: text/plain\r\n\r\nplain text\r\n--XX--\r\n",
			"plain text",
		},
		"multipart html fallback": {
			"Content-Type: multipart/alternative; boundary=XX\r\n\r\n--XX\r\nContent-Type: text/html\r\n\r\n<p>only html</p>\r\n--XX--\r\n",
			"only html",
		},
		"truncated multipart": {
			"Content-Type: multipart/mixed; boundary=XX\r\n\r\n--XX\r\nContent-Type: text/plain\r\n\r\ncut off mid",
			"cut off mid",
		},
		"attachment only": {"Content-Type: application/pdf\r\n\r\n%PDF", ""},
		"not a message":   {"", ""},
	}
	for name, c := range cases {
		if got := previewOf([]byte(c.raw)); got != c.want {
			t.Errorf("%s: previewOf = %q, want %q", name, got, c.want)
		}
	}
}

func TestPreviewOf_CapsLengthByRunes(t *testing.T) {
	got := previewOf([]byte("Subject: s\r\n\r\n" + strings.Repeat("가", 500)))
	if n := len([]rune(got)); n != previewRunes {
		t.Errorf("preview runes = %d, want %d", n, previewRunes)
	}
}
