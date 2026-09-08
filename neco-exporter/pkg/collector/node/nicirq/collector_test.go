package nicirq

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

type fakeIRQ struct {
	irq       int
	actions   []string
	effective string // "" means effective_affinity_list is absent
}

func writeProc(t *testing.T, irqs []fakeIRQ) string {
	t.Helper()

	root := t.TempDir()
	irqRoot := filepath.Join(root, "irq")
	if err := os.MkdirAll(irqRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(irqRoot, "default_smp_affinity"), []byte("ffffffff,ffffffff\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(irqRoot, "not-an-irq"), 0o755); err != nil {
		t.Fatal(err)
	}

	for _, f := range irqs {
		dir := filepath.Join(irqRoot, strconv.Itoa(f.irq))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, a := range f.actions {
			if err := os.MkdirAll(filepath.Join(dir, a), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if f.effective != "" {
			if err := os.WriteFile(filepath.Join(dir, effectiveAffinityFile), []byte(f.effective), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return root
}

type sample struct {
	driver, device, queue string
	cpu                   float64
}

func split(t *testing.T, node string, ms []*exporter.Metric) ([]sample, map[string]float64) {
	t.Helper()

	var samples []sample
	skipped := map[string]float64{}
	for _, m := range ms {
		if m.Labels["node"] != node {
			t.Errorf("node label = %q, want %q", m.Labels["node"], node)
		}
		switch m.Name {
		case "queue_cpu":
			samples = append(samples, sample{
				driver: m.Labels["driver"],
				device: m.Labels["device"],
				queue:  m.Labels["queue"],
				cpu:    m.Value,
			})
		case "skipped_queues":
			if _, dup := skipped[m.Labels["reason"]]; dup {
				t.Errorf("skipped_queues{reason=%q} reported twice", m.Labels["reason"])
			}
			skipped[m.Labels["reason"]] = m.Value
		default:
			t.Errorf("unexpected metric name %q", m.Name)
		}
	}
	slices.SortFunc(samples, func(a, b sample) int {
		if c := strings.Compare(a.device, b.device); c != 0 {
			return c
		}
		qa, _ := strconv.Atoi(a.queue)
		qb, _ := strconv.Atoi(b.queue)
		return qa - qb
	})
	return samples, skipped
}

var noSkips = map[string]float64{skipReadError: 0, skipNotSingleCPU: 0, skipUnmatchedAction: 0}

func TestCollect(t *testing.T) {
	t.Parallel()

	const node = "10.69.6.203"

	tests := []struct {
		name        string
		irqs        []fakeIRQ
		want        []sample
		wantSkipped map[string]float64
	}{
		{
			name: "healthy spread",
			irqs: []fakeIRQ{
				{irq: 178, actions: []string{"ice-eno12399np0-TxRx-0"}, effective: "6\n"},
				{irq: 179, actions: []string{"ice-eno12399np0-TxRx-1"}, effective: "7\n"},
				{irq: 236, actions: []string{"ice-eno12399np0-TxRx-58"}, effective: "0\n"},
				{irq: 314, actions: []string{"ice-eno12409np1-TxRx-0"}, effective: "59\n"},
			},
			want: []sample{
				{"ice", "eno12399np0", "0", 6},
				{"ice", "eno12399np0", "1", 7},
				{"ice", "eno12399np0", "58", 0},
				{"ice", "eno12409np1", "0", 59},
			},
			wantSkipped: noSkips,
		},
		{
			name: "concentrated on cpu 0-7 is reported as-is",
			irqs: []fakeIRQ{
				{irq: 342, actions: []string{"ice-eno12399np0-TxRx-0"}, effective: "0\n"},
				{irq: 343, actions: []string{"ice-eno12399np0-TxRx-1"}, effective: "1\n"},
				{irq: 350, actions: []string{"ice-eno12399np0-TxRx-8"}, effective: "0\n"},
				{irq: 405, actions: []string{"ice-eno12399np0-TxRx-63"}, effective: "7\n"},
			},
			want: []sample{
				{"ice", "eno12399np0", "0", 0},
				{"ice", "eno12399np0", "1", 1},
				{"ice", "eno12399np0", "8", 0},
				{"ice", "eno12399np0", "63", 7},
			},
			wantSkipped: noSkips,
		},
		{
			name: "control queue keeps the pci address in device",
			irqs: []fakeIRQ{
				{irq: 242, actions: []string{"ice-0000:c4:00.0:ctrl-TxRx-0"}, effective: "2\n"},
			},
			want: []sample{
				{"ice", "0000:c4:00.0:ctrl", "0", 2},
			},
			wantSkipped: noSkips,
		},
		{
			name: "non-queue interrupts are ignored",
			irqs: []fakeIRQ{
				{irq: 340, actions: []string{"ice-0000:63:00.0:misc"}, effective: "3\n"},
				{irq: 341, actions: []string{"ice-0000:63:00.0:ll_ts"}, effective: "3\n"},
				{irq: 100, actions: []string{"nvme0q1"}, effective: "1\n"},
				{irq: 0, actions: []string{"timer"}, effective: "0\n"},
				{irq: 50, actions: nil, effective: "0\n"},
			},
			want:        nil,
			wantSkipped: noSkips,
		},
		{
			name: "anything but a single cpu is skipped and counted",
			irqs: []fakeIRQ{
				{irq: 178, actions: []string{"ice-eno12399np0-TxRx-0"}, effective: "0-7\n"},
				{irq: 179, actions: []string{"ice-eno12399np0-TxRx-1"}, effective: "\n"},
				{irq: 180, actions: []string{"ice-eno12399np0-TxRx-2"}, effective: "garbage\n"},
				{irq: 181, actions: []string{"ice-eno12399np0-TxRx-3"}, effective: "-1\n"},
				{irq: 182, actions: []string{"ice-eno12399np0-TxRx-4"}, effective: "9\n"},
			},
			want: []sample{
				{"ice", "eno12399np0", "4", 9},
			},
			wantSkipped: map[string]float64{skipReadError: 0, skipNotSingleCPU: 4, skipUnmatchedAction: 0},
		},
		{
			name: "missing effective_affinity_list is skipped, others still reported",
			irqs: []fakeIRQ{
				{irq: 178, actions: []string{"ice-eno12399np0-TxRx-0"}, effective: ""},
				{irq: 179, actions: []string{"ice-eno12399np0-TxRx-1"}, effective: "7\n"},
			},
			want: []sample{
				{"ice", "eno12399np0", "1", 7},
			},
			wantSkipped: map[string]float64{skipReadError: 1, skipNotSingleCPU: 0, skipUnmatchedAction: 0},
		},
		{
			name: "queue-like action with an unexpected name is counted, not dropped",
			irqs: []fakeIRQ{
				{irq: 178, actions: []string{"ICE-eno12399np0-TXRX-0"}, effective: "6\n"},
				{irq: 179, actions: []string{"ice-eno12399np0-TxRx-1-extra"}, effective: "7\n"},
				{irq: 180, actions: []string{"ice-eno12399np0-TxRx-2"}, effective: "8\n"},
			},
			want: []sample{
				{"ice", "eno12399np0", "2", 8},
			},
			wantSkipped: map[string]float64{skipReadError: 0, skipNotSingleCPU: 0, skipUnmatchedAction: 2},
		},
		{
			name: "two queue actions on one irq are both reported",
			irqs: []fakeIRQ{
				{irq: 178, actions: []string{"ice-eno12399np0-TxRx-0", "ice-eno12409np1-TxRx-0"}, effective: "6\n"},
			},
			want: []sample{
				{"ice", "eno12399np0", "0", 6},
				{"ice", "eno12409np1", "0", 6},
			},
			wantSkipped: noSkips,
		},
		{
			name:        "no interrupts at all",
			irqs:        nil,
			want:        nil,
			wantSkipped: noSkips,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &nicIRQCollector{node: node, procRoot: writeProc(t, tt.irqs), logger: slog.New(slog.DiscardHandler)}
			got, err := c.Collect(context.Background())
			if err != nil {
				t.Fatalf("Collect() error = %v, want nil", err)
			}
			samples, skipped := split(t, node, got)
			if !slices.Equal(samples, tt.want) {
				t.Errorf("Collect() queue_cpu = %v, want %v", samples, tt.want)
			}
			if len(skipped) != len(skipReasons) {
				t.Errorf("Collect() reported skipped_queues for %d reasons, want %d", len(skipped), len(skipReasons))
			}
			for reason, want := range tt.wantSkipped {
				if got := skipped[reason]; got != want {
					t.Errorf("Collect() skipped_queues{reason=%q} = %v, want %v", reason, got, want)
				}
			}
		})
	}
}

func TestSkipLogging(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	c := &nicIRQCollector{
		node:   "n",
		logger: slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}
	collect := func() {
		t.Helper()
		if _, err := c.Collect(context.Background()); err != nil {
			t.Fatalf("Collect() error = %v, want nil", err)
		}
	}
	lines := func() int { return bytes.Count(buf.Bytes(), []byte("\n")) }

	c.procRoot = writeProc(t, []fakeIRQ{{irq: 1, actions: []string{"ice-eno1-TxRx-0"}, effective: "3\n"}})
	collect()
	if got := lines(); got != 0 {
		t.Errorf("healthy first collection wrote %d log lines, want 0: %s", got, buf.String())
	}

	c.procRoot = writeProc(t, []fakeIRQ{{irq: 1, actions: []string{"ice-eno1-TxRx-0"}, effective: "0-7\n"}})
	collect()
	collect()
	collect()
	if got := lines(); got != 1 || !bytes.Contains(buf.Bytes(), []byte("level=WARN")) {
		t.Errorf("persistent skip wrote %d log lines, want exactly 1 WARN: %s", got, buf.String())
	}

	c.procRoot = writeProc(t, []fakeIRQ{
		{irq: 1, actions: []string{"ice-eno1-TxRx-0"}, effective: "0-7\n"},
		{irq: 2, actions: []string{"ice-eno1-TxRx-1"}, effective: "garbage\n"},
	})
	collect()
	if got := lines(); got != 2 {
		t.Errorf("changed skip counts wrote %d log lines in total, want 2: %s", got, buf.String())
	}

	c.procRoot = writeProc(t, []fakeIRQ{{irq: 1, actions: []string{"ice-eno1-TxRx-0"}, effective: "3\n"}})
	collect()
	collect()
	if got := lines(); got != 3 || !bytes.Contains(buf.Bytes(), []byte("level=INFO")) {
		t.Errorf("cleared skip wrote %d log lines in total, want 3 with an INFO: %s", got, buf.String())
	}
}

func TestCollectUnreadableIRQDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission bits are not enforced for root")
	}
	t.Parallel()

	root := writeProc(t, []fakeIRQ{
		{irq: 178, actions: []string{"ice-eno12399np0-TxRx-0"}, effective: "6\n"},
		{irq: 179, actions: []string{"ice-eno12399np0-TxRx-1"}, effective: "7\n"},
	})
	locked := filepath.Join(root, "irq", "178")
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

	c := &nicIRQCollector{node: "n", procRoot: root, logger: slog.New(slog.DiscardHandler)}
	got, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() error = %v, want nil", err)
	}
	samples, skipped := split(t, "n", got)
	want := []sample{{"ice", "eno12399np0", "1", 7}}
	if !slices.Equal(samples, want) {
		t.Errorf("Collect() queue_cpu = %v, want %v", samples, want)
	}
	if skipped[skipReadError] != 1 {
		t.Errorf("Collect() skipped_queues{reason=%q} = %v, want 1", skipReadError, skipped[skipReadError])
	}
}

func TestCollectWithoutProcIRQ(t *testing.T) {
	t.Parallel()

	c := &nicIRQCollector{node: "n", procRoot: t.TempDir(), logger: slog.New(slog.DiscardHandler)}
	if _, err := c.Collect(context.Background()); err == nil {
		t.Fatal("Collect() error = nil without /proc/irq, want error")
	}
}

func TestSetup(t *testing.T) {
	t.Run("node name from env", func(t *testing.T) {
		t.Setenv(nodeNameEnv, "10.69.6.203")

		c := &nicIRQCollector{procRoot: defaultProcRoot}
		if err := c.Setup(context.Background()); err != nil {
			t.Fatalf("Setup() error = %v, want nil", err)
		}
		if c.node != "10.69.6.203" {
			t.Errorf("Setup() node = %q, want %q", c.node, "10.69.6.203")
		}
	})

	t.Run("missing node name", func(t *testing.T) {
		t.Setenv(nodeNameEnv, "")

		c := &nicIRQCollector{procRoot: defaultProcRoot}
		if err := c.Setup(context.Background()); err == nil {
			t.Fatal("Setup() error = nil without NODE_NAME, want error")
		}
	})
}

func TestCollectorMetadata(t *testing.T) {
	t.Parallel()

	c := NewCollector()
	if got := c.Scope(); got != "node" {
		t.Errorf("Scope() = %q, want %q", got, "node")
	}
	if got := c.MetricsPrefix(); got != "nicirq" {
		t.Errorf("MetricsPrefix() = %q, want %q", got, "nicirq")
	}
	if c.IsLeaderMetrics() {
		t.Error("IsLeaderMetrics() = true, want false")
	}
}
