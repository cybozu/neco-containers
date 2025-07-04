package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/cybozu-go/well"
)

var (
	flagListen = flag.String("listen", ":8000", "Listen address and port")
)

func main() {
	flag.Parse()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if s := q.Get("sleep"); s != "" {
			d, err := time.ParseDuration(s)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, "Please specify valid time")
				return
			}
			time.Sleep(d)
			io.WriteString(w, fmt.Sprintf("Hello after sleeping %s", s))
			return
		}
		io.WriteString(w, "Hello")
	})
	s := &well.HTTPServer{
		Server: &http.Server{
			Addr:    *flagListen,
			Handler: mux,
		},
	}
	logger.Info("starting server", "http_host", *flagListen)

	err := s.ListenAndServe()
	if err != nil {
		logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}

	err = well.Wait()
	if err != nil && !well.IsSignaled(err) {
		logger.Error("failed to wait for shutdown", "error", err)
		os.Exit(1)
	}
}
