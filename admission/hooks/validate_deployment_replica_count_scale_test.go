package hooks

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	autoscalingv1 "k8s.io/client-go/applyconfigurations/autoscaling/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("validate deployment replica count scale webhook", func() {
	testcases := []struct {
		scenario          string
		hasAnnotation     bool
		forceReplicaCount int
		expectScalable    bool
	}{
		{
			scenario:          "should deny to scale a Deployment to replicas = 1",
			hasAnnotation:     true,
			forceReplicaCount: 0,
			expectScalable:    false,
		},
		{
			scenario:       "should allow to scale a Deployment without force-replica-count annotation",
			hasAnnotation:  false,
			expectScalable: true,
		},
		{
			scenario:          "should allow to scale a Deployment with annotation force-replica-count != 0 and force-replica-count != replicas",
			hasAnnotation:     true,
			forceReplicaCount: 2,
			expectScalable:    true,
		},
		{
			scenario:          "should allow to scale a Deployment with annotation force-replica-count != 0 and force-replica-count == replicas",
			hasAnnotation:     true,
			forceReplicaCount: 1,
			expectScalable:    true,
		},
	}

	for i, tt := range testcases {
		manifestName := fmt.Sprintf("test-scale-deployment-%d", i)
		tt := tt

		It(tt.scenario, func() {
			var deployManifest string
			if tt.hasAnnotation {
				deployManifest = fmt.Sprintf(deployManifestWithAnnotationTemplate, manifestName, tt.forceReplicaCount, 0)
			} else {
				deployManifest = fmt.Sprintf(deployManifestWithoutAnnotationTemplate, manifestName, 0)
			}

			d := yaml.NewYAMLOrJSONDecoder(strings.NewReader(deployManifest), 4096)
			deploy := &appsv1.Deployment{}
			err := d.Decode(deploy)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Create(testCtx, deploy)
			Expect(err).NotTo(HaveOccurred())

			k8s, err := kubernetes.NewForConfig(k8sConfig)
			Expect(err).NotTo(HaveOccurred())

			deployClient := k8s.AppsV1().Deployments("default")
			scale := autoscalingv1.Scale().WithSpec(autoscalingv1.ScaleSpec().WithReplicas(1))

			_, err = deployClient.ApplyScale(testCtx, manifestName, scale, metav1.ApplyOptions{FieldManager: "dummy", Force: true})

			if tt.expectScalable {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
			}

		})
	}
})
