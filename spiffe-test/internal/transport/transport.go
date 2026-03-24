package transport

import (
	"context"
	"strings"
)

type Kind string

const (
	HTTP Kind = "http"
	GRPC Kind = "grpc"
)

func Parse(s string) Kind {
	switch strings.ToLower(s) {
	case string(HTTP):
		return HTTP
	case string(GRPC):
		return GRPC
	default:
		return ""
	}
}

type HelloClient interface {
	SayHello(ctx context.Context) (string, error)
	SetJWTToken(token string)
	Close() error
}
