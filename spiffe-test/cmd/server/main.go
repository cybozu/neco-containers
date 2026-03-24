package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/spiffeid"

	"github.com/cybozu/neco-containers/spiffe-test/internal/auth"
	"github.com/cybozu/neco-containers/spiffe-test/internal/auth/authenticator"
	"github.com/cybozu/neco-containers/spiffe-test/internal/logging"
	"github.com/cybozu/neco-containers/spiffe-test/internal/service"
	"github.com/cybozu/neco-containers/spiffe-test/internal/spiffe"
	"github.com/cybozu/neco-containers/spiffe-test/internal/transport"
	transportgrpc "github.com/cybozu/neco-containers/spiffe-test/internal/transport/grpc"
	transporthttp "github.com/cybozu/neco-containers/spiffe-test/internal/transport/http"
)

type option struct {
	debug               bool
	transportStr        string
	authModeStr         string
	port                int
	allowedSPIFFEIDStrs []string
	audience            string
	noTLS               bool
	// parsed values
	transportKind transport.Kind
	authMode      authenticator.Mode
	allowedIDs    []spiffeid.ID
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	opt := &option{}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "SPIFFE/SPIRE simple server",
		Long:  "A simple HTTP server that authenticates clients using SPIFFE SVIDs",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runServer(opt); err != nil {
				slog.Error("Server failed", "error", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&opt.transportStr, "transport", "http", "Kind: http or grpc")
	cmd.Flags().StringVar(&opt.authModeStr, "auth", "x509", "Authentication mode: x509 or jwt")
	cmd.Flags().IntVar(&opt.port, "port", 10443, "Server port")
	cmd.Flags().StringSliceVar(&opt.allowedSPIFFEIDStrs, "allowed-client-id", nil, "Allowed client SPIFFE IDs (can be specified multiple times)")
	cmd.Flags().StringVar(&opt.audience, "audience", "simple-server", "JWT audience (for JWT mode)")
	cmd.Flags().BoolVar(&opt.noTLS, "no-tls", false, "Disable TLS (requires --auth jwt, for use behind L7 LB)")
	cmd.Flags().BoolVar(&opt.debug, "debug", false, "Enable debug logging")

	return cmd
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
	if o.noTLS && o.authMode == authenticator.ModeX509 {
		return fmt.Errorf("--no-tls cannot be used with --auth x509 (x509 mTLS requires TLS)")
	}
	for _, idStr := range o.allowedSPIFFEIDStrs {
		id, err := spiffeid.FromString(idStr)
		if err != nil {
			return fmt.Errorf("invalid allowed SPIFFE ID %q: %w", idStr, err)
		}
		o.allowedIDs = append(o.allowedIDs, id)
	}
	if len(o.allowedIDs) == 0 {
		slog.Warn("No allowed SPIFFE IDs specified, all authenticated clients will be rejected")
	}
	slog.Debug("Configuration loaded",
		"transport", string(o.transportKind),
		"authMode", string(o.authMode),
		"port", o.port,
		"allowedSPIFFEIDs", o.allowedSPIFFEIDStrs,
		"audience", o.audience,
		"noTLS", o.noTLS,
	)
	return nil
}

func runServer(opt *option) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
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

	helloService := service.NewHelloService(opt.allowedIDs)

	return startServer(ctx, opt, workloadClient, helloService)
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

func startServer(ctx context.Context, opt *option, workloadClient *spiffe.WorkloadClient, helloService service.HelloService) error {
	addr := fmt.Sprintf(":%d", opt.port)
	tlsProvider := auth.NewX509TLSConfigProvider(workloadClient.X509Source())

	var (
		authn     auth.Authenticator
		tlsConfig *tls.Config
	)
	switch opt.authMode {
	case authenticator.ModeX509:
		authn = authenticator.NewX509Authenticator()
		if !opt.noTLS {
			tlsConfig = tlsProvider.ServerMTLSConfig()
		}
	case authenticator.ModeJWT:
		authn = authenticator.NewJWTAuthenticator(workloadClient.JWTSource(), opt.audience)
		if !opt.noTLS {
			tlsConfig = tlsProvider.ServerTLSConfig()
		}
	}

	slog.Info("Server starting", "port", opt.port, "transport", string(opt.transportKind), "authMode", string(opt.authMode))

	switch opt.transportKind {
	case transport.HTTP:
		return transporthttp.NewServer(addr, tlsConfig, authn, helloService).Start(ctx)
	case transport.GRPC:
		return transportgrpc.NewServer(addr, tlsConfig, authn, helloService).Start(ctx)
	default:
		return fmt.Errorf("unknown transport: %s", opt.transportStr)
	}
}
