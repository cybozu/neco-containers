package main

import (
	"flag"
	"io"
	"net/http"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
)

var (
	flagListen = flag.String("listen", "0.0.0.0:8000", "Listen address and port")
)

func main() {
	flag.Parse()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello")
	})
	s := &well.HTTPServer{
		Server: &http.Server{
			Addr:    *flagListen,
			Handler: mux,
		},
	}
	log.Info("Start listening", map[string]interface{}{
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
