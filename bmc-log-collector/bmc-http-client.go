package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	//"time"
)

type RedfishClient struct {
	user     string
	password string
	//transport	*http.Transport
	client *http.Client
}

// Get from Redfish on iDRAC webserver
func (r *RedfishClient) requestToBmc(url string) ([]byte, error) {

	// Create Request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("%s: accessing %s", err, url))
		return nil, err
	}
	req.SetBasicAuth(r.user, r.password)

	//
	//client := &http.Client{
	//	Timeout:   time.Duration(10) * time.Second,
	//	Transport: r.transport,
	//}

	resp, err := r.client.Do(req)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("HTTP status code = %d", resp.StatusCode)
		slog.Error(fmt.Sprintf("%s: accessing %s", err, url))
		return nil, err
	}

	// response
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(fmt.Sprintf("%s", err))
		return nil, err
	}

	//client.CloseIdleConnections()
	return buf, nil
}
