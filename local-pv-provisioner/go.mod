module github.com/cybozu/neco-containers/local-pv-provisioner

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/google/go-cmp v0.5.4
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	github.com/prometheus/client_golang v1.9.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	k8s.io/api v0.18.14
	k8s.io/apimachinery v0.18.14
	k8s.io/client-go v0.18.14
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29 // indirect
	sigs.k8s.io/controller-runtime v0.6.3
)
