package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
)

var flagListen = flag.String("listen", ":8000", "Listen address and port")

func main() {
	flag.Parse()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if s := q.Get("sleep"); s != "" {
			d, err := time.ParseDuration(s)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, "Please specify valid time")
				return
			}
			time.Sleep(d)
			_, _ = io.WriteString(w, fmt.Sprintf("Hello after sleeping %s", s))
			return
		}
		_, _ = io.WriteString(w, "Hello")
	})
	s := &well.HTTPServer{
		Server: &http.Server{
			Addr:    *flagListen,
			Handler: mux,
		},
	}
	_ = log.Info("Start listening", map[string]any{
		log.FnHTTPHost: *flagListen,
	})

	err := s.ListenAndServe()
	if err != nil {
		log.ErrorExit(err)
	}

	err = well.Wait()
	if err != nil && !well.IsSignaled(err) {
		log.ErrorExit(err)
	}
}
