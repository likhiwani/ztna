package db

import (
	"fmt"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/cowslice"
)

type ServiceEventType byte

func (self ServiceEventType) String() string {
	logtrace.LogWithFunctionName()
	if self == ServiceDialAccessGained {
		return "dial-gained"
	}
	if self == ServiceDialAccessLost {
		return "dial-lost"
	}
	if self == ServiceBindAccessGained {
		return "bind-gained"
	}
	if self == ServiceBindAccessLost {
		return "bind-lost"
	}
	if self == ServiceUpdated {
		return "service-updated"
	}

	return "unknown"
}

const (
	ServiceDialAccessGained ServiceEventType = 1
	ServiceDialAccessLost   ServiceEventType = 2
	ServiceBindAccessGained ServiceEventType = 3
	ServiceBindAccessLost   ServiceEventType = 4
	ServiceUpdated          ServiceEventType = 5
)

var ServiceEvents = &ServiceEventsRegistry{
	handlers: cowslice.NewCowSlice(make([]ServiceEventHandler, 0)),
}

func init() {
	logtrace.LogWithFunctionName()
	ServiceEvents.AddServiceEventHandler(func(event *ServiceEvent) {
		pfxlog.Logger().Tracef("identity %v -> service %v %v", event.IdentityId, event.ServiceId, event.Type.String())
	})
}

type ServiceEvent struct {
	Type       ServiceEventType
	IdentityId string
	ServiceId  string
}

func (self *ServiceEvent) String() string {
	logtrace.LogWithFunctionName()
	return fmt.Sprintf("service event [identity %v -> service %v %v]", self.IdentityId, self.ServiceId, self.Type.String())
}

type ServiceEventHandler func(event *ServiceEvent)

type ServiceEventsRegistry struct {
	handlers *cowslice.CowSlice
}

func (self *ServiceEventsRegistry) AddServiceEventHandler(listener ServiceEventHandler) {
	logtrace.LogWithFunctionName()
	cowslice.Append(self.handlers, listener)
}

func (self *ServiceEventsRegistry) RemoveServiceEventHandler(listener ServiceEventHandler) {
	logtrace.LogWithFunctionName()
	cowslice.Delete(self.handlers, listener)
}

func (self *ServiceEventsRegistry) dispatchEventsAsync(events []*ServiceEvent) {
	logtrace.LogWithFunctionName()
	go self.dispatchEvents(events)
}

func (self *ServiceEventsRegistry) dispatchEvents(events []*ServiceEvent) {
	logtrace.LogWithFunctionName()
	for _, event := range events {
		self.dispatchEvent(event)
	}
}

func (self *ServiceEventsRegistry) dispatchEvent(event *ServiceEvent) {
	logtrace.LogWithFunctionName()
	handlers := self.handlers.Value().([]ServiceEventHandler)
	for _, handler := range handlers {
		handler(event)
	}
}
