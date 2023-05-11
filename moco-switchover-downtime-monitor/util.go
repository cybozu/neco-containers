package main

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
)

// getMySQLCluster gets MySQLCluster object from k8s.
// Note that logging is done in this function.
func getMySQLCluster(ctx context.Context, logger *zap.Logger, name string) (*MySQLCluster, error) {
	stdout, stderr, err := execCmd(ctx, "kubectl", "get", "mysqlcluster", "-n", currentNamespace, name, "-ojson")
	if err != nil {
		logger.Error("could not get MySQLCluster", zap.Error(err), zap.String("stderr", string(stderr)))
		return nil, err
	}
	cluster := &MySQLCluster{}
	err = json.Unmarshal(stdout, cluster)
	if err != nil {
		logger.Error("could not unmarshal MySQLCluster", zap.Error(err))
		return nil, err
	}
	return cluster, nil
}

// getPod gets Pod object from k8s.
// Note that logging is done in this function.
// If notFoundIsNotError is true and the Pod is not found, this function returns (nil, nil).
func getPod(ctx context.Context, logger *zap.Logger, name string, notFoundIsNotError bool) (*Pod, error) {
	args := []string{"get", "pod", "-n", currentNamespace, name, "-ojson"}
	if notFoundIsNotError {
		args = append(args, "--ignore-not-found")
	}
	stdout, stderr, err := execCmd(ctx, "kubectl", args...)
	if err != nil {
		logger.Error("could not get Pod", zap.Error(err), zap.String("stderr", string(stderr)))
		return nil, err
	}
	if len(stdout) == 0 && notFoundIsNotError {
		return nil, nil
	}

	pod := &Pod{}
	err = json.Unmarshal(stdout, pod)
	if err != nil {
		logger.Error("could not unmarshal Pod", zap.Error(err))
		return nil, err
	}
	return pod, nil
}
