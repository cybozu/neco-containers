package main

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProbeHandler(t *testing.T) {
	p := &proxy{}
	handler := p.probeHandler()

	check := func(method, path string, expected int) {
		t.Helper()
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(method, path, nil))
		if w.Code != expected {
			t.Errorf("%s %s: expected %d, got %d", method, path, expected, w.Code)
		}
	}

	// before the resolve map is loaded
	check(http.MethodGet, "/healthz", http.StatusOK)
	check(http.MethodGet, "/readyz", http.StatusServiceUnavailable)

	// after the resolve map is loaded; an empty map is also regarded as loaded
	p.updateResolveMap(map[string]string{})
	check(http.MethodGet, "/healthz", http.StatusOK)
	check(http.MethodGet, "/readyz", http.StatusOK)

	check(http.MethodPost, "/readyz", http.StatusMethodNotAllowed)
	check(http.MethodGet, "/notfound", http.StatusNotFound)
}

func TestProbeHandlerWithTLS(t *testing.T) {
	cases := []struct {
		name           string
		loadResolveMap bool
		loadCert       bool
		expected       int
	}{
		{name: "NothingLoaded", expected: http.StatusServiceUnavailable},
		{name: "OnlyResolveMap", loadResolveMap: true, expected: http.StatusServiceUnavailable},
		{name: "OnlyCertificate", loadCert: true, expected: http.StatusServiceUnavailable},
		{name: "BothLoaded", loadResolveMap: true, loadCert: true, expected: http.StatusOK},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := &proxy{useTLS: true}
			if c.loadResolveMap {
				p.updateResolveMap(map[string]string{})
			}
			if c.loadCert {
				p.updateCertificate(tls.Certificate{})
			}

			w := httptest.NewRecorder()
			p.probeHandler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))
			if w.Code != c.expected {
				t.Errorf("expected %d, got %d", c.expected, w.Code)
			}
		})
	}
}
