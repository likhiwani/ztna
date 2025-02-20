package controller

import (
	"context"
	"sync/atomic"
	"time"
	"ztna-core/ztna/logtrace"

	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/openziti/metrics"
	"github.com/openziti/storage/boltz"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

func (c *Controller) initializeHealthChecks() (gosundheit.Health, error) {
	logtrace.LogWithFunctionName()
	healthChecker := gosundheit.New()
	check, err := checks.NewPingCheck("bolt.read", &boltPinger{
		dbProvider:  c.network.GetDb,
		openReadTxs: c.GetNetwork().GetMetricsRegistry().Gauge("bolt.open_read_txs"),
	})

	if err != nil {
		return nil, err
	}

	err = healthChecker.RegisterCheck(check,
		gosundheit.InitialDelay(c.config.HealthChecks.BoltCheck.InitialDelay),
		gosundheit.ExecutionPeriod(c.config.HealthChecks.BoltCheck.Interval),
		gosundheit.ExecutionTimeout(c.config.HealthChecks.BoltCheck.Timeout),
		gosundheit.InitiallyPassing(true))

	if err != nil {
		return nil, err
	}

	return healthChecker, nil
}

type boltPinger struct {
	dbProvider  func() boltz.Db
	openReadTxs metrics.Gauge
	running     atomic.Bool
}

func (self *boltPinger) PingContext(ctx context.Context) error {
	logtrace.LogWithFunctionName()
	if !self.running.CompareAndSwap(false, true) {
		return errors.Errorf("previous bolt ping is still running")
	}

	deadline, hasDeadline := ctx.Deadline()

	checkFunc := func(tx *bbolt.Tx) error {
		self.openReadTxs.Update(int64(tx.DB().Stats().OpenTxN))
		return nil
	}

	if !hasDeadline {
		defer self.running.Store(false)
		return self.dbProvider().View(checkFunc)
	}

	errC := make(chan error, 1)
	go func() {
		defer self.running.Store(false)
		errC <- self.dbProvider().View(checkFunc)
	}()

	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case err := <-errC:
		return err
	case <-timer.C:
		return errors.Errorf("bolt ping timed out")
	}
}
