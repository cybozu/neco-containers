package hooks

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:verbs=create;update,path=/mutate-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,versions=v1,name=mpod.kb.io

type podMutator struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewPodMutator creates a webhook handler for Pod.
func NewPodMutator(c client.Client, dec *admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &podMutator{c, dec}}
}

func (m *podMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	po := &v1.Pod{}
	err := m.decoder.Decode(req, po)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	poPatched := po.DeepCopy()
	for _, co := range po.Spec.InitContainers {
		if m.isMountedTmp(co) {
			volumeName := m.generateVolumeName(co.Name, po.Spec.Volumes)
			m.appendEmptyDir(volumeName, poPatched)
			m.appendMountTmp(volumeName, &co)
		}
	}
	for _, co := range po.Spec.Containers {
		if m.isMountedTmp(co) {
			volumeName := m.generateVolumeName(co.Name, po.Spec.Volumes)
			m.appendEmptyDir(volumeName, poPatched)
			m.appendMountTmp(volumeName, &co)
		}
	}

	marshaled, err := json.Marshal(poPatched)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}

func (m *podMutator) isMountedTmp(c v1.Container) bool {
	for _, mount := range c.VolumeMounts {
		if mount.MountPath == "/tmp" || strings.HasPrefix(mount.MountPath, "/tmp/") {
			return false
		}
	}
	return true
}

func (m *podMutator) hashString(name string) string {
	h := sha1.New()
	h.Write([]byte(name))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func (m *podMutator) isUniqueVolumeName(volumes []v1.Volume, name string) bool {
	for _, v := range volumes {
		if v.Name == name {
			return false
		}
	}
	return true
}

func (m *podMutator) generateVolumeName(containerName string, volumes []v1.Volume) string {
	for i := 0; ; i++ {
		volumeName := "tmp-" + m.hashString(containerName+strconv.Itoa(i))
		if m.isUniqueVolumeName(volumes, volumeName) {
			return volumeName
		}
	}
}

func (m *podMutator) appendEmptyDir(volumeName string, po *v1.Pod) {
	po.Spec.Volumes = append(po.Spec.Volumes, v1.Volume{
		Name:         volumeName,
		VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
	})
}

func (m *podMutator) appendMountTmp(volumeName string, co *v1.Container) {
	co.VolumeMounts = append(co.VolumeMounts, v1.VolumeMount{Name: volumeName, MountPath: "/tmp"})
}
