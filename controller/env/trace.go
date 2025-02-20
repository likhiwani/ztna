package env

import (
	"fmt"
	"time"
	"ztna-core/ztna/logtrace"

	cmap "github.com/orcaman/concurrent-map/v2"
)

func NewTraceManager(shutdownNotify <-chan struct{}) *TraceManager {
	logtrace.LogWithFunctionName()
	result := &TraceManager{
		traceIdentities: cmap.New[*TraceSpec](),
		shutdownNotify:  shutdownNotify,
	}
	go result.reapExpired()
	return result
}

type TraceSpec struct {
	Until       time.Time
	TraceId     string
	ChannelMask uint32
}

func (self *TraceSpec) String() string {
	logtrace.LogWithFunctionName()
	return fmt.Sprintf("traceId=%v until=%v", self.TraceId, self.Until)
}

type TraceManager struct {
	traceIdentities cmap.ConcurrentMap[string, *TraceSpec]
	shutdownNotify  <-chan struct{}
}

func (self *TraceManager) GetIdentityTrace(identityId string) *TraceSpec {
	logtrace.LogWithFunctionName()
	spec, found := self.traceIdentities.Get(identityId)

	if !found {
		return nil
	}

	specCopy := *spec
	return &specCopy
}

func (self *TraceManager) TraceIdentity(identity string, duration time.Duration, id string, channelMask uint32) *TraceSpec {
	logtrace.LogWithFunctionName()
	spec := &TraceSpec{
		Until:       time.Now().Add(duration),
		TraceId:     id,
		ChannelMask: channelMask,
	}
	self.traceIdentities.Set(identity, spec)
	specCopy := *spec
	return &specCopy
}

func (self *TraceManager) RemoveIdentityTrace(identity string) {
	logtrace.LogWithFunctionName()
	self.traceIdentities.Remove(identity)
}

func (self *TraceManager) reapExpired() {
	logtrace.LogWithFunctionName()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			var toRemove []string
			self.traceIdentities.IterCb(func(key string, spec *TraceSpec) {
				if spec.Until.Before(now) {
					toRemove = append(toRemove, key)
				}
			})

			for _, key := range toRemove {
				self.traceIdentities.Remove(key)
			}
		case <-self.shutdownNotify:
			return
		}
	}
}
