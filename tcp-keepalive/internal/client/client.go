package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"os"
	"time"
)

const clientMessage = "hello"

var log *slog.Logger

type Client struct {
	conn *net.TCPConn

	*Config
}

func init() {
	initLogger()
}

func initLogger() {
	log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func NewClient(cfg *Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	log.Info("new client", slog.Any("config", cfg))
	return &Client{Config: cfg}, nil
}

func (c *Client) Run(ctx context.Context) error {
	log = log.With("dst", c.ServerAddr)

	if c.RetryNum < 0 {
		for ctx.Err() == nil {
			if err := c.run(ctx); err != nil {
				log.Error("run failed", slog.Any("error", err))
				time.Sleep(c.RetryInterval)
			}
		}
		return nil
	}

	for i := 0; i <= c.RetryNum; i++ {
		if ctx.Err() != nil {
			return nil
		}

		if err := c.run(ctx); err != nil {
			log.Error("run failed", slog.Any("error", err))
			time.Sleep(c.RetryInterval)
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
	ap, err := netip.ParseAddrPort(c.ServerAddr)
	if err != nil {
		return err
	}
	c.conn, err = net.DialTCP("tcp", nil, net.TCPAddrFromAddrPort(ap))
	return err
}

func (c *Client) Close() {
	if c.conn == nil {
		return
	}

	if err := c.conn.Close(); err != nil {
		log.Error("failed to close conn", slog.Any("error", err))
	}
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
	go func() {
		if err := c.Send(); err != nil {
			done <- err
		}
		done <- c.Receive()
	}()

	for {
		select {
		case <-t.Done():
			return t.Err()
		case err := <-done:
			return err
		}
	}
}

func (c *Client) Send() error {
	log.Info("send a message", slog.Any("message", clientMessage))
	_, err := c.conn.Write([]byte(clientMessage))
	return err
}

func (c *Client) Receive() error {
	buf := make([]byte, 1024)
	l, err := c.conn.Read(buf)
	if err != nil {
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
		log.Warn("receive an unexpected message")
	} else {
		log.Info("receive a response message")
	}
	return nil
}
