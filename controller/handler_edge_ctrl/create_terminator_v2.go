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
	"ztna-core/ztna/controller/command"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/channel/v3/protobufs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type createTerminatorV2Handler struct {
	baseRequestHandler
}

func NewCreateTerminatorV2Handler(appEnv *env.AppEnv, ch channel.Channel) channel.TypedReceiveHandler {
	logtrace.LogWithFunctionName()
	return &createTerminatorV2Handler{
		baseRequestHandler{
			ch:     ch,
			appEnv: appEnv,
		},
	}
}

func (self *createTerminatorV2Handler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(edge_ctrl_pb.ContentType_CreateTerminatorV2RequestType)
}

func (self *createTerminatorV2Handler) Label() string {
	logtrace.LogWithFunctionName()
	return "create.terminator"
}

func (self *createTerminatorV2Handler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	req := &edge_ctrl_pb.CreateTerminatorV2Request{}
	if err := proto.Unmarshal(msg.Body, req); err != nil {
		pfxlog.ContextLogger(ch.Label()).WithError(err).Error("could not unmarshal CreateTerminatorV2Request")
		return
	}

	ctx := &CreateTerminatorV2RequestContext{
		baseSessionRequestContext: baseSessionRequestContext{handler: self, msg: msg, env: self.appEnv},
		req:                       req,
	}

	go self.CreateTerminatorV2(ctx)
}

func (self *createTerminatorV2Handler) CreateTerminatorV2(ctx *CreateTerminatorV2RequestContext) {
	logtrace.LogWithFunctionName()
	start := time.Now()
	logger := pfxlog.ContextLogger(self.ch.Label()).
		WithField("routerId", self.ch.Id()).
		WithField("terminatorId", ctx.req.Address)

	if !ctx.loadRouter() {
		return
	}
	ctx.verifyTerminatorId(ctx.req.Address)
	ctx.loadSession(ctx.req.SessionToken, ctx.req.ApiSessionToken)
	ctx.checkSessionType(db.SessionTypeBind)
	ctx.checkSessionFingerprints(ctx.req.Fingerprints)
	ctx.verifyIdentityEdgeRouterAccess()
	ctx.loadService()

	if ctx.err != nil {
		errCode := edge_ctrl_pb.CreateTerminatorResult_FailedOther
		if errors.Is(ctx.err, InvalidSessionError{}) {
			errCode = edge_ctrl_pb.CreateTerminatorResult_FailedInvalidSession
		}
		self.returnError(ctx, errCode, ctx.err, logger)
		return
	}

	logger = logger.WithField("serviceId", ctx.service.Id).WithField("service", ctx.service.Name)

	if ctx.req.Cost > math.MaxUint16 {
		ctx.err = invalidCost(fmt.Sprintf("invalid cost %v. cost must be between 0 and %v inclusive", ctx.req.Cost, math.MaxUint16))
		self.returnError(ctx, edge_ctrl_pb.CreateTerminatorResult_FailedOther, ctx.err, logger)
		return
	}

	terminator, _ := self.getNetwork().Terminator.Read(ctx.req.Address)
	if terminator != nil {
		if ctx.err = ctx.validateExistingTerminator(terminator, logger); ctx.err != nil {
			self.returnError(ctx, edge_ctrl_pb.CreateTerminatorResult_FailedIdConflict, ctx.err, logger)
			return
		}

		// if the precedence or cost has changed, update the terminator
		if terminator.Precedence != ctx.req.GetXtPrecedence() || terminator.Cost != uint16(ctx.req.Cost) {
			terminator.Precedence = ctx.req.GetXtPrecedence()
			terminator.Cost = uint16(ctx.req.Cost)
			err := self.appEnv.GetManagers().Terminator.Update(terminator, fields.UpdatedFieldsMap{
				db.FieldTerminatorPrecedence: struct{}{},
				db.FieldTerminatorCost:       struct{}{},
			}, ctx.newChangeContext())

			if err != nil {
				self.returnError(ctx, edge_ctrl_pb.CreateTerminatorResult_FailedOther, err, logger)
				return
			}
		}
	} else {
		terminator = &model.Terminator{
			BaseEntity: models.BaseEntity{
				Id:       ctx.req.Address,
				IsSystem: true,
			},
			Service:        ctx.session.ServiceId,
			Router:         ctx.sourceRouter.Id,
			Binding:        common.EdgeBinding,
			Address:        ctx.req.Address,
			InstanceId:     ctx.req.InstanceId,
			InstanceSecret: ctx.req.InstanceSecret,
			PeerData:       ctx.req.PeerData,
			Precedence:     ctx.req.GetXtPrecedence(),
			Cost:           uint16(ctx.req.Cost),
			HostId:         ctx.session.IdentityId,
			SourceCtrl:     self.appEnv.GetId(),
		}

		cmd := &model.CreateEdgeTerminatorCmd{
			Env:     self.appEnv,
			Entity:  terminator,
			Context: ctx.newChangeContext(),
		}

		createStart := time.Now()
		if err := self.appEnv.GetHostController().GetNetwork().Managers.Dispatcher.Dispatch(cmd); err != nil {
			// terminator might have been created while we were trying to create.
			if terminator, _ = self.getNetwork().Terminator.Read(ctx.req.Address); terminator != nil {
				if validateError := ctx.validateExistingTerminator(terminator, logger); validateError != nil {
					self.returnError(ctx, edge_ctrl_pb.CreateTerminatorResult_FailedIdConflict, validateError, logger)
					return
				}
			} else {
				if command.WasRateLimited(err) {
					self.returnError(ctx, edge_ctrl_pb.CreateTerminatorResult_FailedBusy, err, logger)
					return
				}
				self.returnError(ctx, edge_ctrl_pb.CreateTerminatorResult_FailedOther, err, logger)
				return
			}
		} else {
			logger.WithField("terminator", terminator.Id).
				WithField("createTime", time.Since(createStart)).
				Info("created terminator")
		}
	}

	response := &edge_ctrl_pb.CreateTerminatorV2Response{
		TerminatorId: terminator.Id,
		Result:       edge_ctrl_pb.CreateTerminatorResult_Success,
	}

	body, err := proto.Marshal(response)
	if err != nil {
		logger.WithError(err).Error("failed to marshal CreateTunnelTerminatorResponse")
		return
	}

	responseMsg := channel.NewMessage(response.GetContentType(), body)
	responseMsg.ReplyTo(ctx.msg)
	if err = self.ch.Send(responseMsg); err != nil {
		logger.WithError(err).Error("failed to send CreateTunnelTerminatorResponse")
	}

	logger.WithField("elapsed", time.Since(start)).Info("completed create terminator v2 operation")
}

func (self *createTerminatorV2Handler) returnError(ctx *CreateTerminatorV2RequestContext, resultType edge_ctrl_pb.CreateTerminatorResult, err error, logger *logrus.Entry) {
	logtrace.LogWithFunctionName()
	response := &edge_ctrl_pb.CreateTerminatorV2Response{
		TerminatorId: ctx.req.Address,
		Result:       resultType,
		Msg:          err.Error(),
	}

	if sendErr := protobufs.MarshalTyped(response).ReplyTo(ctx.msg).Send(self.ch); sendErr != nil {
		logger.WithError(err).WithField("sendError", sendErr).Error("failed to send error response")
	} else {
		logger.WithError(err).Error("responded with error")
	}
}

type CreateTerminatorV2RequestContext struct {
	baseSessionRequestContext
	req *edge_ctrl_pb.CreateTerminatorV2Request
}

func (self *CreateTerminatorV2RequestContext) GetSessionToken() string {
	logtrace.LogWithFunctionName()
	return self.req.SessionToken
}

func (self *CreateTerminatorV2RequestContext) validateExistingTerminator(terminator *model.Terminator, log *logrus.Entry) controllerError {
	logtrace.LogWithFunctionName()
	if terminator.Binding != common.EdgeBinding {
		log.WithField("binding", common.EdgeBinding).
			WithField("conflictingBinding", terminator.Binding).
			Error("selected terminator address conflicts with a terminator for a different binding")
		return internalError(errors.New("selected id conflicts with terminator for different binding"))
	}

	if terminator.Service != self.session.ServiceId {
		log.WithField("conflictingService", terminator.Service).
			Error("selected terminator address conflicts with a terminator for a different service")
		return internalError(errors.New("selected id conflicts with terminator for different service"))
	}

	if terminator.Router != self.sourceRouter.Id {
		log.WithField("conflictingRouter", terminator.Router).
			Error("selected terminator address conflicts with a terminator for a different router")
		return internalError(errors.New("selected id conflicts with terminator for different router"))
	}

	if terminator.HostId != self.session.IdentityId {
		log.WithField("identityId", self.session.IdentityId).
			WithField("conflictingIdentity", terminator.HostId).
			Error("selected terminator address conflicts with a terminator for a different identity")
		return internalError(errors.New("selected id conflicts with terminator for different identity"))
	}

	log.Info("terminator already exists")
	return nil
}
