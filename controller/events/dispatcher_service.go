/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package events

import (
	"reflect"
	"strings"
	"ztna-core/ztna/controller/event"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/metrics/metrics_pb"
	"github.com/pkg/errors"
)

func (self *Dispatcher) AddServiceEventHandler(handler event.ServiceEventHandler) {
	logtrace.LogWithFunctionName()
	self.serviceEventHandlers.Append(handler)
}

func (self *Dispatcher) RemoveServiceEventHandler(handler event.ServiceEventHandler) {
	logtrace.LogWithFunctionName()
	self.serviceEventHandlers.Delete(handler)
}

func (self *Dispatcher) AcceptServiceEvent(event *event.ServiceEvent) {
	logtrace.LogWithFunctionName()
	go func() {
		for _, handler := range self.serviceEventHandlers.Value() {
			handler.AcceptServiceEvent(event)
		}
	}()
}

func (self *Dispatcher) registerServiceEventHandler(val interface{}, _ map[string]interface{}) error {
	logtrace.LogWithFunctionName()
	handler, ok := val.(event.ServiceEventHandler)
	if !ok {
		return errors.Errorf("type %v doesn't implement github.com/openziti/edge/event/ServiceEventHandler interface.", reflect.TypeOf(val))
	}

	self.AddServiceEventHandler(handler)
	return nil
}

func (self *Dispatcher) unregisterServiceEventHandler(val interface{}) {
	logtrace.LogWithFunctionName()
	if handler, ok := val.(event.ServiceEventHandler); ok {
		self.RemoveServiceEventHandler(handler)
	}
}

func (self *Dispatcher) initServiceEvents(n *network.Network) {
	logtrace.LogWithFunctionName()
	n.InitServiceCounterDispatch(&serviceEventAdapter{
		Dispatcher: self,
	})
}

// serviceEventAdapter converts service interval counters into service events
type serviceEventAdapter struct {
	*Dispatcher
}

func (self *serviceEventAdapter) AcceptMetrics(message *metrics_pb.MetricsMessage) {
	logtrace.LogWithFunctionName()
	for name, interval := range message.IntervalCounters {
		for _, bucket := range interval.Buckets {
			for combinedId, count := range bucket.Values {
				ids := strings.Split(combinedId, ":")
				serviceId := ids[0]
				terminatorId := ""
				if len(ids) > 1 {
					terminatorId = ids[1]
				}
				evt := &event.ServiceEvent{
					Namespace:        "service.events",
					Version:          2,
					EventType:        name,
					EventSrcId:       self.ctrlId,
					ServiceId:        serviceId,
					TerminatorId:     terminatorId,
					Count:            count,
					IntervalStartUTC: bucket.IntervalStartUTC,
					IntervalLength:   interval.IntervalLength,
				}
				self.Dispatcher.AcceptServiceEvent(evt)
			}
		}
	}
}
