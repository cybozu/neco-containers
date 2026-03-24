package grpc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	hellov1 "github.com/cybozu/neco-containers/spiffe-test/gen/hello/v1"
	"github.com/cybozu/neco-containers/spiffe-test/internal/auth"
	"github.com/cybozu/neco-containers/spiffe-test/internal/service"
)

type Server struct {
	addr       string
	grpcServer *grpc.Server
}

func NewServer(addr string, tlsConfig *tls.Config, authenticator auth.Authenticator, helloService service.HelloService) *Server {
	var opts []grpc.ServerOption
	if tlsConfig != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}
	grpcServer := grpc.NewServer(opts...)

	handler := NewHandler(authenticator, helloService)
	hellov1.RegisterHelloServiceServer(grpcServer, handler)

	return &Server{
		addr:       addr,
		grpcServer: grpcServer,
	}
}

func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	slog.Info("Starting gRPC server", "addr", s.addr)

	errCh := make(chan error, 1)
	go func() {
		if err := s.grpcServer.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		return s.Shutdown()
	case err := <-errCh:
		return err
	}
}

func (s *Server) Shutdown() error {
	slog.Info("Shutting down gRPC server")
	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		slog.Warn("Graceful stop timed out, forcing stop")
		s.grpcServer.Stop()
	}
	return nil
}
