package main

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type proxyHTTPHandler struct {
	orig   http.Handler
	logger *zap.Logger
}

type proxyHTTPResponseWriter struct {
	orig       http.ResponseWriter
	statusCode int
}

func NewProxyHTTPHandler(orig http.Handler, logger *zap.Logger) http.Handler {
	return &proxyHTTPHandler{
		orig:   orig,
		logger: logger,
	}
}

func (h *proxyHTTPHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	start := time.Now()

	prw := &proxyHTTPResponseWriter{
		orig: rw,
	}
	h.orig.ServeHTTP(prw, req)

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

func (rw *proxyHTTPResponseWriter) Header() http.Header {
	return rw.orig.Header()
}

func (rw *proxyHTTPResponseWriter) Write(data []byte) (int, error) {
	return rw.orig.Write(data)
}

func (rw *proxyHTTPResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.orig.WriteHeader(statusCode)
}
