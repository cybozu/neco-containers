package controllers

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// StorageClass is the name of StorageClass. It is set to pv.spec.storageClassName.
	StorageClass = "local-storage"

	localPVProvisionerLabelKey = "local-pv-provisioner.cybozu.com/node"
)

var (
	vpNameRegexp = regexp.MustCompile(`[^.0-9A-Za-z]+`)
)

// NewDeviceDetector creates a new DeviceDetector.
func NewDeviceDetector(client client.Client, log logr.Logger, deviceDir string, deviceNameFilter *regexp.Regexp, nodeName string, interval time.Duration, scheme *runtime.Scheme, deleter Deleter) manager.Runnable {
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
		deleter,
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
	deleter          Deleter
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
		err := dd.deleter.Delete(dev.Path)
		if err != nil {
			return err
		}
		err = dd.createPV(ctx, dev, node)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dd *DeviceDetector) pvName(devPath string) string {
	tmp := strings.Join([]string{"local", dd.nodeName, filepath.Base(devPath)}, "-")
	return strings.ToLower(vpNameRegexp.ReplaceAllString(tmp, "-"))
}

func (dd *DeviceDetector) createPV(ctx context.Context, dev Device, node *corev1.Node) error {
	pvMode := corev1.PersistentVolumeBlock
	log := dd.log
	pv := &corev1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: dd.pvName(dev.Path)}}

	op, err := ctrl.CreateOrUpdate(ctx, dd.Client, pv, func() error {
		pv.ObjectMeta.Labels = map[string]string{localPVProvisionerLabelKey: node.Name}

		pv.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}

		// Workaround because capacity comparison doesn't work well in CreateOrUpdate.
		quantity := *resource.NewQuantity(dev.CapacityBytes, resource.BinarySI)
		_ = quantity.String()
		pv.Spec.Capacity = corev1.ResourceList{
			corev1.ResourceStorage: quantity,
		}

		pv.Spec.NodeAffinity = &corev1.VolumeNodeAffinity{
			Required: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      corev1.LabelHostname,
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{dd.nodeName},
					},
				}},
			}},
		}
		pv.Spec.PersistentVolumeReclaimPolicy = corev1.PersistentVolumeReclaimRetain
		pv.Spec.PersistentVolumeSource = corev1.PersistentVolumeSource{
			Local: &corev1.LocalVolumeSource{Path: dev.Path},
		}
		pv.Spec.StorageClassName = StorageClass
		pv.Spec.VolumeMode = &pvMode

		return ctrl.SetControllerReference(node, pv, dd.scheme)
	})
	if err != nil {
		log.Error(err, "unable to create or update PV", "device", dev)
		return err
	}
	if op != controllerutil.OperationResultNone {
		log.Info("PV successfully created or updated", "operation", op, "device", dev)
	}
	return nil
}
