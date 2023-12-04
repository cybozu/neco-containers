package main

import "github.com/cybozu-go/log"

var logger *log.Logger

func init() {
	logger = log.NewLogger()
	logger.SetFormatter(log.JSONFormat{})
}
