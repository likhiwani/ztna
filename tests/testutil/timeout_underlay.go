package testutil

import (
	"time"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
	"github.com/pkg/errors"
)

func NewTimeoutUnderlayFactory(factory channel.UnderlayFactory, timeout time.Duration) *UnderlayFactoryWrapper {
	logtrace.LogWithFunctionName()
	return &UnderlayFactoryWrapper{
		timeout: timeout,
		wrapped: factory,
	}
}

type UnderlayFactoryWrapper struct {
	timeout time.Duration
	wrapped channel.UnderlayFactory
}

func (self *UnderlayFactoryWrapper) Create(timeout time.Duration) (channel.Underlay, error) {
	logtrace.LogWithFunctionName()
	underlayC := make(chan channel.Underlay, 1)
	errC := make(chan error, 1)
	go func() {
		u, err := self.wrapped.Create(timeout)
		if err != nil {
			errC <- err
		} else {
			underlayC <- u
		}
	}()

	select {
	case underlay := <-underlayC:
		return underlay, nil
	case err := <-errC:
		return nil, err
	case <-time.After(self.timeout):
		return nil, errors.New("timed out")
	}
}
