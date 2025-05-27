package controllers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var k8sClient client.Client
var testEnv *envtest.Environment
var testCtx = context.Background()
var testCancel context.CancelFunc

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	testCtx, testCancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		DownloadBinaryAssets:        true,
		DownloadBinaryAssetsVersion: "v" + os.Getenv("ENVTEST_K8S_VERSION"),
		BinaryAssetsDirectory:       "../bin",
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = corev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("running manager")
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(mgr).ToNot(BeNil())

	_, err = mgr.GetCache().GetInformer(testCtx, &corev1.PersistentVolume{})
	Expect(err).ShouldNot(HaveOccurred())
	pvController := &PersistentVolumeReconciler{
		mgr.GetClient(),
		"worker-1",
		deleterMock{},
	}
	err = pvController.SetupWithManager(mgr, "worker-1")
	Expect(err).ShouldNot(HaveOccurred())

	go func() {
		err = mgr.Start(testCtx)
		if err != nil {
			Expect(err).NotTo(HaveOccurred())
		}
	}()

})

var _ = AfterSuite(func() {
	testCancel()
	time.Sleep(10 * time.Millisecond)
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Test functions", func() {
	Context("do", testDo)
	Context("has annotations set by another configuration", testHasAnnotsSetByAnotherConfiguration)
	Context("parse pv spec configmap", testParsePVSpecConfigMap)
	Context("create pv", testDeviceDetectorCreatePV)
	Context("persistent volume reconciler", testPersistentVolumeReconciler)
	Context("fill deleter", testFillDeleter)
})

type testFS struct {
	pathPrefix string
	realFS     osFS
}

var _ fileSystem = &testFS{}

func NewTestFS(pathPrefix string) (*testFS, error) {
	return &testFS{pathPrefix: pathPrefix}, nil
}
func (fs *testFS) getPath(name string) string     { return filepath.Join(fs.pathPrefix, name) }
func (fs *testFS) Open(name string) (file, error) { return fs.realFS.Open(fs.getPath(name)) }
func (fs *testFS) Stat(name string) (FileInfo, error) {
	return fs.realFS.Stat(fs.getPath(name))
}
func (fs *testFS) OpenFile(name string, flag int, perm FileMode) (file, error) {
	return fs.realFS.OpenFile(fs.getPath(name), flag, perm)
}
func (fs *testFS) Walk(root string, f func(path string, info FileInfo, err error) error) error {
	return fs.realFS.Walk(fs.getPath(root), func(path string, info FileInfo, err error) error {
		testFSPath, err2 := filepath.Rel(fs.pathPrefix, path)
		if err2 != nil {
			panic("filepath.Rel failed: " + path)
		}
		testFSPath = "/" + testFSPath
		return f(testFSPath, info, err)
	})
}
func (fs *testFS) MkdirAll(path string, perm FileMode) error {
	return fs.realFS.MkdirAll(fs.getPath(path), perm)
}
func (fs *testFS) Remove(name string) error {
	return fs.realFS.Remove(fs.getPath(name))
}

func useTestFS(files map[string]string, f func()) {
	originalFS := fs

	pathPrefix, err := os.MkdirTemp("", "lpp-*")
	if err != nil {
		panic("os.MkdirTemp failed")
	}
	defer os.RemoveAll(pathPrefix)

	testFS, err := NewTestFS(pathPrefix)
	if err != nil {
		panic("NewTmpFS failed")
	}
	fs = testFS
	defer func() { fs = originalFS }()

	for path, body := range files {
		if !filepath.IsAbs(path) {
			panic("Not every path of files is absolute: " + path)
		}
		if err := fs.MkdirAll(filepath.Dir(path), 0700); err != nil {
			panic("tmpFS.MkdirAll failed: " + path)
		}
		file, err := fs.OpenFile(path, O_WRONLY|O_CREATE, 0600)
		if err != nil {
			panic("fs.OpenFile failed: " + path)
		}
		if _, err := file.Write([]byte(body)); err != nil {
			panic("file.Write failed: " + path)
		}
		if err := file.Close(); err != nil {
			panic("file.Close failed: " + path)
		}
	}

	f()
}
