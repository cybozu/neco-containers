package main

import (
	"context"

	"go.uber.org/zap"
)

// switchoverOperation implements "switchover" operation using "kubectl moco switchover".
type switchoverOperation struct {
	clusterName          string
	logger               *zap.Logger
	previousPrimaryIndex int
}

func NewSwitchoverOperation(logger *zap.Logger, clusterName string) Operation {
	return &switchoverOperation{
		clusterName: clusterName,
		logger:      logger.With(zap.String("operation", "switchover")),
	}
}

func (o *switchoverOperation) Name() string {
	return "switchover"
}

func (o *switchoverOperation) CheckPreCondition(ctx context.Context) (bool, error) {
	cluster, err := getMySQLCluster(ctx, o.logger, o.clusterName)
	if err != nil {
		return false, err
	}

	if !cluster.Healthy() {
		return false, nil
	}

	o.previousPrimaryIndex = cluster.Status.CurrentPrimaryIndex

	return true, nil
}

func (o *switchoverOperation) Execute(ctx context.Context) error {
	o.logger.Info("executing switchover")

	_, stderr, err := execCmd(ctx, "kubectl", "moco", "switchover", "-n", currentNamespace, o.clusterName)
	if err != nil {
		o.logger.Error("could not switchover", zap.Error(err), zap.String("stderr", string(stderr)))
		return err
	}

	return nil
}

func (o *switchoverOperation) CheckCompletion(ctx context.Context) (bool, error) {
	cluster, err := getMySQLCluster(ctx, o.logger, o.clusterName)
	if err != nil {
		return false, err
	}

	if !cluster.Available() || cluster.Status.CurrentPrimaryIndex == o.previousPrimaryIndex {
		return false, nil
	}

	o.logger.Info("completion confirmed", zap.Int("previous", o.previousPrimaryIndex), zap.Int("current", cluster.Status.CurrentPrimaryIndex))

	return true, nil
}
