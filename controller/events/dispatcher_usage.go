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
	"fmt"
	"reflect"
	"ztna-core/ztna/controller/event"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/metrics/metrics_pb"
	"github.com/pkg/errors"
)

func (self *Dispatcher) AddUsageEventHandler(handler event.UsageEventHandler) {
	logtrace.LogWithFunctionName()
	self.usageEventHandlers.Append(handler)
}

func (self *Dispatcher) RemoveUsageEventHandler(handler event.UsageEventHandler) {
	logtrace.LogWithFunctionName()
	self.usageEventHandlers.Delete(handler)
}

func (self *Dispatcher) AddUsageEventV3Handler(handler event.UsageEventV3Handler) {
	logtrace.LogWithFunctionName()
	self.usageEventV3Handlers.Append(handler)
}

func (self *Dispatcher) RemoveUsageEventV3Handler(handler event.UsageEventV3Handler) {
	logtrace.LogWithFunctionName()
	self.usageEventV3Handlers.DeleteIf(func(val event.UsageEventV3Handler) bool {
		if val == handler {
			return true
		}
		if w, ok := val.(event.UsageEventV3HandlerWrapper); ok {
			return w.IsWrapping(handler)
		}
		return false
	})
}

func (self *Dispatcher) AcceptUsageEvent(event *event.UsageEvent) {
	logtrace.LogWithFunctionName()
	go func() {
		for _, handler := range self.usageEventHandlers.Value() {
			handler.AcceptUsageEvent(event)
		}
	}()
}

func (self *Dispatcher) AcceptUsageEventV3(event *event.UsageEventV3) {
	logtrace.LogWithFunctionName()
	go func() {
		for _, handler := range self.usageEventV3Handlers.Value() {
			handler.AcceptUsageEventV3(event)
		}
	}()
}

func (self *Dispatcher) registerUsageEventHandler(val interface{}, config map[string]interface{}) error {
	logtrace.LogWithFunctionName()
	version := 2

	if configVal, found := config["version"]; found {
		strVal := fmt.Sprintf("%v", configVal)
		if strVal == "2" {
			version = 2
		} else if strVal == "3" {
			version = 3
		} else {
			return errors.Errorf("unsupported usage version: %v", version)
		}
	}

	if version == 2 {
		handler, ok := val.(event.UsageEventHandler)
		if !ok {
			return errors.Errorf("type %v doesn't implement ztna-core/ztna/controller/event/UsageEventHandler interface.", reflect.TypeOf(val))
		}
		self.AddUsageEventHandler(handler)
	} else if version == 3 {
		handler, ok := val.(event.UsageEventV3Handler)
		if !ok {
			return errors.Errorf("type %v doesn't implement ztna-core/ztna/controller/event/UsageEventV3Handler interface.", reflect.TypeOf(val))
		}

		if includeListVal, found := config["include"]; found {
			includes := map[string]struct{}{}
			if list, ok := includeListVal.([]interface{}); ok {
				for _, includeVal := range list {
					if include, ok := includeVal.(string); ok {
						includes[include] = struct{}{}
					} else {
						return errors.Errorf("invalid value type [%T] for usage include list, must be string list", val)
					}
				}
			} else {
				return errors.Errorf("invalid value type [%T] for usage include list, must be string list", val)
			}

			if len(includes) == 0 {
				return errors.Errorf("no values provided in include list for usage events, either drop includes stanza or provide at least one usage type to include")
			}

			handler = &filteredUsageV3EventHandler{
				include: includes,
				wrapped: handler,
			}
		}

		self.AddUsageEventV3Handler(handler)
	} else {
		return errors.Errorf("unsupported usage version: %v", version)
	}
	return nil
}

func (self *Dispatcher) unregisterUsageEventHandler(val interface{}) {
	logtrace.LogWithFunctionName()
	if handler, ok := val.(event.UsageEventHandler); ok {
		self.RemoveUsageEventHandler(handler)
	}

	if handler, ok := val.(event.UsageEventV3Handler); ok {
		self.RemoveUsageEventV3Handler(handler)
	}
}

func (self *Dispatcher) initUsageEvents() {
	logtrace.LogWithFunctionName()
	self.AddMetricsMessageHandler(&usageEventAdapter{
		dispatcher: self,
	})
}

type usageEventAdapter struct {
	dispatcher *Dispatcher
}

func (self *usageEventAdapter) AcceptMetricsMsg(message *metrics_pb.MetricsMessage) {
	logtrace.LogWithFunctionName()
	if message.DoNotPropagate {
		return
	}

	if len(self.dispatcher.usageEventHandlers.Value()) > 0 {
		for name, interval := range message.IntervalCounters {
			for _, bucket := range interval.Buckets {
				for circuitId, usage := range bucket.Values {
					evt := &event.UsageEvent{
						Namespace:        event.UsageEventsNs,
						Version:          event.UsageEventsVersion,
						EventSrcId:       self.dispatcher.ctrlId,
						EventType:        name,
						SourceId:         message.SourceId,
						CircuitId:        circuitId,
						Usage:            usage,
						IntervalStartUTC: bucket.IntervalStartUTC,
						IntervalLength:   interval.IntervalLength,
					}
					self.dispatcher.AcceptUsageEvent(evt)
				}
			}
		}
		for _, interval := range message.UsageCounters {
			for circuitId, bucket := range interval.Buckets {
				for usageType, usage := range bucket.Values {
					evt := &event.UsageEvent{
						Namespace:        event.UsageEventsNs,
						Version:          event.UsageEventsVersion,
						EventType:        "usage." + usageType,
						EventSrcId:       self.dispatcher.ctrlId,
						SourceId:         message.SourceId,
						CircuitId:        circuitId,
						Usage:            usage,
						IntervalStartUTC: interval.IntervalStartUTC,
						IntervalLength:   interval.IntervalLength,
						Tags:             bucket.Tags,
					}
					self.dispatcher.AcceptUsageEvent(evt)
				}
			}
		}
	}

	if len(self.dispatcher.usageEventV3Handlers.Value()) > 0 {
		for name, interval := range message.IntervalCounters {
			for _, bucket := range interval.Buckets {
				for circuitId, usage := range bucket.Values {
					evt := &event.UsageEventV3{
						Namespace:  event.UsageEventsNs,
						Version:    3,
						EventSrcId: self.dispatcher.ctrlId,
						SourceId:   message.SourceId,
						CircuitId:  circuitId,
						Usage: map[string]uint64{
							name: usage,
						},
						IntervalStartUTC: bucket.IntervalStartUTC,
						IntervalLength:   interval.IntervalLength,
					}
					self.dispatcher.AcceptUsageEventV3(evt)
				}
			}
		}

		for _, interval := range message.UsageCounters {
			for circuitId, bucket := range interval.Buckets {
				evt := &event.UsageEventV3{
					Namespace:        event.UsageEventsNs,
					Version:          3,
					SourceId:         message.SourceId,
					EventSrcId:       self.dispatcher.ctrlId,
					CircuitId:        circuitId,
					Usage:            bucket.Values,
					IntervalStartUTC: interval.IntervalStartUTC,
					IntervalLength:   interval.IntervalLength,
					Tags:             bucket.Tags,
				}
				self.dispatcher.AcceptUsageEventV3(evt)
			}
		}
	}
}

type filteredUsageV3EventHandler struct {
	include map[string]struct{}
	wrapped event.UsageEventV3Handler
}

func (self *filteredUsageV3EventHandler) IsWrapping(value event.UsageEventV3Handler) bool {
	logtrace.LogWithFunctionName()
	if self.wrapped == value {
		return true
	}
	if w, ok := self.wrapped.(event.UsageEventV3HandlerWrapper); ok {
		return w.IsWrapping(value)
	}
	return false
}

func (self *filteredUsageV3EventHandler) AcceptUsageEventV3(event *event.UsageEventV3) {
	logtrace.LogWithFunctionName()
	usage := map[string]uint64{}
	for k, v := range event.Usage {
		if _, found := self.include[k]; found {
			usage[k] = v
		}
	}
	// nothing passed filter, skip event
	if len(usage) == 0 {
		return
	}

	// nothing got filtered out, pass through unchanged
	if len(usage) == len(event.Usage) {
		self.wrapped.AcceptUsageEventV3(event)
		return
	}

	newEvent := *event
	newEvent.Usage = usage
	self.wrapped.AcceptUsageEventV3(&newEvent)
}
