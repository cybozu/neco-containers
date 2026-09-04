package main

import (
	"context"
	"io"
	"log/slog"
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

// requestBmcLog requests one page of a BMC log service.
func (c *logCollector) requestBmcLog(ctx context.Context, m Machine, url, logType string) ([]byte, int, error) {
	byteJSON, statusCode, err := requestToBmc(ctx, c.username, c.password, c.httpClient, url)
	if err != nil {
		counterRequestFailed.WithLabelValues(m.Serial, logType).Inc()
		slog.Error("failed access to iDRAC on TCP/IP level.", "url", url, "err", err, "serial", m.Serial)
		return nil, 0, err
	}

	switch statusCode {
	case http.StatusOK:
		counterRequestSuccess.WithLabelValues(m.Serial, logType).Inc()
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		slog.Warn("the log service is not implemented on this BMC", "url", url, "httpStatusCode", statusCode, "serial", m.Serial, "logType", logType)
	default:
		counterRequestFailed.WithLabelValues(m.Serial, logType).Inc()
		slog.Error("failed access to iDRAC on HTTP level.", "url", url, "httpStatusCode", statusCode, "serial", m.Serial, "logType", logType)
	}

	return byteJSON, statusCode, nil
}
