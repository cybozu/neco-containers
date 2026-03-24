package http

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	defaultClientTimeout = 30 * time.Second
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	jwtToken   string
	mu         sync.RWMutex
}

func NewClient(baseURL string, tlsConfig *tls.Config) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultClientTimeout,
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		},
	}
}

func (c *Client) SetJWTToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.jwtToken = token
}

func (c *Client) SayHello(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/hello", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	token := c.jwtToken
	c.mu.RUnlock()
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	var helloResp HelloResponse
	if err := json.NewDecoder(resp.Body).Decode(&helloResp); err != nil {
		return "", fmt.Errorf("failed to parse response (status %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, helloResp.Error)
	}

	return helloResp.Message, nil
}

func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
