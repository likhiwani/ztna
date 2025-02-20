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

	"github.com/pkg/errors"
)

func (self *Dispatcher) AddCircuitEventHandler(handler event.CircuitEventHandler) {
	logtrace.LogWithFunctionName()
	self.circuitEventHandlers.Append(handler)
}

func (self *Dispatcher) RemoveCircuitEventHandler(handler event.CircuitEventHandler) {
	logtrace.LogWithFunctionName()
	self.circuitEventHandlers.DeleteIf(func(val event.CircuitEventHandler) bool {
		if val == handler {
			return true
		}
		if w, ok := val.(event.CircuitEventHandlerWrapper); ok {
			return w.IsWrapping(handler)
		}
		return false
	})
}

func (self *Dispatcher) AcceptCircuitEvent(event *event.CircuitEvent) {
	logtrace.LogWithFunctionName()
	go func() {
		for _, handler := range self.circuitEventHandlers.Value() {
			handler.AcceptCircuitEvent(event)
		}
	}()
}

func (self *Dispatcher) registerCircuitEventHandler(val interface{}, config map[string]interface{}) error {
	logtrace.LogWithFunctionName()
	handler, ok := val.(event.CircuitEventHandler)

	if !ok {
		return errors.Errorf("type %v doesn't implement ztna-core/ztna/controller/event/CircuitEventHandler interface.", reflect.TypeOf(val))
	}

	var includeList []string
	if includeVar, ok := config["include"]; ok {
		if includeStr, ok := includeVar.(string); ok {
			includeList = append(includeList, includeStr)
		} else if includeIntfList, ok := includeVar.([]interface{}); ok {
			for _, val := range includeIntfList {
				includeList = append(includeList, fmt.Sprintf("%v", val))
			}
		} else {
			return errors.Errorf("invalid type %v for fabric.circuits include configuration", reflect.TypeOf(includeVar))
		}
	}

	if len(includeList) == 0 {
		self.AddCircuitEventHandler(handler)
		return nil
	}

	accepted := map[event.CircuitEventType]struct{}{}
	for _, include := range includeList {
		found := false
		for _, t := range event.CircuitEventTypes {
			if include == string(t) {
				accepted[t] = struct{}{}
				found = true
				break
			}
		}
		if !found {
			return errors.Errorf("invalid include %v for fabric.circuits. valid values are %+v", include, event.CircuitEventTypes)
		}
	}
	result := &filteredCircuitEventHandler{
		accepted: accepted,
		wrapped:  handler,
	}
	self.AddCircuitEventHandler(result)
	return nil
}

func (self *Dispatcher) unregisterCircuitEventHandler(val interface{}) {
	logtrace.LogWithFunctionName()
	if handler, ok := val.(event.CircuitEventHandler); ok {
		self.RemoveCircuitEventHandler(handler)
	}
}

type filteredCircuitEventHandler struct {
	accepted map[event.CircuitEventType]struct{}
	wrapped  event.CircuitEventHandler
}

func (self *filteredCircuitEventHandler) IsWrapping(value event.CircuitEventHandler) bool {
	logtrace.LogWithFunctionName()
	if self.wrapped == value {
		return true
	}
	if w, ok := self.wrapped.(event.CircuitEventHandlerWrapper); ok {
		return w.IsWrapping(value)
	}
	return false
}

func (self *filteredCircuitEventHandler) AcceptCircuitEvent(event *event.CircuitEvent) {
	logtrace.LogWithFunctionName()
	if _, found := self.accepted[event.EventType]; found {
		self.wrapped.AcceptCircuitEvent(event)
	}
}
