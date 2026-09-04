package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"slices"
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

// requestBmcLog requests one page of a BMC log service and records the
// outcome in the pointer status fields of the log type. The request metrics
// of logType are counted, and a failure is reported only when it differs
// from the recorded one so that a persistent failure does not flood the log.
// The body is returned only for a 200 reply.
//
// A status listed in notImplemented means that the BMC lacks the log
// service: it is reported as a warning and not counted as a failure, to
// avoid a permanent false alarm.
func (c *logCollector) requestBmcLog(ctx context.Context, m Machine, url, logType string, lastHttpStatusCode *int, lastError *string, notImplemented ...int) ([]byte, bool) {
	byteJSON, statusCode, err := requestToBmc(ctx, c.username, c.password, c.httpClient, url)
	if err != nil {
		counterRequestFailed.WithLabelValues(m.Serial, logType).Inc()
		if *lastError != err.Error() {
			slog.Error("failed access to iDRAC on TCP/IP level.", "url", url, "err", err.Error(), "serial", m.Serial)
		}
		*lastHttpStatusCode = 0
		*lastError = err.Error()
		return nil, false
	}
	if slices.Contains(notImplemented, statusCode) {
		if statusCode != *lastHttpStatusCode {
			slog.Warn("the log service is not implemented on this BMC", "url", url, "httpStatusCode", statusCode, "serial", m.Serial, "logType", logType)
		}
		*lastHttpStatusCode = statusCode
		*lastError = ""
		return nil, false
	}
	if statusCode != http.StatusOK {
		counterRequestFailed.WithLabelValues(m.Serial, logType).Inc()
		if statusCode != *lastHttpStatusCode {
			slog.Error("failed access to iDRAC on HTTP level.", "url", url, "httpStatusCode", statusCode, "serial", m.Serial)
		}
		*lastHttpStatusCode = statusCode
		*lastError = ""
		return nil, false
	}
	counterRequestSuccess.WithLabelValues(m.Serial, logType).Inc()
	// Clear the last error status so that the same error after a recovery
	// is logged again instead of being suppressed by the deduplication
	*lastHttpStatusCode = statusCode
	*lastError = ""
	return byteJSON, true
}
