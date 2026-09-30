package exporter

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

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
