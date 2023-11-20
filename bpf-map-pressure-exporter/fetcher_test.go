package main

import (
	"testing"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/rlimit"
	"github.com/stretchr/testify/assert"
)

func newMap(t *testing.T, name string, maxEntries uint32) (*ebpf.Map, uint32) {
	t.Helper()
	m, err := ebpf.NewMap(&ebpf.MapSpec{
		Name:       name,
		Type:       ebpf.Hash,
		KeySize:    4,
		ValueSize:  4,
		MaxEntries: maxEntries,
	})
	if err != nil {
		t.Fatal(err)
	}
	info, err := m.Info()
	if err != nil {
		t.Fatal(err)
	}
	id, _ := info.ID()
	return m, uint32(id)
}

func setupTestMaps(t *testing.T) (mapIds []uint32) {
	t.Helper()
	if err := rlimit.RemoveMemlock(); err != nil {
		t.Fatal(err)
	}
	m1, id := newMap(t, "cilium_test_1", 10)
	mapIds = append(mapIds, id)
	m2, id := newMap(t, "cilium_test_2", 10)
	mapIds = append(mapIds, id)
	m3, id := newMap(t, "cilium_largemap", 10000)
	mapIds = append(mapIds, id)

	err := m1.Put([]byte("hoge"), []byte("aaaa"))
	if err != nil {
		t.Fatal(err)
	}
	err = m2.Put([]byte("hoge"), []byte("aaaa"))
	if err != nil {
		t.Fatal(err)
	}
	err = m2.Put([]byte("fuga"), []byte("bbbb"))
	if err != nil {
		t.Fatal(err)
	}
	err = m3.Put([]byte("hoge"), []byte("aaaa"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		m1.Close()
		m2.Close()
		m3.Close()
	})

	return mapIds
}

func TestUpdate(t *testing.T) {
	mapIds := setupTestMaps(t)

	cases := []struct {
		name    string
		fetcher *bpfMapPressureFetcher
		expect  []bpfMapPressure
	}{
		{
			name:    "success",
			fetcher: newFetcher([]string{"cilium_test_1"}, 30*time.Second),
			expect: []bpfMapPressure{
				{
					mapId:       mapIds[0],
					mapName:     "cilium_test_1",
					mapPressure: 0.1,
				},
			},
		},
		{
			name:    "multiple maps specified",
			fetcher: newFetcher([]string{"cilium_test_1", "cilium_test_2"}, 30*time.Second),
			expect: []bpfMapPressure{
				{
					mapId:       mapIds[0],
					mapName:     "cilium_test_1",
					mapPressure: 0.1,
				},
				{
					mapId:       mapIds[1],
					mapName:     "cilium_test_2",
					mapPressure: 0.2,
				},
			},
		},
		{
			name:    "multiple maps matched",
			fetcher: newFetcher([]string{"cilium_test"}, 30*time.Second),
			expect: []bpfMapPressure{
				{
					mapId:       mapIds[0],
					mapName:     "cilium_test_1",
					mapPressure: 0.1,
				},
				{
					mapId:       mapIds[1],
					mapName:     "cilium_test_2",
					mapPressure: 0.2,
				},
			},
		},
		{
			name:    "duplicate maps matched",
			fetcher: newFetcher([]string{"cilium_test_1", "ilium_test_1"}, 30*time.Second),
			expect: []bpfMapPressure{
				{
					mapId:       mapIds[0],
					mapName:     "cilium_test_1",
					mapPressure: 0.1,
				},
			},
		},
		{
			name:    "large map",
			fetcher: newFetcher([]string{"cilium_largemap"}, 30*time.Second),
			expect: []bpfMapPressure{
				{
					mapId:       mapIds[2],
					mapName:     "cilium_largemap",
					mapPressure: 0.0001,
				},
			},
		},
		{
			name:    "no maps matched",
			fetcher: newFetcher([]string{"notfound"}, 30*time.Second),
			expect:  []bpfMapPressure{},
		},
		{
			name:    "no maps specified",
			fetcher: newFetcher([]string{}, 30*time.Second),
			expect:  []bpfMapPressure{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.fetcher.update()
			results := tc.fetcher.GetMetrics()
			assert.ElementsMatch(t, tc.expect, results)
		})
	}
}
