package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"time"

	sabakan "github.com/cybozu-go/sabakan/v3"
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
    status {
      state
    }
  }
}
`

// Machine represents a machine registered with sabakan.
type Machine struct {
	Spec struct {
		IPv4        []string `json:"ipv4"`
		Serial      string   `json:"serial"`
		Rack        int      `json:"rack"`
		IndexInRack int      `json:"indexInRack"`
		Role        string   `json:"role"`
		BMC         struct {
			IPv4 string `json:"ipv4"`
		}
	}
	Status struct {
		State string `json:"state"`
	}
}

func (m Machine) isRetired() bool {
	return m.Status.State == sabakan.StateRetired.GQLEnum()
}

func (m Machine) isBootServer() bool {
	return m.Spec.Role == "boot"
}

func (m Machine) hasIPv4() bool {
	return len(m.Spec.IPv4) > 0
}

type sabakanClient struct {
	address string
	http    *http.Client
}

// newSabakanClient creates a sabakanClient for sabakan's GraphQL API at address,
// which must be in the form host:port.
func newSabakanClient(address string) sabakanClient {
	return sabakanClient{
		address: address,
		http: &http.Client{
			// Most of the following values are copied from http.DefaultTransport to workaround a proxy issue.
			// See: https://github.com/golang/go/issues/25793
			Transport: &http.Transport{
				Proxy: nil,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Timeout: 10 * time.Minute,
		},
	}
}

func (c sabakanClient) getMachines(ctx context.Context) ([]Machine, error) {
	endpoint := "http://" + c.address + "/graphql"

	data, err := json.Marshal(struct {
		Query string `json:"query"`
	}{graphQLQuery})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sabakan query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create sabakan request: %w", err)
	}
	// gqlgen 0.9+ requires application/json content-type header.
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query sabakan: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sabakan returned HTTP %d: %s", resp.StatusCode, body)
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
		return nil, fmt.Errorf("sabakan returned error: %v", result.Errors)
	}

	machines := result.Data.Machines
	// sort machines not to generate almost-identical-but-differ-in-order objects, which cause frequent updates
	slices.SortFunc(machines, func(a, b Machine) int { return strings.Compare(a.Spec.Serial, b.Spec.Serial) })
	return machines, nil
}
