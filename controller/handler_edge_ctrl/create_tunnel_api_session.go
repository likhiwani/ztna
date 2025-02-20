package handler_edge_ctrl

import (
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type createApiSessionHandler struct {
	baseRequestHandler
	*TunnelState
}

func NewCreateApiSessionHandler(appEnv *env.AppEnv, ch channel.Channel, tunnelState *TunnelState) channel.TypedReceiveHandler {
	logtrace.LogWithFunctionName()
	return &createApiSessionHandler{
		baseRequestHandler: baseRequestHandler{ch: ch, appEnv: appEnv},
		TunnelState:        tunnelState,
	}
}

func (self *createApiSessionHandler) getTunnelState() *TunnelState {
	logtrace.LogWithFunctionName()
	return self.TunnelState
}

func (self *createApiSessionHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(edge_ctrl_pb.ContentType_CreateApiSessionRequestType)
}

func (self *createApiSessionHandler) Label() string {
	logtrace.LogWithFunctionName()
	return "tunnel.create.api_session"
}

func (self *createApiSessionHandler) HandleReceive(msg *channel.Message, _ channel.Channel) {
	logtrace.LogWithFunctionName()
	req := &edge_ctrl_pb.CreateApiSessionRequest{}
	if err := proto.Unmarshal(msg.Body, req); err != nil {
		logrus.WithField("routerId", self.ch.Id()).WithError(err).Error("could not unmarshal CreateApiSessionRequest")
		return
	}

	logrus.WithField("routerId", self.ch.Id()).Debug("create api session request received")

	ctx := &createApiSessionRequestContext{
		baseTunnelRequestContext: baseTunnelRequestContext{
			baseSessionRequestContext: baseSessionRequestContext{handler: self, msg: msg, env: self.appEnv},
		},
		req: req,
	}

	go self.createApiSession(ctx)
}

func (self *createApiSessionHandler) createApiSession(ctx *createApiSessionRequestContext) {
	logtrace.LogWithFunctionName()
	if !ctx.loadRouter() {
		return
	}

	ctx.loadIdentity()
	ctx.ensureApiSession(ctx.req.ConfigTypes)
	ctx.updateIdentityInfo(ctx.req.EnvInfo, ctx.req.SdkInfo)

	if ctx.err != nil {
		self.returnError(ctx, ctx.err)
		return
	}

	result, err := ctx.getCreateApiSessionResponse()
	if err != nil {
		self.returnError(ctx, internalError(err))
		return
	}

	body, err := proto.Marshal(result)
	if err != nil {
		self.returnError(ctx, internalError(err))
		return
	}

	responseMsg := channel.NewMessage(int32(edge_ctrl_pb.ContentType_CreateApiSessionResponseType), body)
	responseMsg.ReplyTo(ctx.msg)
	if err = self.ch.Send(responseMsg); err != nil {
		logrus.WithField("routerId", self.ch.Id()).WithError(err).Error("failed to send response")
	} else {
		logrus.WithField("routerId", self.ch.Id()).Debug("create api session response sent")
	}
}

type createApiSessionRequestContext struct {
	baseTunnelRequestContext
	req *edge_ctrl_pb.CreateApiSessionRequest
}
