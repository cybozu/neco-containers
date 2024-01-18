package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMonitorReady(t *testing.T) {
	t.Parallel()

	status := http.StatusOK
	var sleepTime time.Duration

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		time.Sleep(sleepTime)
		res.WriteHeader(status)
	}))
	defer server.Close()

	m := monitor{
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
		readyURL: server.URL + "/ready",
	}

	err := m.monitorReady(context.Background())
	if err != nil {
		t.Error("monitorReady returned non-nil when readyURL returns OK", err)
	}

	status = http.StatusInternalServerError
	err = m.monitorReady(context.Background())
	if err == nil {
		t.Error("monitorReady returned nil when readyURL returns NG", err)
	}

	status = http.StatusOK
	sleepTime = time.Second * 2
	err = m.monitorReady(context.Background())
	if err == nil {
		t.Error("monitorReady returned nil when readyURL timed out", err)
	}

	server.Close()

	err = m.monitorReady(context.Background())
	if err == nil {
		t.Error("monitorReady returned nil when readyURL is not reachable")
	}
}

func TestMonitorHTTP(t *testing.T) {
	t.Parallel()

	status := http.StatusOK
	var sleepTime time.Duration

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		time.Sleep(sleepTime)
		res.WriteHeader(status)
	}))
	defer server.Close()

	m := monitor{
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
		httpURL: server.URL + "/",
	}

	err := m.monitorHTTP(context.Background())
	if err != nil {
		t.Error("monitorHTTP returned non-nil when httpURL returns OK", err)
	}

	status = http.StatusInternalServerError
	err = m.monitorHTTP(context.Background())
	if err != nil {
		t.Error("monitorHTTP returned non-nil when httpURL returns NG", err)
	}

	status = http.StatusOK
	sleepTime = time.Second * 2
	err = m.monitorHTTP(context.Background())
	if err == nil {
		t.Error("monitorHTTP returned nil when httpURL timed out", err)
	}

	server.Close()

	err = m.monitorHTTP(context.Background())
	if err == nil {
		t.Error("monitorHTTP returned nil when httpURL is not reachable and http request has succeeded before")
	}

	m2 := monitor{
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
		httpURL: m.httpURL,
	}

	err = m2.monitorHTTP(context.Background())
	if err != nil {
		t.Error("monitorHTTP returned non-nil when httpURL is not reachable and http request has not succeeded before")
	}
}

func TestMonitorHTTPS(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
	}))
	defer server.Close()

	m := monitor{
		httpsAddr: server.Listener.Addr().String(),
	}

	err := m.monitorHTTPS(context.Background())
	if err != nil {
		t.Error("monitorHTTPS returned non-nil when httpsAddr is reachable", err)
	}

	server.Close()

	err = m.monitorHTTPS(context.Background())
	if err == nil {
		t.Error("monitorHTTPS returned nil when httpsAddr is not reachable and connection has succeeded before")
	}

	m2 := monitor{
		httpsAddr: m.httpsAddr,
	}

	err = m2.monitorHTTPS(context.Background())
	if err != nil {
		t.Error("monitorHTTPS returned non-nil when httpsAddr is not reachable and connection has not succeeded before")
	}
}
