package bpf

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strconv"

	"github.com/cilium/cilium/pkg/client"
	"github.com/cilium/ebpf"
	"github.com/cybozu/neco-containers/neco-server-exporter/pkg/exporter"
)

type collector struct {
	ciliumClient *client.Client
}

var _ exporter.Collector = &collector{}

func NewCollector() exporter.Collector {
	return &collector{}
}

func (c *collector) SectionName() string {
	return "bpf"
}

func (c *collector) Collect(ctx context.Context) ([]*exporter.Metric, error) {
	if err := CheckBPFStatsEnabled(); err != nil {
		return nil, err
	}
	if c.ciliumClient == nil {
		var err error
		if c.ciliumClient, err = client.NewClient(""); err != nil {
			return nil, fmt.Errorf("failed to open Cilium socket: %w", err)
		}
	}

	tcxMeta, err := CollectTCXMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to read TCX info: %w", err)
	}

	endpointMeta, err := CollectEndpointMetadata(c.ciliumClient)
	if err != nil {
		// maybe disconnected? connect again on the next iteration
		c.ciliumClient = nil
		return nil, fmt.Errorf("failed to read from Cilium socket: %w", err)
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

		prog, err := ebpf.NewProgramFromID(id)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			// possibly asynchronously removed, continue to the next one
			continue ProgramLoop
		case err != nil:
			return nil, fmt.Errorf("failed to open BPF Program: %w", err)
		}
		defer prog.Close()

		// skip programs without execution
		stats, err := prog.Stats()
		if err != nil {
			return nil, fmt.Errorf("failed to get BPF stats: %w", err)
		}
		if stats.RunCount == 0 {
			continue ProgramLoop
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

		// run_time_seconds_total
		m := &exporter.Metric{
			Name:   "run_time_seconds_total",
			Value:  stats.Runtime.Seconds(),
			Labels: labels,
		}
		ret = append(ret, m)

		// run_count_total
		m = &exporter.Metric{
			Name:   "run_count_total",
			Value:  float64(stats.RunCount),
			Labels: labels,
		}
		ret = append(ret, m)

		// interrupt program enumeration
		select {
		case <-ctx.Done():
			return nil, errors.New("context is done")
		default:
		}
	}
	return ret, nil
}
