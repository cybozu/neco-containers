package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
	"ttypdb/common"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func pollPods(ctx context.Context, logger *zap.Logger, clientset *kubernetes.Clientset) {
	ticker := time.NewTicker(time.Second * time.Duration(*flagPollIntervalSeconds))
	defer ticker.Stop()

	m := sync.Mutex{}

	for {
		select {
		case <-ctx.Done():
			m.Lock() // wait for doPollPods completion
			return
		case <-ticker.C:
			if m.TryLock() {
				go func() {
					defer m.Unlock()
					doPollPods(ctx, logger, clientset)
				}()
			} else {
				metricsPollingSkipsCounter.Inc()
				logger.Warn(fmt.Sprintf("The previous polling takes more than %d seconds. Skip polling this time.", *flagPollIntervalSeconds))
			}
		}
	}
}

func doPollPods(ctx context.Context, logger *zap.Logger, clientset *kubernetes.Clientset) {
	logger.Info("polling start")
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		metricsPollingDurationSecondsHistogram.Observe(duration.Seconds())
		logger.Info("polling completed", zap.Duration("duration", duration.Round(time.Millisecond)))
	}()

	podList, err := clientset.CoreV1().Pods(currentNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: *flagSelector,
	})
	if err != nil {
		logger.Error("failed to list Pods", zap.Error(err))
		return
	}

	logger.Info("listed Pods", zap.Int("count", len(podList.Items)))
	for _, pod := range podList.Items {
		checkPod(ctx, logger.With(zap.String("namespace", pod.Namespace), zap.String("pod", pod.Name)), clientset, &pod)
	}
}

func checkPod(ctx context.Context, logger *zap.Logger, clientset *kubernetes.Clientset, pod *corev1.Pod) {
	if pod.DeletionTimestamp != nil {
		logger.Debug("the Pod is about to be deleted. skipping.")
		return
	}

	podIP := pod.Status.PodIP

	var container *corev1.Container
	for _, c := range pod.Spec.Containers {
		if c.Name == "ttypdb-sidecar" {
			container = &c
			break
		}
	}
	if container == nil {
		logger.Error("failed to find sidecar container")
		return
	}
	if len(container.Ports) < 1 {
		logger.Error("failed to get sidecar container port")
		return
	}

	resp, err := http.Get(fmt.Sprintf("http://%s:%d/status", podIP, container.Ports[0].ContainerPort))
	if err != nil {
		logger.Error("failed to get status", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	status := common.Status{}
	statusBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to get status", zap.Error(err))
		return
	}
	err = json.Unmarshal(statusBytes, &status)
	if err != nil {
		logger.Error("failed to unmarshal status", zap.Error(err))
		return
	}
	if status.TTYs < 0 {
		logger.Error("broken status")
		return
	}

	pdbInterface := clientset.PolicyV1().PodDisruptionBudgets(pod.Namespace)
	foundPdb := false
	_, err = pdbInterface.Get(ctx, pod.Name, metav1.GetOptions{})
	if err != nil {
		var k8serr *k8serrors.StatusError
		if !errors.As(err, &k8serr) || k8serr.Status().Reason != metav1.StatusReasonNotFound {
			logger.Error("failed to check PDB", zap.Error(err))
			return
		}
	} else {
		foundPdb = true
	}

	if status.TTYs == 0 {
		// no controlling terminals are observed. delete PDB.
		if !foundPdb {
			logger.Debug("PDB does not exist")
			return
		}

		err := pdbInterface.Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if err != nil {
			logger.Error("failed to delete PDB", zap.Error(err))
		} else {
			logger.Info("deleted PDB")
		}
	} else {
		// some controlling terminals are observed. create PDB.
		if foundPdb {
			logger.Debug("PDB already exists")
			return
		}

		zeroIntstr := intstr.FromInt(0)
		pdb := &policyv1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "v1",
						Kind:       "Pod",
						Name:       pod.GetName(),
						UID:        pod.GetUID(),
					},
				},
			},
			Spec: policyv1.PodDisruptionBudgetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: pod.Labels,
				},
				MaxUnavailable: &zeroIntstr,
			},
		}
		_, err = pdbInterface.Create(ctx, pdb, metav1.CreateOptions{})
		if err != nil {
			logger.Error("failed to create PDB", zap.Error(err))
		} else {
			logger.Info("created PDB")
		}
	}
}
