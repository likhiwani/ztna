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

package handler_mgmt

import (
	"ztna-core/ztna/common/handler_common"
	"ztna-core/ztna/common/pb/mgmt_pb"
	"ztna-core/ztna/common/trace"
	"ztna-core/ztna/controller/network"
	logtrace "ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	trace_pb "github.com/openziti/channel/v3/trace/pb"
	"google.golang.org/protobuf/proto"
)

type streamTracesHandler struct {
	network        *network.Network
	streamHandlers []trace.EventHandler
}

func newStreamTracesHandler(network *network.Network) *streamTracesHandler {
	logtrace.LogWithFunctionName()
	return &streamTracesHandler{network: network}
}

func (*streamTracesHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(mgmt_pb.ContentType_StreamTracesRequestType)
}

func (handler *streamTracesHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	request := &mgmt_pb.StreamTracesRequest{}
	if err := proto.Unmarshal(msg.Body, request); err != nil {
		handler_common.SendFailure(msg, ch, err.Error())
		return
	}

	filter := createFilter(request)
	eventHandler := &traceEventsHandler{ch, filter}

	handler.streamHandlers = append(handler.streamHandlers, eventHandler)
	trace.AddTraceEventHandler(eventHandler)
}

func (handler *streamTracesHandler) HandleClose(channel.Channel) {
	logtrace.LogWithFunctionName()
	for _, streamHandler := range handler.streamHandlers {
		trace.RemoveTraceEventHandler(streamHandler)
	}
}

func createFilter(request *mgmt_pb.StreamTracesRequest) trace.Filter {
	logtrace.LogWithFunctionName()
	if !request.EnabledFilter {
		return trace.NewAllowAllFilter()
	}
	if request.FilterType == mgmt_pb.TraceFilterType_INCLUDE {
		return trace.NewIncludeFilter(request.ContentTypes)
	}
	return trace.NewExcludeFilter(request.ContentTypes)
}

type traceEventsHandler struct {
	ch     channel.Channel
	filter trace.Filter
}

func (handler *traceEventsHandler) Accept(event *trace_pb.ChannelMessage) {
	logtrace.LogWithFunctionName()
	if !handler.filter.Accept(event) {
		return
	}
	body, err := proto.Marshal(event)
	if err != nil {
		pfxlog.Logger().Errorf("unexpected error unmarshalling ChannelMessage (%s)", err)
		return
	}

	responseMsg := channel.NewMessage(int32(mgmt_pb.ContentType_StreamTracesEventType), body)
	if err := handler.ch.Send(responseMsg); err != nil {
		pfxlog.Logger().Errorf("unexpected error sending ChannelMessage (%s)", err)
		handler.close()
	}
}

func (handler *traceEventsHandler) close() {
	logtrace.LogWithFunctionName()
	if err := handler.ch.Close(); err != nil {
		pfxlog.Logger().WithError(err).Error("unexpected error while closing mgmt channel")
	}
}
