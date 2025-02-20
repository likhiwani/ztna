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

package handler_peer_ctrl

import (
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"google.golang.org/protobuf/proto"
)

type inspectHandler struct {
	network *network.Network
}

func newInspectHandler(n *network.Network) *inspectHandler {
	logtrace.LogWithFunctionName()
	return &inspectHandler{
		network: n,
	}
}

func (*inspectHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(ctrl_pb.ContentType_InspectRequestType)
}

func (handler *inspectHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	go func() {
		context := &inspectRequestContext{
			handler:  handler,
			msg:      msg,
			ch:       ch,
			request:  &ctrl_pb.InspectRequest{},
			response: &ctrl_pb.InspectResponse{Success: true},
		}

		var err error
		if err = proto.Unmarshal(msg.Body, context.request); err != nil {
			context.appendError(err.Error())
		}

		if !context.response.Success {
			context.sendResponse()
			return
		}

		context.processLocal()
		context.sendResponse()
	}()
}

type inspectRequestContext struct {
	handler  *inspectHandler
	msg      *channel.Message
	ch       channel.Channel
	request  *ctrl_pb.InspectRequest
	response *ctrl_pb.InspectResponse
}

func (context *inspectRequestContext) processLocal() {
	logtrace.LogWithFunctionName()
	result := context.handler.network.Inspections.Inspect(context.handler.network.GetAppId(), context.request.RequestedValues)
	for _, value := range result.Results {
		context.appendValue(value.Name, value.Value)
	}

	for _, err := range result.Errors {
		context.appendError(err)
	}
}

func (context *inspectRequestContext) sendResponse() {
	logtrace.LogWithFunctionName()
	body, err := proto.Marshal(context.response)
	if err != nil {
		pfxlog.Logger().WithError(err).Error("unexpected error serializing InspectResponse")
		return
	}

	responseMsg := channel.NewMessage(int32(ctrl_pb.ContentType_InspectResponseType), body)
	responseMsg.ReplyTo(context.msg)
	if err := context.ch.Send(responseMsg); err != nil {
		pfxlog.Logger().WithError(err).Error("unexpected error sending InspectResponse")
	}
}

func (context *inspectRequestContext) appendValue(name string, value string) {
	logtrace.LogWithFunctionName()
	context.response.Values = append(context.response.Values, &ctrl_pb.InspectResponse_InspectValue{
		Name:  name,
		Value: value,
	})
}

func (context *inspectRequestContext) appendError(err string) {
	logtrace.LogWithFunctionName()
	context.response.Success = false
	context.response.Errors = append(context.response.Errors, err)
}
