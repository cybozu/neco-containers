package exporter

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

func BuildMetricName(scope, prefix, name string, labels map[string]string) string {
	lbls := ""
	if len(labels) > 0 {
		for _, k := range slices.Sorted(maps.Keys(labels)) {
			v := labels[k]
			// Label values would need escaping in the Prometheus text format, but they come from
			// Kubernetes resources (e.g. label values, names), which cannot contain these characters.
			// Seeing one means something is critically wrong, so crash instead of exporting broken data.
			// Note that VictoriaMetrics/metrics does not reject all of them (e.g. `\` is accepted as-is).
			if strings.ContainsAny(v, "\\\"\n") {
				panic(fmt.Sprintf("BUG: label value contains a character that requires escaping: %s=%q", k, v))
			}
			lbls = lbls + fmt.Sprintf(`,%s="%s"`, k, v)
		}
		lbls = "{" + lbls[1:] + "}"
	}
	return fmt.Sprintf("neco_%s_%s_%s%s", scope, prefix, name, lbls)
}
