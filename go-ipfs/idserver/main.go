package main

import (
	"flag"
	"net/http"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
)

var (
	flagListen = flag.String("listen", ":8000", "Listen address and port")
	flagDir    = flag.String("dir", "/doc", "Content Directory")
)

func main() {
	flag.Parse()
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir(*flagDir)))
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
