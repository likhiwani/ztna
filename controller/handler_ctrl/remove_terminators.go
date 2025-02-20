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

package handler_ctrl

import (
	"ztna-core/ztna/common/handler_common"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/controller/command"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"google.golang.org/protobuf/proto"
)

type removeTerminatorsHandler struct {
	baseHandler
}

func newRemoveTerminatorsHandler(network *network.Network, router *model.Router) *removeTerminatorsHandler {
	logtrace.LogWithFunctionName()
	return &removeTerminatorsHandler{
		baseHandler: baseHandler{
			router:  router,
			network: network,
		},
	}
}

func (self *removeTerminatorsHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(ctrl_pb.ContentType_RemoveTerminatorsRequestType)
}

func (self *removeTerminatorsHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label())

	request := &ctrl_pb.RemoveTerminatorsRequest{}
	if err := proto.Unmarshal(msg.Body, request); err != nil {
		log.WithError(err).Error("failed to unmarshal remove terminator message")
		return
	}

	go self.handleRemoveTerminators(msg, ch, request)
}

func (self *removeTerminatorsHandler) handleRemoveTerminators(msg *channel.Message, ch channel.Channel, request *ctrl_pb.RemoveTerminatorsRequest) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label())

	if err := self.network.Terminator.DeleteBatch(request.TerminatorIds, self.newChangeContext(ch, "fabric.remove.terminators.batch")); err == nil {
		log.
			WithField("routerId", ch.Id()).
			WithField("terminatorIds", request.TerminatorIds).
			Info("removed terminators")
		handler_common.SendSuccess(msg, ch, "")
	} else if command.WasRateLimited(err) {
		handler_common.SendServerBusy(msg, ch, "remove.terminators")
	} else {
		handler_common.SendFailure(msg, ch, err.Error())
	}
}
