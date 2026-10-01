// Package cpupinning reports the CPUs exclusively allocated to each container by kubelet's CPU manager.
package cpupinning

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	podresourcesv1 "k8s.io/kubelet/pkg/apis/podresources/v1"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

const (
	defaultSocketPath = "/var/lib/kubelet/pod-resources/kubelet.sock"
	nodeNameEnv       = "NODE_NAME"
	listTimeout       = 10 * time.Second
)

type cpuPinningCollector struct {
	node       string
	socketPath string
}

var _ exporter.Collector = &cpuPinningCollector{}

// NewCollector returns a node-scope collector that reports exclusively allocated CPUs
// read from kubelet's PodResources API.
func NewCollector() exporter.Collector {
	return &cpuPinningCollector{socketPath: defaultSocketPath}
}

func (c *cpuPinningCollector) Scope() string {
	return constants.ScopeNode
}

func (c *cpuPinningCollector) MetricsPrefix() string {
	return "cpupinning"
}

func (c *cpuPinningCollector) IsLeaderMetrics() bool {
	return false
}

func (c *cpuPinningCollector) Setup(_ context.Context) error {
	node := os.Getenv(nodeNameEnv)
	if node == "" {
		return fmt.Errorf("%s environment variable is not set", nodeNameEnv)
	}
	c.node = node
	return nil
}

// Collect connects to the socket on every call because kubelet recreates it when it restarts.
// cpu_ids of List contains only exclusively allocated CPUs, so containers in the shared pool produce no metrics.
func (c *cpuPinningCollector) Collect(ctx context.Context) ([]*exporter.Metric, error) {
	ctx, cancel := context.WithTimeout(ctx, listTimeout)
	defer cancel()

	conn, err := grpc.NewClient("unix://"+c.socketPath, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create PodResources client for %s: %w", c.socketPath, err)
	}
	defer conn.Close()

	resp, err := podresourcesv1.NewPodResourcesListerClient(conn).List(ctx, &podresourcesv1.ListPodResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pod resources via %s: %w", c.socketPath, err)
	}

	var ret []*exporter.Metric
	for _, pod := range resp.GetPodResources() {
		for _, container := range pod.GetContainers() {
			for _, cpu := range container.GetCpuIds() {
				ret = append(ret, &exporter.Metric{
					Name:  "info",
					Value: 1,
					Labels: map[string]string{
						"node":             c.node,
						"pinned_namespace": pod.GetNamespace(),
						"pinned_pod":       pod.GetName(),
						"pinned_container": container.GetName(),
						"cpu":              strconv.FormatInt(cpu, 10),
					},
				})
			}
		}
	}
	return ret, nil
}
