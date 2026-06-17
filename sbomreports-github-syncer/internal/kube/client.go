package kube

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewDynamicClient(kubeconfig string) (dynamic.Interface, error) {
	cfg, err := buildConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create Kubernetes dynamic client: %w", err)
	}
	return client, nil
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("load kubeconfig %q: %w", kubeconfig, err)
		}
		return cfg, nil
	}

	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("load Kubernetes config: no in-cluster config and no home directory: %w", err)
	}
	defaultKubeconfig := filepath.Join(home, ".kube", "config")
	if _, err := os.Stat(defaultKubeconfig); err != nil {
		return nil, fmt.Errorf("load Kubernetes config: no in-cluster config and %s is unavailable: %w", defaultKubeconfig, err)
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", defaultKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("load default kubeconfig %q: %w", defaultKubeconfig, err)
	}
	return cfg, nil
}
