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
	"fmt"
	"sync/atomic"
	"time"
	logtrace "ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/xgress"

	"github.com/openziti/channel/v3"
	trace_pb "github.com/openziti/channel/v3/trace/pb"
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/openziti/identity"
)

type XgressPeekHandler struct {
	appId      *identity.TokenId
	enabled    atomic.Bool
	controller Controller
	decoders   []channel.TraceMessageDecoder
	eventSinks concurrenz.CopyOnWriteSlice[EventHandler]
}

func (self *XgressPeekHandler) EnableTracing(sourceType SourceType, matcher SourceMatcher, handler EventHandler, resultChan chan<- ToggleApplyResult) {
	logtrace.LogWithFunctionName()
	self.ToggleTracing(sourceType, matcher, true, handler, resultChan)
}

func (self *XgressPeekHandler) DisableTracing(sourceType SourceType, matcher SourceMatcher, handler EventHandler, resultChan chan<- ToggleApplyResult) {
	logtrace.LogWithFunctionName()
	self.ToggleTracing(sourceType, matcher, false, handler, resultChan)
}

func (self *XgressPeekHandler) ToggleTracing(sourceType SourceType, matcher SourceMatcher, enable bool, handler EventHandler, resultChan chan<- ToggleApplyResult) {
	logtrace.LogWithFunctionName()
	name := "xgress"
	matched := sourceType == SourceTypePipe && matcher.Matches(name)
	prevState := self.IsEnabled()
	nextState := prevState

	if matched {
		nextState = enable
		if enable {
			self.enabled.Store(true)
			self.eventSinks.Append(handler)
		} else {
			self.eventSinks.Delete(handler)
			if len(self.eventSinks.Value()) == 0 {
				self.enabled.Store(false)
			}
		}
	}

	resultChan <- &ToggleApplyResultImpl{matched,
		fmt.Sprintf("Link %v.%v matched? %v. Old trace state: %v, New trace state: %v",
			self.appId.Token, name, matched, prevState, nextState)}
}

func (self *XgressPeekHandler) Rx(x *xgress.Xgress, payload *xgress.Payload) {
	logtrace.LogWithFunctionName()
	self.trace(x, payload, true)
}

func (self *XgressPeekHandler) Tx(x *xgress.Xgress, payload *xgress.Payload) {
	logtrace.LogWithFunctionName()
	self.trace(x, payload, false)
}

func (self *XgressPeekHandler) Close(*xgress.Xgress) {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func NewXgressPeekHandler(appId *identity.TokenId, controller Controller) *XgressPeekHandler {
	logtrace.LogWithFunctionName()
	handler := &XgressPeekHandler{
		appId:      appId,
		controller: controller,
		decoders:   decoders,
	}
	controller.AddSource(handler)
	return handler
}

func (self *XgressPeekHandler) IsEnabled() bool {
	logtrace.LogWithFunctionName()
	return self.enabled.Load()
}

func (self *XgressPeekHandler) trace(x *xgress.Xgress, payload *xgress.Payload, rx bool) {
	logtrace.LogWithFunctionName()
	decode, _ := xgress.DecodePayload(payload)

	traceMsg := &trace_pb.ChannelMessage{
		Timestamp:   time.Now().UnixNano(),
		Identity:    self.appId.Token,
		Channel:     x.Label(),
		IsRx:        rx,
		ContentType: xgress.ContentTypePayloadType,
		Sequence:    -1,
		ReplyFor:    -1,
		Length:      int32(len(payload.Data)),
		Decode:      decode,
	}

	// This can result in a message send. Doing a send from inside a peekhandler can cause deadlocks, so it's best avoided
	for _, eventSink := range self.eventSinks.Value() {
		go eventSink.Accept(traceMsg)
	}
}
