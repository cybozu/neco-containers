package components

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/cilium/cilium/api/v1/client/endpoint"
	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/client"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/btf"
	"github.com/cilium/ebpf/link"
)

func CheckBPFStatsEnabled() error {
	flg, err := os.ReadFile("/proc/sys/kernel/bpf_stats_enabled")
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(flg)) == "0" {
		return errors.New("BPF stats is not enabled")
	}
	return nil
}

type TCXMetadata struct {
	Ifindex   uint32
	Direction string
}

func CollectTCXMetadata() (map[ebpf.ProgramID]TCXMetadata, error) {
	var it link.Iterator
	defer it.Close()
	ret := make(map[ebpf.ProgramID]TCXMetadata)
	for it.Next() {
		li := it.Take()
		defer li.Close()

		info, err := li.Info()
		if err != nil {
			return nil, err
		}

		tcx := info.TCX()
		if tcx == nil {
			continue
		}

		direction := "unknown"
		switch ebpf.AttachType(tcx.AttachType) {
		case ebpf.AttachTCXIngress:
			direction = "ingress"
		case ebpf.AttachTCXEgress:
			direction = "egress"
		}

		ret[info.Program] = TCXMetadata{
			Ifindex:   tcx.Ifindex,
			Direction: direction,
		}
	}
	return ret, nil
}

func CollectEndpointMetadata(client *client.Client) (map[uint32]*models.Endpoint, error) {
	ret := make(map[uint32]*models.Endpoint)
	params := &endpoint.GetEndpointParams{}

	response, err := client.Endpoint.GetEndpoint(params)
	if err != nil {
		return nil, err
	}

	for _, ep := range response.Payload {
		ifindex := uint32(ep.Status.Networking.InterfaceIndex)
		ret[ifindex] = ep
	}
	return ret, nil
}

func GetLongProgramName(info *ebpf.ProgramInfo) (string, error) {
	id, ok := info.BTFID()
	if !ok {
		return "", errors.New("no BTFID found")
	}

	handle, err := btf.NewHandleFromID(id)
	if err != nil {
		return "", err
	}
	defer handle.Close()

	spec, err := handle.Spec(nil)
	if err != nil {
		return "", err
	}

	li := slices.Collect(maps.Keys(maps.Collect(spec.All())))
	li = slices.DeleteFunc(li, func(t btf.Type) bool {
		_, ok := t.(*btf.Func)
		return !ok
	})

	switch len(li) {
	case 1:
		return li[0].(*btf.Func).Name, nil
	default:
		return "", errors.New("unsupported BTF info")
	}
}

func StartBPFPerformanceExporter(ctx context.Context, interval time.Duration) error {
	const section = "bpf"
	if err := CheckBPFStatsEnabled(); err != nil {
		return err
	}

	ciliumClient, err := client.NewClient("")
	if err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	for {
		var id ebpf.ProgramID
		var err error
		MarkMetricsSection(section)
		labels := make(map[string]string)

		tcxMeta, err := CollectTCXMetadata()
		if err != nil {
			return err
		}

		endpointMeta, err := CollectEndpointMetadata(ciliumClient)
		if err != nil {
			return err
		}

	ProgramLoop:
		for {
			id, err = ebpf.ProgramGetNextID(id)
			switch {
			case errors.Is(err, fs.ErrNotExist):
				// no next program, finish iterating
				break ProgramLoop
			case err != nil:
				return fmt.Errorf("failed to iterate BPF Program: %w", err)
			}

			prog, err := ebpf.NewProgramFromID(id)
			switch {
			case errors.Is(err, fs.ErrNotExist):
				// possibly asynchronously removed, continue to the next one
				continue ProgramLoop
			case err != nil:
				return err
			}

			info, err := prog.Info()
			if err != nil {
				return err
			}

			stats, err := prog.Stats()
			if err != nil {
				return err
			}
			if stats.RunCount == 0 {
				continue ProgramLoop
			}

			clear(labels)
			labels["id"] = strconv.Itoa(int(id))
			labels["type"] = info.Type.String()
			labels["name"] = info.Name
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

			metricsName := GetMetricsName("bpf", "run_time_seconds_total", labels)
			counter := metrics.GetOrCreateFloatCounter(metricsName)
			counter.Set(stats.Runtime.Seconds())
			UnmarkMetrics(metricsName)

			metricsName = GetMetricsName("bpf", "run_count_total", labels)
			counter = metrics.GetOrCreateFloatCounter(metricsName)
			counter.Set(float64(stats.RunCount))
			UnmarkMetrics(metricsName)

			// interrupt program enumeration if context is done
			select {
			case <-ctx.Done():
				return nil
			default:
			}
		}
		CleanMetricsSection(section)

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
