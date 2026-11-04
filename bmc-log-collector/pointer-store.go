package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"maps"
	"path"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

// pointerStore keeps the last read position of SEL for each machine serial.
type pointerStore interface {
	// get returns the pointer of the serial. It returns the zero value if no pointer exists yet.
	get(serial string) (LastPointer, error)
	// put records the pointer of the serial.
	put(serial string, ptr LastPointer) error
	// sync removes the pointers of the machines not in machinesList and persists the pointers.
	sync(ctx context.Context, machinesList []Machine) error
}

// filePointerStore stores each pointer in a file named after the serial.
type filePointerStore struct {
	dir string
}

func newFilePointerStore(dir string) *filePointerStore {
	return &filePointerStore{dir: dir}
}

func (s *filePointerStore) get(serial string) (LastPointer, error) {
	filePath := path.Join(s.dir, serial)
	if err := checkAndCreatePointerFile(filePath); err != nil {
		return LastPointer{}, err
	}
	return readLastPointer(filePath)
}

func (s *filePointerStore) put(serial string, ptr LastPointer) error {
	return updateLastPointer(ptr, path.Join(s.dir, serial))
}

func (s *filePointerStore) sync(_ context.Context, machinesList []Machine) error {
	return deletePtrFileDisappearedSerial(s.dir, machinesList)
}

const configMapPointerFieldManager = "bmc-log-collector"

// configMapPointerStore keeps the pointers in memory and writes them to a ConfigMap on sync.
// The keys of the ConfigMap are the serials and the values are the pointers in JSON.
// It assumes that only one process writes the ConfigMap.
type configMapPointerStore struct {
	client    corev1client.ConfigMapInterface
	namespace string
	name      string

	mu        sync.Mutex
	pointers  map[string]LastPointer
	persisted map[string]string
}

func newConfigMapPointerStore(ctx context.Context, client corev1client.ConfigMapInterface, namespace, name string) (*configMapPointerStore, error) {
	s := &configMapPointerStore{
		client:    client,
		namespace: namespace,
		name:      name,
		pointers:  make(map[string]LastPointer),
		persisted: make(map[string]string),
	}

	cm, err := client.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}

	for serial, value := range cm.Data {
		var ptr LastPointer
		if err := json.Unmarshal([]byte(value), &ptr); err != nil {
			slog.Error("ignore a broken pointer in the ConfigMap", "err", err, "serial", serial, "configmap", name)
			continue
		}
		s.pointers[serial] = ptr
	}
	s.persisted = maps.Clone(cm.Data)
	return s, nil
}

func (s *configMapPointerStore) get(serial string) (LastPointer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pointers[serial], nil
}

func (s *configMapPointerStore) put(serial string, ptr LastPointer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pointers[serial] = ptr
	return nil
}

func (s *configMapPointerStore) sync(ctx context.Context, machinesList []Machine) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	machineExist := make(map[string]bool, len(machinesList))
	for _, m := range machinesList {
		machineExist[m.Serial] = true
	}

	data := make(map[string]string, len(s.pointers))
	for serial, ptr := range s.pointers {
		if !machineExist[serial] {
			delete(s.pointers, serial)
			deleteMetrics(serial)
			continue
		}
		// An invalid key makes the whole ConfigMap update fail, so skip it.
		if errs := validation.IsConfigMapKey(serial); len(errs) > 0 {
			slog.Error("can't store the pointer in the ConfigMap because the serial is not a valid key", "serial", serial, "errs", errs)
			continue
		}
		byteJSON, err := json.Marshal(ptr)
		if err != nil {
			return err
		}
		data[serial] = string(byteJSON)
	}

	if maps.Equal(data, s.persisted) {
		return nil
	}

	cm := corev1apply.ConfigMap(s.name, s.namespace).WithData(data)
	_, err := s.client.Apply(ctx, cm, metav1.ApplyOptions{FieldManager: configMapPointerFieldManager, Force: true})
	if err != nil {
		return err
	}
	s.persisted = data
	return nil
}
