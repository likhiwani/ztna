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
	"ztna-core/ztna/common"
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type updateTerminatorHandler struct {
	baseRequestHandler
}

func NewUpdateTerminatorHandler(appEnv *env.AppEnv, ch channel.Channel) channel.TypedReceiveHandler {
	logtrace.LogWithFunctionName()
	return &updateTerminatorHandler{
		baseRequestHandler{
			ch:     ch,
			appEnv: appEnv,
		},
	}
}

func (self *updateTerminatorHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(edge_ctrl_pb.ContentType_UpdateTerminatorRequestType)
}

func (self *updateTerminatorHandler) Label() string {
	logtrace.LogWithFunctionName()
	return "update.terminator"
}

func (self *updateTerminatorHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	req := &edge_ctrl_pb.UpdateTerminatorRequest{}
	if err := proto.Unmarshal(msg.Body, req); err != nil {
		pfxlog.ContextLogger(ch.Label()).WithError(err).Error("could not unmarshal UpdateTerminator")
		return
	}

	ctx := &UpdateTerminatorRequestContext{
		baseSessionRequestContext: baseSessionRequestContext{handler: self, msg: msg, env: self.appEnv},
		req:                       req,
	}

	go self.UpdateTerminator(ctx)
}

func (self *updateTerminatorHandler) UpdateTerminator(ctx *UpdateTerminatorRequestContext) {
	logtrace.LogWithFunctionName()
	if !ctx.loadRouter() {
		return
	}

	logger := logrus.
		WithField("routerId", self.ch.Id()).
		WithField("token", ctx.req.SessionToken).
		WithField("terminatorId", ctx.req.TerminatorId).
		WithField("cost", ctx.req.Cost).
		WithField("updateCost", ctx.req.UpdateCost).
		WithField("precedence", ctx.req.Precedence).
		WithField("updatePrecedence", ctx.req.UpdatePrecedence)

	logger.Debug("update request received")

	ctx.loadSession(ctx.req.SessionToken, ctx.req.ApiSessionToken)
	ctx.checkSessionType(db.SessionTypeBind)
	ctx.checkSessionFingerprints(ctx.req.Fingerprints)
	terminator := ctx.verifyTerminator(ctx.req.TerminatorId, common.EdgeBinding)

	ctx.updateTerminator(terminator, ctx.req, ctx.newChangeContext())

	if ctx.err != nil {
		self.returnError(ctx, ctx.err)
		return
	}

	logger = logger.WithField("serviceId", terminator.Service)
	logger.Info("updated terminator")

	responseMsg := channel.NewMessage(int32(edge_ctrl_pb.ContentType_UpdateTerminatorResponseType), nil)
	responseMsg.ReplyTo(ctx.msg)
	if err := self.ch.Send(responseMsg); err != nil {
		logger.WithError(err).Error("failed to send update terminator response")
	}
}

type UpdateTerminatorRequestContext struct {
	baseSessionRequestContext
	req *edge_ctrl_pb.UpdateTerminatorRequest
}

func (self *UpdateTerminatorRequestContext) GetSessionToken() string {
	logtrace.LogWithFunctionName()
	return self.req.SessionToken
}
