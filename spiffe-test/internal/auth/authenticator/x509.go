package authenticator

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	"github.com/cybozu/neco-containers/spiffe-test/internal/auth"
)

var _ auth.Authenticator = &X509Authenticator{}

// X509Authenticator extracts identity from mTLS client certificate.
type X509Authenticator struct{}

func NewX509Authenticator() *X509Authenticator {
	return &X509Authenticator{}
}

func (a *X509Authenticator) GetCallerID(r *http.Request) (spiffeid.ID, error) {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return spiffeid.ID{}, auth.ErrNoClientCert
	}
	return a.validateCertificate(r.TLS.PeerCertificates[0])
}

func (a *X509Authenticator) GetCallerIDFromContext(ctx context.Context) (spiffeid.ID, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return spiffeid.ID{}, fmt.Errorf("failed to get peer from context")
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return spiffeid.ID{}, fmt.Errorf("failed to get TLS info from peer")
	}

	if len(tlsInfo.State.PeerCertificates) == 0 {
		return spiffeid.ID{}, auth.ErrNoClientCert
	}

	return a.validateCertificate(tlsInfo.State.PeerCertificates[0])
}

func (a *X509Authenticator) validateCertificate(cert *x509.Certificate) (spiffeid.ID, error) {
	id, err := x509svid.IDFromCert(cert)
	if err != nil {
		return spiffeid.ID{}, fmt.Errorf("failed to extract SPIFFE ID from certificate: %w", err)
	}
	return id, nil
}
