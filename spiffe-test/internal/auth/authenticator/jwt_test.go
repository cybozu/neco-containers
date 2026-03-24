package authenticator

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/cybozu/neco-containers/spiffe-test/internal/auth"
)

func Test_NewJWTAuthenticator(t *testing.T) {
	tests := []struct {
		name     string
		audience string
	}{
		{
			name:     "valid configuration",
			audience: "my-audience",
		},
		{
			name:     "empty audience is allowed",
			audience: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewJWTAuthenticator(nil, tt.audience)
			if a == nil {
				t.Error("NewJWTAuthenticator() returned nil")
			}
		})
	}
}

func TestJWTAuthenticator_GetCallerID(t *testing.T) {
	jwtAuth := NewJWTAuthenticator(nil, "test-audience")

	t.Run("missing authorization header", func(t *testing.T) {
		req := &http.Request{Header: http.Header{}}
		_, err := jwtAuth.GetCallerID(req)
		if err == nil {
			t.Error("GetCallerID() expected error for missing authorization header, got nil")
			return
		}
		if !errors.Is(err, auth.ErrInvalidToken) {
			t.Errorf("GetCallerID() error = %v, want error wrapping %v", err, auth.ErrInvalidToken)
		}
		if !strings.Contains(err.Error(), "missing authorization header") {
			t.Errorf("GetCallerID() error = %v, want error containing 'missing authorization header'", err)
		}
	})

	t.Run("invalid authorization format - no space", func(t *testing.T) {
		req := &http.Request{Header: http.Header{"Authorization": []string{"InvalidHeader"}}}
		_, err := jwtAuth.GetCallerID(req)
		if err == nil {
			t.Error("GetCallerID() expected error for invalid authorization format, got nil")
			return
		}
		if !errors.Is(err, auth.ErrInvalidToken) {
			t.Errorf("GetCallerID() error = %v, want error wrapping %v", err, auth.ErrInvalidToken)
		}
	})

	t.Run("invalid authorization format - wrong scheme", func(t *testing.T) {
		req := &http.Request{Header: http.Header{"Authorization": []string{"Basic dXNlcjpwYXNz"}}}
		_, err := jwtAuth.GetCallerID(req)
		if err == nil {
			t.Error("GetCallerID() expected error for wrong auth scheme, got nil")
			return
		}
		if !errors.Is(err, auth.ErrInvalidToken) {
			t.Errorf("GetCallerID() error = %v, want error wrapping %v", err, auth.ErrInvalidToken)
		}
	})
}
