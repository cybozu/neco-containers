package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	hellov1 "github.com/cybozu/neco-containers/spiffe-test/gen/hello/v1"
)

type Client struct {
	conn     *grpc.ClientConn
	client   hellov1.HelloServiceClient
	jwtToken string
	mu       sync.RWMutex
}

func NewClient(addr string, tlsConfig *tls.Config) (*Client, error) {
	creds := credentials.NewTLS(tlsConfig)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return &Client{
		conn:   conn,
		client: hellov1.NewHelloServiceClient(conn),
	}, nil
}

func (c *Client) SetJWTToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.jwtToken = token
}

func (c *Client) SayHello(ctx context.Context) (string, error) {
	c.mu.RLock()
	token := c.jwtToken
	c.mu.RUnlock()
	if token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	}

	resp, err := c.client.SayHello(ctx, &hellov1.SayHelloRequest{})
	if err != nil {
		return "", fmt.Errorf("SayHello RPC failed: %w", err)
	}

	return resp.Message, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
