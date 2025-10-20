package components

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/VictoriaMetrics/metrics"
)

var (
	mu     sync.Mutex
	marked map[string]struct{}
)

func init() {
	marked = make(map[string]struct{})
}

func GetMetricsName(section, name string, labels map[string]string) string {
	lbls := ""
	if labels != nil {
		for _, k := range slices.Sorted(maps.Keys(labels)) {
			lbls = lbls + fmt.Sprintf(`,%s="%s"`, k, labels[k])
		}
		lbls = "{" + lbls[1:] + "}"
	}
	return fmt.Sprintf("neco_server_%s_%s%s", section, name, lbls)
}

// Mark all the metrices as stale in a section for future GC
func MarkMetricsSection(section string) {
	names := metrics.ListMetricNames()
	mu.Lock()
	defer mu.Unlock()

	prefix := "neco_server_" + section
	for _, name := range names {
		if strings.HasPrefix(name, prefix) {
			marked[name] = struct{}{}
		}
	}
}

// Unmark an active metrics as stale
func UnmarkMetrics(metricsName string) {
	mu.Lock()
	defer mu.Unlock()

	delete(marked, metricsName)
}

// Garbage-collect all the marked-as-stale metrices in a section
func CleanMetricsSection(section string) {
	mu.Lock()
	defer mu.Unlock()

	prefix := "neco_server_" + section
	for k := range marked {
		if strings.HasPrefix(k, prefix) {
			metrics.UnregisterMetric(k)
			delete(marked, k)
		}
	}
}
