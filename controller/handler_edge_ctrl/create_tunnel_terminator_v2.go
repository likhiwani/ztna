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
	"fmt"
	"math"
	"time"
	"ztna-core/ztna/common"
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type createTunnelTerminatorV2Handler struct {
	baseRequestHandler
	*TunnelState
}

func NewCreateTunnelTerminatorV2Handler(appEnv *env.AppEnv, ch channel.Channel) channel.TypedReceiveHandler {
	logtrace.LogWithFunctionName()
	return &createTunnelTerminatorV2Handler{
		baseRequestHandler: baseRequestHandler{ch: ch, appEnv: appEnv},
	}
}

func (self *createTunnelTerminatorV2Handler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(edge_ctrl_pb.ContentType_CreateTunnelTerminatorRequestV2Type)
}

func (self *createTunnelTerminatorV2Handler) Label() string {
	logtrace.LogWithFunctionName()
	return "tunnel.create.terminator.v2"
}

func (self *createTunnelTerminatorV2Handler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	startTime := time.Now()
	req := &edge_ctrl_pb.CreateTunnelTerminatorRequestV2{}
	if err := proto.Unmarshal(msg.Body, req); err != nil {
		pfxlog.ContextLogger(ch.Label()).WithError(err).Error("could not unmarshal CreateTunnelTerminatorRequestV2")
		return
	}

	ctx := &createTunnelTerminatorV2RequestContext{
		baseTunnelRequestContext: baseTunnelRequestContext{
			baseSessionRequestContext: baseSessionRequestContext{handler: self, msg: msg, env: self.appEnv},
		},
		req: req,
	}

	go self.CreateTerminator(ctx, startTime)
}

func (self *createTunnelTerminatorV2Handler) CreateTerminator(ctx *createTunnelTerminatorV2RequestContext, startTime time.Time) {
	logtrace.LogWithFunctionName()
	logger := logrus.
		WithField("routerId", self.ch.Id()).
		WithField("terminatorId", ctx.req.Address)

	if !ctx.loadRouter() {
		return
	}
	ctx.loadIdentity()
	ctx.loadServiceForId(ctx.req.ServiceId)
	ctx.verifyEdgeRouterServiceBindAccess()
	ctx.verifyRouterEdgeRouterAccess()

	if ctx.err != nil {
		self.logResult(ctx, ctx.err)
		return
	}

	logger = logger.WithField("serviceId", ctx.req.ServiceId).WithField("service", ctx.service.Name)

	if ctx.req.Cost > math.MaxUint16 {
		self.returnError(ctx, invalidCost(fmt.Sprintf("invalid cost %v. cost must be between 0 and %v inclusive", ctx.req.Cost, math.MaxUint16)))
		return
	}

	terminator, _ := self.getNetwork().Terminator.Read(ctx.req.Address)
	if terminator != nil {
		if err := ctx.validateExistingTerminator(terminator, logger); err != nil {
			self.returnError(ctx, err)
			return
		}
	} else {
		terminator = &model.Terminator{
			BaseEntity: models.BaseEntity{
				Id:       ctx.req.Address,
				IsSystem: true,
			},
			Service:        ctx.req.ServiceId,
			Router:         ctx.sourceRouter.Id,
			Binding:        common.TunnelBinding,
			Address:        ctx.req.Address,
			InstanceId:     ctx.req.InstanceId,
			InstanceSecret: ctx.req.InstanceSecret,
			PeerData:       ctx.req.PeerData,
			Precedence:     ctx.req.GetXtPrecedence(),
			Cost:           uint16(ctx.req.Cost),
			HostId:         ctx.identity.Id,
			SourceCtrl:     self.appEnv.GetId(),
		}

		if err := self.appEnv.Managers.Terminator.Create(terminator, ctx.newTunnelChangeContext()); err != nil {
			// terminator might have been created while we were trying to create.
			if terminator, _ = self.getNetwork().Terminator.Read(ctx.req.Address); terminator != nil {
				if validateError := ctx.validateExistingTerminator(terminator, logger); validateError != nil {
					self.returnError(ctx, validateError)
					return
				}
			} else {
				self.returnError(ctx, internalError(err))
				return
			}
		} else {
			logger.Info("created terminator")
		}
	}

	response := &edge_ctrl_pb.CreateTunnelTerminatorResponseV2{
		TerminatorId: ctx.req.Address,
		StartTime:    ctx.req.StartTime,
	}

	body, err := proto.Marshal(response)
	if err != nil {
		logger.WithError(err).Error("failed to marshal CreateTunnelTerminatorResponseV2")
		return
	}

	responseMsg := channel.NewMessage(response.GetContentType(), body)
	responseMsg.ReplyTo(ctx.msg)
	if err = self.ch.Send(responseMsg); err != nil {
		logger.WithError(err).Error("failed to send CreateTunnelTerminatorResponseV2")
	}

	logger.WithField("elapsedTime", time.Since(startTime)).Info("completed create tunnel terminator operation")
}

type createTunnelTerminatorV2RequestContext struct {
	baseTunnelRequestContext
	req *edge_ctrl_pb.CreateTunnelTerminatorRequestV2
}

func (self *createTunnelTerminatorV2RequestContext) validateExistingTerminator(terminator *model.Terminator, log *logrus.Entry) controllerError {
	logtrace.LogWithFunctionName()
	if terminator.Binding != common.TunnelBinding {
		log.WithField("binding", common.TunnelBinding).
			WithField("conflictingBinding", terminator.Binding).
			Error("selected terminator address conflicts with a terminator for a different binding")
		return internalError(errors.New("selected id conflicts with terminator for different binding"))
	}

	if terminator.Service != self.req.ServiceId {
		log.WithField("conflictingService", terminator.Service).
			Error("selected terminator address conflicts with a terminator for a different service")
		return internalError(errors.New("selected id conflicts with terminator for different service"))
	}

	if terminator.Router != self.sourceRouter.Id {
		log.WithField("conflictingRouter", terminator.Router).
			Error("selected terminator address conflicts with a terminator for a different router")
		return internalError(errors.New("selected id conflicts with terminator for different router"))
	}

	if terminator.HostId != self.identity.Id {
		log.WithField("identityId", self.identity.Id).
			WithField("conflictingIdentity", terminator.HostId).
			Error("selected terminator address conflicts with a terminator for a different identity")
		return internalError(errors.New("selected id conflicts with terminator for different identity"))
	}

	log.Info("terminator already exists")
	return nil
}
