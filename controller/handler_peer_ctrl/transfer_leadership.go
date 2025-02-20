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
	"ztna-core/ztna/common/pb/cmd_pb"
	"ztna-core/ztna/controller/peermsg"
	"ztna-core/ztna/controller/raft"
	"ztna-core/ztna/logtrace"

	raft2 "github.com/hashicorp/raft"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

func newTransferLeadershipHandler(controller *raft.Controller) channel.TypedReceiveHandler {
	logtrace.LogWithFunctionName()
	return &transferLeadershipHandler{
		controller: controller,
	}
}

type transferLeadershipHandler struct {
	controller *raft.Controller
}

func (self *transferLeadershipHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(cmd_pb.ContentType_TransferLeadershipRequestType)
}

func (self *transferLeadershipHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label())
	request := &cmd_pb.TransferLeadershipRequest{}
	if err := proto.Unmarshal(msg.Body, request); err != nil {
		log.WithError(err).Error("failed to unmarshal transfer leadership message")
		go sendErrorResponse(msg, ch, err, peermsg.ErrorCodeBadMessage)
		return
	}
	go self.handleTransferLeadership(msg, ch, request)
}

func (self *transferLeadershipHandler) handleTransferLeadership(m *channel.Message, ch channel.Channel, req *cmd_pb.TransferLeadershipRequest) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label())

	log.Infof("received transfer leadership id: %v", req.Id)

	if err := self.controller.HandleTransferLeadership(req); err != nil {
		if errors.Is(err, raft2.ErrNotLeader) {
			sendErrorResponse(m, ch, err, peermsg.ErrorCodeNotLeader)
		} else {
			sendErrorResponse(m, ch, err, peermsg.ErrorCodeGeneric)
		}
	} else {
		resp := channel.NewMessage(int32(cmd_pb.ContentType_SuccessResponseType), nil)
		resp.ReplyTo(m)
		if sendErr := ch.Send(resp); sendErr != nil {
			logrus.WithError(sendErr).Error("error while sending success response")
		}
	}
}
