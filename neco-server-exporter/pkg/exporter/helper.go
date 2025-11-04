package exporter

import (
	"fmt"
	"maps"
	"slices"
)

func BuildMetricName(section, name string, labels map[string]string) string {
	lbls := ""
	if labels != nil {
		for _, k := range slices.Sorted(maps.Keys(labels)) {
			lbls = lbls + fmt.Sprintf(`,%s="%s"`, k, labels[k])
		}
		lbls = "{" + lbls[1:] + "}"
	}
	return fmt.Sprintf("neco_server_%s_%s%s", section, name, lbls)
}
