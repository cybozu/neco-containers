package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	_ "github.com/go-sql-driver/mysql"
)

var username = "moco-writable"
var databaseName = "test"

var cooldownDurationMin = time.Second * 5
var cooldownDurationMax = time.Second * 125

var preconditionCheckInterval = time.Second
var preconditionWarningInterval = time.Minute * 10

var completionCheckInterval = time.Minute
var completionWarningInterval = time.Minute * 10
var completionTimeout = time.Minute * 30

var pingerInterval = time.Millisecond * 100
var pingerTimeout = time.Millisecond * 500

var errCancelledByCompletion = errors.New("check completed")
var errCancelledByFailure = errors.New("check failed")

type Checker struct {
	clusterName string
	logger      *zap.Logger
	username    string
	password    string
	operations  []Operation
}

// NewChecker creates a new Checker object.
// ctx is the context.
// clusterName is the name of the checked MySQLCluster.
// logger is the logger.
func NewChecker(ctx context.Context, clusterName string, logger *zap.Logger) (*Checker, error) {
	c := &Checker{}
	c.clusterName = clusterName
	c.logger = logger.With(zap.String("cluster", clusterName))
	c.username = username

	_, stderr, err := execCmd(ctx, "kubectl", "get", "mysqlcluster", "-n", currentNamespace, clusterName)
	if err != nil {
		c.logger.Error("could not get MySQLCluster", zap.Error(err), zap.String("stderr", string(stderr)))
		return nil, err
	}

	stdout, stderr, err := execCmd(ctx, "kubectl", "moco", "credential", "-n", currentNamespace, clusterName, "-u", username)
	if err != nil {
		c.logger.Error("could not get credential", zap.Error(err), zap.String("stderr", string(stderr)))
		return nil, err
	}
	c.password = strings.Trim(string(stdout), "\n")

	for _, oc := range operationConstructors {
		c.operations = append(c.operations, oc(c.logger, c.clusterName))
	}

	return c, nil
}

// `Run` repeatedly checks the specified MySQLCluster.
// `ctx` is the context.
func (c *Checker) Run(ctx context.Context) {
	c.logger.Info("starting checker")
	defer c.logger.Info("stopping checker")

	for {
		rand.Shuffle(len(c.operations), func(i, j int) {
			c.operations[i], c.operations[j] = c.operations[j], c.operations[i]
		})
		for _, operation := range c.operations {
			err := c.cooldown(ctx)
			if err != nil {
				return
			}
			err = c.check(ctx, operation)
			if err != nil && ctx.Err() != nil {
				return
			}
		}
	}
}

func (c *Checker) cooldown(ctx context.Context) error {
	cooldownDuration := cooldownDurationMin + time.Duration(rand.Int63n(int64(cooldownDurationMax-cooldownDurationMin)))
	c.logger.Info("begin cooldown", zap.Stringer("duration", cooldownDuration))
	cooldownTimer := time.NewTimer(cooldownDuration)
	defer cooldownTimer.Stop()
	select {
	case <-ctx.Done():
		c.logger.Info("cancel cooldown")
		return ctx.Err()
	case <-cooldownTimer.C:
	}
	c.logger.Info("end cooldown")
	return nil
}

func (c *Checker) check(ctx context.Context, operation Operation) error {
	logger := c.logger.With(zap.String("operation", operation.Name()))
	logger.Info("begin checking")

	incrementCheckFailureCounter := func(reason string) {
		checkFailureCounter.With(map[string]string{
			"cluster":   c.clusterName,
			"operation": operation.Name(),
			"reason":    reason,
		}).Add(1)
	}

	err := c.waitForPreCondition(ctx, logger, operation)
	if err != nil {
		logger.Info("cancel checking")
		return err
	}

	timedCtx, cancel := context.WithTimeout(ctx, completionTimeout)
	defer cancel()
	pingerCtx, cancelPinger := context.WithCancelCause(timedCtx)
	defer cancelPinger(nil)

	wg := &sync.WaitGroup{}

	pingerParams := []struct {
		Suffix  string
		DoWrite bool
	}{
		{
			Suffix:  "all",
			DoWrite: false,
		},
		{
			Suffix:  "primary",
			DoWrite: true,
		},
		{
			Suffix:  "primary",
			DoWrite: false,
		},
		{
			Suffix:  "replica",
			DoWrite: false,
		},
	}
	for _, param := range pingerParams {
		param := param
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.pinger(pingerCtx, logger, operation.Name(), param.Suffix, param.DoWrite)
		}()
	}

	time.Sleep(time.Second) // in order to execute the operation after the first ping

	err = func() error {
		logger.Info("executing operation")
		err := operation.Execute(timedCtx)
		if err != nil {
			if ctx.Err() == nil {
				if timedCtx.Err() != nil {
					logger.Error("operation execution timed out")
					incrementCheckFailureCounter("execution_timeout")
				} else {
					logger.Error("operation execution failed", zap.Error(err))
					incrementCheckFailureCounter("execution_failure")
				}
				cancelPinger(errCancelledByFailure)
			}
			return err
		}
		logger.Info("operation executed")

		startTime := time.Now()
		completionCheckTicker := time.NewTicker(completionCheckInterval)
		defer completionCheckTicker.Stop()
		completionWarningTicker := time.NewTicker(completionWarningInterval)
		defer completionWarningTicker.Stop()
		for {
			select {
			case <-timedCtx.Done():
				if ctx.Err() == nil {
					logger.Error("operation completion timed out")
					incrementCheckFailureCounter("completion_timeout")
					cancelPinger(errCancelledByFailure)
				}
				return timedCtx.Err()
			case <-completionWarningTicker.C:
				logger.Warn("operation is not completed yet", zap.Stringer("duration", time.Since(startTime).Round(time.Second)))
				continue
			case <-completionCheckTicker.C:
			}

			completed, err := operation.CheckCompletion(timedCtx)
			if err != nil {
				if ctx.Err() == nil {
					logger.Error("operation completion check failed")
					incrementCheckFailureCounter("completion_failure")
					cancelPinger(errCancelledByFailure)
				}
				return err
			}
			if completed {
				break
			}
		}

		cancelPinger(errCancelledByCompletion)
		logger.Info("operation completed")
		return nil
	}()

	wg.Wait()

	if err != nil {
		if ctx.Err() != nil {
			logger.Info("cancel checking")
		} else {
			logger.Error("check failed")
		}
		return err
	}

	logger.Info("end checking")
	return nil
}

func (c *Checker) waitForPreCondition(ctx context.Context, logger *zap.Logger, operation Operation) error {
	startTime := time.Now()
	preconditionCheckTicker := time.NewTicker(preconditionCheckInterval)
	defer preconditionCheckTicker.Stop()
	preconditionWarningTicker := time.NewTicker(preconditionWarningInterval)
	defer preconditionWarningTicker.Stop()
	for {
		met, err := operation.CheckPreCondition(ctx)
		if err != nil && ctx.Err() != nil {
			return err
		}
		if met {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-preconditionWarningTicker.C:
			logger.Warn("precondition is not met yet", zap.Stringer("duration", time.Since(startTime).Round(time.Second)))
		case <-preconditionCheckTicker.C:
		}
	}

	logger.Info("precondition is met")
	return nil
}

// pinger periodically checks availability of a MySQL endpoint.
// ctx is the context. If ctx is cancelled and its cause is errCancelledByCompletion, this function finish checking and report the result.
// logger is the logger. This function internally creates a child logger with endpoint field.
// operation is the operation name. It is used only for metrics disposition.
// endpoint is the endpoint kind. If endpoint is "all", this function uses the endpoint which all replicas belong to. Otherwise, this function uses the endpoint as address suffix.
// doWrite specifies whether to write or read. If true, execute an UPDATE statement. If false, execute a SELECT statement.
func (c *Checker) pinger(ctx context.Context, logger *zap.Logger, operation string, endpoint string, doWrite bool) {
	logger = logger.With(zap.String("endpoint", endpoint), zap.Bool("write", doWrite))
	logger.Debug("begin pinger")

	serviceSuffix := ""
	if endpoint != "all" {
		serviceSuffix = "-" + endpoint
	}
	serviceFQDN := "moco-" + c.clusterName + serviceSuffix + "." + currentNamespace + ".svc"

	mtx := sync.Mutex{}
	unixEpoch := time.Unix(0, 0)
	firstFailed := unixEpoch
	lastFailed := firstFailed
	failCount := 0

	pingerTicker := time.NewTicker(pingerInterval)
	defer pingerTicker.Stop()

	wg := &sync.WaitGroup{}

pingLoop:
	for {
		wg.Add(1)
		go func() {
			defer wg.Done()

			now := time.Now()
			succeeded := c.pingOnce(ctx, logger, serviceFQDN, doWrite)

			if !succeeded {
				func() {
					mtx.Lock()
					defer mtx.Unlock()
					failCount++
					if lastFailed.Before(now) {
						lastFailed = now
					}
					if firstFailed == unixEpoch || firstFailed.After(now) {
						firstFailed = now
					}
				}()
			}
		}()

		select {
		case <-ctx.Done():
			break pingLoop
		case <-pingerTicker.C:
		}
	}

	logger.Debug("waiting for goroutines")
	wg.Wait()
	logger.Debug("goroutines finished")

	if context.Cause(ctx) != errCancelledByCompletion {
		logger.Debug("cancel pinger")
		return
	}

	grossDuration := time.Duration(0)
	if firstFailed != unixEpoch {
		grossDuration = lastFailed.Sub(firstFailed).Round(pingerInterval) + pingerInterval
	}
	netDuration := pingerInterval * time.Duration(failCount)

	logger.Info("downtime",
		zap.Int("count", failCount),
		zap.Float64("gross", grossDuration.Seconds()),
		zap.Float64("net", netDuration.Seconds()))

	labels := map[string]string{
		"cluster":   c.clusterName,
		"operation": operation,
		"endpoint":  endpoint,
		"write":     fmt.Sprint(doWrite),
	}
	downtimeGrossHistogramVec.With(labels).Observe(grossDuration.Seconds())
	downtimeNetHistogramVec.With(labels).Observe(netDuration.Seconds())

	logger.Debug("end pinger")
}

// pingOnce checks the availability of the endpoint once.
func (c *Checker) pingOnce(ctx context.Context, logger *zap.Logger, serviceFQDN string, doWrite bool) bool {
	pingCtx, cancel := context.WithTimeout(ctx, pingerTimeout)
	defer cancel()

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s)/%s", username, c.password, serviceFQDN, databaseName))
	if err != nil {
		// sql.Open just create an object and does not connect to the database.
		// So, its error should be reported.
		logger.Error("sql.Open failed", zap.Error(err))
		return false
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if doWrite {
		stmt, err := db.PrepareContext(pingCtx, "UPDATE tbl SET val=? WHERE id=1")
		if err != nil {
			if ctx.Err() == nil {
				logger.Debug("sql.DB.PrepareContext failed", zap.Error(err))
			}
			return false
		}
		defer stmt.Close()

		result, err := stmt.ExecContext(pingCtx, rand.Int31())
		if err != nil {
			if ctx.Err() == nil {
				logger.Debug("sql.Stmt.ExecContext failed", zap.Error(err))
			}
			return false
		}

		num, err := result.RowsAffected()
		if err != nil {
			if ctx.Err() == nil {
				logger.Debug("sql.Result.RowsAffected failed", zap.Error(err))
			}
			return false
		}
		if num != 1 {
			logger.Debug("the number of rows affected is not 1")
			return false
		}
	} else {
		stmt, err := db.PrepareContext(pingCtx, "SELECT val FROM tbl WHERE id=0")
		if err != nil {
			if ctx.Err() == nil {
				logger.Debug("sql.DB.PrepareContext failed", zap.Error(err))
			}
			return false
		}
		defer stmt.Close()

		row := stmt.QueryRowContext(pingCtx)
		var val int
		err = row.Scan(&val)
		if err != nil {
			if ctx.Err() == nil {
				if err == sql.ErrNoRows {
					logger.Debug("sql.Row.Scan failed", zap.Error(err))
				} else {
					logger.Debug("the number of rows returned is not 1")
				}
			}
			return false
		}
	}

	return true
}
