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

func (self *Dispatcher) AddLinkEventHandler(handler event.LinkEventHandler) {
	logtrace.LogWithFunctionName()
	self.linkEventHandlers.Append(handler)
}

func (self *Dispatcher) RemoveLinkEventHandler(handler event.LinkEventHandler) {
	logtrace.LogWithFunctionName()
	self.linkEventHandlers.Delete(handler)
}

func (self *Dispatcher) AcceptLinkEvent(event *event.LinkEvent) {
	logtrace.LogWithFunctionName()
	go func() {
		for _, handler := range self.linkEventHandlers.Value() {
			handler.AcceptLinkEvent(event)
		}
	}()
}

func (self *Dispatcher) registerLinkEventHandler(val interface{}, _ map[string]interface{}) error {
	logtrace.LogWithFunctionName()
	handler, ok := val.(event.LinkEventHandler)

	if !ok {
		return errors.Errorf("type %v doesn't implement ztna-core/ztna/controller/event/LinkEventHandler interface.", reflect.TypeOf(val))
	}

	self.linkEventHandlers.Append(handler)

	return nil
}

func (self *Dispatcher) unregisterLinkEventHandler(val interface{}) {
	logtrace.LogWithFunctionName()
	if handler, ok := val.(event.LinkEventHandler); ok {
		self.RemoveLinkEventHandler(handler)
	}
}
