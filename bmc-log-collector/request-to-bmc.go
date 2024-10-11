package main

import (
	"context"
	"io"
	"net/http"
)

// Get from Redfish API on BMC REST service
func requestToBmc(ctx context.Context, username string, password string, client *http.Client, url string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.SetBasicAuth(username, password)
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return buf, resp.StatusCode, nil
}
