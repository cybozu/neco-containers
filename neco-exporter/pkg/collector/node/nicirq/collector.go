// Package nicirq reports which CPU handles each NIC queue interrupt.
package nicirq

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

const (
	defaultProcRoot       = "/proc"
	nodeNameEnv           = "NODE_NAME"
	effectiveAffinityFile = "effective_affinity_list"
)

// Label values of skipped_queues.
const (
	skipReadError       = "read_error"
	skipNotSingleCPU    = "not_single_cpu"
	skipUnmatchedAction = "unmatched_action"
)

var skipReasons = []string{skipReadError, skipNotSingleCPU, skipUnmatchedAction}

// actionPattern matches e.g. "ice-eno12399np0-TxRx-17" and "ice-0000:c4:00.0:ctrl-TxRx-0".
var (
	actionPattern    = regexp.MustCompile(`^([a-z0-9_]+)-(.+)-TxRx-([0-9]+)$`)
	queueLikePattern = regexp.MustCompile(`(?i)txrx`)
)

type nicIRQCollector struct {
	node        string
	procRoot    string
	logger      *slog.Logger
	lastSkipped map[string]int
}

var _ exporter.Collector = &nicIRQCollector{}

type queueIRQ struct {
	irq                   int
	action                string
	driver, device, queue string
}

type skip struct {
	q      queueIRQ
	reason string
	err    error
}

// NewCollector returns a node-scope collector that reports the CPU handling
// each NIC queue interrupt, read from the host's /proc/irq.
func NewCollector() exporter.Collector {
	return &nicIRQCollector{procRoot: defaultProcRoot, logger: slog.Default()}
}

func (c *nicIRQCollector) Scope() string {
	return constants.ScopeNode
}

func (c *nicIRQCollector) MetricsPrefix() string {
	return "nicirq"
}

func (c *nicIRQCollector) IsLeaderMetrics() bool {
	return false
}

func (c *nicIRQCollector) Setup(_ context.Context) error {
	node := os.Getenv(nodeNameEnv)
	if node == "" {
		return fmt.Errorf("%s environment variable is not set", nodeNameEnv)
	}
	c.node = node
	return nil
}

// Collect walks /proc/irq rather than reading /proc/interrupts, which is
// empty inside containers. Only an unreadable /proc/irq is an error; a queue
// interrupt that cannot be attributed to one CPU is counted in skipped_queues.
func (c *nicIRQCollector) Collect(ctx context.Context) ([]*exporter.Metric, error) {
	irqRoot := filepath.Join(c.procRoot, "irq")
	entries, err := os.ReadDir(irqRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to list %s: %w", irqRoot, err)
	}

	var ret []*exporter.Metric
	var skipped []skip
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		irq, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		irqDir := filepath.Join(irqRoot, e.Name())

		queues, unmatched, err := c.listQueueIRQs(irqDir, irq)
		if err != nil {
			skipped = append(skipped, skip{q: queueIRQ{irq: irq}, reason: skipReadError, err: err})
			continue
		}
		skipped = append(skipped, unmatched...)
		for _, q := range queues {
			m, sk := c.collectQueue(irqDir, q)
			if sk != nil {
				skipped = append(skipped, *sk)
				continue
			}
			ret = append(ret, m)
		}
	}

	ret = append(ret, c.skippedMetrics(ctx, skipped)...)
	return ret, nil
}

func (c *nicIRQCollector) listQueueIRQs(irqDir string, irq int) ([]queueIRQ, []skip, error) {
	actions, err := os.ReadDir(irqDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list %s: %w", irqDir, err)
	}

	var queues []queueIRQ
	var unmatched []skip
	for _, a := range actions {
		if !a.IsDir() {
			continue
		}
		name := a.Name()
		m := actionPattern.FindStringSubmatch(name)
		if m == nil {
			if queueLikePattern.MatchString(name) {
				unmatched = append(unmatched, skip{
					q:      queueIRQ{irq: irq, action: name},
					reason: skipUnmatchedAction,
					err:    fmt.Errorf("action %q does not match %s", name, actionPattern),
				})
			}
			continue
		}
		queues = append(queues, queueIRQ{irq: irq, action: name, driver: m[1], device: m[2], queue: m[3]})
	}
	return queues, unmatched, nil
}

func (c *nicIRQCollector) collectQueue(irqDir string, q queueIRQ) (*exporter.Metric, *skip) {
	data, err := os.ReadFile(filepath.Join(irqDir, effectiveAffinityFile))
	if err != nil {
		return nil, &skip{q: q, reason: skipReadError, err: err}
	}

	// On x86 a vector is bound to exactly one CPU, so the cpulist is a single number.
	cpu, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || cpu < 0 {
		return nil, &skip{q: q, reason: skipNotSingleCPU, err: fmt.Errorf("%s is %q, not a single cpu", effectiveAffinityFile, strings.TrimSpace(string(data)))}
	}

	return &exporter.Metric{
		Name:  "queue_cpu",
		Value: float64(cpu),
		Labels: map[string]string{
			"node":   c.node,
			"driver": q.driver,
			"device": q.device,
			"queue":  q.queue,
		},
	}, nil
}

func (c *nicIRQCollector) skippedMetrics(ctx context.Context, skipped []skip) []*exporter.Metric {
	counts := make(map[string]int, len(skipReasons))
	for _, s := range skipped {
		counts[s.reason]++
	}
	c.logSkipChanges(ctx, counts, skipped)

	ret := make([]*exporter.Metric, 0, len(skipReasons))
	for _, reason := range skipReasons {
		ret = append(ret, &exporter.Metric{
			Name:  "skipped_queues",
			Value: float64(counts[reason]),
			Labels: map[string]string{
				"node":   c.node,
				"reason": reason,
			},
		})
	}
	return ret
}

// logSkipChanges logs only when the counts differ from the previous collection.
func (c *nicIRQCollector) logSkipChanges(ctx context.Context, counts map[string]int, skipped []skip) {
	changed := false
	for _, reason := range skipReasons {
		if counts[reason] != c.lastSkipped[reason] {
			changed = true
			break
		}
	}
	c.lastSkipped = counts
	if !changed {
		return
	}

	attrs := []any{slog.Int("skipped", len(skipped))}
	for _, reason := range skipReasons {
		attrs = append(attrs, slog.Int(reason, counts[reason]))
	}
	if len(skipped) == 0 {
		c.logger.InfoContext(ctx, "all queue interrupts are reported in nicirq metrics again", attrs...)
		return
	}
	first := skipped[0]
	attrs = append(attrs,
		slog.Int("first_irq", first.q.irq),
		slog.String("first_action", first.q.action),
		slog.Any("first_error", first.err))
	c.logger.WarnContext(ctx, "some queue interrupts are left out of nicirq metrics", attrs...)
}
