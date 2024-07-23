package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	CILIUM_GROUP   string = "cilium.io"
	CILIUM_VERSION string = "v2"
	CEP_NAME       string = "ciliumendpoints"

	METRICS_MISSING_NAME string = "cep_checker_missing"
)

var (
	cmd = &cobra.Command{
		Use:     "cep-checker",
		Short:   "cep-checker checks missing Pods or CiliumEndpoints",
		RunE:    cmdMain,
		Version: "1.0.0",
	}

	cfg Config

	log *slog.Logger

	missingMap = make(map[string]string)
)

type Config struct {
	interval      time.Duration
	metricsServer string
}

func init() {
	cmd.Flags().DurationVarP(&cfg.interval, "interval", "i", time.Second*30, "Interval to check missing CEPs or Pods")
	cmd.Flags().StringVarP(&cfg.metricsServer, "metrics-server", "m", "0.0.0.0:8080", "Metrics server address and port")

	log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func cmdMain(cmd *cobra.Command, args []string) error {

	ticker := time.NewTicker(cfg.interval)

	config, err := config.GetConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	ctx := context.Background()

	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WritePrometheus(w, false)
	})

	log.InfoContext(ctx, "start metrics server", slog.String("server", cfg.metricsServer))
	go http.ListenAndServe(cfg.metricsServer, nil)

	for {
		<-ticker.C

		newMissings, err := checkAll(ctx, clientset, dynamicClient)
		if err != nil {
			return err
		}

		// delete resolved metrics
		for k, v := range missingMap {
			if _, ok := newMissings[k]; ok {
				continue
			}
			s := strings.Split(k, "/")
			ns := s[0]
			name := s[1]
			target := fmt.Sprintf(`%s{name="%s", namespace="%s", resource="%s"}`, METRICS_MISSING_NAME, name, ns, v)
			log.InfoContext(ctx, "resolve", slog.String("namespace", ns), slog.String("name", name), slog.String("resource", v))
			metrics.UnregisterMetric(target)
			delete(missingMap, k)
		}
		for k, v := range newMissings {
			missingMap[k] = v
		}
	}
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Error("failed to run", slog.Any("error", err))
		os.Exit(1)
	}
}

func checkAll(ctx context.Context, clientset *kubernetes.Clientset, dynamicClient *dynamic.DynamicClient) (map[string]string, error) {
	pods, err := getPods(ctx, clientset)
	if err != nil {
		return nil, err
	}

	ceps, err := getCeps(ctx, dynamicClient)
	if err != nil {
		return nil, err
	}

	missings := make(map[string]string)

	// pod -> cep
	for key := range pods {
		if _, ok := ceps[key]; !ok {
			// To avoid a miss detection, check again
			res, err := check(ctx, clientset, dynamicClient, key)
			if err != nil {
				return nil, err
			}
			if res != "" {
				missings[key] = res
			}
		}
	}

	// cep -> pod
	for key := range ceps {
		if _, ok := pods[key]; !ok {
			// To avoid a miss detection, check again
			res, err := check(ctx, clientset, dynamicClient, key)
			if err != nil {
				return nil, err
			}
			if res != "" {
				missings[key] = res
			}
		}
	}
	return missings, nil
}

// Check checks the consistency for given key(namespace/name).
// If there is the inconsistency(which one is missing), output a log and create a metric.
func check(ctx context.Context, clientset *kubernetes.Clientset, dynamicClient *dynamic.DynamicClient, key string) (string, error) {

	podExist := false
	cepExist := false

	s := strings.Split(key, "/")
	ns := s[0]
	name := s[1]

	_, err := clientset.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return "", err
		}
	} else {
		podExist = true
	}

	_, err = dynamicClient.Resource(schema.GroupVersionResource{Group: CILIUM_GROUP, Version: CILIUM_VERSION, Resource: CEP_NAME}).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return "", err
		}
	} else {
		cepExist = true
	}

	if podExist != cepExist {
		if podExist {
			log.WarnContext(ctx, "find a missing CEP", slog.String("namespace", ns), slog.String("name", name))
			m := fmt.Sprintf(`%s{name="%s", namespace="%s", resource="cep"}`, METRICS_MISSING_NAME, name, ns)
			metrics.GetOrCreateGauge(m, func() float64 { return 1 })
			return "cep", nil
		}
		if cepExist {
			log.WarnContext(ctx, "find a missing Pod", slog.String("namespace", ns), slog.String("name", name))
			m := fmt.Sprintf(`%s{name="%s", namespace="%s", resource="pod"}`, METRICS_MISSING_NAME, name, ns)
			metrics.GetOrCreateGauge(m, func() float64 { return 1 })
			return "pod", nil
		}
	}

	return "", nil
}

// list pods that is not in host network and that is running all namespaces
func getPods(ctx context.Context, clientset *kubernetes.Clientset) (map[string]struct{}, error) {
	pods := make(map[string]struct{})

	podList, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range podList.Items {
		if pod.Spec.HostNetwork {
			continue
		}
		if pod.Status.Phase == "Succeeded" {
			continue
		}
		if pod.Status.Phase == "Pending" && pod.Status.PodIP == "" {
			continue
		}
		key := fmt.Sprintf("%s/%s", pod.GetNamespace(), pod.GetName())
		pods[key] = struct{}{}
	}

	return pods, nil
}

// list all CiliumEndpoints
func getCeps(ctx context.Context, dynamicClient *dynamic.DynamicClient) (map[string]struct{}, error) {
	ceps := make(map[string]struct{})

	cepList, err := dynamicClient.Resource(schema.GroupVersionResource{Group: CILIUM_GROUP, Version: CILIUM_VERSION, Resource: CEP_NAME}).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, cep := range cepList.Items {
		key := fmt.Sprintf("%s/%s", cep.GetNamespace(), cep.GetName())
		ceps[key] = struct{}{}
	}

	return ceps, nil
}
