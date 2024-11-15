package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cybozu-go/log"
	"golang.org/x/sync/errgroup"
)

const calcMapPressureConcurrency = 10

type IBpfMapPressureFetcher interface {
	GetMetrics() []bpfMapPressure
	Start(context.Context)
}

type bpfMapPressureFetcher struct {
	mapNameStrings []string
	fetchInterval  time.Duration
	mutex          sync.RWMutex
	metrics        []bpfMapPressure
}

type bpfMapPressure struct {
	mapId       uint32
	mapName     string
	mapPressure float64
}

func newFetcher(mapNameStrings []string, interval time.Duration) *bpfMapPressureFetcher {
	return &bpfMapPressureFetcher{
		mapNameStrings: mapNameStrings,
		fetchInterval:  interval,
	}
}

func (f *bpfMapPressureFetcher) Start(ctx context.Context) {
	f.update()

	ticker := time.NewTicker(f.fetchInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			f.update()
		}
	}
}

func (f *bpfMapPressureFetcher) update() {
	metrics := f.fetch()
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.metrics = metrics
}

func (f *bpfMapPressureFetcher) GetMetrics() []bpfMapPressure {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	res := make([]bpfMapPressure, len(f.metrics))
	copy(res, f.metrics)
	return res
}

func createBuffer(structSize int, elemNum int) (typ reflect.Type, buf any) {
	fields := []reflect.StructField{}
	for i := 0; i < structSize; i++ {
		fields = append(fields, reflect.StructField{
			Name: fmt.Sprintf("Field%d", i),
			Type: reflect.TypeOf(uint8(0)),
		})
	}
	typ = reflect.StructOf(fields)
	buf = reflect.MakeSlice(reflect.SliceOf(typ), elemNum, elemNum).Interface()
	return typ, buf
}

func calcMapPressure(id ebpf.MapID, m *ebpf.Map, minfo *ebpf.MapInfo) bpfMapPressure {
	// The logic here is based on:
	// https://github.com/cilium/cilium/pull/28183/files#diff-866773192f1b66200105da12d7cbb35f6bee9ee9ef2499d64ef1dfca6908eba4R264-R315
	const chunkSize uint32 = 4096

	mx := minfo.MaxEntries
	_, kout := createBuffer(int(minfo.KeySize), int(chunkSize))
	_, vout := createBuffer(int(minfo.ValueSize), int(chunkSize))

	var cursor ebpf.MapBatchCursor

	cnt := 0
	for {
		c, err := m.BatchLookup(&cursor, kout, vout, nil)
		cnt += c
		if err != nil {
			if errors.Is(err, ebpf.ErrKeyNotExist) {
				break
			}
			_ = logger.Warn("failed to execute BatchLookup", map[string]interface{}{
				"id":        id,
				log.FnError: err,
			})
		}
	}

	return bpfMapPressure{
		mapId:       uint32(id),
		mapName:     minfo.Name,
		mapPressure: float64(cnt) / float64(mx),
	}
}

func (f *bpfMapPressureFetcher) fetch() []bpfMapPressure {
	results := []bpfMapPressure{}
	var id ebpf.MapID = 0
	var err error

	var mu sync.Mutex
	eg := new(errgroup.Group)
	eg.SetLimit(calcMapPressureConcurrency)

	for {
		id, err = ebpf.MapGetNextID(id)
		if errors.Is(err, os.ErrNotExist) {
			break
		}
		if err != nil {
			_ = logger.Warn("failed to get next map id", map[string]interface{}{
				"id":        id,
				log.FnError: err,
			})
			return results
		}
		m, err := ebpf.NewMapFromID(id)
		if err != nil {
			_ = logger.Warn("failed to get map", map[string]interface{}{
				"id":        id,
				log.FnError: err,
			})
			return results
		}
		minfo, err := m.Info()
		if err != nil {
			_ = logger.Warn("failed to get map info", map[string]interface{}{
				"id":        id,
				log.FnError: err,
			})
			return results
		}
		for _, str := range f.mapNameStrings {
			if !strings.Contains(minfo.Name, str) {
				continue
			}
			id := id
			m := m
			minfo := minfo
			eg.Go(func() error {
				res := calcMapPressure(id, m, minfo)
				mu.Lock()
				results = append(results, res)
				mu.Unlock()
				return nil
			})
			break
		}
	}
	_ = eg.Wait()
	return results
}
