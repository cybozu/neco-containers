package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cybozu/neco-containers/websocket-keepalive/internal/metrics"
	"github.com/gorilla/websocket"
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
	Host string
	Port int
}

func Run(host string, port int) error {
	config := &Config{
		Host: host,
		Port: port,
	}

	metricsConfig := &metrics.Config{
		Export:   true,
		AddrPort: "0.0.0.0:8081",
	}

	return RunWithConfig(config, metricsConfig)
}

func RunWithConfig(config *Config, metricsConfig *metrics.Config) error {

	http.HandleFunc("/ws", handleWebSocket)

	server := &http.Server{
		Addr: fmt.Sprintf("%s:%d", config.Host, config.Port),
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	shutdownTimeout := time.Second * 30

	m, err := NewMetrics(metricsConfig)
	if err != nil {
		return err
	}
	if metricsConfig.Export {
		slog.Info("Start metrics server", "listen", m.AddrPort)
		go func() {
			if err := m.Metrics.Serve(); err != nil {
				slog.Error("metrics server failed", "error", err)
				// When metrics server doesn't run correctly, program shouldn't do any more. Just panic here.
				panic(err)
			}
		}()
	}

	go func() {
		slog.Info("WebSocket server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
		}
		slog.Info("WebSocket server stopping", "addr", server.Addr)
	}()

	<-stop
	slog.Info("Shutting down server...")

	connections.Lock()
	for remote, c := range connections.db {
		slog.Debug("Notifying the closing signal", "connection", remote)
		c.serverCloseCh <- struct{}{}
	}
	connections.Unlock()

	now := time.Now()
	for !connections.isEmpty() {
		if time.Since(now) > shutdownTimeout {
			slog.Error("shutdown timeout is exceeded")
			connections.Lock()
			for remote, c := range connections.db {
				slog.Error("remaining connection is reset forcely", "remote_addr", remote)
				c.Close()
			}
			connections.Unlock()
			break
		}
		time.Sleep(time.Second)
	}

	slog.Info("all connections are closed")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade connection", "error", err)
		return
	}

	conn := newConn(c, r.RemoteAddr)

	connections.register(r.RemoteAddr, conn)

	if err := conn.handleWebsocketConnection(context.Background()); err != nil {
		slog.Error("faile to handle websocket connectioin", "remote_addr", r.RemoteAddr)
	}

}

type conn struct {
	*websocket.Conn
	remoteAddr    string
	serverCloseCh chan struct{}
}

func newConn(c *websocket.Conn, remoteAddr string) *conn {
	return &conn{
		Conn:          c,
		remoteAddr:    remoteAddr,
		serverCloseCh: make(chan struct{}),
	}
}

func (c *conn) handleWebsocketConnection(ctx context.Context) error {

	slog.Info("start to handle new connection", "remote_addr", c.remoteAddr)

	done := make(chan struct{})
	closeTimeout := time.Second * 10

	m := initServerMetrics(c.LocalAddr().String(), c.remoteAddr)
	m.setEstablished()

	c.SetPingHandler(func(appData string) error {
		m.incrementPingTotal()
		return func() error {
			if err := c.WriteControl(websocket.PongMessage, []byte("from server"), time.Now().Add(5*time.Second)); err != nil {
				return err
			}
			m.incrementPongTotal()
			return nil
		}()
	})

	defer func() error {
		c.Close()
		return connections.unregister(c.remoteAddr)
	}()

	go func() {
		defer close(done)
		// Message reading loop
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
	case <-done:
		slog.Info("connection closed.", "remote_addr", c.remoteAddr)
	case <-c.serverCloseCh:
		slog.Info("shutdown signal received, closing.", "remote_addr", c.remoteAddr)
		closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "from server")
		if err := c.WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(5*time.Second)); err != nil {
			slog.Error("failed to write close message", "error", err)
			return nil
		}
		select {
		case <-done:
		case <-time.After(closeTimeout):
		}
	}
	m.setUnestablished()
	return nil
}
