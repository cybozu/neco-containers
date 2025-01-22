package k8s

import (
	"log/slog"
	"os"

	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

const FieldManagerName = "lbip-manager"

var client *kubernetes.Clientset

func init() {
	config, err := ctrl.GetConfig()
	if err != nil {
		log.Error("cannot get a kubernetes config", slog.Any("error", err))
		os.Exit(1)
	}

	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Error("cannot get a kubernetes client", slog.Any("error", err))
		os.Exit(1)
	}
}
