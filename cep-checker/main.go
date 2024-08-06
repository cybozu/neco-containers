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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
)

const (
	CiliumGroup   string = "cilium.io"
	CiliumVersion string = "v2"
	CepName       string = "ciliumendpoints"

	MetricsMissingName string = "cep_checker_missing"
)

var (
	cmd = &cobra.Command{
		Use:     "cep-checker",
		Short:   "cep-checker checks missing Pods or CiliumEndpoints",
		Run:     cmdMain,
		Version: "1.0.0",
	}

	cfg Config

	log *slog.Logger

	missingResourceNameToKind = make(map[string]string)
)

type Config struct {
	interval      time.Duration
	metricsServer string
	ignoreJobPod  bool
}

func init() {
	cmd.Flags().DurationVarP(&cfg.interval, "interval", "i", time.Second*30, "Interval to check missing CEPs or Pods")
	cmd.Flags().StringVarP(&cfg.metricsServer, "metrics-server", "m", "0.0.0.0:8080", "Metrics server address and port")
	cmd.Flags().BoolVar(&cfg.ignoreJobPod, "ignore-job-pod", true, "Ignore Pods created by Job")

	log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func cmdMain(cmd *cobra.Command, args []string) {

	ticker := time.NewTicker(cfg.interval)

	config, err := config.GetConfig()
	if err != nil {
		log.Error("failed to get config", slog.Any("error", err))
		os.Exit(1)
	}

	scheme := runtime.NewScheme()
	ciliumv2.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	client, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		log.Error("failed to create k8s client", slog.Any("error", err))
		os.Exit(1)
	}

	ctx := context.Background()

	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics.WritePrometheus(w, false)
	})

	go func() {

		for {
			<-ticker.C

			newMissings, err := checkAll(ctx, client)
			if err != nil {
				log.ErrorContext(ctx, "failed to check resources", slog.Any("error", err))
				continue
			}

			// delete resolved metrics
			for k, v := range missingResourceNameToKind {
				if _, ok := newMissings[k]; ok {
					continue
				}
				s := strings.Split(k, "/")
				ns := s[0]
				name := s[1]
				target := fmt.Sprintf(`%s{name="%s", namespace="%s", resource="%s"}`, MetricsMissingName, name, ns, v)
				log.InfoContext(ctx, "resolve", slog.String("namespace", ns), slog.String("name", name), slog.String("resource", v))
				metrics.UnregisterMetric(target)
				delete(missingResourceNameToKind, k)
			}
			for k, v := range newMissings {
				missingResourceNameToKind[k] = v
			}
		}
	}()

	log.InfoContext(ctx, "start metrics server", slog.String("server", cfg.metricsServer))
	if err := http.ListenAndServe(cfg.metricsServer, nil); err != nil {
		log.Error("failed to server the metrics server", slog.Any("error", err))
		os.Exit(1)
	}
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Error("failed to run", slog.Any("error", err))
		os.Exit(1)
	}
}

func checkAll(ctx context.Context, client client.Client) (map[string]string, error) {
	pods, err := getPods(ctx, client)
	if err != nil {
		return nil, err
	}

	ceps, err := getCeps(ctx, client)
	if err != nil {
		return nil, err
	}

	missings := make(map[string]string)

	// pod -> cep
	for key, val := range pods {
		if cfg.ignoreJobPod && val {
			// skip job pod
			continue
		}
		if _, ok := ceps[key]; !ok {
			// To avoid a miss detection, check again
			res, err := check(ctx, client, key)
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
			res, err := check(ctx, client, key)
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
func check(ctx context.Context, c client.Client, key string) (string, error) {

	podExist := false
	cepExist := false

	s := strings.Split(key, "/")
	ns := s[0]
	name := s[1]

	pod := &corev1.Pod{}
	if err := c.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, pod); err != nil {
		if !apierrors.IsNotFound(err) {
			return "", err
		}
	} else {
		podExist = true
	}

	cep := &ciliumv2.CiliumEndpoint{}
	if err := c.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, cep); err != nil {
		if !apierrors.IsNotFound(err) {
			return "", err
		}
	} else {
		cepExist = true
	}

	if podExist != cepExist {
		if podExist {
			log.InfoContext(ctx, "find a missing CEP", slog.String("namespace", ns), slog.String("name", name))
			m := fmt.Sprintf(`%s{name="%s", namespace="%s", resource="cep"}`, MetricsMissingName, name, ns)
			metrics.GetOrCreateGauge(m, func() float64 { return 1 })
			return "cep", nil
		}
		if cepExist {
			log.InfoContext(ctx, "find a missing Pod", slog.String("namespace", ns), slog.String("name", name))
			m := fmt.Sprintf(`%s{name="%s", namespace="%s", resource="pod"}`, MetricsMissingName, name, ns)
			metrics.GetOrCreateGauge(m, func() float64 { return 1 })
			return "pod", nil
		}
	}

	return "", nil
}

// getPods returns map[string]bool.
// This lists pods that is not in host network and that is running all namespaces.
// The map key is the pair of namespace and name(namespace/name).
// The map value is the bool value where its pod is owned by the Job resource; it is set to true.
func getPods(ctx context.Context, client client.Client) (map[string]bool, error) {
	pods := make(map[string]bool)

	podList := &corev1.PodList{}
	if err := client.List(ctx, podList); err != nil {
		return nil, err
	}

	for _, pod := range podList.Items {
		if pod.Spec.HostNetwork {
			continue
		}
		if pod.Status.Phase == "Succeeded" {
			continue
		}
		if pod.Status.Phase != "Running" && pod.Status.PodIP == "" {
			continue
		}
		key := fmt.Sprintf("%s/%s", pod.GetNamespace(), pod.GetName())
		// Check the pod is owned by Job
		ownedByJob := false
		for _, ref := range pod.OwnerReferences {
			if ref.Kind == "Job" {
				ownedByJob = true
			}
		}
		pods[key] = ownedByJob
	}

	return pods, nil
}

// list all CiliumEndpoints
func getCeps(ctx context.Context, client client.Client) (map[string]struct{}, error) {
	ceps := make(map[string]struct{})

	cepList := &ciliumv2.CiliumEndpointList{}
	if err := client.List(ctx, cepList); err != nil {
		return nil, err
	}

	for _, cep := range cepList.Items {
		key := fmt.Sprintf("%s/%s", cep.GetNamespace(), cep.GetName())
		ceps[key] = struct{}{}
	}

	return ceps, nil
}
