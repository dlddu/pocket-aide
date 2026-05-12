// Command oidcmock is a standalone HTTP server that exposes the oidcmock
// package as a real listener on a configurable port. Used for local development
// and CI integration tests where the mock IdP must be reachable from outside
// the test process (e.g. iOS simulator, the backend binary).
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/dlddu/pocket-aide/backend/internal/oidcmock"
)

func main() {
	addr := flag.String("addr", ":5556", "listen address")
	clientID := flag.String("client-id", "pocket-aide-ios-dev", "expected client_id")
	audience := flag.String("audience", "pocket-aide-dev", "audience claim")
	subject := flag.String("subject", "mock-user-123", "subject claim")
	issuerOverride := flag.String("issuer", "", "issuer URL override (default: http://localhost<addr>)")
	flag.Parse()

	srv, err := oidcmock.New(oidcmock.Options{
		ClientID: *clientID,
		Audience: *audience,
		Subject:  *subject,
	})
	if err != nil {
		log.Fatalf("oidcmock new: %v", err)
	}

	issuer := *issuerOverride
	if issuer == "" {
		issuer = fmt.Sprintf("http://localhost%s", *addr)
	}
	srv.SetIssuer(issuer)

	log.Printf("oidcmock listening on %s, issuer=%s, client_id=%s, audience=%s, subject=%s",
		*addr, issuer, *clientID, *audience, *subject)
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
