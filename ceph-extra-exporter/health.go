package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// healthChecker reports whether every worker is still executing update().
//
// It deliberately ignores whether the commands of a worker succeed or not: a
// command which keeps failing is usually caused by a problem outside of this
// exporter (e.g. a Ceph cluster failure) and is not fixed by restarting the
// pod. Such a failure is exposed as the ceph_extra_failed_total metric
// instead. Only a worker stuck in update(), which a restart does fix, makes
// this check fail.
type healthChecker struct {
	executers []*cephExecuter
	threshold time.Duration
	// unhealthy holds the health state reported last, so that a liveness probe
	// running every few seconds logs only when the state changes.
	unhealthy atomic.Bool
}

func newHealthChecker(executers []*cephExecuter, threshold time.Duration) *healthChecker {
	return &healthChecker{
		executers: executers,
		threshold: threshold,
	}
}

func (hc *healthChecker) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	for _, executer := range hc.executers {
		elapsed := executer.timeSinceLastUpdate()
		if elapsed <= hc.threshold {
			continue
		}
		if hc.unhealthy.CompareAndSwap(false, true) {
			logger.Warn("worker seems to be stuck", "rule", executer.rule.name, "elapsed", elapsed.String())
		}
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(rw, "rule %q has not finished an update for %s\n", executer.rule.name, elapsed)
		return
	}

	if hc.unhealthy.CompareAndSwap(true, false) {
		logger.Info("all workers are alive again")
	}
	rw.WriteHeader(http.StatusOK)
}

// minHealthCheckThreshold returns the longest possible interval between two
// consecutive ends of update() while all the enabled workers are alive: the
// slowest update() plus the wait until the next tick.
func minHealthCheckThreshold(rules []rule, options exportOptions, executionInterval, commandTimeout time.Duration) time.Duration {
	maxUpdate := time.Duration(0)
	for _, r := range enabledRules(rules, options) {
		if d := r.maxUpdateDuration(commandTimeout); d > maxUpdate {
			maxUpdate = d
		}
	}
	return executionInterval + maxUpdate
}
