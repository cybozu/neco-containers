package transport

import (
	"context"
	"strings"
)

type Transport string

const (
	HTTP Transport = "http"
	GRPC Transport = "grpc"
)

func Parse(s string) Transport {
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
