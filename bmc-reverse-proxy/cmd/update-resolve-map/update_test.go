package main

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cybozu/neco-containers/bmc-reverse-proxy/internal/resolvemap"
)

func TestUpdate(t *testing.T) {
	body := `{"data":{"searchMachines":[{"spec":{"ipv4":["198.51.100.196"],"serial":"CS1","rack":1,"indexInRack":4,"role":"cs","bmc":{"ipv4":"192.0.2.4"}}}]}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body)) //nolint:errcheck
	}))
	defer server.Close()
	sabakan := newSabakanClient(strings.TrimPrefix(server.URL, "http://"))

	output := filepath.Join(t.TempDir(), "resolve-map.json")
	if err := update(t.Context(), sabakan, output); err != nil {
		t.Fatal(err)
	}

	actual, err := resolvemap.Read(output)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{
		"rack1-cs4":      "192.0.2.4",
		"198-51-100-196": "192.0.2.4",
		"CS1":            "192.0.2.4",
	}
	if !maps.Equal(actual, expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestUpdateKeepsFileOnError(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{
			name:   "SabakanError",
			status: http.StatusInternalServerError,
		},
		{
			name:   "NoMachines",
			status: http.StatusOK,
			body:   `{"data":{"searchMachines":[]}}`,
		},
		{
			name:   "NoBMCAddresses",
			status: http.StatusOK,
			body:   `{"data":{"searchMachines":[{"spec":{"ipv4":["198.51.100.196"],"serial":"CS1","rack":1,"indexInRack":4,"role":"cs","bmc":{"ipv4":""}}}]}}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(c.status)
				w.Write([]byte(c.body)) //nolint:errcheck
			}))
			defer server.Close()
			sabakan := newSabakanClient(strings.TrimPrefix(server.URL, "http://"))

			// The existing file must be kept on errors.
			output := filepath.Join(t.TempDir(), "resolve-map.json")
			old := map[string]string{"foo": "203.0.113.1"}
			if err := resolvemap.Write(output, old); err != nil {
				t.Fatal(err)
			}
			if err := update(t.Context(), sabakan, output); err == nil {
				t.Fatal("expected an error")
			}
			actual, err := resolvemap.Read(output)
			if err != nil {
				t.Fatal(err)
			}
			if !maps.Equal(actual, old) {
				t.Errorf("the file was changed: %v", actual)
			}
		})
	}
}

func newMachine(serial, role string, rack, index int, nodeIPv4, bmcIPv4 string) Machine {
	var m Machine
	m.Spec.Serial = serial
	m.Spec.Role = role
	m.Spec.Rack = rack
	m.Spec.IndexInRack = index
	if nodeIPv4 != "" {
		m.Spec.IPv4 = []string{nodeIPv4, "198.51.100.254"}
	}
	m.Spec.BMC.IPv4 = bmcIPv4
	return m
}

func TestResolveMapData(t *testing.T) {
	machines := []Machine{
		newMachine("BOOT0", "boot", 0, 3, "198.51.100.3", "192.0.2.3"),
		newMachine("CS1", "cs", 1, 4, "198.51.100.196", "192.0.2.4"),
		// sabakan returns them as is, but check the conversion of the case anyway
		newMachine("ss2", "SS", 2, 5, "198.51.100.105", "192.0.2.5"),
		// skipped: no BMC address
		newMachine("NOBMC", "cs", 2, 5, "198.51.100.105", ""),
		// skipped: no node address
		newMachine("NONODE", "cs", 2, 6, "", "192.0.2.6"),
	}

	expected := map[string]string{
		"boot-0":         "192.0.2.3",
		"198-51-100-3":   "192.0.2.3",
		"BOOT0":          "192.0.2.3",
		"rack1-cs4":      "192.0.2.4",
		"198-51-100-196": "192.0.2.4",
		"CS1":            "192.0.2.4",
		"rack2-ss5":      "192.0.2.5",
		"198-51-100-105": "192.0.2.5",
		"SS2":            "192.0.2.5",
	}

	actual := resolveMapData(machines)
	if !maps.Equal(actual, expected) {
		t.Errorf("unexpected resolve map:\nexpected: %v\nactual:   %v", expected, actual)
	}
}
