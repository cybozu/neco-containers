package auth

import (
	"crypto/tls"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

type TLSConfigProvider interface {
	ServerMTLSConfig() *tls.Config
	ServerTLSConfig() *tls.Config
	ClientMTLSConfig(serverID spiffeid.ID) *tls.Config
	ClientTLSConfig(serverID spiffeid.ID) *tls.Config
}

type x509TLSConfigProvider struct {
	x509Source *workloadapi.X509Source
}

func NewX509TLSConfigProvider(x509Source *workloadapi.X509Source) TLSConfigProvider {
	return &x509TLSConfigProvider{
		x509Source: x509Source,
	}
}

// ServerMTLSConfig returns a TLS config that requires client certificates but
// accepts any valid SPIFFE ID at the TLS layer (AuthorizeAny). The actual
// authorization of caller identity is performed by auth.Authenticator and the
// service layer.
func (p *x509TLSConfigProvider) ServerMTLSConfig() *tls.Config {
	return tlsconfig.MTLSServerConfig(
		p.x509Source,
		p.x509Source,
		tlsconfig.AuthorizeAny(),
	)
}

func (p *x509TLSConfigProvider) ServerTLSConfig() *tls.Config {
	return tlsconfig.TLSServerConfig(p.x509Source)
}

func (p *x509TLSConfigProvider) ClientMTLSConfig(serverID spiffeid.ID) *tls.Config {
	return tlsconfig.MTLSClientConfig(
		p.x509Source,
		p.x509Source,
		tlsconfig.AuthorizeID(serverID),
	)
}

func (p *x509TLSConfigProvider) ClientTLSConfig(serverID spiffeid.ID) *tls.Config {
	return tlsconfig.TLSClientConfig(
		p.x509Source,
		tlsconfig.AuthorizeID(serverID),
	)
}
