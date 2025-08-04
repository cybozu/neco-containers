package client

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cybozu/neco-containers/websocket-keepalive/internal/common"
	"github.com/gorilla/websocket"
)

type Config struct {
	Host           string
	Port           int
	PingInterval   time.Duration
	MaxPingRetries int
}

func Run(host string, port int, interval time.Duration) error {
	config := &Config{
		Host:           host,
		Port:           port,
		PingInterval:   interval,
		MaxPingRetries: 3,
	}
	return RunWithConfig(config)
}

func RunWithConfig(config *Config) error {
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", config.Host, config.Port), Path: "/ws"}
	slog.Info("Connecting to WebSocket server", "url", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	slog.Info("Connected to WebSocket server")

	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	type message struct {
		messageType int
		message     []byte
	}
	msgCh := make(chan message)

	// Set pong handler
	conn.SetPongHandler(func(appData string) error {
		msgCh <- message{
			messageType: websocket.PongMessage,
			message:     []byte(appData),
		}
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(config.PingInterval)
	defer ticker.Stop()

	go func() {
		retryCount := 0
		waitingForReply := false
		waitingCount := 0
		waitingLimit := 1

		defer close(done)

		for {
			select {
			case <-ticker.C:
				if waitingForReply {
					if waitingCount < waitingLimit {
						slog.Warn("Still waiting for reply")
						waitingCount += 1
						continue
					}
					if retryCount < config.MaxPingRetries {
						retryCount += 1
						slog.Debug("Sending ping message for retry", "retry", retryCount, "max-retry", config.MaxPingRetries)
						err := conn.WriteMessage(websocket.PingMessage, []byte("hello from client"))
						if err != nil {
							slog.Error("Failed to send ping message", "error", err)
							return
						}
						waitingCount = 0
					} else {
						// decide the connection is broken.
						slog.Error("Reached to retry limit. The connection is broken.")
						return
					}
				} else {
					slog.Debug("Sending ping message")
					err := conn.WriteMessage(websocket.PingMessage, []byte("hello from client"))
					if err != nil {
						slog.Error("Failed to send ping message", "error", err)
						return
					}
				}
				waitingForReply = true
			case msg := <-msgCh:
				slog.Debug("Received message", "message", string(msg.message), "message-type", common.WebSocketMessageType(msg.messageType))
				retryCount = 0
				waitingForReply = false
			case <-done:
				return
			}
		}
	}()

	go func() {
		defer close(done)
		for {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				slog.Error("Failed to read message", "error", err)
				return
			}
			slog.Debug("Received message", "message", string(message), "message-type", common.WebSocketMessageType(msgType))
		}
	}()

	for {
		select {
		case <-done:
			return fmt.Errorf("connection closed")
		case <-interrupt:
			slog.Info("Interrupt received, closing connection")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				return err
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}
}
