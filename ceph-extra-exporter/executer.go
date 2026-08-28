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
const commandTimeout time.Duration = 60 * time.Second

// defaultHealthCheckThreshold must be longer than
// minHealthCheckThreshold(rules, ...), which TestDefaultHealthCheckThreshold
// guards.
const defaultHealthCheckThreshold time.Duration = 2*executionInterval + commandTimeout

type metric struct {
	metricType prometheus.ValueType
	help       string
	jqFilter   string
	labelKeys  []string
}

type ruleTarget string

const (
	ruleTargetCommon ruleTarget = ""
	ruleTargetRGW    ruleTarget = "rgw"
	ruleTargetRBD    ruleTarget = "rbd"
)

type rule struct {
	name    string
	command []string
	target  ruleTarget
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
	rule              *rule
	executionInterval time.Duration
	commandTimeout    time.Duration
	metricValues      map[string][]metricValue
	mutex             sync.RWMutex
	failedCounter     map[string]int
	// lastUpdateTime is the time when update() finished last, regardless of
	// whether the commands in it succeeded or not. It is used to detect a
	// worker stuck in update().
	lastUpdateTime time.Time
}

func newExecuter(rule *rule, executionInterval, commandTimeout time.Duration) *cephExecuter {
	return &cephExecuter{
		rule:              rule,
		executionInterval: executionInterval,
		commandTimeout:    commandTimeout,
		metricValues:      make(map[string][]metricValue),
		failedCounter:     map[string]int{"command": 0, "jq": 0, "parse": 0},
		lastUpdateTime:    time.Now(),
	}
}

func (ce *cephExecuter) start(ctx context.Context) {
	ce.update(ctx)

	ticker := time.NewTicker(ce.executionInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			ce.update(ctx)
		}
	}
}

func (ce *cephExecuter) timeSinceLastUpdate() time.Duration {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()
	return time.Since(ce.lastUpdateTime)
}

func (ce *cephExecuter) incrementFailedCounter(reason string) {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()
	ce.failedCounter[reason] += 1
}

// maxUpdateDuration returns the longest time update() can take for this rule.
// update() runs the command of the rule once and one jq command per metric,
// each of which is aborted after commandTimeout.
func (r *rule) maxUpdateDuration(commandTimeout time.Duration) time.Duration {
	return time.Duration(1+len(r.metrics)) * commandTimeout
}

func (ce *cephExecuter) update(ctx context.Context) {
	logger.Info("starting update", "rule", ce.rule.name)
	values := make(map[string][]metricValue)

	defer func() {
		ce.mutex.Lock()
		defer ce.mutex.Unlock()
		ce.metricValues = values
		ce.lastUpdateTime = time.Now()
	}()

	jsonBytes, err := executeCommand(ctx, ce.rule.command, nil, ce.commandTimeout)
	if err != nil {
		logger.Warn("command execution failed", "command", ce.rule.command)
		ce.incrementFailedCounter("command")
		return
	}

	for name, metric := range ce.rule.metrics {
		result, err := executeCommand(ctx, []string{"jq", "-r", metric.jqFilter}, bytes.NewBuffer(jsonBytes), ce.commandTimeout)
		if err != nil {
			logger.Warn("jq command failed", "filter", metric.jqFilter)
			ce.incrementFailedCounter("jq")
			continue
		}

		mv := []metricValue{}
		if err := json.Unmarshal(result, &mv); err != nil {
			logger.Warn("parse value failed", "value", string(result), "error", err)
			ce.incrementFailedCounter("parse")
			continue
		}
		values[name] = mv
	}
}

func executeCommand(ctx context.Context, command []string, input io.Reader, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
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
				logger.Error("failed to io.Copy", "error", err)
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
