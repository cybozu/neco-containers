package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const graphQLQuery = `
query search {
  searchMachines(having: null, notHaving: null) {
    spec {
      ipv4
      serial
      rack
      indexInRack
      role
      bmc {
        ipv4
      }
    }
  }
}
`

// Machine represents a machine registered with sabakan.
// Only the fields needed to build the resolve map are included.
type Machine struct {
	Spec struct {
		IPv4        []string `json:"ipv4"`
		Serial      string   `json:"serial"`
		Rack        int      `json:"rack"`
		IndexInRack int      `json:"indexInRack"`
		Role        string   `json:"role"`
		BMC         struct {
			IPv4 string `json:"ipv4"`
		} `json:"bmc"`
	} `json:"spec"`
}

func (m Machine) isBootServer() bool {
	return m.Spec.Role == "boot"
}

func (m Machine) hasIPv4() bool {
	return len(m.Spec.IPv4) > 0
}

// queryBody is the request body of graphQLQuery.
// json.Marshal never fails for a struct with a string field, so the error is ignored.
var queryBody, _ = json.Marshal(struct {
	Query string `json:"query"`
}{graphQLQuery})

// maxErrorBodySize is the maximum size of the response body included in an error message,
// so that a large error page, e.g. from an HTTP proxy, does not flood the log.
// It applies only to error responses (non-200), and does not limit successful responses,
// which can be large with many machines.
const maxErrorBodySize = 1024

type sabakanClient struct {
	endpoint string
	http     *http.Client
}

// newSabakanClient creates a sabakanClient for sabakan's GraphQL API at address,
// which must be in the form host:port.
func newSabakanClient(address string) sabakanClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Do not use HTTP proxies because sabakan runs on the boot servers in the same network.
	// See: https://github.com/golang/go/issues/25793
	transport.Proxy = nil

	endpoint := url.URL{Scheme: "http", Host: address, Path: "/graphql"}
	return sabakanClient{
		endpoint: endpoint.String(),
		http: &http.Client{
			Transport: transport,
			// Shorter than the update interval, so that a hung request does not block the next update for long.
			Timeout: 30 * time.Second,
		},
	}
}

func (c sabakanClient) getMachines(ctx context.Context) ([]Machine, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(queryBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create sabakan request: %w", err)
	}
	// gqlgen 0.9+ requires application/json content-type header.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query sabakan: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		return nil, fmt.Errorf("sabakan returned HTTP %d: %s", resp.StatusCode, bytes.TrimSpace(body))
	}

	var result struct {
		Data struct {
			Machines []Machine `json:"searchMachines"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode sabakan response: %w", err)
	}
	if len(result.Errors) > 0 {
		messages := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			messages[i] = e.Message
		}
		return nil, fmt.Errorf("sabakan returned errors: %s", strings.Join(messages, "; "))
	}
	return result.Data.Machines, nil
}
