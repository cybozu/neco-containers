package hooks

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	storagev1 "k8s.io/api/storage/v1"

	//+kubebuilder:scaffold:imports
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	podMutatingWebhookPath              = "/mutate-pod"
	podValidatingWebhookPath            = "/validate-pod"
	contourMutatingWebhookPath          = "/mutate-projectcontour-io-httpproxy"
	calicoValidateWebhookPath           = "/validate-projectcalico-org-networkpolicy"
	contourValidateWebhookPath          = "/validate-projectcontour-io-httpproxy"
	grafanaDashboardValidateWebhookPath = "/validate-integreatly-org-grafanadashboard"
	deleteValidateWebhookPath           = "/validate-delete"
	preventDeleteValidateWebhookPath    = "/validate-preventdelete"
)

var scheme *runtime.Scheme
var k8sConfig *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var testCtx = context.Background()
var testCancel context.CancelFunc

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Webhook Suite")
}

func isHTTPProxyMutationDisabled() bool {
	return os.Getenv("NO_HTTPPROXY_MUTATION") == "true"
}

type nullWebhook struct{}

func (wh *nullWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("ok")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	testCtx, testCancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "config", "webhook")},
		},
	}

	var err error
	k8sConfig, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sConfig).NotTo(BeNil())

	scheme = runtime.NewScheme()
	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = admissionv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())
	err = admissionv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(k8sConfig, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("setting up resources")
	setupNetworkPolicyResources()
	sc := &storagev1.StorageClass{}
	sc.Name = "local-storage"
	sc.Provisioner = "kubernetes.io/no-provisioner"
	waitForFirstConsumer := storagev1.VolumeBindingWaitForFirstConsumer
	sc.VolumeBindingMode = &waitForFirstConsumer
	err = k8sClient.Create(testCtx, sc)
	Expect(err).NotTo(HaveOccurred())

	// start webhook server using Manager
	webhookInstallOptions := &testEnv.WebhookInstallOptions
	mgr, err := ctrl.NewManager(k8sConfig, ctrl.Options{
		Scheme:             scheme,
		Host:               webhookInstallOptions.LocalServingHost,
		Port:               webhookInstallOptions.LocalServingPort,
		CertDir:            webhookInstallOptions.LocalServingCertDir,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	Expect(err).NotTo(HaveOccurred())

	dec, err := admission.NewDecoder(scheme)
	Expect(err).NotTo(HaveOccurred())
	wh := mgr.GetWebhookServer()
	wh.Register(podMutatingWebhookPath, NewPodMutator(mgr.GetClient(), dec))
	permissive := os.Getenv("TEST_PERMISSIVE") == "true"
	wh.Register(podValidatingWebhookPath, NewPodValidator(mgr.GetClient(), dec, []string{"quay.io/cybozu/"}, permissive))
	wh.Register(calicoValidateWebhookPath, NewCalicoNetworkPolicyValidator(mgr.GetClient(), dec, 1000))
	if isHTTPProxyMutationDisabled() {
		wh.Register(contourMutatingWebhookPath, &webhook.Admission{Handler: &nullWebhook{}})
	} else {
		wh.Register(contourMutatingWebhookPath, NewContourHTTPProxyMutator(mgr.GetClient(), dec, "secured"))
	}
	wh.Register(contourValidateWebhookPath, NewContourHTTPProxyValidator(mgr.GetClient(), dec))
	wh.Register(grafanaDashboardValidateWebhookPath, NewGrafanaDashboardValidator(mgr.GetClient(), dec))
	wh.Register(deleteValidateWebhookPath, NewDeleteValidator(mgr.GetClient(), dec))
	wh.Register(preventDeleteValidateWebhookPath, NewPreventDeleteValidator(mgr.GetClient(), dec))

	//+kubebuilder:scaffold:webhook

	go func() {
		err = mgr.Start(testCtx)
		if err != nil {
			Expect(err).NotTo(HaveOccurred())
		}
	}()

	// wait for the webhook server to get ready
	dialer := &net.Dialer{Timeout: time.Second}
	addrPort := fmt.Sprintf("%s:%d", webhookInstallOptions.LocalServingHost, webhookInstallOptions.LocalServingPort)
	Eventually(func() error {
		conn, err := tls.DialWithDialer(dialer, "tcp", addrPort, &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}).Should(Succeed())

}, 60)

var _ = AfterSuite(func() {
	testCancel()
	time.Sleep(10 * time.Millisecond)
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
