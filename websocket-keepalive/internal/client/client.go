package client

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cybozu/neco-containers/websocket-keepalive/internal/common"
	"github.com/cybozu/neco-containers/websocket-keepalive/internal/metrics"
	"github.com/gorilla/websocket"
)

type Config struct {
	Host           string
	Port           int
	PingInterval   time.Duration
	MaxPingRetries int
}

type message struct {
	messageType int
	message     []byte
}

type conn struct {
	*websocket.Conn
	closeCh chan struct{}
	closing atomic.Bool
	ctrlC   chan os.Signal
}

func newConn(c *websocket.Conn) *conn {
	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt, syscall.SIGTERM)
	return &conn{
		Conn:    c,
		closing: atomic.Bool{},
		closeCh: make(chan struct{}),
		ctrlC:   ctrlC,
	}
}

func Run(host string, port int, interval time.Duration) error {
	config := &Config{
		Host:           host,
		Port:           port,
		PingInterval:   interval,
		MaxPingRetries: 3,
	}

	metricsConfig := &metrics.Config{
		Export:   true,
		AddrPort: "0.0.0.0:8080",
	}
	return RunWithConfig(config, metricsConfig)
}

func RunWithConfig(config *Config, metricsConfig *metrics.Config) error {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		return err
	}
	u := url.URL{Scheme: "ws", Host: addr.AddrPort().String(), Path: "/ws"}

	// ctx := context.Background()

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

	slog.Info("Connecting to WebSocket server", "url", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	initMetrics(c.LocalAddr().String(), fmt.Sprintf("%s:%d", config.Host, config.Port))

	m.setUnestablished()
	m.setEstablished()

	conn := newConn(c)
	slog.Info("Connected to WebSocket server", "remote_addr", conn.RemoteAddr().String())
	defer conn.Close()

	return conn.handleWebsocketConnection(context.Background(), config, m)

}

func (c *conn) handleWebsocketConnection(ctx context.Context, config *Config, m *Metrics) error {
	msgCh := make(chan message)

	// Set pong handler
	c.SetPongHandler(func(appData string) error {
		msgCh <- message{
			messageType: websocket.PongMessage,
			message:     []byte(appData),
		}
		m.incrementPongTotal()
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(config.PingInterval)
	defer func() {
		ticker.Stop()
		m.setUnestablished()
	}()

	go func() {
		for {
			msgType, message, err := c.ReadMessage()
			if err != nil {
				if !websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					// This code will be run in both active and passive close sequence.
					// But, in active closing, this normal closure error must be a reply message, so do nothing.
					// When passive close, we must send the reply for its close message, but default close handler will do it, so we also do nothing in this case.
					slog.Info("Received close message", "msg", message, "remote_addr", c.RemoteAddr().String())
					c.closeCh <- struct{}{}
					return
				}
				slog.Error("Failed to read message", "error", err, "remote_addr", c.RemoteAddr().String())
				return
			}
			slog.Debug("Received message", "message", string(message), "message-type", common.WebSocketMessageType(msgType), "remote_addr", c.RemoteAddr().String())
		}
	}()

	retryCount := 0
	waitingForReply := false
	waitingCount := 0
	waitingLimit := 1

	for {
		select {
		case <-c.closeCh:
			slog.Info("connection closed.", "remote_addr", c.RemoteAddr().String())
			return nil
		case <-c.ctrlC:
			slog.Info("Interrupt received, closing connection", "remote_addr", c.RemoteAddr().String())
			// acitve close sequence.
			// we must send a close message.
			if err := c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "from client"), time.Now().Add(time.Second*5)); err != nil {
				return err
			}
			select {
			case <-c.closeCh:
				slog.Info("the final close message was got. closed.", "remote_addr", c.RemoteAddr().String())
			case <-time.After(time.Second * 10):
				slog.Info("dead line for waiting the final close message, close.", "remote_addr", c.RemoteAddr().String())
			}
			return nil
		case <-ticker.C:
			if waitingForReply {
				if waitingCount < waitingLimit {
					slog.Warn("Still waiting for reply", "remote_addr", c.RemoteAddr().String())
					waitingCount += 1
					continue
				}
				if retryCount >= config.MaxPingRetries {
					// decide the connection is broken.
					slog.Error("Reached to retry limit. The connection is broken.", "remote_addr", c.RemoteAddr().String())
					c.closeCh <- struct{}{}
					return nil
				}
				retryCount += 1
				waitingCount = 0
				m.incrementRetryCount()
				slog.Debug("Sending ping message for retry", "retry", retryCount, "max-retry", config.MaxPingRetries, "remote_addr", c.RemoteAddr().String())
			}
			slog.Debug("Sending ping message", "remote_addr", c.RemoteAddr())
			err := c.WriteMessage(websocket.PingMessage, []byte("hello from client"))
			if err != nil {
				slog.Error("Failed to send ping message", "error", err, "remote_addr", c.RemoteAddr().String())
				return nil
			}
			m.incrementPingTotal()
			waitingForReply = true
		case msg := <-msgCh:
			slog.Debug("Received message", "message", string(msg.message), "message-type", common.WebSocketMessageType(msg.messageType), "remote_addr", c.RemoteAddr().String())
			retryCount = 0
			waitingForReply = false
		}
	}
}
