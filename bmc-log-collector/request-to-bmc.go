package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"slices"
)

// errBMCRequestFailed is returned by requestBmcLog after the failure was counted,
// recorded and reported.
var errBMCRequestFailed = errors.New("request to the BMC failed")

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

// requestBmcLog requests the entries of a BMC log service and counts the request
// metrics of logType. A failure is recorded in lastHttpStatusCode/lastError and
// reported only when it differs from the recorded one; the caller clears them on
// success. A status in notImplemented (the BMC lacks the log service) is a
// warning, not a failure.
func (c *logCollector) requestBmcLog(ctx context.Context, m Machine, url, logType string, lastHttpStatusCode *int, lastError *string, notImplemented ...int) ([]byte, error) {
	byteJSON, statusCode, err := requestToBmc(ctx, c.username, c.password, c.httpClient, url)
	if err != nil {
		counterRequestFailed.WithLabelValues(m.Serial, logType).Inc()
		if *lastError != err.Error() {
			slog.Error("failed access to iDRAC on TCP/IP level.", "url", url, "err", err.Error(), "serial", m.Serial)
		}
		*lastHttpStatusCode = 0
		*lastError = err.Error()
		return nil, errBMCRequestFailed
	}
	if slices.Contains(notImplemented, statusCode) {
		if statusCode != *lastHttpStatusCode {
			slog.Warn("the log service is not implemented on this BMC", "url", url, "httpStatusCode", statusCode, "serial", m.Serial, "logType", logType)
		}
		*lastHttpStatusCode = statusCode
		*lastError = ""
		return nil, errBMCRequestFailed
	}
	if statusCode != http.StatusOK {
		counterRequestFailed.WithLabelValues(m.Serial, logType).Inc()
		if statusCode != *lastHttpStatusCode {
			slog.Error("failed access to iDRAC on HTTP level.", "url", url, "httpStatusCode", statusCode, "serial", m.Serial)
		}
		*lastHttpStatusCode = statusCode
		*lastError = ""
		return nil, errBMCRequestFailed
	}
	counterRequestSuccess.WithLabelValues(m.Serial, logType).Inc()
	return byteJSON, nil
}
