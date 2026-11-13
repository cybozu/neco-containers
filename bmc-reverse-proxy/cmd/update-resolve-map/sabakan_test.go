package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetMachines(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		body     string
		expected []string // serials
		wantErr  string   // substring of the error message
	}{
		{
			name:     "Success",
			status:   http.StatusOK,
			body:     `{"data":{"searchMachines":[{"spec":{"ipv4":["198.51.100.3"],"serial":"ABC","rack":1,"indexInRack":3,"role":"cs","bmc":{"ipv4":"192.0.2.3"}}}]}}`,
			expected: []string{"ABC"},
		},
		{
			name:    "GraphQLError",
			status:  http.StatusOK,
			body:    `{"errors":[{"message":"something wrong"},{"message":"another error"}]}`,
			wantErr: "sabakan returned errors: something wrong; another error",
		},
		{
			name:    "HTTPError",
			status:  http.StatusInternalServerError,
			body:    "internal error\n",
			wantErr: "sabakan returned HTTP 500: internal error",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/graphql" ||
					r.Header.Get("Content-Type") != "application/json" || r.Header.Get("Accept") != "application/json" {
					t.Errorf("unexpected request: %s %s %v", r.Method, r.URL.Path, r.Header)
				}
				w.WriteHeader(c.status)
				w.Write([]byte(c.body)) //nolint:errcheck
			}))
			defer server.Close()

			address := strings.TrimPrefix(server.URL, "http://")
			machines, err := newSabakanClient(address).getMachines(t.Context())
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Errorf("expected an error containing %q, got %v", c.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if len(machines) != len(c.expected) {
				t.Fatalf("unexpected machines: %#v", machines)
			}
			for i, m := range machines {
				if m.Spec.Serial != c.expected[i] {
					t.Errorf("unexpected serial: %s", m.Spec.Serial)
				}
			}
			m := machines[0]
			if m.Spec.BMC.IPv4 != "192.0.2.3" || m.Spec.Rack != 1 || m.Spec.IndexInRack != 3 || m.Spec.Role != "cs" || m.Spec.IPv4[0] != "198.51.100.3" {
				t.Errorf("unexpected machine: %#v", m)
			}
		})
	}
}

// TestGetMachinesLargeErrorBody checks that a large error body is truncated in the error message.
func TestGetMachinesLargeErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(strings.Repeat("x", 100*1024))) //nolint:errcheck
	}))
	defer server.Close()

	_, err := newSabakanClient(strings.TrimPrefix(server.URL, "http://")).getMachines(t.Context())
	if err == nil {
		t.Fatal("expected an error")
	}
	if len(err.Error()) > maxErrorBodySize+100 {
		t.Errorf("the error message is too long: %d bytes", len(err.Error()))
	}
}
