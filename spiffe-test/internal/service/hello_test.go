package service

import (
	"context"
	"errors"
	"testing"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func TestNewHelloService(t *testing.T) {
	tests := []struct {
		name       string
		allowedIDs []spiffeid.ID
	}{
		{
			name:       "with allowed IDs",
			allowedIDs: mustParseIDs(t, "spiffe://example.com/workload1", "spiffe://example.com/workload2"),
		},
		{
			name:       "with empty list",
			allowedIDs: []spiffeid.ID{},
		},
		{
			name:       "with nil list",
			allowedIDs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewHelloService(tt.allowedIDs)
			if svc == nil {
				t.Error("NewHelloService() returned nil")
			}
		})
	}
}

func TestHelloService_SayHello(t *testing.T) {
	allowedIDs := mustParseIDs(t, "spiffe://example.com/workload1", "spiffe://example.com/workload2")

	tests := []struct {
		name        string
		callerID    string
		wantErr     bool
		errType     error
		wantMessage string
	}{
		{
			name:        "allowed caller",
			callerID:    "spiffe://example.com/workload1",
			wantErr:     false,
			wantMessage: "Hello, spiffe://example.com/workload1!",
		},
		{
			name:        "another allowed caller",
			callerID:    "spiffe://example.com/workload2",
			wantErr:     false,
			wantMessage: "Hello, spiffe://example.com/workload2!",
		},
		{
			name:     "unauthorized caller",
			callerID: "spiffe://example.com/unknown",
			wantErr:  true,
			errType:  ErrUnauthorized,
		},
		{
			name:     "caller from different trust domain",
			callerID: "spiffe://other.org/workload1",
			wantErr:  true,
			errType:  ErrUnauthorized,
		},
	}

	svc := NewHelloService(allowedIDs)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callerID, err := spiffeid.FromString(tt.callerID)
			if err != nil {
				t.Fatalf("Failed to parse caller ID: %v", err)
			}

			message, err := svc.SayHello(ctx, callerID)
			if tt.wantErr {
				if err == nil {
					t.Error("SayHello() expected error, got nil")
					return
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("SayHello() error = %v, want error wrapping %v", err, tt.errType)
				}
				return
			}
			if err != nil {
				t.Errorf("SayHello() unexpected error = %v", err)
				return
			}
			if message != tt.wantMessage {
				t.Errorf("SayHello() = %q, want %q", message, tt.wantMessage)
			}
		})
	}
}

func TestHelloService_SayHello_EmptyAllowList(t *testing.T) {
	svc := NewHelloService([]spiffeid.ID{})
	ctx := context.Background()

	callerID, err := spiffeid.FromString("spiffe://example.com/workload1")
	if err != nil {
		t.Fatalf("Failed to parse caller ID: %v", err)
	}

	_, err = svc.SayHello(ctx, callerID)
	if err == nil {
		t.Error("SayHello() expected error with empty allow list, got nil")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("SayHello() error = %v, want error wrapping %v", err, ErrUnauthorized)
	}
}

func mustParseIDs(t *testing.T, ids ...string) []spiffeid.ID {
	t.Helper()
	result := make([]spiffeid.ID, len(ids))
	for i, s := range ids {
		id, err := spiffeid.FromString(s)
		if err != nil {
			t.Fatalf("Failed to parse SPIFFE ID %q: %v", s, err)
		}
		result[i] = id
	}
	return result
}
