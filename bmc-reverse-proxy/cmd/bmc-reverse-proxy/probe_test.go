package main

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProbeHandler(t *testing.T) {
	cases := []struct {
		name           string
		method         string
		path           string
		loadResolveMap bool
		loadCert       bool
		expected       int
	}{
		{name: "HealthzBeforeLoaded", method: http.MethodGet, path: "/healthz", expected: http.StatusOK},
		{name: "ReadyzNothingLoaded", method: http.MethodGet, path: "/readyz", expected: http.StatusServiceUnavailable},
		{name: "ReadyzOnlyResolveMap", method: http.MethodGet, path: "/readyz", loadResolveMap: true, expected: http.StatusServiceUnavailable},
		{name: "ReadyzOnlyCertificate", method: http.MethodGet, path: "/readyz", loadCert: true, expected: http.StatusServiceUnavailable},
		// An empty resolve map is also regarded as loaded.
		{name: "ReadyzBothLoaded", method: http.MethodGet, path: "/readyz", loadResolveMap: true, loadCert: true, expected: http.StatusOK},
		{name: "MethodNotAllowed", method: http.MethodPost, path: "/readyz", expected: http.StatusMethodNotAllowed},
		{name: "NotFound", method: http.MethodGet, path: "/notfound", expected: http.StatusNotFound},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := &proxy{}
			if c.loadResolveMap {
				p.updateResolveMap(map[string]string{})
			}
			if c.loadCert {
				p.updateCertificate(tls.Certificate{})
			}

			w := httptest.NewRecorder()
			p.probeHandler().ServeHTTP(w, httptest.NewRequest(c.method, c.path, nil))
			if w.Code != c.expected {
				t.Errorf("%s %s: expected %d, got %d", c.method, c.path, c.expected, w.Code)
			}
		})
	}
}
