module github.com/cybozu/neco-containers/local-pv-provisioner

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/prometheus/client_golang v1.0.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.3.2
	k8s.io/api v0.17.6
	k8s.io/apimachinery v0.17.6
	k8s.io/client-go v0.17.6
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.5.2
)
