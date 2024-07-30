package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type RedfishClient struct {
	user     string
	password string
	client   *http.Client
}

// Get from Redfish on BMC REST service
func (r *RedfishClient) requestToBmc(url string) ([]byte, error) {

	// create Redfish request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("%s: accessing %s", err, url))
		return nil, err
	}
	req.SetBasicAuth(r.user, r.password)

	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// execute access
	resp, err := r.client.Do(req)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return nil, err
	}
	defer resp.Body.Close()

	// check status
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("HTTP status code = %d", resp.StatusCode)
		slog.Error(fmt.Sprintf("%s: accessing %s", err, url))
		return nil, err
	}

	// read body
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return nil, err
	}

	return buf, nil
}
