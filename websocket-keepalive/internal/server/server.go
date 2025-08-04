package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cybozu/neco-containers/websocket-keepalive/internal/common"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Config struct {
	Host         string
	Port         int
	PingInterval time.Duration
}

var pingInterval time.Duration

func Run(host string, port int, interval time.Duration) error {

	pingInterval = interval

	http.HandleFunc("/ws", handleWebSocket)

	server := &http.Server{
		Addr: fmt.Sprintf("%s:%d", host, port),
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("WebSocket server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
		}
	}()

	<-stop
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade connection", "error", err)
		return
	}
	defer conn.Close()

	slog.Info("New WebSocket connection established", "remote_addr", r.RemoteAddr, "ping_interval", pingInterval)

	// Set ping handler (automatic pong response)
	conn.SetPingHandler(func(appData string) error {
		slog.Debug("Received ping, sending pong", "message", appData, "remote_addr", conn.RemoteAddr().String())
		return conn.WriteMessage(websocket.PongMessage, []byte("hello from server"))
	})

	// Message reading loop
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("WebSocket error", "error", err)
			} else {
				slog.Info("Connection closed", "remote_addr", r.RemoteAddr)
			}
			break
		}

		switch messageType {
		case websocket.TextMessage:
			slog.Debug("Received text message", "message", string(message), "remote_addr", r.RemoteAddr)
		default:
			slog.Debug("Received message", "type", common.WebSocketMessageType(messageType), "remote_addr", r.RemoteAddr)
		}
	}
}
