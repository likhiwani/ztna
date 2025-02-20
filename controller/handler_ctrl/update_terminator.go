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
	"fmt"
	"math"
	"ztna-core/ztna/common/handler_common"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/controller/xt"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type updateTerminatorHandler struct {
	baseHandler
}

func newUpdateTerminatorHandler(network *network.Network, router *model.Router) *updateTerminatorHandler {
	logtrace.LogWithFunctionName()
	return &updateTerminatorHandler{
		baseHandler: baseHandler{
			router:  router,
			network: network,
		},
	}
}

func (self *updateTerminatorHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(ctrl_pb.ContentType_UpdateTerminatorRequestType)
}

func (self *updateTerminatorHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	request := &ctrl_pb.UpdateTerminatorRequest{}
	if err := proto.Unmarshal(msg.Body, request); err != nil {
		log.WithError(err).Error("failed to unmarshal update terminator message")
		return
	}

	go self.handleUpdateTerminator(msg, ch, request)
}

func (self *updateTerminatorHandler) handleUpdateTerminator(msg *channel.Message, ch channel.Channel, request *ctrl_pb.UpdateTerminatorRequest) {
	logtrace.LogWithFunctionName()
	terminator, err := self.network.Terminator.Read(request.TerminatorId)
	if err != nil {
		handler_common.SendFailure(msg, ch, err.Error())
		return
	}

	if !request.UpdateCost && !request.UpdatePrecedence {
		// nothing to do
		handler_common.SendSuccess(msg, ch, "")
		return
	}

	checker := fields.UpdatedFieldsMap{}

	if request.UpdateCost {
		if request.Cost > math.MaxUint16 {
			handler_common.SendFailure(msg, ch, fmt.Sprintf("invalid static cost %v. Must be less than %v", request.Cost, math.MaxUint16))
			return
		}
		terminator.Cost = uint16(request.Cost)
		checker[db.FieldTerminatorCost] = struct{}{}
	}

	if request.UpdatePrecedence {
		if request.UpdatePrecedence {
			if request.Precedence == ctrl_pb.TerminatorPrecedence_Default {
				terminator.Precedence = xt.Precedences.Default
			} else if request.Precedence == ctrl_pb.TerminatorPrecedence_Required {
				terminator.Precedence = xt.Precedences.Required
			} else if request.Precedence == ctrl_pb.TerminatorPrecedence_Failed {
				terminator.Precedence = xt.Precedences.Failed
			} else {
				handler_common.SendFailure(msg, ch, fmt.Sprintf("invalid precedence: %v", request.Precedence))
				return
			}
		}
		checker[db.FieldTerminatorPrecedence] = struct{}{}
	}

	if err := self.network.Terminator.Update(terminator, checker, self.newChangeContext(ch, "fabric.update.terminator")); err != nil {
		handler_common.SendFailure(msg, ch, err.Error())
		return
	}

	handler_common.SendSuccess(msg, ch, "")
}
