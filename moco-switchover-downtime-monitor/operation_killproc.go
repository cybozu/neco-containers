package main

import (
	"context"
	"fmt"
	"math/rand"

	"go.uber.org/zap"
)

// killprocOperation implements "killproc" operation using "kubectl exec -- kill".
type killprocOperation struct {
	clusterName                 string
	logger                      *zap.Logger
	primary                     bool
	targetIndex                 int
	targetPodName               string
	targetPodMysqldRestartCount int
	previousPrimaryIndex        int
}

func newKillprocOperation(logger *zap.Logger, clusterName string, primary bool) Operation {
	o := &killprocOperation{
		clusterName: clusterName,
		logger:      logger, // temporary
		primary:     primary,
	}

	o.logger = logger.With(zap.String("operation", o.Name()))

	return o
}

func NewKillprocPrimaryOperation(logger *zap.Logger, clusterName string) Operation {
	return newKillprocOperation(logger, clusterName, true)
}

func NewKillprocReplicaOperation(logger *zap.Logger, clusterName string) Operation {
	return newKillprocOperation(logger, clusterName, false)
}

func (o *killprocOperation) Name() string {
	if o.primary {
		return "killproc-primary"
	} else {
		return "killproc-replica"
	}
}

func (o *killprocOperation) CheckPreCondition(ctx context.Context) (bool, error) {
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
	o.targetPodMysqldRestartCount = pod.RestartCount("mysqld")

	return true, nil
}

func (o *killprocOperation) Execute(ctx context.Context) error {
	o.logger.Info("executing killproc", zap.Int("index", o.targetIndex))

	_, stderr, err := execCmdWithInput(ctx,
		[]byte(`exec kill -SEGV $(ps axo pid,command | awk '$2=="mysqld" {print $1}')`),
		"kubectl", "exec", "-n", currentNamespace, o.targetPodName, "-c", "mysqld", "-i", "--", "bash", "/dev/stdin")
	if err != nil {
		o.logger.Error("could not kill process", zap.Error(err), zap.String("stderr", string(stderr)))
		return err
	}

	return nil
}

func (o *killprocOperation) CheckCompletion(ctx context.Context) (bool, error) {
	cluster, err := getMySQLCluster(ctx, o.logger, o.clusterName)
	if err != nil {
		return false, err
	}
	if !cluster.Available() {
		return false, nil
	}

	pod, err := getPod(ctx, o.logger, o.targetPodName, false)
	if err != nil {
		return false, err
	}
	if pod.RestartCount("mysqld") != o.targetPodMysqldRestartCount+1 {
		return false, nil
	}

	o.logger.Info("completion confirmed", zap.Int("previous", o.previousPrimaryIndex), zap.Int("current", cluster.Status.CurrentPrimaryIndex))

	return true, nil
}
