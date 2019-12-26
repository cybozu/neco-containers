package controllers

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const nodeNameLabel = "cybozu.com/node-name"

// NewDeviceDetector creates a new DeviceDetector.
func NewDeviceDetector(client client.Client, log logr.Logger, deviceDir string, deviceNameFilter *regexp.Regexp, nodeName string, interval time.Duration, scheme *runtime.Scheme) manager.Runnable {
	return &DeviceDetector{client, log, deviceDir, deviceNameFilter, nodeName, interval, scheme}
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
		if !apierrs.IsNotFound(err) {
			log.Error(err, "unable to fetch Node", "node", dd.nodeName)
			return err
		}
	}

	nodeGVK, err := apiutil.GVKForObject(node, dd.scheme)
	if err != nil {
		log.Error(err, "unable to get GVK")
		return err
	}

	var pvList corev1.PersistentVolumeList
	if err := dd.List(ctx, &pvList, client.MatchingLabels{nodeNameLabel: dd.nodeName}); err != nil {
		log.Error(err, "unable to list PV")
		return err
	}

	devices, err := dd.listLocalDevices()
	if err != nil {
		log.Error(err, "unable to list local devices")
		return err
	}
	log.Info("local devices", "devices", devices)

	nodeRef := &metav1.OwnerReference{
		APIVersion: nodeGVK.GroupVersion().String(),
		Kind:       nodeGVK.Kind,
		Name:       node.GetName(),
		UID:        node.GetUID(),
	}
	var errStrings []string
	for _, dev := range devices {
		if !dd.pvExists(pvList, dev) {
			err := dd.createPV(ctx, dev, nodeRef)
			if err != nil {
				errStrings = append(errStrings, err.Error())
			}
		}
	}
	if len(errStrings) > 0 {
		err := errors.New(strings.Join(errStrings, "; "))
		log.Error(err, "unable to create pv")
		return err
	}
	return nil
}
