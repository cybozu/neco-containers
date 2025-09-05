package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cybozu/neco-containers/websocket-keepalive/internal/common"
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

	closing := atomic.Bool{}
	closeTimeout := time.Second * 10

	activeClose := make(chan struct{})
	pasiveClose := make(chan struct{})

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
		// When pasive closing, we don't have to send close message maually.
		// If server routines have already recieved, closing variable must be true.
		if !closing.Load() {
			closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "from server")
			if err := c.WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(5*time.Second)); err != nil {
				slog.Error("failed to write close message", "error", err)
			}
			now := time.Now()
			for !closing.Load() {
				// waiting for a close message from client
				if time.Since(now) > closeTimeout {
					slog.Error("close timeout is expire, close connection")
					break
				}
				time.Sleep(time.Millisecond)
			}

		}
		c.Close()
		return connections.unregister(c.remoteAddr)
	}()

	go func() {
		// Message reading loop
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
						slog.Error("Got websocket abnormal closure", "error", err)
					} else {
						slog.Error("WebSocket error", "error", err)
					}
					activeClose <- struct{}{}
				} else {
					if !closing.Load() {
						slog.Info("Connection will be closed from client", "error", err, "messageType", messageType, "message", message)
						closing.Store(true)
						pasiveClose <- struct{}{}
					}
					return
				}
			}
			switch messageType {
			case websocket.TextMessage:
				slog.Debug("Received text message", "message", string(message), "remote_addr", c.remoteAddr)
			case websocket.CloseMessage:
				slog.Info("Received close message", "message", string(message), "remote_addr", c.remoteAddr)
				return
			default:
				slog.Debug("Received message", "type", common.WebSocketMessageType(messageType), "remote_addr", c.remoteAddr)
			}
		}
	}()

	select {
	case <-c.serverCloseCh:
		slog.Info("shutdown signal received, closing.", "remote_addr", c.remoteAddr)
	case <-activeClose:
		slog.Error("something went wrong, close connection", "remote_addr", c.remoteAddr)
	case <-pasiveClose:
		slog.Info("closed passively.", "remote_addr", c.remoteAddr)
	}
	m.setUnestablished()
	return nil
}
