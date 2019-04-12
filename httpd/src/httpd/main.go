package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
)

var (
	flagPort    = flag.Int("p", 8000, "Listen port")
	flagAddress = flag.String("a", "localhost", "Listen address")
)

func main() {
	flag.Parse()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello")
	})
	addr := fmt.Sprintf("%s:%d", *flagAddress, *flagPort)
	s := &well.HTTPServer{
		Server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
	log.Info("Start listening", map[string]interface{}{
		log.FnHTTPHost: addr,
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
