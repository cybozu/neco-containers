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

	RunE: func(cmd *cobra.Command, args []string) error {
		mux := http.NewServeMux()
		mux.HandleFunc("/", handler)
		serv := &well.HTTPServer{
			Server: &http.Server{
				Addr:    config.listenAddr,
				Handler: mux,
			},
		}
		err := serv.ListenAndServe()
		if err != nil {
			return err
		}

		err = well.Wait()
		if err != nil && !well.IsSignaled(err) {
			return err
		}
		return nil
	},
}

// Execute executes the command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handler(rw http.ResponseWriter, req *http.Request) {
	env := well.NewEnvironment(req.Context())
	env.Go(monitorReady)
	env.Go(monitorHTTP)
	env.Go(monitorHTTPS)

	env.Stop()
	err := env.Wait()
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("ok"))
	return
}

func monitorReady(ctx context.Context) error {
	c := well.HTTPClient{
		Client: &http.Client{
			Timeout: config.timeout,
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", config.readyURL, nil)
	if err != nil {
		log.Error("failed to build HTTP request for readyURL", map[string]interface{}{
			log.FnError: err,
		})
		return err
	}
	resp, err := c.Do(req)
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

func monitorHTTP(ctx context.Context) error {
	c := well.HTTPClient{
		Client: &http.Client{
			Timeout: config.timeout,
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", config.httpURL, nil)
	if err != nil {
		log.Error("failed to build HTTP request for httpURL", map[string]interface{}{
			log.FnError: err,
		})
		return err
	}
	_, err = c.Do(req)
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

func monitorHTTPS(ctx context.Context) error {
	conn, err := net.DialTimeout("tcp", config.httpsAddr, config.timeout)
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
	fs.StringVar(&config.listenAddr, "listen-addr", ":8003", "Listen address for probe")
	fs.DurationVar(&config.timeout, "timeout", time.Second*5, "Timeout")
	fs.StringVar(&config.readyURL, "ready-url", "http://0.0.0.0:8002/ready", "URL of Envoy readiness probe")
	fs.StringVar(&config.httpURL, "http-url", "http://0.0.0.0:8080/", "URL for checking HTTP behavior")
	fs.StringVar(&config.httpsAddr, "https-addr", "0.0.0.0:8443", "Address for checking HTTPS behavior")
}
