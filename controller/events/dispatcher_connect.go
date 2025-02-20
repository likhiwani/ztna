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
	"ztna-core/ztna/controller/event"
	"ztna-core/ztna/logtrace"

	"github.com/pkg/errors"
)

func (self *Dispatcher) AddConnectEventHandler(handler event.ConnectEventHandler) {
	logtrace.LogWithFunctionName()
	self.connectEventHandlers.Append(handler)
}

func (self *Dispatcher) RemoveConnectEventHandler(handler event.ConnectEventHandler) {
	logtrace.LogWithFunctionName()
	self.connectEventHandlers.DeleteIf(func(val event.ConnectEventHandler) bool {
		if val == handler {
			return true
		}
		if w, ok := val.(event.ConnectEventHandlerWrapper); ok {
			return w.IsWrapping(handler)
		}
		return false
	})
}

func (self *Dispatcher) AcceptConnectEvent(evt *event.ConnectEvent) {
	logtrace.LogWithFunctionName()
	evt.EventSrcId = self.ctrlId
	for _, handler := range self.connectEventHandlers.Value() {
		go handler.AcceptConnectEvent(evt)
	}
}

func (self *Dispatcher) registerConnectEventHandler(val interface{}, _ map[string]interface{}) error {
	logtrace.LogWithFunctionName()
	handler, ok := val.(event.ConnectEventHandler)

	if !ok {
		return errors.Errorf("type %v doesn't implement ztna-core/ztna/controller/event/ConnectEventHandler interface.", reflect.TypeOf(val))
	}

	self.AddConnectEventHandler(handler)
	return nil
}

func (self *Dispatcher) unregisterConnectEventHandler(val interface{}) {
	logtrace.LogWithFunctionName()
	if handler, ok := val.(event.ConnectEventHandler); ok {
		self.RemoveConnectEventHandler(handler)
	}
}
