module github.com/cybozu/neco-containers/local-pv-provisioner

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/google/go-cmp v0.4.0
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/prometheus/client_golang v1.1.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.3.2
	golang.org/x/crypto v0.0.0-20200221231518-2aa609cf4a9d // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29 // indirect
	sigs.k8s.io/controller-runtime v0.6.3
)
