package exporter

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// labelValueEscaper escapes label values as defined by the Prometheus text exposition format.
//
// This is an additional layer of safety. Label values of most metrics come from Kubernetes
// resources (e.g. label values, names), which cannot contain these characters today.
// However, VictoriaMetrics/metrics validates the metric name string built here and panics
// on a malformed one, which would crash the whole exporter. Each of an unescaped `"`,
// a trailing `\`, or a line break in a value causes such a panic.
var labelValueEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`)

func BuildMetricName(scope, prefix, name string, labels map[string]string) string {
	lbls := ""
	if len(labels) > 0 {
		for _, k := range slices.Sorted(maps.Keys(labels)) {
			lbls = lbls + fmt.Sprintf(`,%s="%s"`, k, labelValueEscaper.Replace(labels[k]))
		}
		lbls = "{" + lbls[1:] + "}"
	}
	return fmt.Sprintf("neco_%s_%s_%s%s", scope, prefix, name, lbls)
}
