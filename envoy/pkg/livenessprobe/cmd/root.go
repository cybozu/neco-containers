package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var config struct {
	timeout    time.Duration
	listenAddr string
	readyURL   string
	httpURL    string
	httpsAddr  string
}

type monitor struct {
	client    *http.Client
	timeout   time.Duration
	readyURL  string
	httpURL   string
	httpsAddr string
}

var rootCmd = &cobra.Command{
	Use:   "livenessprobe",
	Short: "liveness probe for Envoy",
	Long:  `Liveness probe for Envoy.`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true

		err := well.LogConfig{}.Apply()
		if err != nil {
			log.ErrorExit(err)
		}
	},

	Run: func(cmd *cobra.Command, args []string) {
		mux := http.NewServeMux()

		m := &monitor{
			client: &http.Client{
				Transport: &http.Transport{
					DisableKeepAlives: true,

					// rest are copied from http.DefaultTransport
					Proxy: http.ProxyFromEnvironment,
					DialContext: (&net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 30 * time.Second,
						DualStack: true,
					}).DialContext,
					ForceAttemptHTTP2:     true,
					MaxIdleConns:          100,
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
				},
				Timeout: config.timeout,
			},
			readyURL:  config.readyURL,
			httpURL:   config.httpURL,
			httpsAddr: config.httpsAddr,
			timeout:   config.timeout,
		}
		mux.Handle("/", m)

		serv := &http.Server{
			Addr:    config.listenAddr,
			Handler: mux,
		}
		well.Go(func(ctx context.Context) error {
			<-ctx.Done()
			return serv.Shutdown(ctx)
		})
		err := serv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.ErrorExit(err)
		}
	},
}

// Execute executes the command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (m *monitor) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	env := well.NewEnvironment(req.Context())
	env.Go(m.monitorReady)
	env.Go(m.monitorHTTP)
	env.Go(m.monitorHTTPS)

	env.Stop()
	err := env.Wait()
	if err != nil {
		log.Error("returning failure result", map[string]interface{}{
			log.FnError: err,
		})
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	log.Debug("returning success result", nil)
	return
}

func (m *monitor) monitorReady(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", m.readyURL, nil)
	if err != nil {
		log.Error("failed to build HTTP request for readyURL", map[string]interface{}{
			log.FnError: err,
		})
		return err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		log.Error("failed to access readiness probe", map[string]interface{}{
			log.FnError: err,
		})
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Error("readiness probe returned non-OK", map[string]interface{}{
			"status": resp.StatusCode,
		})
		return fmt.Errorf("readiness probe returned %d", resp.StatusCode)
	}

	return nil
}

func (m *monitor) monitorHTTP(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", m.httpURL, nil)
	if err != nil {
		log.Error("failed to build HTTP request for httpURL", map[string]interface{}{
			log.FnError: err,
		})
		return err
	}
	_, err = m.client.Do(req)
	if err != nil {
		log.Error("failed to access HTTP endpoint", map[string]interface{}{
			log.FnError: err,
		})
		return err
	}

	// Status code is not checked.
	// The current implementation of Envoy returns 404, but this can be changed.
	return nil
}

func (m *monitor) monitorHTTPS(ctx context.Context) error {
	conn, err := net.DialTimeout("tcp", m.httpsAddr, m.timeout)
	if err != nil {
		log.Error("failed to connect to HTTPS endpoint", map[string]interface{}{
			log.FnError: err,
		})
		return err
	}

	if conn != nil {
		conn.Close()
	}
	return nil
}

func init() {
	fs := rootCmd.Flags()
	fs.StringVar(&config.listenAddr, "listen-addr", ":8502", "Listen address for probe")
	fs.DurationVar(&config.timeout, "timeout", time.Second*5, "Timeout")
	fs.StringVar(&config.readyURL, "ready-url", "http://localhost:8002/ready", "URL of Envoy readiness probe")
	fs.StringVar(&config.httpURL, "http-url", "http://localhost:8080/", "URL for checking HTTP behavior")
	fs.StringVar(&config.httpsAddr, "https-addr", "localhost:8443", "Address for checking HTTPS behavior")
}
