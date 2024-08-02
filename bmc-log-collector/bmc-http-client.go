package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type RedfishClient struct {
	user     string
	password string
	client   *http.Client
}

// Get from Redfish on BMC REST service
// func (r *RedfishClient) requestToBmc(ctx context.Context, url string) ([]byte, error) {
func requestToBmc(ctx context.Context, url string, r RedfishClient) ([]byte, error) {
	// create Redfish request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("http.NewRequest()", "err", err, "URL", url)
		return nil, err
	}
	req.SetBasicAuth(r.user, r.password)
	req = req.WithContext(ctx)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("HTTP status code = %d", resp.StatusCode)
		slog.Error("resp.StatusCode", "err", err)
		return nil, err
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("io.ReadAll()", "err", err)
		return nil, err
	}

	return buf, nil
}
