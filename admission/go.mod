module github.com/cybozu/neco-containers/admission

go 1.13

require (
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/projectcalico/libcalico-go v1.7.2-0.20191008175127-399044ecb659
	github.com/projectcontour/contour v1.3.0
	github.com/spf13/cobra v0.0.5
	go.uber.org/atomic v1.4.0 // indirect
	golang.org/x/crypto v0.0.0-20200221231518-2aa609cf4a9d // indirect
	k8s.io/api v0.17.6
	k8s.io/apimachinery v0.17.6
	k8s.io/client-go v0.17.6
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.5.2
	sigs.k8s.io/yaml v1.1.0
)
