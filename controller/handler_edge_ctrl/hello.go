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

package handler_edge_ctrl

import (
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"google.golang.org/protobuf/proto"
)

type helloHandler struct {
	appEnv   *env.AppEnv
	callback func(routerId string, msg *channel.Message, respHello *edge_ctrl_pb.ClientHello)
}

func NewHelloHandler(appEnv *env.AppEnv, callback func(routerId string, msg *channel.Message, respHello *edge_ctrl_pb.ClientHello)) *helloHandler {
	logtrace.LogWithFunctionName()
	return &helloHandler{
		appEnv:   appEnv,
		callback: callback,
	}
}

func (h *helloHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return env.ClientHelloType
}

func (h *helloHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	respHello := &edge_ctrl_pb.ClientHello{}
	if err := proto.Unmarshal(msg.Body, respHello); err != nil {
		pfxlog.Logger().WithError(err).Error("could not unmarshal clientHello after serverHello")
		return
	}

	h.callback(ch.Id(), msg, respHello)
}
