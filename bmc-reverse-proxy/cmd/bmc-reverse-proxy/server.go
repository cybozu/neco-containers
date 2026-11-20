package main

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

// serve serves HTTP on ln until ctx is canceled, and then shuts down the server gracefully.
// It returns after all handlers have returned.
//
// http.Server.Shutdown does not wait for hijacked connections such as WebSockets
// used by the virtual console. httputil.ReverseProxy does not return from the handler
// until it finishes relaying a hijacked connection, so serve waits for the handlers by itself.
// To do so, serve wraps server.Handler.
func serve(ctx context.Context, server *http.Server, ln net.Listener) error {
	var handlers sync.WaitGroup
	handler := server.Handler
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.Add(1)
		defer handlers.Done()
		handler.ServeHTTP(w, r)
	})

	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		if err := server.Shutdown(context.Background()); err != nil {
			slog.Warn("unclean shutdown", "addr", server.Addr, "error", err)
		}
	}()

	err := server.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) {
		// Serve returns as soon as Shutdown is called. Wait for active requests to finish,
		// and then for the handlers of hijacked connections to return.
		// No handler starts after Shutdown returns, so handlers.Wait does not race with handlers.Add.
		<-shutdownDone
		handlers.Wait()
		return nil
	}
	return err
}

// loggingResponseWriter records the status code and response size for access logs.
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int64
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += int64(n)
	return n, err
}

// Unwrap allows http.ResponseController to access Flush of the underlying writer.
func (w *loggingResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Hijack records the status as 101 Switching Protocols when the connection is hijacked.
// httputil.ReverseProxy hijacks the connection only to switch protocols such as WebSocket,
// and writes the 101 response directly to the connection without calling WriteHeader.
// http.ResponseController uses this method before Unwrap.
func (w *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, brw, err := http.NewResponseController(w.ResponseWriter).Hijack()
	if err == nil {
		w.status = http.StatusSwitchingProtocols
	}
	return conn, brw, err
}

func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		lw := newLoggingResponseWriter(w)
		next.ServeHTTP(lw, r)

		attrs := []slog.Attr{
			slog.Float64("response_time", time.Since(startTime).Seconds()),
			slog.String("protocol", r.Proto),
			slog.Int("http_status_code", lw.status),
			slog.String("http_method", r.Method),
			slog.String("url", r.RequestURI),
			slog.String("http_host", r.Host),
			slog.Int64("request_size", r.ContentLength),
			slog.Int64("response_size", lw.size),
		}
		if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			attrs = append(attrs, slog.String("remote_ipaddr", ip))
		}
		if ua := r.Header.Get("User-Agent"); ua != "" {
			attrs = append(attrs, slog.String("http_user_agent", ua))
		}

		lv := slog.LevelInfo
		switch {
		case lw.status >= 500:
			lv = slog.LevelError
		case lw.status >= 400:
			lv = slog.LevelWarn
		}
		slog.LogAttrs(r.Context(), lv, "access", attrs...)
	})
}
