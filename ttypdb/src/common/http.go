package common

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type proxyHTTPHandler struct {
	http.Handler
	logger *zap.Logger
}

type proxyHTTPResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewProxyHTTPHandler(orig http.Handler, logger *zap.Logger) http.Handler {
	return &proxyHTTPHandler{
		Handler: orig,
		logger:  logger,
	}
}

func (h *proxyHTTPHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	start := time.Now()

	prw := &proxyHTTPResponseWriter{
		ResponseWriter: rw,
	}
	h.Handler.ServeHTTP(prw, req)

	if prw.statusCode == 0 {
		prw.statusCode = http.StatusOK
	}

	logfn := h.logger.Info
	if prw.statusCode >= 500 {
		logfn = h.logger.Error
	} else if prw.statusCode >= 400 {
		logfn = h.logger.Warn
	}
	logfn("http access",
		zap.String("path", req.URL.Path),
		zap.Int("status", prw.statusCode),
		zap.Float64("duration", time.Since(start).Seconds()))
}

func (rw *proxyHTTPResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
