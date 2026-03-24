package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/spiffeid"

	"github.com/cybozu/neco-containers/spiffe-test/internal/auth"
	"github.com/cybozu/neco-containers/spiffe-test/internal/auth/authenticator"
	"github.com/cybozu/neco-containers/spiffe-test/internal/logging"
	"github.com/cybozu/neco-containers/spiffe-test/internal/spiffe"
	"github.com/cybozu/neco-containers/spiffe-test/internal/transport"
	transportgrpc "github.com/cybozu/neco-containers/spiffe-test/internal/transport/grpc"
	transporthttp "github.com/cybozu/neco-containers/spiffe-test/internal/transport/http"
)

type option struct {
	debug        bool
	transportStr string
	authModeStr  string
	serverAddr   string
	serverIDStr  string
	audience     string
	loopInterval time.Duration
	loopCount    int
	// parsed values
	authMode      authenticator.Mode
	transportKind transport.Kind
	serverID      spiffeid.ID
}

func (o *option) parse() error {
	o.authMode = authenticator.ParseMode(o.authModeStr)
	if o.authMode == "" {
		return fmt.Errorf("unknown auth mode: %s (must be 'x509' or 'jwt')", o.authModeStr)
	}
	o.transportKind = transport.Parse(o.transportStr)
	if o.transportKind == "" {
		return fmt.Errorf("unknown transport: %s (must be 'http' or 'grpc')", o.transportStr)
	}

	var err error
	if o.serverID, err = spiffeid.FromString(o.serverIDStr); err != nil {
		return fmt.Errorf("invalid server SPIFFE ID: %w", err)
	}

	slog.Debug("Configuration loaded",
		"transport", string(o.transportKind),
		"authMode", string(o.authMode),
		"serverAddr", o.serverAddr,
		"serverID", o.serverIDStr,
		"audience", o.audience,
		"interval", o.loopInterval,
		"count", o.loopCount,
	)
	return nil
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	opt := &option{}

	cmd := &cobra.Command{
		Use:   "client",
		Short: "SPIFFE/SPIRE simple client",
		Long:  "A simple client that authenticates to servers using SPIFFE SVIDs",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runClient(cmd.Context(), opt); err != nil {
				slog.Error("Client failed", "error", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&opt.transportStr, "transport", "http", "Kind: http or grpc")
	cmd.Flags().StringVar(&opt.authModeStr, "auth", "x509", "Authentication mode: x509 or jwt")
	cmd.Flags().StringVar(&opt.serverAddr, "server", "https://simple-server:10443", "Server address (http: full URL, grpc: host:port)")
	cmd.Flags().StringVar(&opt.serverIDStr, "server-id", "spiffe://example.com/server", "Expected server SPIFFE ID")
	cmd.Flags().StringVar(&opt.audience, "audience", "simple-server", "JWT audience (for JWT mode)")
	cmd.Flags().DurationVar(&opt.loopInterval, "interval", 5*time.Second, "Interval between requests (0 for single request)")
	cmd.Flags().IntVar(&opt.loopCount, "count", 0, "Number of requests (0 for infinite)")
	cmd.Flags().BoolVar(&opt.debug, "debug", false, "Enable debug logging")

	return cmd
}

func runClient(ctx context.Context, opt *option) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logging.Setup(opt.debug)

	if err := opt.parse(); err != nil {
		return err
	}

	workloadClient, err := initWorkloadClient(ctx, opt)
	if err != nil {
		return err
	}
	defer func() {
		if err := workloadClient.Close(); err != nil {
			slog.Error("Failed to close workload client", "error", err)
		}
	}()

	client, err := createClient(opt, workloadClient)
	if err != nil {
		return err
	}
	defer func() {
		if err := client.Close(); err != nil {
			slog.Error("Failed to close client", "error", err)
		}
	}()

	slog.Info("Client created", "transport", string(opt.transportKind), "authMode", string(opt.authMode))

	return runRequestLoop(ctx, opt, client, workloadClient)
}

func initWorkloadClient(ctx context.Context, opt *option) (*spiffe.WorkloadClient, error) {
	var opts []spiffe.WorkloadClientOption
	if opt.authMode == authenticator.ModeJWT {
		opts = append(opts, spiffe.WithJWTSource())
	}

	workloadClient, err := spiffe.NewWorkloadClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create workload client: %w", err)
	}

	return workloadClient, nil
}

func createClient(opt *option, workloadClient *spiffe.WorkloadClient) (transport.HelloClient, error) {
	tlsProvider := auth.NewX509TLSConfigProvider(workloadClient.X509Source())
	tlsConfig := clientTLSConfig(opt.authMode, tlsProvider, opt.serverID)

	switch opt.transportKind {
	case transport.HTTP:
		return transporthttp.NewClient(opt.serverAddr, tlsConfig), nil
	case transport.GRPC:
		grpcClient, err := transportgrpc.NewClient(opt.serverAddr, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC client: %w", err)
		}
		return grpcClient, nil
	default:
		return nil, fmt.Errorf("unknown transport: %s", opt.transportStr)
	}
}

func clientTLSConfig(mode authenticator.Mode, provider auth.TLSConfigProvider, srvID spiffeid.ID) *tls.Config {
	if mode == authenticator.ModeX509 {
		return provider.ClientMTLSConfig(srvID)
	}
	return provider.ClientTLSConfig(srvID)
}

func runRequestLoop(ctx context.Context, opt *option, client transport.HelloClient, workloadClient *spiffe.WorkloadClient) error {
	count := 0
	for {
		count++

		if err := prepareJWTIfNeeded(ctx, opt, client, workloadClient); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			// Wait and continue to next iteration instead of making unauthenticated request
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(opt.loopInterval):
				continue
			}
		}

		slog.Info("Making request", "count", count, "server", opt.serverAddr, "transport", string(opt.transportKind))
		message, err := client.SayHello(ctx)
		if err != nil {
			slog.Error("Request failed", "error", err)
		} else {
			slog.Info("Response received", "message", message)
		}

		if shouldStop(opt, count) {
			break
		}

		// Wait for next iteration
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(opt.loopInterval):
		}
	}

	return nil
}

// prepareJWTIfNeeded fetches a fresh JWT SVID if using JWT auth mode.
// JWT SVIDs have short TTLs (typically ~5 minutes) to minimize risk from token leakage.
func prepareJWTIfNeeded(ctx context.Context, opt *option, client transport.HelloClient, workloadClient *spiffe.WorkloadClient) error {
	if opt.authMode != authenticator.ModeJWT {
		return nil
	}

	jwtSVID, err := workloadClient.FetchJWTSVID(ctx, opt.audience)
	if err != nil {
		slog.Error("Failed to fetch JWT SVID, skipping request", "error", err)
		return err
	}
	client.SetJWTToken(jwtSVID.Marshal())
	slog.Debug("Fetched JWT SVID", "spiffeID", jwtSVID.ID.String())
	return nil
}

func shouldStop(opt *option, count int) bool {
	if opt.loopCount > 0 && count >= opt.loopCount {
		slog.Info("Reached request count limit, stopping")
		return true
	}
	if opt.loopInterval == 0 {
		return true
	}
	return false
}
