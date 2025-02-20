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

func (self *Dispatcher) AddClusterEventHandler(handler event.ClusterEventHandler) {
	logtrace.LogWithFunctionName()
	self.clusterEventHandlers.Append(handler)
}

func (self *Dispatcher) RemoveClusterEventHandler(handler event.ClusterEventHandler) {
	logtrace.LogWithFunctionName()
	self.clusterEventHandlers.Delete(handler)
}

func (self *Dispatcher) AcceptClusterEvent(event *event.ClusterEvent) {
	logtrace.LogWithFunctionName()
	event.EventSrcId = self.ctrlId
	go func() {
		for _, handler := range self.clusterEventHandlers.Value() {
			handler.AcceptClusterEvent(event)
		}
	}()
}

func (self *Dispatcher) registerClusterEventHandler(val interface{}, _ map[string]interface{}) error {
	logtrace.LogWithFunctionName()
	handler, ok := val.(event.ClusterEventHandler)

	if !ok {
		return errors.Errorf("type %v doesn't implement ztna-core/ztna/controller/event/ClusterEventHandler interface.", reflect.TypeOf(val))
	}

	self.clusterEventHandlers.Append(handler)

	return nil
}

func (self *Dispatcher) unregisterClusterEventHandler(val interface{}) {
	logtrace.LogWithFunctionName()
	if handler, ok := val.(event.ClusterEventHandler); ok {
		self.RemoveClusterEventHandler(handler)
	}
}
