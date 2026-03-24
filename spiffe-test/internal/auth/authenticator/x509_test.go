package authenticator

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/cybozu/neco-containers/spiffe-test/internal/auth"
)

func Test_NewX509Authenticator(t *testing.T) {
	a := NewX509Authenticator()
	if a == nil {
		t.Error("NewX509Authenticator() returned nil")
	}
}

func TestX509Authenticator_validateCertificate(t *testing.T) {
	tests := []struct {
		name     string
		spiffeID string
		wantErr  bool
	}{
		{
			name:     "valid certificate with SPIFFE ID",
			spiffeID: "spiffe://example.com/workload",
			wantErr:  false,
		},
		{
			name:     "valid certificate from another trust domain",
			spiffeID: "spiffe://another.org/workload",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewX509Authenticator()

			cert := createTestCertWithSPIFFEID(t, tt.spiffeID)
			id, err := a.validateCertificate(cert)

			if tt.wantErr {
				if err == nil {
					t.Error("validateCertificate() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("validateCertificate() unexpected error = %v", err)
				return
			}
			if id.String() != tt.spiffeID {
				t.Errorf("validateCertificate() returned ID = %v, want %v", id.String(), tt.spiffeID)
			}
		})
	}
}

func TestX509Authenticator_validateCertificate_NoSPIFFEID(t *testing.T) {
	a := NewX509Authenticator()

	cert := createTestCertWithoutSPIFFEID(t)
	_, err := a.validateCertificate(cert)
	if err == nil {
		t.Error("validateCertificate() expected error for cert without SPIFFE ID, got nil")
	}
}

func createTestCert(t *testing.T, uris []*url.URL) *x509.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(1 * time.Hour),
		URIs:      uris,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	return cert
}

func createTestCertWithSPIFFEID(t *testing.T, spiffeID string) *x509.Certificate {
	t.Helper()
	spiffeURI, err := url.Parse(spiffeID)
	if err != nil {
		t.Fatalf("Failed to parse SPIFFE ID: %v", err)
	}
	return createTestCert(t, []*url.URL{spiffeURI})
}

func createTestCertWithoutSPIFFEID(t *testing.T) *x509.Certificate {
	t.Helper()
	return createTestCert(t, nil)
}

func TestX509Authenticator_GetCallerID(t *testing.T) {
	a := NewX509Authenticator()

	t.Run("valid mTLS request", func(t *testing.T) {
		cert := createTestCertWithSPIFFEID(t, "spiffe://example.com/workload")
		req := &http.Request{
			TLS: &tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{cert},
			},
		}
		id, err := a.GetCallerID(req)
		if err != nil {
			t.Errorf("GetCallerID() unexpected error = %v", err)
			return
		}
		if id.String() != "spiffe://example.com/workload" {
			t.Errorf("GetCallerID() returned ID = %v, want %v", id.String(), "spiffe://example.com/workload")
		}
	})

	t.Run("request without TLS", func(t *testing.T) {
		req := &http.Request{TLS: nil}
		_, err := a.GetCallerID(req)
		if err == nil {
			t.Error("GetCallerID() expected error for request without TLS, got nil")
			return
		}
		if !errors.Is(err, auth.ErrNoClientCert) {
			t.Errorf("GetCallerID() error = %v, want %v", err, auth.ErrNoClientCert)
		}
	})

	t.Run("TLS without peer certificates", func(t *testing.T) {
		req := &http.Request{TLS: &tls.ConnectionState{PeerCertificates: nil}}
		_, err := a.GetCallerID(req)
		if err == nil {
			t.Error("GetCallerID() expected error for TLS without peer certs, got nil")
			return
		}
		if !errors.Is(err, auth.ErrNoClientCert) {
			t.Errorf("GetCallerID() error = %v, want %v", err, auth.ErrNoClientCert)
		}
	})

	t.Run("TLS with empty peer certificates", func(t *testing.T) {
		req := &http.Request{TLS: &tls.ConnectionState{PeerCertificates: []*x509.Certificate{}}}
		_, err := a.GetCallerID(req)
		if err == nil {
			t.Error("GetCallerID() expected error for TLS with empty peer certs, got nil")
			return
		}
		if !errors.Is(err, auth.ErrNoClientCert) {
			t.Errorf("GetCallerID() error = %v, want %v", err, auth.ErrNoClientCert)
		}
	})
}
