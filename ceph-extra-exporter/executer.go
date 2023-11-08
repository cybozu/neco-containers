package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const executionInterval time.Duration = 300 * time.Second

type metric struct {
	metricType prometheus.ValueType
	help       string
	jqFilter   string
	labelKeys  []string
}

type rule struct {
	name    string
	command []string
	metrics map[string]metric
}

type metricValue struct {
	labelValues []string
	value       float64
}

func (mv *metricValue) UnmarshalJSON(b []byte) error {
	var x struct {
		LabelValues []string `json:"labels"`
		Value       *float64 `json:"value"`
	}
	err := json.Unmarshal(b, &x)
	if err != nil {
		return err
	}
	if x.LabelValues == nil {
		return errors.New("no labels found")
	}
	if x.Value == nil {
		return errors.New("no value found")
	}
	mv.labelValues = x.LabelValues
	mv.value = *x.Value
	return nil
}

type cephExecuter struct {
	rule          *rule
	metricValues  map[string][]metricValue
	mutex         sync.RWMutex
	failedCounter map[string]int
}

func newExecuter(rule *rule) *cephExecuter {
	return &cephExecuter{
		rule:          rule,
		metricValues:  make(map[string][]metricValue),
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
	values := make(map[string][]metricValue)

	defer func() {
		ce.mutex.Lock()
		defer ce.mutex.Unlock()
		ce.metricValues = values
	}()

	jsonBytes, err := executeCommand(ce.rule.command, nil)
	if err != nil {
		_ = logger.Warn("command execution failed", map[string]interface{}{
			"command": ce.rule.command,
		})
		ce.failedCounter["command"] += 1
		return
	}

	for name, metric := range ce.rule.metrics {
		result, err := executeCommand([]string{"jq", "-r", metric.jqFilter}, bytes.NewBuffer(jsonBytes))
		if err != nil {
			_ = logger.Warn("jq command failed", map[string]interface{}{
				"filter": metric.jqFilter,
			})
			ce.failedCounter["jq"] += 1
			continue
		}

		mv := []metricValue{}
		if err := json.Unmarshal(result, &mv); err != nil {
			_ = logger.Warn("parse value failed", map[string]interface{}{
				"value": string(result),
				"error": err,
			})
			ce.failedCounter["parse"] += 1
			continue
		}
		values[name] = mv
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
				_ = logger.Error("failed to io.Copy", map[string]interface{}{"error": err})
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
