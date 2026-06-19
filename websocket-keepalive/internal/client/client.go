package client

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"time"

	"github.com/gorilla/websocket"

	"github.com/cybozu/neco-containers/websocket-keepalive/internal/common"
	"github.com/cybozu/neco-containers/websocket-keepalive/internal/metrics"
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
}

func newConn(c *websocket.Conn) *conn {
	return &conn{Conn: c}
}

func RunWithConfig(ctx context.Context, config *Config, metricsConfig *metrics.Config) error {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		return err
	}
	remote := addr.AddrPort().String()
	u := url.URL{Scheme: "ws", Host: remote, Path: "/ws"}

	m, err := NewMetrics(metricsConfig)
	if err != nil {
		return err
	}
	if metricsConfig.Export {
		slog.Info("Start metrics server", "listen", m.AddrPort)
		go serveMetrics(ctx, m.Metrics)
	}

	slog.Info("Connecting to WebSocket server", "url", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	initMetrics(c.LocalAddr().String(), remote)

	m.setUnestablished()
	m.setEstablished()

	conn := newConn(c)
	slog.Info("Connected to WebSocket server", "remote_addr", conn.RemoteAddr().String())
	defer conn.Close()

	return conn.handleWebsocketConnection(ctx, config, m)
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

func (c *conn) handleWebsocketConnection(ctx context.Context, config *Config, m *Metrics) error {
	pongWait := 2 * config.PingInterval
	if err := c.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return fmt.Errorf("failed to set initial read deadline: %w", err)
	}

	pongReceived := make(chan message, 1)

	c.SetPongHandler(func(appData string) error {
		if err := c.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			slog.Warn("failed to extend read deadline on pong", "error", err, "remote_addr", c.RemoteAddr().String())
		}
		select {
		case pongReceived <- message{messageType: websocket.PongMessage, message: []byte(appData)}:
		default:
		}
		m.incrementPongTotal()
		return nil
	})

	ticker := time.NewTicker(config.PingInterval)
	defer func() {
		ticker.Stop()
		m.setUnestablished()
	}()

	var readerErr error
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		for {
			msgType, data, err := c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					// Runs in both active and passive close sequences.
					// Active close: this is the server's close reply.
					// Passive close: gorilla's default close handler already sent the reply.
					slog.Info("Connection closed normally", "remote_addr", c.RemoteAddr().String())
					return
				}
				// Anything else (abnormal close code, read deadline exceeded, TCP RST, etc.)
				// must be surfaced so the main loop returns it and the process exits non-zero.
				slog.Error("Reader terminated abnormally", "error", err, "remote_addr", c.RemoteAddr().String())
				readerErr = err
				return
			}
			slog.Debug("Received message", "message", string(data), "message-type", common.WebSocketMessageType(msgType), "remote_addr", c.RemoteAddr().String())
		}
	}()

	const waitingLimit = 1

	retryCount := 0
	waitingForReply := false
	waitingCount := 0

	for {
		select {
		case <-ctx.Done():
			slog.Info("Interrupt received, closing connection", "remote_addr", c.RemoteAddr().String())
			// active close sequence.
			// we must send a close message. safe: gorilla serializes WriteControl internally.
			if err := c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "from client"), time.Now().Add(time.Second*5)); err != nil {
				return err
			}
			select {
			case <-readerDone:
				slog.Info("the final close message was got. closed.", "remote_addr", c.RemoteAddr().String())
			case <-time.After(time.Second * 10):
				slog.Info("dead line for waiting the final close message, close.", "remote_addr", c.RemoteAddr().String())
			}
			return nil
		case <-readerDone:
			slog.Info("connection closed.", "remote_addr", c.RemoteAddr().String())
			return readerErr
		case <-ticker.C:
			if waitingForReply {
				if waitingCount < waitingLimit {
					slog.Warn("Still waiting for reply", "remote_addr", c.RemoteAddr().String())
					waitingCount += 1
					continue
				}
				if retryCount >= config.MaxPingRetries {
					slog.Error("Reached to retry limit. The connection is broken.", "remote_addr", c.RemoteAddr().String())
					return fmt.Errorf("ping retry limit exceeded (%d)", config.MaxPingRetries)
				}
				retryCount += 1
				waitingCount = 0
				m.incrementRetryCount()
				slog.Debug("Sending ping message for retry", "retry", retryCount, "max-retry", config.MaxPingRetries, "remote_addr", c.RemoteAddr().String())
			}
			slog.Debug("Sending ping message", "remote_addr", c.RemoteAddr().String())
			if err := c.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
				return fmt.Errorf("failed to set write deadline: %w", err)
			}
			err := c.WriteMessage(websocket.PingMessage, []byte("hello from client"))
			if err != nil {
				slog.Error("Failed to send ping message", "error", err, "remote_addr", c.RemoteAddr().String())
				return err
			}
			m.incrementPingTotal()
			waitingForReply = true
		case msg := <-pongReceived:
			slog.Debug("Received pong", "message", string(msg.message), "message-type", common.WebSocketMessageType(msg.messageType), "remote_addr", c.RemoteAddr().String())
			retryCount = 0
			waitingForReply = false
		}
	}
}
