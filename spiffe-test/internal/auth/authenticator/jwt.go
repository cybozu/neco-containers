package authenticator

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/svid/jwtsvid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc/metadata"

	"github.com/cybozu/neco-containers/spiffe-test/internal/auth"
)

var _ auth.Authenticator = &JWTAuthenticator{}

// JWTAuthenticator extracts identity from JWT SVID in Authorization header.
type JWTAuthenticator struct {
	jwtSource *workloadapi.JWTSource
	audience  string
}

func NewJWTAuthenticator(jwtSource *workloadapi.JWTSource, audience string) *JWTAuthenticator {
	return &JWTAuthenticator{
		jwtSource: jwtSource,
		audience:  audience,
	}
}

func (a *JWTAuthenticator) GetCallerID(r *http.Request) (spiffeid.ID, error) {
	token, err := a.extractBearerToken(r.Header.Get("Authorization"))
	if err != nil {
		return spiffeid.ID{}, err
	}
	return a.validateToken(token)
}

func (a *JWTAuthenticator) GetCallerIDFromContext(ctx context.Context) (spiffeid.ID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return spiffeid.ID{}, fmt.Errorf("failed to get metadata from context")
	}

	authValues := md.Get("authorization")
	if len(authValues) == 0 {
		return spiffeid.ID{}, fmt.Errorf("%w: missing authorization metadata", auth.ErrInvalidToken)
	}

	token, err := a.extractBearerToken(authValues[0])
	if err != nil {
		return spiffeid.ID{}, err
	}

	return a.validateToken(token)
}

func (a *JWTAuthenticator) extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("%w: missing authorization header", auth.ErrInvalidToken)
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("%w: invalid authorization header format", auth.ErrInvalidToken)
	}
	return parts[1], nil
}

func (a *JWTAuthenticator) validateToken(token string) (spiffeid.ID, error) {
	svid, err := jwtsvid.ParseAndValidate(token, a.jwtSource, []string{a.audience})
	if err != nil {
		return spiffeid.ID{}, fmt.Errorf("%w: %v", auth.ErrInvalidToken, err)
	}
	return svid.ID, nil
}
