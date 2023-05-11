package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// rolloutOperation implements "rollout" operation using "kubectl rollout restart".
type rolloutOperation struct {
	clusterName           string
	logger                *zap.Logger
	podCreationTimestamps []string
	previousPrimaryIndex  int
}

func NewRolloutOperation(logger *zap.Logger, clusterName string) Operation {
	return &rolloutOperation{
		clusterName: clusterName,
		logger:      logger.With(zap.String("operation", "rollout")),
	}
}

func (o *rolloutOperation) Name() string {
	return "rollout"
}

func (o *rolloutOperation) CheckPreCondition(ctx context.Context) (bool, error) {
	cluster, err := getMySQLCluster(ctx, o.logger, o.clusterName)
	if err != nil {
		return false, err
	}

	if !cluster.Healthy() {
		return false, nil
	}

	o.previousPrimaryIndex = cluster.Status.CurrentPrimaryIndex

	o.podCreationTimestamps = make([]string, cluster.Spec.Replicas)
	for i := 0; i < cluster.Spec.Replicas; i++ {
		pod, err := getPod(ctx, o.logger, fmt.Sprintf("moco-%s-%d", o.clusterName, i), false)
		if err != nil {
			return false, err
		}
		o.podCreationTimestamps[i] = pod.ObjectMeta.CreationTimestamp
	}

	return true, nil
}

func (o *rolloutOperation) Execute(ctx context.Context) error {
	o.logger.Info("executing rollout")

	_, stderr, err := execCmd(ctx, "kubectl", "rollout", "restart", "-n", currentNamespace, "sts/moco-"+o.clusterName)
	if err != nil {
		o.logger.Error("could not rollout", zap.Error(err), zap.String("stderr", string(stderr)))
		return err
	}

	return nil
}

func (o *rolloutOperation) CheckCompletion(ctx context.Context) (bool, error) {
	cluster, err := getMySQLCluster(ctx, o.logger, o.clusterName)
	if err != nil {
		return false, err
	}
	if !cluster.Available() {
		return false, nil
	}

	for i := 0; i < cluster.Spec.Replicas; i++ {
		pod, err := getPod(ctx, o.logger, fmt.Sprintf("moco-%s-%d", o.clusterName, i), true)
		if err != nil {
			return false, err
		}
		if pod == nil {
			return false, nil
		}

		if pod.CreationTimestamp == o.podCreationTimestamps[i] {
			return false, nil
		}
	}

	o.logger.Info("completion confirmed", zap.Int("previous", o.previousPrimaryIndex), zap.Int("current", cluster.Status.CurrentPrimaryIndex))

	return true, nil
}
