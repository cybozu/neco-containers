package main

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cybozu-go/log"
	"github.com/prometheus/client_golang/prometheus"
)

const executionInterval time.Duration = 300 * time.Second

type metric struct {
	metricType prometheus.ValueType
	help       string
	jqFilter   string
}

type rule struct {
	name    string
	command []string
	metrics map[string]metric
}

type cephExecuter struct {
	rule          *rule
	values        map[string]float64
	mutex         sync.RWMutex
	failedCounter map[string]int
}

func newExecuter(rule *rule) *cephExecuter {
	return &cephExecuter{
		rule:          rule,
		values:        make(map[string]float64),
		failedCounter: map[string]int{"command": 0, "jq": 0, "parse": 0},
	}
}

func (ce *cephExecuter) start(ctx context.Context) {
	ce.update()

	ticker := time.NewTicker(executionInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			ce.update()
		}
	}
}

func (ce *cephExecuter) update() {
	values := make(map[string]float64)

	defer func() {
		ce.mutex.Lock()
		defer ce.mutex.Unlock()
		ce.values = values
	}()

	json, err := executeCommand(ce.rule.command, nil)
	if err != nil {
		_ = logger.Warn("command execution failed", map[string]interface{}{
			"command": ce.rule.command,
		})
		ce.failedCounter["command"] += 1
		return
	}

	for name, metric := range ce.rule.metrics {
		result, err := executeCommand([]string{"jq", "-r", metric.jqFilter}, bytes.NewBuffer(json))
		if err != nil {
			_ = logger.Warn("jq command failed", map[string]interface{}{
				"filter": metric.jqFilter,
			})
			ce.failedCounter["jq"] += 1
			continue
		}

		value, err := strconv.ParseFloat(strings.TrimSpace(string(result)), 64)
		if err != nil {
			_ = logger.Warn("parse value failed", map[string]interface{}{
				"value": string(result),
			})
			ce.failedCounter["parse"] += 1
			continue
		}
		values[name] = value
	}
}

func executeCommand(command []string, input io.Reader) ([]byte, error) {
	cmd := exec.Command(command[0], command[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer stdout.Close()

	if input != nil {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		go func() {
			defer stdin.Close()
			if _, err = io.Copy(stdin, input); err != nil {
				_ = logger.Error("failed to io.Copy", map[string]interface{}{log.FnError: err})
			}
		}()
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	r, err := io.ReadAll(stdout)
	if err != nil {
		return r, err
	}

	if err := cmd.Wait(); err != nil {
		return r, err
	}

	return r, nil
}
