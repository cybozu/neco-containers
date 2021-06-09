package hooks

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-pod,mutating=true,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create,versions=v1,name=mpod.kb.io,admissionReviewVersions={v1,v1beta1}

type podMutator struct {
	client  client.Client
	decoder *admission.Decoder
}

// NewPodMutator creates a webhook handler for Pod.
func NewPodMutator(c client.Client, dec *admission.Decoder) http.Handler {
	return &webhook.Admission{Handler: &podMutator{c, dec}}
}

func (m *podMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	po := &corev1.Pod{}
	err := m.decoder.Decode(req, po)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	poPatched := po.DeepCopy()
	for i, co := range po.Spec.Containers {
		if !m.isMountedTmp(&co) {
			volumeName := m.generateVolumeName(co.Name, po.Spec.Volumes)
			m.appendEmptyDir(volumeName, poPatched)
			poPatched.Spec.Containers[i].VolumeMounts = append(poPatched.Spec.Containers[i].VolumeMounts,
				corev1.VolumeMount{Name: volumeName, MountPath: "/tmp"})
		}
	}
	for i, co := range po.Spec.InitContainers {
		if !m.isMountedTmp(&co) {
			volumeName := m.generateVolumeName(co.Name, po.Spec.Volumes)
			m.appendEmptyDir(volumeName, poPatched)
			poPatched.Spec.InitContainers[i].VolumeMounts = append(poPatched.Spec.InitContainers[i].VolumeMounts,
				corev1.VolumeMount{Name: volumeName, MountPath: "/tmp"})
		}
	}

	marshaled, err := json.Marshal(poPatched)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}

func (m *podMutator) isMountedTmp(co *corev1.Container) bool {
	for _, mount := range co.VolumeMounts {
		if mount.MountPath == "/tmp" || strings.HasPrefix(mount.MountPath, "/tmp/") {
			return true
		}
	}
	return false
}

func (m *podMutator) hashString(name string) string {
	sum := sha1.Sum([]byte(name))
	return hex.EncodeToString(sum[:])
}

func (m *podMutator) isUniqueVolumeName(volumes []corev1.Volume, name string) bool {
	for _, v := range volumes {
		if v.Name == name {
			return false
		}
	}
	return true
}

func (m *podMutator) generateVolumeName(containerName string, volumes []corev1.Volume) string {
	for i := 0; ; i++ {
		volumeName := "tmp-" + m.hashString(containerName+strconv.Itoa(i))
		if m.isUniqueVolumeName(volumes, volumeName) {
			return volumeName
		}
	}
}

func (m *podMutator) appendEmptyDir(volumeName string, po *corev1.Pod) {
	po.Spec.Volumes = append(po.Spec.Volumes, corev1.Volume{
		Name:         volumeName,
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	})
}
