package bpf

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strconv"

	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/client"
	"github.com/cilium/ebpf"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

type bpfCollector struct {
	ciliumClient *client.Client
}

var _ exporter.Collector = &bpfCollector{}

func NewCollector() exporter.Collector {
	return &bpfCollector{}
}

func (c *bpfCollector) Scope() string {
	return constants.ScopeNode
}

func (c *bpfCollector) MetricsPrefix() string {
	return "bpf"
}

func (c *bpfCollector) Setup() error {
	var cli *client.Client
	var err error
	if cli, err = client.NewClient(""); err != nil {
		return fmt.Errorf("failed to open Cilium socket: %w", err)
	}
	c.ciliumClient = cli
	return nil
}

func (c *bpfCollector) collectProgramMetrics(
	id ebpf.ProgramID,
	tcxMeta map[ebpf.ProgramID]TCXMetadata, endpointMeta map[uint32]*models.Endpoint,
) ([]*exporter.Metric, error) {

	prog, err := ebpf.NewProgramFromID(id)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// possibly asynchronously removed, considered normal
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("failed to open BPF Program: %w", err)
	}
	defer prog.Close()

	// omit metrices for programs without execution
	stats, err := prog.Stats()
	if err != nil {
		return nil, fmt.Errorf("failed to get BPF stats: %w", err)
	}
	if stats.RunCount == 0 {
		return nil, nil
	}

	info, err := prog.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get BPF program info: %w", err)
	}

	labels := map[string]string{
		"id":   strconv.Itoa(int(id)),
		"type": info.Type.String(),
		"name": info.Name,
	}
	if n, err := GetLongProgramName(info); err == nil {
		labels["name"] = n
	}

	if meta, ok := tcxMeta[id]; ok {
		labels["ifindex"] = strconv.Itoa(int(meta.Ifindex))
		labels["direction"] = meta.Direction
		if ep, ok := endpointMeta[meta.Ifindex]; ok {
			// reading deprecated fields, but think it later
			// https://github.com/cilium/cilium/pull/26894
			labels["namespace"] = ep.Status.ExternalIdentifiers.K8sNamespace
			labels["pod"] = ep.Status.ExternalIdentifiers.K8sPodName
			labels["container"] = ep.Status.ExternalIdentifiers.ContainerName
		}
	}

	timeMetric := &exporter.Metric{
		Name:   "run_time_seconds_total",
		Value:  stats.Runtime.Seconds(),
		Labels: labels,
	}

	countMetric := &exporter.Metric{
		Name:   "run_count_total",
		Value:  float64(stats.RunCount),
		Labels: labels,
	}

	return []*exporter.Metric{timeMetric, countMetric}, nil
}

func (c *bpfCollector) Collect(ctx context.Context) ([]*exporter.Metric, error) {
	if err := CheckBPFStatsEnabled(); err != nil {
		return nil, err
	}

	tcxMeta, err := CollectTCXMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to read TCX info: %w", err)
	}

	endpointMeta, err := CollectEndpointMetadata(c.ciliumClient)
	if err != nil {
		// maybe disconnected? connect again
		if err := c.Setup(); err != nil {
			return nil, fmt.Errorf("failed to reopen Cilium socket: %w", err)
		}
		// retry
		endpointMeta, err = CollectEndpointMetadata(c.ciliumClient)
		if err != nil {
			return nil, fmt.Errorf("failed to read from Cilium socket: %w", err)
		}
	}

	ret := make([]*exporter.Metric, 0)
	var id ebpf.ProgramID

ProgramLoop:
	for {
		id, err = ebpf.ProgramGetNextID(id)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			// no next program, finish iterating
			break ProgramLoop
		case err != nil:
			return nil, fmt.Errorf("failed to iterate BPF Program: %w", err)
		}

		sub, err := c.collectProgramMetrics(id, tcxMeta, endpointMeta)
		if err != nil {
			return nil, err
		}
		ret = append(ret, sub...)

		// interrupt program enumeration
		select {
		case <-ctx.Done():
			return nil, errors.New("context is done")
		default:
		}
	}
	return ret, nil
}
