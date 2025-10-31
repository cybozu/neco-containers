package bpf

import (
	"errors"
	"maps"
	"os"
	"slices"
	"strings"

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
