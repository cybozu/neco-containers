package controllers

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

	lppLegacyLabelKey = "local-pv-provisioner.cybozu.com/node"

	lppDomain                = "local-pv-provisioner.cybozu.io/"
	lppAnnotNode             = lppDomain + "node"
	lppAnnotPVSpecConfigMap  = lppDomain + "pv-spec-configmap"
	lppAnnotVolumeMode       = lppDomain + "volumeMode"
	lppAnnotFSType           = lppDomain + "fsType"
	lppAnnotDeviceDir        = lppDomain + "deviceDir"
	lppAnnotDeviceNameFilter = lppDomain + "deviceNameFilter"

	pvSpecCMKeyVolumeMode       = "volumeMode"
	pvSpecCMKeyFsType           = "fsType"
	pvSpecCMKeyDeviceDir        = "deviceDir"
	pvSpecCMKeyDeviceNameFilter = "deviceNameFilter"
)

var (
	vpNameRegexp = regexp.MustCompile(`[^.0-9A-Za-z]+`)

	supportedFsTypes = []string{"ext4", "xfs", "btrfs"}
)

func isFilesystem(volumeMode string) bool {
	return volumeMode == "Filesystem"
}

type pvSpec struct {
	volumeMode               string
	fsType                   string
	deviceDir                string
	deviceNameFilter         string
	deviceNameFilterCompiled *regexp.Regexp
}

func parsePVSpecConfigMap(cm *corev1.ConfigMap) (*pvSpec, error) {
	// Parse the values of the config map. If a required value is missing, return an error.
	pvSpecVolumeMode, ok1 := cm.Data[pvSpecCMKeyVolumeMode]
	pvSpecFSType, ok2 := cm.Data[pvSpecCMKeyFsType]
	pvSpecDeviceDir, ok3 := cm.Data[pvSpecCMKeyDeviceDir]
	pvSpecDeviceNameFilter, ok4 := cm.Data[pvSpecCMKeyDeviceNameFilter]
	if !ok1 || (isFilesystem(pvSpecVolumeMode) && !ok2) || !ok3 || !ok4 {
		return nil, errors.New("required value not found in pv spec ConfigMap")
	}

	// Make sure that the parsed values are valid.
	if !isFilesystem(pvSpecVolumeMode) && pvSpecVolumeMode != "Block" {
		return nil, errors.New("volumeMode should be either 'Filesystem' or 'Block'")
	}
	if isFilesystem(pvSpecVolumeMode) && !slices.Contains(supportedFsTypes, pvSpecFSType) {
		return nil, fmt.Errorf("fsType should be some of %v if volumeMode is 'Filesystem'", supportedFsTypes)
	}
	if !filepath.IsAbs(pvSpecDeviceDir) {
		return nil, errors.New("deviceDir must be an absolute path")
	}
	info, err := fs.Stat(pvSpecDeviceDir)
	if err != nil {
		return nil, fmt.Errorf("unable to get status of device directory: %s: %w", pvSpecDeviceDir, err)
	}
	if !info.Mode().IsDir() {
		return nil, errors.New("deviceDir is not a directory")
	}
	pvSpecDeviceNameFilterCompiled, err := regexp.Compile(pvSpecDeviceNameFilter)
	if err != nil {
		return nil, fmt.Errorf("unable to compile deviceNameFilter: %s: %w", pvSpecDeviceNameFilter, err)
	}

	return &pvSpec{
		volumeMode:               pvSpecVolumeMode,
		fsType:                   pvSpecFSType,
		deviceDir:                pvSpecDeviceDir,
		deviceNameFilter:         pvSpecDeviceNameFilter,
		deviceNameFilterCompiled: pvSpecDeviceNameFilterCompiled,
	}, nil
}

// hasAnnotsSetByAnotherConfiguration checks that pvSpec's settings are the same as those in alreadyCreatedPVs.
func hasAnnotsSetByAnotherConfiguration(pvSpec *pvSpec, alreadyCreatedPVs []corev1.PersistentVolume) bool {
	conflicted := false
	for _, pv := range alreadyCreatedPVs {
		annot := pv.GetAnnotations()
		volumeMode, ok1 := annot[lppAnnotVolumeMode]
		fsType, ok2 := annot[lppAnnotFSType]
		deviceDir, ok3 := annot[lppAnnotDeviceDir]
		deviceNameFilter, ok4 := annot[lppAnnotDeviceNameFilter]
		if !ok1 || (isFilesystem(volumeMode) && !ok2) || !ok3 || !ok4 ||
			volumeMode != pvSpec.volumeMode || (isFilesystem(volumeMode) && fsType != pvSpec.fsType) ||
			deviceDir != pvSpec.deviceDir || deviceNameFilter != pvSpec.deviceNameFilter {
			conflicted = true
			break
		}
	}
	return conflicted
}

// NewDeviceDetector creates a new DeviceDetector.
func NewDeviceDetector(client client.Client, reader client.Reader, log logr.Logger, nodeName string, interval time.Duration, scheme *runtime.Scheme, deleter Deleter, defaultPVSpecConfigMap, workingNamespace string) manager.Runnable {
	dd := &DeviceDetector{
		client, reader, log, nodeName, interval, scheme,
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
		defaultPVSpecConfigMap,
		workingNamespace,
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
	reader                 client.Reader
	log                    logr.Logger
	nodeName               string
	interval               time.Duration
	scheme                 *runtime.Scheme
	availableDevices       prometheus.Gauge
	errorDevices           prometheus.Gauge
	deleter                Deleter
	defaultPVSpecConfigMap string
	workingNamespace       string
}

// Start implements controller-runtime's manager.Runnable.
func (dd *DeviceDetector) Start(ctx context.Context) error {
	dd.do()

	tick := time.NewTicker(dd.interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			dd.do()
		case <-ctx.Done():
			return nil
		}
	}
}

// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:namespace=default,groups="",resources=configmaps,verbs=get;list;watch

func (dd *DeviceDetector) do() {
	ctx := context.Background()
	log := dd.log

	node := new(corev1.Node)
	if err := dd.Get(ctx, types.NamespacedName{Name: dd.nodeName}, node); err != nil {
		log.Error(err, "unable to fetch Node", "node", dd.nodeName)
		return
	}

	// Get pv-spec-configmap name
	annot := node.GetAnnotations()
	pvSpecCMName, ok := annot[lppAnnotPVSpecConfigMap]
	if !ok {
		// There's no pv-spec-configmap annotation on this Node.
		// If the user sets its default value, use it. Otherwise, ignore this node.
		if dd.defaultPVSpecConfigMap == "" {
			log.Info("pv-spec-configmap annotation not found", "node", dd.nodeName)
			return
		} else {
			pvSpecCMName = dd.defaultPVSpecConfigMap
		}
	}

	// Fetch the pv spec configmap.
	var pvSpecCM corev1.ConfigMap
	if err := dd.Get(ctx, types.NamespacedName{Name: pvSpecCMName, Namespace: dd.workingNamespace}, &pvSpecCM); err != nil {
		log.Error(err, "unable to fetch ConfigMap", "configmap", pvSpecCMName, "namespace", dd.workingNamespace)
		return
	}

	// Parse and validate the values of the config map
	pvSpec, err := parsePVSpecConfigMap(&pvSpecCM)
	if err != nil {
		log.Error(err, "unable to parse spec configmap")
		return
	}

	// Make sure that the PVs to be deployed will not conflict.
	var alreadyCreatedPVs corev1.PersistentVolumeList
	if err := dd.List(ctx, &alreadyCreatedPVs, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			lppLegacyLabelKey: node.Name,
		}),
	}); err != nil {
		log.Error(err, "unable to fetch pv list")
		return
	}
	if hasAnnotsSetByAnotherConfiguration(pvSpec, alreadyCreatedPVs.Items) {
		log.Error(
			errors.New("there are already some PVs that are created with different settings"),
			"there are already some PVs that are created with different settings",
		)
		return
	}

	// Make the PVs according to the specified spec.
	devices, errDevices, err := dd.listLocalDevices(pvSpec.deviceDir, pvSpec.deviceNameFilterCompiled)
	if err != nil {
		log.Error(err, "unable to list local devices")
		return
	}

	availDevices := make([]Device, 0)

	for _, dev := range devices {
		var pv corev1.PersistentVolume
		err := dd.reader.Get(ctx, types.NamespacedName{Name: dd.pvName(dev.Path)}, &pv)
		if apierrors.IsNotFound(err) {
			err := dd.deleter.Delete(dev.Path)
			if err != nil {
				log.Error(err, "unable to cleanup device", "path", dev.Path)
				errDevices = append(errDevices, dev)
				continue
			}
		}

		err = dd.createPV(ctx, dev, node, pvSpec.volumeMode, pvSpec.fsType, pvSpec.deviceDir, pvSpec.deviceNameFilter)
		if err != nil {
			log.Error(err, "unable to create or update PV", "path", dev.Path)
			errDevices = append(errDevices, dev)
		}

		availDevices = append(availDevices, dev)
	}

	dd.availableDevices.Set(float64(len(availDevices)))
	dd.errorDevices.Set(float64(len(errDevices)))
}

func (dd *DeviceDetector) pvName(devPath string) string {
	tmp := strings.Join([]string{"local", dd.nodeName, filepath.Base(devPath)}, "-")
	return strings.ToLower(vpNameRegexp.ReplaceAllString(tmp, "-"))
}

func (dd *DeviceDetector) createPV(ctx context.Context, dev Device, node *corev1.Node, volumeMode, fsType, deviceDir, deviceNameFilter string) error {
	log := dd.log
	pv := &corev1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: dd.pvName(dev.Path)}}

	pvMode := corev1.PersistentVolumeBlock
	if isFilesystem(volumeMode) {
		pvMode = corev1.PersistentVolumeFilesystem
	}

	op, err := ctrl.CreateOrUpdate(ctx, dd.Client, pv, func() error {
		if pv.ObjectMeta.Labels == nil {
			pv.ObjectMeta.Labels = make(map[string]string)
		}
		pv.ObjectMeta.Labels[lppLegacyLabelKey] = node.Name
		pv.ObjectMeta.Labels[lppAnnotNode] = node.Name

		if pv.ObjectMeta.Annotations == nil {
			pv.ObjectMeta.Annotations = make(map[string]string)
		}
		pv.ObjectMeta.Annotations[lppAnnotVolumeMode] = volumeMode
		if isFilesystem(volumeMode) {
			pv.ObjectMeta.Annotations[lppAnnotFSType] = fsType
		}
		pv.ObjectMeta.Annotations[lppAnnotDeviceDir] = deviceDir
		pv.ObjectMeta.Annotations[lppAnnotDeviceNameFilter] = deviceNameFilter

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
		if isFilesystem(volumeMode) {
			pv.Spec.PersistentVolumeSource.Local.FSType = &fsType
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
