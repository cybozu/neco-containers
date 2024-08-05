package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Request of SEL in Redfish API
type redfishSelRequest struct {
	username string
	password string
	client   *http.Client
	url      string
}

// Get from Redfish API on BMC REST service
// func requestToBmc(ctx context.Context, url string, r RedfishClient) ([]byte, error) {
func requestToBmc(ctx context.Context, r redfishSelRequest) ([]byte, error) {
	// create Redfish request
	req, err := http.NewRequest("GET", r.url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(r.username, r.password)
	req = req.WithContext(ctx)
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("HTTP status code = %d", resp.StatusCode)
		return nil, err
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
