package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/neco-containers/tcp-keepalive/internal/metrics"
)

const clientMessage = "hello"

var log *slog.Logger

type Client struct {
	conn    *net.TCPConn
	metrics *Metrics

	*Config
}

func init() {
	initLogger()
}

func initLogger() {
	log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func NewClient(cfg *Config, mcfg *metrics.Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	m, err := NewMetrics(mcfg)
	if err != nil {
		return nil, err
	}

	log.Info("new client", slog.Any("config", cfg), slog.Any("metrics", m))
	return &Client{Config: cfg, metrics: m}, nil
}

func (c *Client) Run(ctx context.Context) error {
	log = log.With("dst", c.ServerAddr)
	c.metrics.setConnStateUnestablished()

	if c.metrics.Export {
		go func() {
			for {
				if err := c.metrics.Serve(); err != nil {
					log.Error("serving metrics failed", slog.Any("error", err))
				}
			}
		}()
	}

	if c.RetryNum < 0 {
		cnt := uint64(0)
		for ctx.Err() == nil {
			retryTotal.Set(cnt)
			retryCount.Set(cnt)

			if err := c.run(ctx); err != nil {
				log.Error("run failed", slog.Any("error", err))
				time.Sleep(c.RetryInterval)
			}
			cnt++
		}
		return nil
	}

	for i := 0; i <= c.RetryNum; i++ {
		retryCount.Set(uint64(i))

		if ctx.Err() != nil {
			return nil
		}

		if err := c.run(ctx); err != nil {
			log.Error("run failed", slog.Any("error", err))
			time.Sleep(c.RetryInterval)
			retryTotal.Inc()
			continue
		}
		i = 0
	}
	return errors.New("retry limit exceeded")
}

func (c *Client) run(ctx context.Context) error {
	if err := c.Dial(); err != nil {
		return err
	}
	defer c.Close()

	return c.SendByPeriod(ctx)
}

func (c *Client) Dial() error {
	ap, err := net.ResolveTCPAddr("tcp", c.ServerAddr)
	if err != nil {
		return err
	}
	if c.conn, err = net.DialTCP("tcp", nil, ap); err != nil {
		return err
	}
	c.metrics.setConnStateEstablished()
	return nil
}

func (c *Client) Close() {
	if c.conn == nil {
		return
	}

	if err := c.conn.Close(); err != nil {
		log.Error("failed to close conn", slog.Any("error", err))
	}
	c.metrics.setConnStateClosed()
}

func (c *Client) SendByPeriod(ctx context.Context) error {
	si := time.NewTicker(c.SendInterval)
	for {
		if err := c.SendAndReceive(); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case <-si.C:
			continue
		}
	}
}

func (c *Client) SendAndReceive() error {
	t, cancel := context.WithTimeout(context.Background(), c.ReceiveTimeout)
	defer cancel()

	done := make(chan error)
	sendDone := false
	go func() {
		if err := c.Send(); err != nil {
			done <- err
			return
		}
		sendDone = true
		done <- c.Receive()
	}()

	for {
		select {
		case <-t.Done():
			if !sendDone {
				sendTimeoutTotal.Inc()
			} else {
				receiveTimeoutTotal.Inc()
			}
			return t.Err()
		case err := <-done:
			return err
		}
	}
}

func (c *Client) Send() error {
	log.Info("send a message", slog.Any("message", clientMessage))
	if _, err := c.conn.Write([]byte(clientMessage)); err != nil {
		sendErrorTotal.Inc()
		return err
	}
	sendSuccessTotal.Inc()
	return nil
}

func (c *Client) Receive() error {
	buf := make([]byte, 1024)
	l, err := c.conn.Read(buf)
	if err != nil {
		receiveErrorTotal.Inc()

		if errors.Is(err, io.EOF) {
			return errors.New("got EOF")
		}
		netErr := errors.Unwrap(err)
		if errors.Is(netErr, net.ErrClosed) {
			return errors.New("connection is closed")
		}
		return fmt.Errorf("failed to read the response: %w", err)
	}

	msg := string(buf[:l])
	log := log.With("message", msg)
	if msg != clientMessage {
		receiveErrorTotal.Inc()
		log.Warn("receive an unexpected message")
	} else {
		receiveSuccessTotal.Inc()
		log.Info("receive a response message")
	}
	return nil
}
