package main

import (
	"errors"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/cybozu-go/cke-tools/etcdbackup"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	yaml "gopkg.in/yaml.v2"
)

var flgConfig = flag.String("config", "", "path to configuration file")

func main() {
	flag.Parse()
	well.LogConfig{}.Apply()

	if *flgConfig == "" {
		log.ErrorExit(errors.New("usage: etcdbackup -config=<CONFIGFILE>"))
	}

	f, err := os.Open(*flgConfig)
	if err != nil {
		log.ErrorExit(err)
	}
	cfg := etcdbackup.NewConfig()
	err = yaml.NewDecoder(f).Decode(cfg)
	if err != nil {
		log.ErrorExit(err)
	}

	server := etcdbackup.NewServer(cfg)
	s := &well.HTTPServer{
		Server: &http.Server{
			Addr:    cfg.Listen,
			Handler: server,
		},
		ShutdownTimeout: 3 * time.Minute,
	}

	log.Info("started", map[string]interface{}{
		"listen": cfg.Listen,
	})

	err = s.ListenAndServe()
	if err != nil {
		log.ErrorExit(err)
	}

	err = well.Wait()
	if err != nil && !well.IsSignaled(err) {
		log.ErrorExit(err)
	}
}
