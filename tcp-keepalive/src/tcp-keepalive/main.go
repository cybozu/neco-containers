package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = cobra.Command{
	Use:   "tcp-keepalive",
	Short: "tcp-keepalive is a simple TCP server and client program to confirm the long live connectivity.",
	RunE:  rootMain,
}

var ServerCmd = cobra.Command{
	Use:   "server",
	Short: "run the TCP server",
	RunE:  serverMain,
}

var ClientCmd = cobra.Command{
	Use:   "client",
	Short: "run the TCP client",
	RunE:  clientMain,
}

func init() {
	ServerCmd.Flags().StringP("listen", "l", ":8000", "Listen address and port")
	ServerCmd.Flags().DurationP("interval", "i", time.Second*5, "Interval to send a keepalive message")
	ServerCmd.Flags().DurationP("timeout", "t", time.Second*15, "Deadline to receive a keepalive message")
	ServerCmd.Flags().Int("retry-limit", 0, "The limit to retry, 0 is no limit")
	ServerCmd.Flags().Bool("silent", false, "Server doesn't send keepalive message")
	ClientCmd.Flags().StringP("server", "s", "127.0.0.1:8000", "Server running host")
	ClientCmd.Flags().DurationP("interval", "i", time.Second*5, "Interval to send a keepalive message")
	ClientCmd.Flags().DurationP("timeout", "t", time.Second*15, "Deadline to receive a keepalive message")
	ClientCmd.Flags().BoolP("retry", "y", false, "Try to connect after a previous connection is closed")
	ClientCmd.Flags().DurationP("retry-interval", "r", time.Second, "Connect retry interval")
	ClientCmd.Flags().Bool("ignore-server-msg", false, "Ignore whether receiving the message from server or not")
	rootCmd.AddCommand(&ServerCmd)
	rootCmd.AddCommand(&ClientCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootMain(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func serverMain(cmd *cobra.Command, args []string) error {
	logger := initLogger()

	addr, err := cmd.Flags().GetString("listen")
	if err != nil {
		logger.Error("Failed to get listen flag", slog.Any("error", err))
		return err
	}

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		logger.Error("Failed to get interval flag", slog.Any("error", err))
		return err
	}

	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		logger.Error("Failed to get timeout flag", slog.Any("error", err))
		return err
	}

	retryLimit, err := cmd.Flags().GetInt("retry-limit")
	if err != nil {
		logger.Error("Failed to get retry-limit flag", slog.Any("error", err))
		return err
	}

	noMsg, err := cmd.Flags().GetBool("silent")
	if err != nil {
		logger.Error("Failed to get silent flag", slog.Any("error", err))
		return err
	}

	logger = logger.With("local", addr)

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logger.Error("Failed to resolve host", slog.Any("error", err))
		return err
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Start TCP server", slog.Int("retry-limit", retryLimit), slog.Duration("interval", interval), slog.Duration("timeout", timeout))

	retryCount := 0

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		logger.Error("Failed to listen TCP", slog.Any("error", err))
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	connections := make(chan *net.TCPConn)
	closeNotifyChan := make(chan struct{})

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := listener.AcceptTCP()
				if err != nil {
					logger.ErrorContext(ctx, "Failed to accept new TCP connection", slog.Any("error", err))
				}
				logger.InfoContext(ctx, "Accept new connection.", slog.Any("remote", conn.RemoteAddr()), slog.Int("retry-count", retryCount))
				connections <- conn
				retryCount += 1
			}
		}

	}(ctx)

	for {
		select {
		case <-sig:
			// close all connections
			logger.InfoContext(ctx, "Signal received, close all connections...")
			cancel()
			return nil
		case conn := <-connections:
			go handleConnection(ctx, conn, interval, timeout, closeNotifyChan, noMsg, false)
		case <-closeNotifyChan:
			if retryLimit != 0 && retryCount >= retryLimit {
				logger.WarnContext(ctx, "Exceed retry limit. exit.")
				cancel()
				return nil
			}
		}

	}
}

func handleConnection(ctx context.Context, conn net.Conn, interval, timeout time.Duration, closeNotifyChan chan struct{}, notSendMsg, ignoreRecvMsg bool) {
	logger := initLogger().With("remote", conn.RemoteAddr())

	intervalTicker := time.NewTicker(interval)
	timeoutTicker := time.NewTicker(timeout)

	closeChan := make(chan struct{})
	receiveChan := make(chan struct{})

	go receive(ctx, logger, conn, closeChan, receiveChan)
	for {
		select {
		case <-ctx.Done():
			logger.InfoContext(ctx, "Close connection")
			conn.Close()
			if closeNotifyChan != nil {
				closeNotifyChan <- struct{}{}
			}
			return
		case <-closeChan:
			logger.InfoContext(ctx, "Close connection")
			conn.Close()
			if closeNotifyChan != nil {
				closeNotifyChan <- struct{}{}
			}
			return
		case <-intervalTicker.C:
			if !notSendMsg {
				if _, err := conn.Write([]byte("keepalive")); err != nil {
					logger.ErrorContext(ctx, "Failed to send keepalive message", slog.Any("error", err))
					if closeNotifyChan != nil {
						closeNotifyChan <- struct{}{}
					}
					return
				}
				logger.InfoContext(ctx, "Send a keepalive message")
			}
		case <-timeoutTicker.C:
			if !ignoreRecvMsg {
				logger.WarnContext(ctx, "Deadline exceeded to receive a keepalive message. Close connection")
				conn.Close()
				if closeNotifyChan != nil {
					closeNotifyChan <- struct{}{}
				}
				return
			}
		case <-receiveChan:
			timeoutTicker.Reset(timeout)
		}
	}
}

func clientMain(cmd *cobra.Command, args []string) error {
	logger := initLogger()

	addr, err := cmd.Flags().GetString("server")
	if err != nil {
		logger.Error("Failed to get server flag", slog.Any("error", err))
		return err
	}

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		logger.Error("Failed to get interval flag", slog.Any("error", err))
		return err
	}

	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		logger.Error("Failed to get timeout flag", slog.Any("error", err))
		return err
	}

	retryInterval, err := cmd.Flags().GetDuration("retry-interval")
	if err != nil {
		logger.Error("Failed to get retry-interval flag", slog.Any("error", err))
		return err
	}

	retry, err := cmd.Flags().GetBool("retry")
	if err != nil {
		logger.Error("Failed to get retry flag", slog.Any("error", err))
		return err
	}

	ignoreRecvMsg, err := cmd.Flags().GetBool("ignore-server-msg")
	if err != nil {
		logger.Error("Failed to get ignore-server-msg flag", slog.Any("error", err))
		return err
	}

	logger = logger.With("local", addr)
	logger.Info("Start TCP client")

	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	nextChan := make(chan struct{}, 1)
	connections := make(chan net.Conn)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
			case <-nextChan:
				conn, err := net.Dial("tcp", addr)
				if err != nil {
					logger.Error("Failed to dial to TCP server", slog.Any("error", err))
					connections <- nil
				} else {
					connections <- conn
				}
			}
		}

	}(ctx)

	nextChan <- struct{}{}
	waitChan := make(chan struct{}, 1)

	for {
		select {
		case <-sig:
			logger.WarnContext(ctx, "Signal received. Stop the client")
			cancel()
			return nil
		case conn := <-connections:
			if conn == nil {
				if !retry {
					cancel()
					return fmt.Errorf("Got nil from the connection channel")
				}
				time.Sleep(retryInterval)
				nextChan <- struct{}{}
				continue
			}
			go func() {
				logger.InfoContext(ctx, "Start to handle connection")
				handleConnection(ctx, conn, interval, timeout, nil, false, ignoreRecvMsg)
				waitChan <- struct{}{}
			}()
		case <-waitChan:
			if !retry {
				cancel()
				return nil
			}
			time.Sleep(retryInterval)
			nextChan <- struct{}{}
		}
	}

}

func receive(ctx context.Context, logger *slog.Logger, conn net.Conn, closeChan, receiveChan chan struct{}) {
	for {
		buf := make([]byte, 1024)
		l, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.WarnContext(ctx, "Got EOF. Close connection.")
				closeChan <- struct{}{}
				return
			}
			netErr := errors.Unwrap(err)
			if errors.Is(netErr, net.ErrClosed) {
				return
			}
			logger.ErrorContext(ctx, "Failed to read the message", slog.Any("error", err))
		}
		msg := string(buf[:l])
		if msg != "keepalive" {
			logger.WarnContext(ctx, "Receive the other message", slog.String("message", msg))
		} else {
			logger.InfoContext(ctx, "Receive the keepalive")
			receiveChan <- struct{}{}
		}
	}
}

func initLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
