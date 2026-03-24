package spiffe

import (
	"context"
	"errors"
	"fmt"

	"github.com/spiffe/go-spiffe/v2/svid/jwtsvid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// WorkloadClient is a client of the SPIFFE Workload API. Both the application's
// client and server use this to fetch SVIDs from the Workload API.
// The socket address is resolved from the SPIFFE_ENDPOINT_SOCKET environment variable.
type WorkloadClient struct {
	x509Source *workloadapi.X509Source
	jwtSource  *workloadapi.JWTSource
}

// WorkloadClientOption configures a WorkloadClient.
type WorkloadClientOption func(ctx context.Context, w *WorkloadClient) error

// WithJWTSource enables JWT source initialization.
func WithJWTSource() WorkloadClientOption {
	return func(ctx context.Context, w *WorkloadClient) error {
		source, err := workloadapi.NewJWTSource(ctx)
		if err != nil {
			return fmt.Errorf("failed to create JWTSource: %w", err)
		}
		w.jwtSource = source
		return nil
	}
}

// NewWorkloadClient creates a new WorkloadClient.
// X509Source is always initialized. Additional sources can be initialized via options.
func NewWorkloadClient(ctx context.Context, opts ...WorkloadClientOption) (*WorkloadClient, error) {
	x509Source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create X509Source: %w", err)
	}

	w := &WorkloadClient{
		x509Source: x509Source,
	}

	for _, opt := range opts {
		if err := opt(ctx, w); err != nil {
			_ = w.Close()
			return nil, err
		}
	}

	return w, nil
}

func (w *WorkloadClient) X509Source() *workloadapi.X509Source {
	return w.x509Source
}

func (w *WorkloadClient) JWTSource() *workloadapi.JWTSource {
	return w.jwtSource
}

func (w *WorkloadClient) FetchJWTSVID(ctx context.Context, audience string) (*jwtsvid.SVID, error) {
	if w.jwtSource == nil {
		return nil, fmt.Errorf("JWT source not initialized")
	}
	svid, err := w.jwtSource.FetchJWTSVID(ctx, jwtsvid.Params{
		Audience: audience,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWT SVID: %w", err)
	}
	return svid, nil
}

func (w *WorkloadClient) Close() error {
	var errs []error
	if w.x509Source != nil {
		if err := w.x509Source.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close X509Source: %w", err))
		}
	}
	if w.jwtSource != nil {
		if err := w.jwtSource.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close JWTSource: %w", err))
		}
	}
	return errors.Join(errs...)
}
