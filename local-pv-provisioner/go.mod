module github.com/cybozu/neco-containers/local-pv-provisioner

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.3.2
	k8s.io/api v0.16.4
	k8s.io/apimachinery v0.16.4
	k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
	k8s.io/klog v0.4.0
	sigs.k8s.io/controller-runtime v0.4.0
)
