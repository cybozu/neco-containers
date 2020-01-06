package controllers

import (
	"context"
	"regexp"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewDeviceDetector creates a new DeviceDetector.
func NewDeviceDetector(client client.Client, log logr.Logger, deviceDir string, deviceNameFilter *regexp.Regexp, nodeName string, interval time.Duration, scheme *runtime.Scheme) manager.Runnable {
	return &DeviceDetector{client, log, deviceDir, deviceNameFilter, nodeName, interval, scheme}
}

// Device is containing information of a device.
type Device struct {
	Path          string
	CapacityBytes int64
	HasError      bool
}

// DeviceDetector monitors local devices.
type DeviceDetector struct {
	client.Client
	log              logr.Logger
	deviceDir        string
	deviceNameFilter *regexp.Regexp
	nodeName         string
	interval         time.Duration
	scheme           *runtime.Scheme
}

// Start implements controller-runtime's manager.Runnable.
func (dd *DeviceDetector) Start(ch <-chan struct{}) error {
	err := dd.do()
	if err != nil {
		return err
	}

	tick := time.NewTicker(dd.interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			err := dd.do()
			if err != nil {
				return err
			}
		case <-ch:
			return nil
		}
	}
}

// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;update;patch;watch
// +kubebuilder:rbac:groups="",resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;create;update;patch;delete;watch
// +kubebuilder:rbac:groups="",resources=persistentvolumes/status,verbs=get;update;patch

func (dd *DeviceDetector) do() error {
	ctx := context.Background()
	log := dd.log.WithValues("node", dd.nodeName)

	node := new(corev1.Node)
	if err := dd.Get(ctx, types.NamespacedName{Name: dd.nodeName}, node); err != nil {
		log.Error(err, "unable to fetch Node", "node", dd.nodeName)
		return err
	}

	devices, err := dd.listLocalDevices()
	if err != nil {
		log.Error(err, "unable to list local devices")
		return err
	}
	log.Info("local devices", "devices", devices)

	for _, dev := range devices {
		if !dev.HasError {
			err := dd.createPV(ctx, dev, node)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
