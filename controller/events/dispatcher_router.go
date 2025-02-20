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
	"time"
	"ztna-core/ztna/controller/event"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/logtrace"

	"github.com/pkg/errors"
)

func (self *Dispatcher) AddRouterEventHandler(handler event.RouterEventHandler) {
	logtrace.LogWithFunctionName()
	self.routerEventHandlers.Append(handler)
}

func (self *Dispatcher) RemoveRouterEventHandler(handler event.RouterEventHandler) {
	logtrace.LogWithFunctionName()
	self.routerEventHandlers.Delete(handler)
}

func (self *Dispatcher) AcceptRouterEvent(event *event.RouterEvent) {
	logtrace.LogWithFunctionName()
	go func() {
		for _, handler := range self.routerEventHandlers.Value() {
			handler.AcceptRouterEvent(event)
		}
	}()
}

func (self *Dispatcher) initRouterEvents(n *network.Network) {
	logtrace.LogWithFunctionName()
	routerEvtAdapter := &routerEventAdapter{
		Dispatcher: self,
	}
	n.AddRouterPresenceHandler(routerEvtAdapter)
}

func (self *Dispatcher) registerRouterEventHandler(val interface{}, _ map[string]interface{}) error {
	logtrace.LogWithFunctionName()
	handler, ok := val.(event.RouterEventHandler)

	if !ok {
		return errors.Errorf("type %v doesn't implement ztna-core/ztna/controller/event/RouterEventHandler interface.", reflect.TypeOf(val))
	}

	self.AddRouterEventHandler(handler)

	return nil
}

func (self *Dispatcher) unregisterRouterEventHandler(val interface{}) {
	logtrace.LogWithFunctionName()
	if handler, ok := val.(event.RouterEventHandler); ok {
		self.RemoveRouterEventHandler(handler)
	}
}

// routerEventAdapter converts network router presence events to event.RouterEvent
type routerEventAdapter struct {
	*Dispatcher
}

func (self *routerEventAdapter) RouterConnected(r *model.Router) {
	logtrace.LogWithFunctionName()
	self.routerChange(event.RouterOnline, r, true)
}

func (self *routerEventAdapter) RouterDisconnected(r *model.Router) {
	logtrace.LogWithFunctionName()
	self.routerChange(event.RouterOffline, r, false)
}

func (self *routerEventAdapter) routerChange(eventType event.RouterEventType, r *model.Router, online bool) {
	logtrace.LogWithFunctionName()
	evt := &event.RouterEvent{
		Namespace:    event.RouterEventsNs,
		EventSrcId:   self.ctrlId,
		EventType:    eventType,
		Timestamp:    time.Now(),
		RouterId:     r.Id,
		RouterOnline: online,
	}

	self.Dispatcher.AcceptRouterEvent(evt)

	if eventType == event.RouterOnline {
		srcAddr := ""
		dstAddr := ""
		if ctrl := r.Control; ctrl != nil {
			srcAddr = r.Control.Underlay().GetRemoteAddr().String()
			dstAddr = r.Control.Underlay().GetLocalAddr().String()
		}

		connectEvent := &event.ConnectEvent{
			Namespace: event.ConnectEventNS,
			SrcType:   event.ConnectSourceRouter,
			DstType:   event.ConnectDestinationController,
			SrcId:     r.Id,
			SrcAddr:   srcAddr,
			DstId:     self.Dispatcher.ctrlId,
			DstAddr:   dstAddr,
			Timestamp: time.Now(),
		}
		self.Dispatcher.AcceptConnectEvent(connectEvent)
	}
}
