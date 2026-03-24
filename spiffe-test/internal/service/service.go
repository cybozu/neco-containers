package service

import (
	"context"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

// HelloService defines the core business logic
type HelloService interface {
	// SayHello returns a greeting message for the caller
	// Returns error if the caller is not authorized
	SayHello(ctx context.Context, callerID spiffeid.ID) (string, error)
}
