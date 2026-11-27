package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Machine represents a machine registered with sabakan.
// It has only the fields needed for the resolve map.
// See GET /api/v1/machines in https://github.com/cybozu-go/sabakan/blob/main/docs/api.md.
type Machine struct {
	Spec struct {
		Serial string `json:"serial"`
		BMC    struct {
			IPv4 string `json:"ipv4"`
		} `json:"bmc"`
	} `json:"spec"`
}

// maxErrorBodySize limits the size of the response body in an error message.
// It prevents a large error page, e.g. from an HTTP proxy, from flooding the log.
// It applies only to error responses (non-200).
// Successful responses can be large with many machines.
const maxErrorBodySize = 1024

type sabakanClient struct {
	endpoint string
	http     *http.Client
}

// newSabakanClient creates a sabakanClient for sabakan's REST API.
// baseURL is the URL of the REST API, e.g. http://localhost:10080/.
func newSabakanClient(baseURL *url.URL) sabakanClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Do not use HTTP proxies. sabakan runs on the boot servers in the same network.
	// See: https://github.com/golang/go/issues/25793
	transport.Proxy = nil

	endpoint := baseURL.JoinPath("api/v1/machines")
	return sabakanClient{
		endpoint: endpoint.String(),
		http: &http.Client{
			Transport: transport,
			// Shorter than the default update interval (1m).
			// A hung request does not block the next update for long.
			Timeout: 30 * time.Second,
		},
	}
}

// getMachines gets all the machines registered with sabakan.
func (c sabakanClient) getMachines(ctx context.Context) ([]Machine, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create sabakan request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query sabakan: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		// sabakan returns 404, not an empty list, if no machines are registered.
		return nil, nil
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		return nil, fmt.Errorf("sabakan returned HTTP %d: %s", resp.StatusCode, bytes.TrimSpace(body))
	}

	var machines []Machine
	if err := json.NewDecoder(resp.Body).Decode(&machines); err != nil {
		return nil, fmt.Errorf("failed to decode sabakan response: %w", err)
	}
	return machines, nil
}
