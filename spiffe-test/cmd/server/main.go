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
	debug            bool
	transportStr     string
	transportMode    transport.Transport
	authModeStr      string
	authMode         authenticator.Mode
	port             int
	allowedSPIFFEIDs []string
	socketPath       string
	audience         string
	noTLS            bool
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

	cmd.Flags().StringVar(&opt.transportStr, "transport", "http", "Transport: http or grpc")
	cmd.Flags().StringVar(&opt.authModeStr, "auth", "x509", "Authentication mode: x509 or jwt")
	cmd.Flags().IntVar(&opt.port, "port", 10443, "Server port")
	cmd.Flags().StringSliceVar(&opt.allowedSPIFFEIDs, "allowed-spiffe-id", nil, "Allowed SPIFFE IDs (can be specified multiple times)")
	cmd.Flags().StringVar(&opt.socketPath, "socket-path", "unix:///spiffe-workload-api/spire-agent.sock", "SPIFFE Workload API socket path")
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
	o.transportMode = transport.Parse(o.transportStr)
	if o.transportMode == "" {
		return fmt.Errorf("unknown transport: %s (must be 'http' or 'grpc')", o.transportStr)
	}
	if envSocket := os.Getenv("SPIFFE_ENDPOINT_SOCKET"); envSocket != "" {
		o.socketPath = envSocket
	}
	if o.noTLS && o.authMode == authenticator.ModeX509 {
		return fmt.Errorf("--no-tls cannot be used with --auth x509 (x509 mTLS requires TLS)")
	}
	if len(o.allowedSPIFFEIDs) == 0 {
		slog.Warn("No allowed SPIFFE IDs specified, all authenticated clients will be rejected")
	}
	slog.Debug("Configuration loaded",
		"transport", string(o.transportMode),
		"authMode", string(o.authMode),
		"port", o.port,
		"allowedSPIFFEIDs", o.allowedSPIFFEIDs,
		"socketPath", o.socketPath,
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

	helloService := service.NewHelloService(opt.allowedSPIFFEIDs)

	return startServer(ctx, opt, workloadClient, helloService)
}

func initWorkloadClient(ctx context.Context, opt *option) (*spiffe.WorkloadClient, error) {
	var opts []spiffe.WorkloadClientOption
	if opt.authMode == authenticator.ModeJWT {
		opts = append(opts, spiffe.WithJWTSource())
	}

	workloadClient, err := spiffe.NewWorkloadClient(ctx, opt.socketPath, opts...)
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

	slog.Info("Server starting", "port", opt.port, "transport", string(opt.transportMode), "authMode", string(opt.authMode))

	switch opt.transportMode {
	case transport.HTTP:
		return transporthttp.NewServer(addr, tlsConfig, authn, helloService).Start(ctx)
	case transport.GRPC:
		return transportgrpc.NewServer(addr, tlsConfig, authn, helloService).Start(ctx)
	default:
		return fmt.Errorf("unknown transport: %s", opt.transportStr)
	}
}
