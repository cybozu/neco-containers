package main

import (
	"context"
	"fmt"
	"math/rand"

	"go.uber.org/zap"
)

// poddeleteOperation implements "poddelete" operation using "kubectl delete pod".
type poddeleteOperation struct {
	clusterName                string
	logger                     *zap.Logger
	primary                    bool
	targetIndex                int
	targetPodName              string
	targetPodCreationTimestamp string // not necessarily be time.Time
	previousPrimaryIndex       int
}

func newPoddeleteOperation(logger *zap.Logger, clusterName string, primary bool) Operation {
	o := &poddeleteOperation{
		clusterName: clusterName,
		logger:      logger, // temporary
		primary:     primary,
	}

	o.logger = logger.With(zap.String("operation", o.Name()))

	return o
}

func NewPoddeletePrimaryOperation(logger *zap.Logger, clusterName string) Operation {
	return newPoddeleteOperation(logger, clusterName, true)
}

func NewPoddeleteReplicaOperation(logger *zap.Logger, clusterName string) Operation {
	return newPoddeleteOperation(logger, clusterName, false)
}

func (o *poddeleteOperation) Name() string {
	if o.primary {
		return "poddelete-primary"
	} else {
		return "poddelete-replica"
	}
}

func (o *poddeleteOperation) CheckPreCondition(ctx context.Context) (bool, error) {
	cluster, err := getMySQLCluster(ctx, o.logger, o.clusterName)
	if err != nil {
		return false, err
	}

	if !cluster.Healthy() {
		return false, nil
	}

	o.previousPrimaryIndex = cluster.Status.CurrentPrimaryIndex

	if o.primary {
		o.targetIndex = cluster.Status.CurrentPrimaryIndex
	} else {
		o.targetIndex = rand.Intn(cluster.Spec.Replicas - 1)
		if o.targetIndex >= cluster.Status.CurrentPrimaryIndex {
			o.targetIndex++
		}
	}
	o.targetPodName = fmt.Sprintf("moco-%s-%d", o.clusterName, o.targetIndex)

	pod, err := getPod(ctx, o.logger, o.targetPodName, false)
	if err != nil {
		return false, err
	}
	o.targetPodCreationTimestamp = pod.CreationTimestamp

	return true, nil
}

func (o *poddeleteOperation) Execute(ctx context.Context) error {
	o.logger.Info("executing poddelete", zap.Int("index", o.targetIndex))

	_, stderr, err := execCmd(ctx, "kubectl", "delete", "pod", "-n", currentNamespace, o.targetPodName, "--wait=false")
	if err != nil {
		o.logger.Error("could not delete pod", zap.Error(err), zap.String("stderr", string(stderr)))
		return err
	}

	return nil
}

func (o *poddeleteOperation) CheckCompletion(ctx context.Context) (bool, error) {
	cluster, err := getMySQLCluster(ctx, o.logger, o.clusterName)
	if err != nil {
		return false, err
	}
	if !cluster.Available() {
		return false, nil
	}

	pod, err := getPod(ctx, o.logger, o.targetPodName, true)
	if err != nil {
		return false, err
	}
	if pod == nil {
		return false, nil
	}

	if pod.CreationTimestamp == o.targetPodCreationTimestamp {
		return false, nil
	}

	o.logger.Info("completion confirmed", zap.Int("previous", o.previousPrimaryIndex), zap.Int("current", cluster.Status.CurrentPrimaryIndex))

	return true, nil
}
