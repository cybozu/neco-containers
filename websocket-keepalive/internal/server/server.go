package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/cybozu/neco-containers/websocket-keepalive/internal/metrics"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type connDB struct {
	sync.Mutex
	db map[string]*conn
}

var connections = connDB{
	Mutex: sync.Mutex{},
	db:    make(map[string]*conn),
}

func (db *connDB) register(key string, c *conn) error {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.db[key]; ok {
		return fmt.Errorf("%s is already registered", key)
	}
	db.db[key] = c
	return nil
}

func (db *connDB) unregister(key string) error {
	db.Lock()
	defer db.Unlock()
	if _, ok := db.db[key]; !ok {
		return fmt.Errorf("%s is not registered", key)
	}
	delete(db.db, key)

	return nil
}

func (db *connDB) isEmpty() bool {
	db.Lock()
	defer db.Unlock()
	return len(db.db) == 0
}

type Config struct {
	Host         string
	Port         int
	PingInterval time.Duration
}

func RunWithConfig(ctx context.Context, config *Config, metricsConfig *metrics.Config) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(ctx, config, w, r)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler: mux,
	}

	shutdownTimeout := 30 * time.Second

	m, err := NewMetrics(metricsConfig)
	if err != nil {
		return err
	}
	if metricsConfig.Export {
		slog.Info("Start metrics server", "listen", m.AddrPort)
		go serveMetrics(ctx, m.Metrics)
	}

	// Bind synchronously so port conflicts fail fast instead of silently zombieing.
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", server.Addr, err)
	}

	serveErrCh := make(chan error, 1)
	go func() {
		slog.Info("WebSocket server starting", "addr", server.Addr)
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			serveErrCh <- err
			return
		}
		close(serveErrCh)
	}()

	select {
	case <-ctx.Done():
		slog.Info("Shutting down server...")
	case err := <-serveErrCh:
		return fmt.Errorf("websocket server failed: %w", err)
	}

	// Each per-connection handler selects on ctx.Done() and starts its own active close
	// sequence, so shutdown just waits for them to unregister themselves.
	deadline := time.Now().Add(shutdownTimeout)
	for !connections.isEmpty() {
		if time.Now().After(deadline) {
			slog.Error("shutdown timeout is exceeded")
			connections.Lock()
			for remote, c := range connections.db {
				slog.Error("remaining connection is reset forcibly", "remote_addr", remote)
				c.Close()
			}
			connections.Unlock()
			break
		}
		time.Sleep(time.Second)
	}

	slog.Info("all connections are closed")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return nil
}

func serveMetrics(ctx context.Context, m *metrics.Metrics) {
	const maxBackoff = 30 * time.Second
	backoff := time.Second
	for {
		err := m.Serve(ctx)
		if ctx.Err() != nil {
			return
		}
		if err == nil {
			return
		}
		slog.Error("serving metrics failed", "error", err, "retry_in", backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < maxBackoff {
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

func handleWebSocket(ctx context.Context, config *Config, w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade connection", "error", err)
		return
	}
	// Defers run in LIFO order: Close() fires first to unblock the reader goroutine
	// (and any pending Writes), then unregister removes the entry from the DB.
	defer func() {
		if err := connections.unregister(r.RemoteAddr); err != nil {
			slog.Error("failed to unregister connection", "error", err, "remote_addr", r.RemoteAddr)
		}
	}()
	defer c.Close()

	wsConn := newConn(c, r.RemoteAddr)
	if err := connections.register(r.RemoteAddr, wsConn); err != nil {
		slog.Error("failed to register connection", "error", err, "remote_addr", r.RemoteAddr)
		return
	}

	if err := wsConn.handleWebsocketConnection(ctx, config); err != nil {
		slog.Error("failed to handle websocket connection", "error", err, "remote_addr", r.RemoteAddr)
	}
}

type conn struct {
	*websocket.Conn
	remoteAddr string
}

func newConn(c *websocket.Conn, remoteAddr string) *conn {
	return &conn{
		Conn:       c,
		remoteAddr: remoteAddr,
	}
}

func (c *conn) handleWebsocketConnection(ctx context.Context, config *Config) error {

	slog.Info("start to handle new connection", "remote_addr", c.remoteAddr)

	closeTimeout := time.Second * 10
	pongWait := 2 * config.PingInterval

	if err := c.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return fmt.Errorf("failed to set initial read deadline: %w", err)
	}

	m := initServerMetrics(c.LocalAddr().String(), c.remoteAddr)
	m.setEstablished()
	defer m.setUnestablished()

	c.SetPingHandler(func(appData string) error {
		if err := c.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			slog.Warn("failed to extend read deadline on ping", "error", err, "remote_addr", c.remoteAddr)
		}
		m.incrementPingTotal()
		// safe: gorilla serializes WriteControl internally.
		// Match gorilla's default PingHandler policy: swallow write errors (including
		// ErrCloseSent after we've sent a close frame, and transient net errors) and
		// let the next ReadMessage surface real connection failures instead.
		err := c.WriteControl(websocket.PongMessage, []byte("from server"), time.Now().Add(5*time.Second))
		if err == nil {
			m.incrementPongTotal()
			return nil
		}
		if err == websocket.ErrCloseSent {
			return nil
		}
		slog.Warn("failed to send pong", "error", err, "remote_addr", c.remoteAddr)
		return nil
	})

	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					slog.Info("Connection will be closed normally", "error", err, "remote_addr", c.remoteAddr)
					return
				}
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					slog.Error("Got websocket abnormal closure", "error", err, "remote_addr", c.remoteAddr)
					return
				}
				slog.Error("Websocket error", "error", err, "messageType", messageType, "message", string(message), "remote_addr", c.remoteAddr)
				return
			}
		}
	}()

	select {
	case <-readerDone:
		slog.Info("connection closed.", "remote_addr", c.remoteAddr)
	case <-ctx.Done():
		slog.Info("shutdown signal received, closing.", "remote_addr", c.remoteAddr)
		closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "from server")
		// safe: gorilla serializes WriteControl internally.
		if err := c.WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(5*time.Second)); err != nil {
			slog.Error("failed to write close message", "error", err, "remote_addr", c.remoteAddr)
		} else {
			select {
			case <-readerDone:
			case <-time.After(closeTimeout):
			}
		}
	}
	return nil
}
