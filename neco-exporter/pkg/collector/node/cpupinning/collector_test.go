package cpupinning

import (
	"cmp"
	"context"
	"errors"
	"net"
	"path/filepath"
	"slices"
	"testing"

	"google.golang.org/grpc"
	podresourcesv1 "k8s.io/kubelet/pkg/apis/podresources/v1"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

type fakeServer struct {
	podresourcesv1.UnimplementedPodResourcesListerServer
	resp *podresourcesv1.ListPodResourcesResponse
	err  error
}

func (s *fakeServer) List(context.Context, *podresourcesv1.ListPodResourcesRequest) (*podresourcesv1.ListPodResourcesResponse, error) {
	return s.resp, s.err
}

func startServer(t *testing.T, srv *fakeServer) string {
	t.Helper()

	socketPath := filepath.Join(t.TempDir(), "kubelet.sock")
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}

	s := grpc.NewServer()
	podresourcesv1.RegisterPodResourcesListerServer(s, srv)
	go func() {
		_ = s.Serve(lis)
	}()
	t.Cleanup(s.Stop)

	return socketPath
}

func pod(namespace, name string, containers ...*podresourcesv1.ContainerResources) *podresourcesv1.PodResources {
	return &podresourcesv1.PodResources{Namespace: namespace, Name: name, Containers: containers}
}

func container(name string, cpus ...int64) *podresourcesv1.ContainerResources {
	return &podresourcesv1.ContainerResources{Name: name, CpuIds: cpus}
}

type sample struct {
	namespace, pod, container, cpu string
}

func toSamples(t *testing.T, node string, ms []*exporter.Metric) []sample {
	t.Helper()

	ret := make([]sample, 0, len(ms))
	for _, m := range ms {
		if m.Name != "info" {
			t.Errorf("unexpected metric name %q", m.Name)
		}
		if m.Value != 1 {
			t.Errorf("value = %v, want 1", m.Value)
		}
		if m.Labels["node"] != node {
			t.Errorf("node label = %q, want %q", m.Labels["node"], node)
		}
		if len(m.Labels) != 5 {
			t.Errorf("unexpected labels %v", m.Labels)
		}
		ret = append(ret, sample{
			namespace: m.Labels["pinned_namespace"],
			pod:       m.Labels["pinned_pod"],
			container: m.Labels["pinned_container"],
			cpu:       m.Labels["cpu"],
		})
	}
	slices.SortFunc(ret, func(a, b sample) int {
		return cmp.Or(
			cmp.Compare(a.namespace, b.namespace),
			cmp.Compare(a.pod, b.pod),
			cmp.Compare(a.container, b.container),
			cmp.Compare(a.cpu, b.cpu),
		)
	})
	return ret
}

func TestCollect(t *testing.T) {
	t.Parallel()

	const node = "10.69.6.203"

	tests := []struct {
		name string
		pods []*podresourcesv1.PodResources
		want []sample
	}{
		{
			name: "multiple cpus",
			pods: []*podresourcesv1.PodResources{
				pod("app-a", "moco-db-0", container("mysqld", 4, 5, 68, 69)),
			},
			want: []sample{
				{"app-a", "moco-db-0", "mysqld", "4"},
				{"app-a", "moco-db-0", "mysqld", "5"},
				{"app-a", "moco-db-0", "mysqld", "68"},
				{"app-a", "moco-db-0", "mysqld", "69"},
			},
		},
		{
			name: "shared pool containers are omitted",
			pods: []*podresourcesv1.PodResources{
				pod("app-a", "moco-db-0", container("mysqld", 4), container("agent")),
				pod("default", "burstable", container("main")),
			},
			want: []sample{
				{"app-a", "moco-db-0", "mysqld", "4"},
			},
		},
		{
			name: "multiple pinned containers and pods",
			pods: []*podresourcesv1.PodResources{
				pod("app-a", "moco-db-0", container("mysqld", 4, 5), container("sidecar", 6)),
				pod("app-b", "moco-db-1", container("mysqld", 10)),
			},
			want: []sample{
				{"app-a", "moco-db-0", "mysqld", "4"},
				{"app-a", "moco-db-0", "mysqld", "5"},
				{"app-a", "moco-db-0", "sidecar", "6"},
				{"app-b", "moco-db-1", "mysqld", "10"},
			},
		},
		{
			name: "no pods",
			want: []sample{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			socketPath := startServer(t, &fakeServer{
				resp: &podresourcesv1.ListPodResourcesResponse{PodResources: tt.pods},
			})
			c := &cpuPinningCollector{node: node, socketPath: socketPath}

			ms, err := c.Collect(t.Context())
			if err != nil {
				t.Fatalf("Collect() error = %v", err)
			}
			if got := toSamples(t, node, ms); !slices.Equal(got, tt.want) {
				t.Errorf("Collect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		socketPath func(t *testing.T) string
	}{
		{
			name: "List fails",
			socketPath: func(t *testing.T) string {
				return startServer(t, &fakeServer{err: errors.New("boom")})
			},
		},
		{
			name: "socket does not exist",
			socketPath: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "kubelet.sock")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &cpuPinningCollector{node: "10.69.6.203", socketPath: tt.socketPath(t)}
			if _, err := c.Collect(t.Context()); err == nil {
				t.Fatal("Collect() error = nil, want error")
			}
		})
	}
}
