package hooks

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	podMutatingWebhookPath              = "/mutate-pod"
	contourMutatingWebhookPath          = "/mutate-projectcontour-io-httpproxy"
	calicoValidateWebhookPath           = "/validate-projectcalico-org-networkpolicy"
	contourValidateWebhookPath          = "/validate-projectcontour-io-httpproxy"
	argocdValidateWebhookPath           = "/validate-argoproj-io-application"
	grafanaDashboardValidateWebhookPath = "/validate-integreatly-org-grafanadashboard"
	deleteValidateWebhookPath           = "/validate-delete"
)

var k8sClient client.Client
var testEnv *envtest.Environment
var testCtx = context.Background()
var stopCh = make(chan struct{})

func setupCommonResources() {
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	SetDefaultEventuallyTimeout(time.Minute)
	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	By("bootstrapping test environment")
	failPolicy := admissionregistrationv1beta1.Fail
	sideEffect := admissionregistrationv1beta1.SideEffectClassNone
	webhookInstallOptions := envtest.WebhookInstallOptions{
		MutatingWebhooks: []runtime.Object{
			&admissionregistrationv1beta1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: "neco-admission",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "MutatingWebhookConfiguration",
					APIVersion: "admissionregistration.k8s.io/v1beta1",
				},
				Webhooks: []admissionregistrationv1beta1.MutatingWebhook{
					{
						Name:          "mpod.kb.io",
						FailurePolicy: &failPolicy,
						ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
							Service: &admissionregistrationv1beta1.ServiceReference{
								Path: &podMutatingWebhookPath,
							},
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
							Service: &admissionregistrationv1beta1.ServiceReference{
								Path: &contourMutatingWebhookPath,
							},
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
				},
			},
		},
		ValidatingWebhooks: []runtime.Object{
			&admissionregistrationv1beta1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: "neco-admission",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "ValidatingWebhookConfiguration",
					APIVersion: "admissionregistration.k8s.io/v1beta1",
				},
				Webhooks: []admissionregistrationv1beta1.ValidatingWebhook{
					{
						Name:          "vnetworkpolicy.kb.io",
						FailurePolicy: &failPolicy,
						SideEffects:   &sideEffect,
						ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
							Service: &admissionregistrationv1beta1.ServiceReference{
								Path: &calicoValidateWebhookPath,
							},
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
							Service: &admissionregistrationv1beta1.ServiceReference{
								Path: &contourValidateWebhookPath,
							},
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
							Service: &admissionregistrationv1beta1.ServiceReference{
								Path: &argocdValidateWebhookPath,
							},
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
					{
						Name:          "vgrafanadashboard.kb.io",
						FailurePolicy: &failPolicy,
						SideEffects:   &sideEffect,
						ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
							Service: &admissionregistrationv1beta1.ServiceReference{
								Path: &grafanaDashboardValidateWebhookPath,
							},
						},
						Rules: []admissionregistrationv1beta1.RuleWithOperations{
							{
								Operations: []admissionregistrationv1beta1.OperationType{
									admissionregistrationv1beta1.Create,
									admissionregistrationv1beta1.Update,
								},
								Rule: admissionregistrationv1beta1.Rule{
									APIGroups:   []string{"integreatly.org"},
									APIVersions: []string{"v1alpha1"},
									Resources:   []string{"grafanadashboards"},
								},
							},
						},
					},
					{
						Name:          "vdelete.kb.io",
						FailurePolicy: &failPolicy,
						SideEffects:   &sideEffect,
						ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
							Service: &admissionregistrationv1beta1.ServiceReference{
								Path: &deleteValidateWebhookPath,
							},
						},
						Rules: []admissionregistrationv1beta1.RuleWithOperations{
							{
								Operations: []admissionregistrationv1beta1.OperationType{
									admissionregistrationv1beta1.Delete,
								},
								Rule: admissionregistrationv1beta1.Rule{
									APIGroups:   []string{""},
									APIVersions: []string{"v1"},
									Resources:   []string{"namespaces"},
								},
							},
						},
					},
				},
			},
		},
	}
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		WebhookInstallOptions: webhookInstallOptions,
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
	go run(stopCh, cfg, &testEnv.WebhookInstallOptions)
	d := &net.Dialer{Timeout: time.Second}
	Eventually(func() error {
		serverURL := fmt.Sprintf("%s:%d", testEnv.WebhookInstallOptions.LocalServingHost, testEnv.WebhookInstallOptions.LocalServingPort)
		conn, err := tls.DialWithDialer(d, "tcp", serverURL, &tls.Config{
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
