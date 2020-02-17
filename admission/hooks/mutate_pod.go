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
	for ic, c := range po.Spec.Containers {
		if m.isTargetContainer(c) {
			poPatched = m.appendMountTmp(ic, *po)
		}
	}

	marshaled, err := json.Marshal(poPatched)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}

func (m *podMutator) isTargetContainer(c v1.Container) bool {
	for _, mount := range c.VolumeMounts {
		if mount.MountPath == "/tmp" || strings.HasPrefix(mount.MountPath, "/tmp/") {
			return false
		}
	}
	return true
}

func (m *podMutator) appendMountTmp(indexContainer int, po v1.Pod) *v1.Pod {
	volumeName := m.generateVolumeName(po.Spec.Containers[indexContainer].Name, po.Spec.Volumes)

	po.Spec.Volumes = append(po.Spec.Volumes, v1.Volume{
		Name:         volumeName,
		VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}},
	})
	po.Spec.Containers[indexContainer].VolumeMounts = append(po.Spec.Containers[indexContainer].VolumeMounts,
		v1.VolumeMount{Name: volumeName, MountPath: "/tmp"})
	return &po
}

func (m *podMutator) hashString(name string) string {
	h := sha1.New()
	h.Write([]byte(name))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func (m *podMutator) generateVolumeName(containerName string, volumes []v1.Volume) string {
	for i := 0; ; i++ {
		volumeName := "tmp-" + m.hashString(containerName+strconv.Itoa(i))
		if m.isUniqueVolumeName(volumes, volumeName) {
			return volumeName
		}
	}
}

func (m *podMutator) isUniqueVolumeName(volumes []v1.Volume, name string) bool {
	for _, v := range volumes {
		if v.Name == name {
			return false
		}
	}
	return true
}
