package render

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type File struct {
	Path    string
	Content []byte
}

type Options struct {
	PathPrefix          string
	ClusterName         string
	NamespaceTeamLabels map[string]string
}

type indexFile struct {
	GeneratedAt string       `json:"generatedAt"`
	ClusterName string       `json:"clusterName,omitempty"`
	Reports     []indexEntry `json:"reports"`
}

type indexEntry struct {
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	File            string `json:"file"`
	ResourceKind    string `json:"resourceKind,omitempty"`
	ResourceName    string `json:"resourceName,omitempty"`
	ResourceNS      string `json:"resourceNamespace,omitempty"`
	ContainerName   string `json:"containerName,omitempty"`
	ArtifactRepo    string `json:"artifactRepository,omitempty"`
	ArtifactTag     string `json:"artifactTag,omitempty"`
	ArtifactDigest  string `json:"artifactDigest,omitempty"`
	CreationTime    string `json:"creationTimestamp,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`

	NamespaceTeamLabel string `json:"namespaceTeamLabel,omitempty"`
}

var unsafePathChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func RenderSbomReports(reports []unstructured.Unstructured, opts Options) ([]File, error) {
	prefix := cleanPrefix(opts.PathPrefix)
	files := make([]File, 0, len(reports)+1)
	index := indexFile{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		ClusterName: opts.ClusterName,
		Reports:     make([]indexEntry, 0, len(reports)),
	}

	sort.Slice(reports, func(i, j int) bool {
		if reports[i].GetNamespace() == reports[j].GetNamespace() {
			return reports[i].GetName() < reports[j].GetName()
		}
		return reports[i].GetNamespace() < reports[j].GetNamespace()
	})

	for _, report := range reports {
		payload, err := payloadFor(report)
		if err != nil {
			return nil, fmt.Errorf("render %s/%s: %w", report.GetNamespace(), report.GetName(), err)
		}
		content, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshal %s/%s: %w", report.GetNamespace(), report.GetName(), err)
		}
		content = append(content, '\n')

		filePath := reportFilePath(prefix, opts.ClusterName, report)
		files = append(files, File{Path: filePath, Content: content})
		index.Reports = append(index.Reports, buildIndexEntry(report, filePath, opts))
	}

	content, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal index: %w", err)
	}
	content = append(content, '\n')
	indexPath := path.Join(prefix, "index.json")
	files = append(files, File{Path: indexPath, Content: content})
	return files, nil
}

func payloadFor(report unstructured.Unstructured) (any, error) {
	m, found, err := unstructured.NestedMap(report.Object, "report", "components")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("missing .report.components")
	}
	return m, nil
}

func buildIndexEntry(report unstructured.Unstructured, filePath string, opts Options) indexEntry {
	labels := report.GetLabels()
	ns := report.GetNamespace()

	entry := indexEntry{
		Namespace:          ns,
		Name:               report.GetName(),
		File:               filePath,
		ResourceKind:       labels["trivy-operator.resource.kind"],
		ResourceName:       labels["trivy-operator.resource.name"],
		ResourceNS:         labels["trivy-operator.resource.namespace"],
		ContainerName:      labels["trivy-operator.container.name"],
		CreationTime:       report.GetCreationTimestamp().Time.Format(time.RFC3339),
		ResourceVersion:    report.GetResourceVersion(),
		NamespaceTeamLabel: opts.NamespaceTeamLabels[ns],
	}

	if v, found, _ := unstructured.NestedString(report.Object, "report", "artifact", "repository"); found {
		entry.ArtifactRepo = v
	}
	if v, found, _ := unstructured.NestedString(report.Object, "report", "artifact", "tag"); found {
		entry.ArtifactTag = v
	}
	if v, found, _ := unstructured.NestedString(report.Object, "report", "artifact", "digest"); found {
		entry.ArtifactDigest = v
	}
	return entry
}

func reportFilePath(prefix, cluster string, report unstructured.Unstructured) string {
	ns := sanitize(report.GetNamespace())
	clusterName := sanitize(cluster)
	name := sanitize(report.GetName())
	if ns == "" {
		ns = "unknown-namespace"
	}
	return path.Join(prefix, clusterName, ns, name+".json")
}

func cleanPrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	prefix = strings.Trim(prefix, "/")
	if prefix == "" || prefix == "." {
		return ""
	}
	return path.Clean(prefix)
}

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	s = unsafePathChars.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-.")
	return s
}
