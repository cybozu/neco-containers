package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync/atomic"
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
	client         *http.Client
	timeout        time.Duration
	readyURL       string
	httpURL        string
	httpsAddr      string
	httpActivated  atomicBool
	httpsActivated atomicBool
}

func (m *monitor) Probe(ctx context.Context) error {
	env := well.NewEnvironment(ctx)
	env.Go(m.monitorReady)
	env.Go(m.monitorHTTP)
	env.Go(m.monitorHTTPS)

	env.Stop()
	return env.Wait()
}

func (m *monitor) monitorReady(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", m.readyURL, nil)
	if err != nil {
		return fmt.Errorf("failed to build HTTP request for readyURL: %v", err)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to access readiness probe: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("envoy readiness endpoint returned non-OK with status %d", resp.StatusCode)
	}

	return nil
}

func (m *monitor) monitorHTTP(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", m.httpURL, nil)
	if err != nil {
		return fmt.Errorf("failed to build HTTP request for httpURL: %v", err)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		if !m.httpActivated.get() {
			return nil
		}
		return fmt.Errorf("failed to access HTTP endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Status code is not checked.
	// The current implementation of Envoy returns 404, but this can be changed.
	m.httpActivated.set(true)
	return nil
}

func (m *monitor) monitorHTTPS(ctx context.Context) error {
	conn, err := net.DialTimeout("tcp", m.httpsAddr, m.timeout)
	if err != nil {
		if !m.httpsActivated.get() {
			return nil
		}
		return fmt.Errorf("failed to connect to HTTPS endpoint: %v", err)
	}

	if conn != nil {
		conn.Close()
	}
	m.httpsActivated.set(true)
	return nil
}

func (m *monitor) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := m.Probe(req.Context()); err != nil {
		_ = log.Error("monitor failed", map[string]interface{}{
			log.FnError: err,
		})
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	_ = log.Debug("monitor succeeded", nil)
}

type livenessMonitor struct {
	*monitor
}

func (m *livenessMonitor) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := m.Probe(req.Context()); err != nil {
		_ = log.Error("liveness monitor failed", map[string]interface{}{
			log.FnError: err,
		})
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	_ = log.Debug("liveness monitor succeeded", nil)
}

type readinessMonitor struct {
	*monitor
}

func (m *readinessMonitor) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := m.Probe(req.Context()); err != nil {
		_ = log.Error("readiness monitor failed", map[string]interface{}{
			log.FnError: err,
		})
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	_ = log.Debug("readiness monitor succeeded", nil)
}

type atomicBool struct {
	flag int32
}

func (b *atomicBool) set(flag bool) {
	var val int32
	if flag {
		val = 1
	}
	atomic.StoreInt32(&b.flag, val)
}

func (b *atomicBool) get() bool {
	return atomic.LoadInt32(&b.flag) != 0
}

var rootCmd = &cobra.Command{
	Use:   "probe",
	Short: "liveness/readiness probe for Envoy",
	Long:  `Liveness/Readiness probe for Envoy.`,

	Run: func(cmd *cobra.Command, args []string) {
		err := well.LogConfig{}.Apply()
		if err != nil {
			log.ErrorExit(err)
		}

		mux := http.NewServeMux()

		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.DisableKeepAlives = true

		m := &monitor{
			client: &http.Client{
				Transport: transport,
				Timeout:   config.timeout,
			},
			readyURL:  config.readyURL,
			httpURL:   config.httpURL,
			httpsAddr: config.httpsAddr,
			timeout:   config.timeout,
		}
		lm := &livenessMonitor{m}
		rm := &readinessMonitor{m}

		mux.Handle("/", m)
		mux.Handle("/healthz", lm)
		mux.Handle("/readyz", rm)

		serv := &http.Server{
			Addr:    config.listenAddr,
			Handler: mux,
		}
		well.Go(func(ctx context.Context) error {
			<-ctx.Done()
			return serv.Shutdown(ctx)
		})
		err = serv.ListenAndServe()
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

func init() {
	fs := rootCmd.Flags()
	fs.StringVar(&config.listenAddr, "listen-addr", ":8502", "Listen address for probes")
	fs.DurationVar(&config.timeout, "timeout", time.Second*5, "Timeout")
	fs.StringVar(&config.readyURL, "ready-url", "http://localhost:8002/ready", "URL of Envoy readiness probe")
	fs.StringVar(&config.httpURL, "http-url", "http://localhost:8080/", "URL for checking HTTP behavior")
	fs.StringVar(&config.httpsAddr, "https-addr", "localhost:8443", "Address for checking HTTPS behavior")
}
