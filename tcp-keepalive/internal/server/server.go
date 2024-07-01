package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
)

var log *slog.Logger

type Server struct {
	listener *net.TCPListener

	*Config
}

func init() {
	initLogger()
}

func initLogger() {
	log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func NewServer(cfg *Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	log.Info("new server", slog.Any("config", cfg))
	return &Server{Config: cfg}, nil
}

func (s *Server) Run(ctx context.Context) error {
	log = log.With("listen", s.ListenAddr)

	if err := s.Listen(); err != nil {
		return err
	}

	defer s.Close()

	log.Info("start tcp-keepalive server")

	for {
		if ctx.Err() != nil {
			return nil
		}

		conn, err := s.Accept()
		if err != nil {
			log.Warn("failed to accept", slog.Any("error", err))
			continue
		}

		go func(conn net.Conn) {
			log := log.With("client", conn.RemoteAddr().String())
			log.Info("accepted a connection")
			defer func() {
				if err := conn.Close(); err != nil {
					log.Error("failed to close connection", slog.Any("error", err))
				}
			}()

			if err := s.Handle(ctx, conn); err != nil {
				log.Error("failed to handle connection", slog.Any("error", err))
			}
		}(conn)
	}
}

func (s *Server) Listen() error {
	addr, err := net.ResolveTCPAddr("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve address: %w", err)
	}

	if s.listener, err = net.ListenTCP("tcp", addr); err != nil {
		return fmt.Errorf("failed to start listen: %w", err)
	}
	return nil
}

func (s *Server) Close() {
	if s.listener == nil {
		return
	}

	if err := s.listener.Close(); err != nil {
		log.Error("failed to close listener", slog.Any("error", err))
	}
}

func (s *Server) Accept() (net.Conn, error) {
	return s.listener.AcceptTCP()
}

func (s *Server) Handle(ctx context.Context, conn net.Conn) error {
	done := make(chan error)
	go func() {
		for {
			done <- s.handle(conn)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return <-done
		case err := <-done:
			if err != nil {
				return err
			}
		}
	}
}

func (s *Server) handle(conn net.Conn) error {
	msg, err := s.Recieve(conn)
	if err != nil {
		return err
	}
	log := log.With("client", conn.RemoteAddr().String())
	log.Info("received a message", slog.String("message", msg))

	return s.Send(msg, conn)
}

func (s *Server) Recieve(conn net.Conn) (string, error) {
	buf := make([]byte, 1024)
	l, err := conn.Read(buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return "", errors.New("got EOF")
		}
		netErr := errors.Unwrap(err)
		if errors.Is(netErr, net.ErrClosed) {
			return "", errors.New("connection is closed")
		}
		return "", fmt.Errorf("failed to read the response: %w", err)
	}

	return string(buf[:l]), nil

}

func (s *Server) Send(msg string, conn net.Conn) error {
	_, err := conn.Write([]byte(msg))
	return err
}
