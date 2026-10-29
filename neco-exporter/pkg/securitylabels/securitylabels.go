// Package securitylabels handles security labels of Cilium identities.
// It is shared by collectors that need to expose the same identity, so that their metrics can be joined.
package securitylabels

import (
	"slices"
	"strings"
)

// Format formats security labels in the same way as Hubble metrics with the "identity" context,
// i.e. sorted "<source>:<key>=<value>" (or "<source>:<key>" if the value is empty) joined by commas.
// The keys are "<source>:<key>", the same as the keys of CiliumIdentity's security-labels.
func Format(securityLabels map[string]string) string {
	entries := make([]string, 0, len(securityLabels))
	for k, v := range securityLabels {
		if v == "" {
			entries = append(entries, k)
		} else {
			entries = append(entries, k+"="+v)
		}
	}
	slices.Sort(entries)
	return strings.Join(entries, ",")
}

// Parse converts labels formatted by Cilium's Label.String() (e.g. "k8s:app=foo", or "k8s:app"
// for an empty value) into the map accepted by Format.
// Cilium's endpoint API reports identity labels in this form.
// The value starts after the first "=", which cannot appear in a source or a Kubernetes label key.
func Parse(labels []string) map[string]string {
	ret := make(map[string]string, len(labels))
	for _, l := range labels {
		k, v, _ := strings.Cut(l, "=")
		ret[k] = v
	}
	return ret
}
