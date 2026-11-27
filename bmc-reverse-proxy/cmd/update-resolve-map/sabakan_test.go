package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// startFakeSabakan starts a fake sabakan that responds with status and body.
// It returns a sabakanClient for it. It also checks the request from the client.
func startFakeSabakan(t *testing.T, status int, body string) sabakanClient {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/machines" || r.Header.Get("Accept") != "application/json" {
			t.Errorf("unexpected request: %s %s %v", r.Method, r.URL.Path, r.Header)
		}
		w.WriteHeader(status)
		w.Write([]byte(body)) //nolint:errcheck
	}))
	t.Cleanup(server.Close)
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	return newSabakanClient(u)
}

func TestGetMachines(t *testing.T) {
	sabakan := startFakeSabakan(t, http.StatusOK, `[
  {
    "spec": {
      "serial": "ABC",
      "bmc": {"ipv4": "192.0.2.3"}
    }
  }
]`)

	machines, err := sabakan.getMachines(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(machines) != 1 {
		t.Fatalf("expected 1 machine, got %#v", machines)
	}
	m := machines[0]
	if m.Spec.Serial != "ABC" || m.Spec.BMC.IPv4 != "192.0.2.3" {
		t.Errorf("unexpected machine: %#v", m)
	}
}

// TestGetMachinesNotFound checks that 404 is regarded as no machines.
// sabakan returns 404, not an empty list, if no machines are registered.
func TestGetMachinesNotFound(t *testing.T) {
	sabakan := startFakeSabakan(t, http.StatusNotFound, `{"status":404,"error":"not found"}`)

	machines, err := sabakan.getMachines(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(machines) != 0 {
		t.Errorf("expected no machines, got %#v", machines)
	}
}

func TestGetMachinesError(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		body    string
		wantErr string // substring of the error message
	}{
		{
			name:    "HTTPError",
			status:  http.StatusInternalServerError,
			body:    "internal error\n",
			wantErr: "sabakan returned HTTP 500: internal error",
		},
		{
			name:    "BrokenJSON",
			status:  http.StatusOK,
			body:    `[{"spec":`,
			wantErr: "failed to decode sabakan response",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := startFakeSabakan(t, c.status, c.body).getMachines(t.Context())
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("expected an error containing %q, got %v", c.wantErr, err)
			}
		})
	}
}

// TestGetMachinesLargeErrorBody checks that a large error body is truncated in the error message.
func TestGetMachinesLargeErrorBody(t *testing.T) {
	sabakan := startFakeSabakan(t, http.StatusBadGateway, strings.Repeat("x", 100*1024))

	_, err := sabakan.getMachines(t.Context())
	if err == nil {
		t.Fatal("expected an error")
	}
	if len(err.Error()) > maxErrorBodySize+100 {
		t.Errorf("the error message is too long: %d bytes", len(err.Error()))
	}
}
