module github.com/cybozu/neco-containers/admission

go 1.13

require (
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/projectcalico/libcalico-go v1.7.2-0.20191008175127-399044ecb659
	github.com/projectcontour/contour v1.0.1
	github.com/spf13/cobra v0.0.5
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	k8s.io/api v0.16.4
	k8s.io/apimachinery v0.16.4
	k8s.io/client-go v0.16.4
	k8s.io/klog v0.4.0
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/yaml v1.1.0
)
