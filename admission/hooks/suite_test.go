package hooks

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	storagev1 "k8s.io/api/storage/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var k8sClient client.Client
var testEnv *envtest.Environment
var testCtx = context.Background()
var stopCh = make(chan struct{})

func strPtr(s string) *string { return &s }

func modePtr(m storagev1.VolumeBindingMode) *storagev1.VolumeBindingMode { return &m }

func setupCommonResources() {
	caBundle, err := ioutil.ReadFile("testdata/ca.crt")
	Expect(err).ShouldNot(HaveOccurred())
	vwh := &admissionregistrationv1beta1.ValidatingWebhookConfiguration{}
	vwh.Name = "neco-admission"
	_, err = ctrl.CreateOrUpdate(testCtx, k8sClient, vwh, func() error {
		failPolicy := admissionregistrationv1beta1.Fail
		sideEffect := admissionregistrationv1beta1.SideEffectClassNone
		vwh.Webhooks = []admissionregistrationv1beta1.ValidatingWebhook{
			{
				Name:          "vnetworkpolicy.kb.io",
				FailurePolicy: &failPolicy,
				SideEffects:   &sideEffect,
				ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
					CABundle: caBundle,
					URL:      strPtr("https://127.0.0.1:8443/validate-projectcalico-org-networkpolicy"),
				},
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1beta1.OperationType{
							admissionregistrationv1beta1.Create,
							admissionregistrationv1beta1.Update,
						},
						Rule: admissionregistrationv1beta1.Rule{
							APIGroups:   []string{"crd.projectcalico.org"},
							APIVersions: []string{"v1"},
							Resources:   []string{"networkpolicies"},
						},
					},
				},
			},
			{
				Name:          "vhttpproxy.kb.io",
				FailurePolicy: &failPolicy,
				SideEffects:   &sideEffect,
				ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
					CABundle: caBundle,
					URL:      strPtr("https://127.0.0.1:8443/validate-projectcontour-io-httpproxy"),
				},
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1beta1.OperationType{
							admissionregistrationv1beta1.Create,
							admissionregistrationv1beta1.Update,
						},
						Rule: admissionregistrationv1beta1.Rule{
							APIGroups:   []string{"projectcontour.io"},
							APIVersions: []string{"v1"},
							Resources:   []string{"httpproxies"},
						},
					},
				},
			},
			{
				Name:          "vapplication.kb.io",
				FailurePolicy: &failPolicy,
				SideEffects:   &sideEffect,
				ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
					CABundle: caBundle,
					URL:      strPtr("https://127.0.0.1:8443/validate-argoproj-io-application"),
				},
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1beta1.OperationType{
							admissionregistrationv1beta1.Create,
							admissionregistrationv1beta1.Update,
						},
						Rule: admissionregistrationv1beta1.Rule{
							APIGroups:   []string{"argoproj.io"},
							APIVersions: []string{"v1alpha1"},
							Resources:   []string{"applications"},
						},
					},
				},
			},
		}
		return nil
	})
	Expect(err).ShouldNot(HaveOccurred())

	mwh := &admissionregistrationv1beta1.MutatingWebhookConfiguration{}
	mwh.Name = "neco-admission"
	_, err = ctrl.CreateOrUpdate(testCtx, k8sClient, mwh, func() error {
		failPolicy := admissionregistrationv1beta1.Fail
		mwh.Webhooks = []admissionregistrationv1beta1.MutatingWebhook{
			{
				Name:          "mpod.kb.io",
				FailurePolicy: &failPolicy,
				ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
					CABundle: caBundle,
					URL:      strPtr("https://127.0.0.1:8443/mutate-pod"),
				},
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1beta1.OperationType{
							admissionregistrationv1beta1.Create,
						},
						Rule: admissionregistrationv1beta1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"pods"},
						},
					},
				},
			},
			{
				Name:          "mhttpproxy.kb.io",
				FailurePolicy: &failPolicy,
				ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
					CABundle: caBundle,
					URL:      strPtr("https://127.0.0.1:8443/mutate-projectcontour-io-httpproxy"),
				},
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1beta1.OperationType{
							admissionregistrationv1beta1.Create,
						},
						Rule: admissionregistrationv1beta1.Rule{
							APIGroups:   []string{"projectcontour.io"},
							APIVersions: []string{"v1"},
							Resources:   []string{"httpproxies"},
						},
					},
				},
			},
		}
		return nil
	})
	Expect(err).ShouldNot(HaveOccurred())
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	SetDefaultEventuallyTimeout(time.Minute)
	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	apiServerFlags := append(envtest.DefaultKubeAPIServerFlags,
		"--admission-control=MutatingAdmissionWebhook",
		"--admission-control=ValidatingAdmissionWebhook",
	)
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:  []string{filepath.Join("..", "config", "crd", "bases")},
		KubeAPIServerFlags: apiServerFlags,
	}

	var err error
	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	By("setting up resources")
	setupCommonResources()
	setupNetworkPolicyResources()

	By("running webhook server")
	go run(stopCh, cfg, "127.0.0.1", 8443)
	d := &net.Dialer{Timeout: time.Second}
	Eventually(func() error {
		conn, err := tls.DialWithDialer(d, "tcp", "127.0.0.1:8443", &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}).Should(Succeed())
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	close(stopCh)
	time.Sleep(10 * time.Millisecond)
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
