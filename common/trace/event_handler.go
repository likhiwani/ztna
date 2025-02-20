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

package trace

import (
	logtrace "ztna-core/ztna/logtrace"

	trace_pb "github.com/openziti/channel/v3/trace/pb"
	"github.com/openziti/foundation/v2/concurrenz"
)

var EventHandlerRegistry = concurrenz.CopyOnWriteSlice[EventHandler]{}

func AddTraceEventHandler(handler EventHandler) {
	logtrace.LogWithFunctionName()
	EventHandlerRegistry.Append(handler)
}

func RemoveTraceEventHandler(handler EventHandler) {
	logtrace.LogWithFunctionName()
	EventHandlerRegistry.Delete(handler)
}

// EventHandler is for types wishing to receive trace messages
type EventHandler interface {
	Accept(event *trace_pb.ChannelMessage)
}

type eventWrapper struct {
	wrapped *trace_pb.ChannelMessage
}

func (event *eventWrapper) handle(impl *controllerImpl) {
	logtrace.LogWithFunctionName()
	for _, handler := range EventHandlerRegistry.Value() {
		handler.Accept(event.wrapped)
	}
}
