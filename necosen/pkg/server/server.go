package server

import (
	"context"
	"crypto/tls"
	"net"
	"net/netip"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/code"
	status "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Listener     net.Listener
	Reflection   bool
	TLSCertFile  string
	TLSKeyFile   string
	AllowedCIDRs []netip.Prefix
}

type NecosenServer interface {
	authv3.AuthorizationServer
	Start() error
}

type authServer struct {
	config Config
	logger *zap.SugaredLogger
}

var _ authv3.AuthorizationServer = &authServer{}

func NewAuthorizationServer(c Config, l *zap.Logger) NecosenServer {
	return &authServer{
		config: c,
		logger: l.Sugar(),
	}
}

func (a *authServer) Start() error {
	var opts []grpc.ServerOption
	if a.config.TLSKeyFile != "" {
		a.logger.Info("gRPC uses TLS")
		cert, err := tls.LoadX509KeyPair(a.config.TLSCertFile, a.config.TLSKeyFile)
		if err != nil {
			return err
		}
		opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(&cert)))
	}

	a.logger.Info("Start authorization server")
	s := grpc.NewServer(opts...)
	if a.config.Reflection {
		reflection.Register(s)
	}
	authv3.RegisterAuthorizationServer(s, a)
	return s.Serve(a.config.Listener)
}

func (a *authServer) Check(ctx context.Context, req *authv3.CheckRequest) (*authv3.CheckResponse, error) {
	a.logger.Debug("Check incoming request")

	resp, err := a.CheckCIDR(req)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return resp, nil
	}

	// Additional checks can be written here

	return &authv3.CheckResponse{
		HttpResponse: &authv3.CheckResponse_OkResponse{},
	}, nil
}

func (a *authServer) CheckCIDR(req *authv3.CheckRequest) (*authv3.CheckResponse, error) {
	a.logger.Debug("Check CIDR")

	if len(a.config.AllowedCIDRs) == 0 {
		// Default allow
		return nil, nil
	}

	sa := req.Attributes.Source.Address.GetSocketAddress()
	if sa == nil {
		a.logger.Debugw("Incoming request does not have a socket address", "address", req.Attributes.Source.Address.String())
		return &authv3.CheckResponse{
			Status: &status.Status{
				Code:    int32(code.Code_PERMISSION_DENIED),
				Message: "Invalid Request",
			},
		}, nil
	}

	clientIP, err := netip.ParseAddr(sa.Address)
	if err != nil {
		a.logger.Debugw("Failed to parse address", "address", sa.Address, "error", err)
		return &authv3.CheckResponse{
			Status: &status.Status{
				Code:    int32(code.Code_PERMISSION_DENIED),
				Message: "Invalid Request",
			},
		}, nil
	}

	for _, cidr := range a.config.AllowedCIDRs {
		if cidr.Contains(clientIP) {
			a.logger.Debugw("Client meets the CIDR requirements", "address", clientIP.String())
			return nil, nil
		}
	}

	a.logger.Debugw("Client does not meet the CIDR requirements", "address", clientIP.String())
	return &authv3.CheckResponse{
		Status: &status.Status{
			Code:    int32(code.Code_PERMISSION_DENIED),
			Message: "Invalid Request",
		},
	}, nil
}
