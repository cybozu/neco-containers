package grpc

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	hellov1 "github.com/cybozu/neco-containers/spiffe-test/gen/hello/v1"
	"github.com/cybozu/neco-containers/spiffe-test/internal/auth"
	"github.com/cybozu/neco-containers/spiffe-test/internal/service"
)

type Handler struct {
	hellov1.UnimplementedHelloServiceServer
	authenticator auth.Authenticator
	service       service.HelloService
}

func NewHandler(authenticator auth.Authenticator, svc service.HelloService) *Handler {
	return &Handler{
		authenticator: authenticator,
		service:       svc,
	}
}

func (h *Handler) SayHello(ctx context.Context, req *hellov1.SayHelloRequest) (*hellov1.SayHelloResponse, error) {
	callerID, err := h.authenticator.GetCallerIDFromContext(ctx)
	if err != nil {
		slog.Warn("Authentication failed", "error", err)
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	slog.Info("gRPC request received", "callerID", callerID.String())

	message, err := h.service.SayHello(ctx, callerID)
	if err != nil {
		slog.Warn("Service error", "error", err, "callerID", callerID.String())
		if errors.Is(err, service.ErrUnauthorized) {
			return nil, status.Errorf(codes.PermissionDenied, "access denied")
		}
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &hellov1.SayHelloResponse{Message: message}, nil
}
