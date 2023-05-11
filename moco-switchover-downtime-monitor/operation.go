package main

import (
	"context"

	"go.uber.org/zap"
)

// Operation represents an instance of switchover-like operation including the target MySQLCluster
type Operation interface {
	// Name returns the name of this operation.
	// For example, "switchover"
	Name() string

	// `CheckPrecondition` checks the precondition for this operation.
	// If the context is done, returns ctx.Err().
	// If the precondition is met, returns true. In this case, this object may save the current state for later use.
	// If the precondition is not met yet, returns false.
	CheckPreCondition(ctx context.Context) (bool, error)

	// Execute executes this operation.
	Execute(ctx context.Context) error

	// `CheckCompletion` checks the completion of the side effect of the execution.
	// If the context is done, returns ctx.Err().
	// After the completion, returns true.
	// Otherwise, returns false.
	CheckCompletion(ctx context.Context) (bool, error)
}

var operationConstructors = []func(*zap.Logger, string) Operation{
	NewSwitchoverOperation,
	NewPoddeletePrimaryOperation,
	NewPoddeleteReplicaOperation,
	NewRolloutOperation,
	NewKillprocPrimaryOperation,
	NewKillprocReplicaOperation,
}
