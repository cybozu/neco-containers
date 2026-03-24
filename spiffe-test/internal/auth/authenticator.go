package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

var ErrNoClientCert = fmt.Errorf("no client certificate provided")

var ErrInvalidToken = fmt.Errorf("invalid token")

// Authenticator extracts and verifies caller identity.
// GetCallerID is for HTTP handlers, GetCallerIDFromContext is for gRPC handlers.
type Authenticator interface {
	GetCallerID(r *http.Request) (spiffeid.ID, error)
	GetCallerIDFromContext(ctx context.Context) (spiffeid.ID, error)
}
