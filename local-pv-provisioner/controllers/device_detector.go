package controllers

import (
	"context"
	"regexp"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// NewDeviceDetector creates a new DeviceDetector.
func NewDeviceDetector(client client.Client, log logr.Logger, deviceDir string, deviceNameFilter *regexp.Regexp, nodeName string, interval time.Duration, scheme *runtime.Scheme) manager.Runnable {

	dd := &DeviceDetector{
		client, log, deviceDir, deviceNameFilter, nodeName, interval, scheme,
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "local_pv_provisioner_available_devices",
			Help:        "The number of devices recognized by local pv provisioner without errors.",
			ConstLabels: prometheus.Labels{"node": nodeName},
		}),
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "local_pv_provisioner_error_devices",
			Help:        "The number of error devices recognized by local pv provisioner.",
			ConstLabels: prometheus.Labels{"node": nodeName},
		}),
	}

	metrics.Registry.MustRegister(dd.availableDevices)
	metrics.Registry.MustRegister(dd.errorDevices)

	return dd
}

// Device is containing information of a device.
type Device struct {
	Path          string
	CapacityBytes int64
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
	availableDevices prometheus.Gauge
	errorDevices     prometheus.Gauge
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

// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch;create;update;patch

func (dd *DeviceDetector) do() error {
	ctx := context.Background()
	log := dd.log

	node := new(corev1.Node)
	if err := dd.Get(ctx, types.NamespacedName{Name: dd.nodeName}, node); err != nil {
		log.Error(err, "unable to fetch Node", "node", dd.nodeName)
		return err
	}

	devices, errDevices, err := dd.listLocalDevices()
	if err != nil {
		log.Error(err, "unable to list local devices")
		return err
	}

	dd.availableDevices.Set(float64(len(devices)))
	dd.errorDevices.Set(float64(len(errDevices)))

	for _, dev := range devices {
		err := dd.createPV(ctx, dev, node)
		if err != nil {
			return err
		}
	}

	return nil
}
