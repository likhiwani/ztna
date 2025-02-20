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
	"net"
	"syscall"
	"time"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/env"

	"ztna-core/ztna/common/ctrl_msg"
	"ztna-core/ztna/common/logcontext"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/controller/xt"
	"ztna-core/ztna/router/forwarder"
	"ztna-core/ztna/router/handler_xgress"
	"ztna-core/ztna/router/xgress"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/foundation/v2/goroutines"
	"github.com/openziti/identity"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type routeHandler struct {
	id        *identity.TokenId
	ch        channel.Channel
	env       env.RouterEnv
	dialerCfg map[string]xgress.OptionsData
	forwarder *forwarder.Forwarder
	pool      goroutines.Pool
}

func newRouteHandler(ch channel.Channel, env env.RouterEnv, forwarder *forwarder.Forwarder, pool goroutines.Pool) *routeHandler {
	logtrace.LogWithFunctionName()
	handler := &routeHandler{
		id:        env.GetRouterId(),
		ch:        ch,
		env:       env,
		forwarder: forwarder,
		pool:      pool,
		dialerCfg: env.GetDialerCfg(),
	}

	return handler
}

func (rh *routeHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(ctrl_pb.ContentType_RouteType)
}

func (rh *routeHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	route := &ctrl_pb.Route{}

	if err := proto.Unmarshal(msg.Body, route); err != nil {
		pfxlog.ContextLogger(ch.Label()).WithError(err).Error("error unmarshaling")
		return
	}

	var ctx logcontext.Context
	if route.Context != nil {
		ctx = logcontext.NewContextWith(route.Context.ChannelMask, route.Context.Fields)
	} else {
		ctx = logcontext.NewContext()
	}

	log := pfxlog.ChannelLogger(logcontext.EstablishPath).Wire(ctx).
		WithField("context", ch.Label()).
		WithField("circuitId", route.CircuitId).
		WithField("attempt", route.Attempt)

	if route.Egress != nil {
		log = log.WithField("binding", route.Egress.Binding).WithField("destination", route.Egress.Destination)
	}

	log.Debugf("attempt [#%d] for [s/%s]", route.Attempt, route.CircuitId)

	workF := func() {
		if route.Egress != nil {
			if rh.forwarder.HasDestination(xgress.Address(route.Egress.Address)) {
				log.Warnf("destination exists for [%s]", route.Egress.Address)
				rh.completeRoute(msg, int(route.Attempt), route, nil, log)
				return
			} else {
				rh.connectEgress(msg, int(route.Attempt), ch, route, ctx, time.Now().Add(time.Duration(route.Timeout)))
				return
			}
		} else {
			rh.completeRoute(msg, int(route.Attempt), route, nil, log)
		}
	}

	// if the queue is full, don't wait, we can't hold up the control channel processing
	if err := rh.pool.QueueOrError(workF); err != nil {
		log.WithError(err).Error("error queuing route processing to pool")
		// don't send failure back. we can't delegate to another goroutine and if we sent from
		// here we could block processing of incoming messages
	}
}

func (rh *routeHandler) completeRoute(msg *channel.Message, attempt int, route *ctrl_pb.Route, peerData xt.PeerData, log *logrus.Entry) {
	logtrace.LogWithFunctionName()
	if err := rh.forwarder.Route(rh.ch.Id(), route); err != nil {
		rh.fail(msg, attempt, route, err, ctrl_msg.ErrorTypeGeneric, log)
		return
	}

	log.Debug("forwarder updated with route")

	response := ctrl_msg.NewRouteResultSuccessMsg(route.CircuitId, attempt)
	for k, v := range peerData {
		response.Headers[int32(k)] = v
	}

	response.ReplyTo(msg)

	log.Debug("sending success response")
	if err := response.WithTimeout(rh.env.GetNetworkControllers().DefaultRequestTimeout()).Send(rh.ch); err == nil {
		log.Debug("handled route")
	} else {
		log.WithError(err).Error("send response failed")
	}
}

func (rh *routeHandler) fail(msg *channel.Message, attempt int, route *ctrl_pb.Route, err error, errorHeader byte, log *logrus.Entry) {
	logtrace.LogWithFunctionName()
	log.WithError(err).Error("failure while handling route update")

	response := ctrl_msg.NewRouteResultFailedMessage(route.CircuitId, attempt, err.Error())
	response.PutByteHeader(ctrl_msg.RouteResultErrorCodeHeader, errorHeader)

	response.ReplyTo(msg)
	if err = response.WithTimeout(rh.env.GetNetworkControllers().DefaultRequestTimeout()).Send(rh.ch); err != nil {
		log.WithError(err).Error("send failure response failed")
	}
}

func (rh *routeHandler) connectEgress(msg *channel.Message, attempt int, ch channel.Channel, route *ctrl_pb.Route, ctx logcontext.Context, deadline time.Time) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ChannelLogger(logcontext.EstablishPath).Wire(ctx).
		WithField("context", ch.Label()).
		WithField("circuitId", route.CircuitId).
		WithField("binding", route.Egress.Binding).
		WithField("destination", route.Egress.Destination).
		WithField("attempt", route.Attempt)

	log.Debug("route request received")

	if factory, err := xgress.GlobalRegistry().Factory(route.Egress.Binding); err == nil {
		if dialer, err := factory.CreateDialer(rh.dialerCfg[route.Egress.Binding]); err == nil {
			bindHandler := handler_xgress.NewBindHandler(
				handler_xgress.NewReceiveHandler(rh.forwarder),
				handler_xgress.NewCloseHandler(rh.env.GetNetworkControllers(), rh.forwarder),
				rh.forwarder)

			if rh.forwarder.Options.XgressDialDwellTime > 0 {
				log.Infof("dwelling [%s] on dial", rh.forwarder.Options.XgressDialDwellTime)
				time.Sleep(rh.forwarder.Options.XgressDialDwellTime)
			}

			params := newDialParams(rh.ch.Id(), route, bindHandler, ctx, deadline)
			if peerData, err := dialer.Dial(params); err == nil {
				rh.completeRoute(msg, attempt, route, peerData, log)
			} else {
				var errCode byte

				switch {
				case errors.Is(err, syscall.ECONNREFUSED):
					errCode = ctrl_msg.ErrorTypeConnectionRefused
				case isNetworkTimeout(err) || errors.Is(err, syscall.ETIMEDOUT):
					errCode = ctrl_msg.ErrorTypeDialTimedOut
				case errors.As(err, &xgress.MisconfiguredTerminatorError{}):
					errCode = ctrl_msg.ErrorTypeMisconfiguredTerminator
				case errors.As(err, &xgress.InvalidTerminatorError{}):
					errCode = ctrl_msg.ErrorTypeInvalidTerminator
				default:
					errCode = ctrl_msg.ErrorTypeGeneric
				}

				rh.fail(msg, attempt, route, errors.Wrapf(err, "error creating route for [c/%s]", route.CircuitId), errCode, log)
			}
		} else {
			var errCode byte = ctrl_msg.ErrorTypeMisconfiguredTerminator
			rh.fail(msg, attempt, route, errors.Wrapf(err, "unable to create dialer for [c/%s]", route.CircuitId), errCode, log)
		}
	} else {
		var errCode byte = ctrl_msg.ErrorTypeMisconfiguredTerminator
		rh.fail(msg, attempt, route, errors.Wrapf(err, "error creating route for [c/%s]", route.CircuitId), errCode, log)
	}
}

func isNetworkTimeout(err error) bool {
	logtrace.LogWithFunctionName()
	var netErr net.Error
	return errors.As(err, &netErr)
}

func newDialParams(ctrlId string, route *ctrl_pb.Route, bindHandler xgress.BindHandler, logContext logcontext.Context, deadline time.Time) *dialParams {
	logtrace.LogWithFunctionName()
	return &dialParams{
		ctrlId:      ctrlId,
		Route:       route,
		circuitId:   &identity.TokenId{Token: route.CircuitId, Data: route.Egress.PeerData},
		bindHandler: bindHandler,
		logContext:  logContext,
		deadline:    deadline,
	}
}

type dialParams struct {
	*ctrl_pb.Route
	ctrlId      string
	circuitId   *identity.TokenId
	bindHandler xgress.BindHandler
	logContext  logcontext.Context
	deadline    time.Time
}

func (self *dialParams) GetCtrlId() string {
	logtrace.LogWithFunctionName()
	return self.ctrlId
}

func (self *dialParams) GetDestination() string {
	logtrace.LogWithFunctionName()
	return self.Egress.Destination
}

func (self *dialParams) GetCircuitId() *identity.TokenId {
	logtrace.LogWithFunctionName()
	return self.circuitId
}

func (self *dialParams) GetAddress() xgress.Address {
	logtrace.LogWithFunctionName()
	return xgress.Address(self.Egress.Address)
}

func (self *dialParams) GetBindHandler() xgress.BindHandler {
	logtrace.LogWithFunctionName()
	return self.bindHandler
}

func (self *dialParams) GetLogContext() logcontext.Context {
	logtrace.LogWithFunctionName()
	return self.logContext
}

func (self *dialParams) GetDeadline() time.Time {
	logtrace.LogWithFunctionName()
	return self.deadline
}

func (self *dialParams) GetCircuitTags() map[string]string {
	logtrace.LogWithFunctionName()
	return self.Tags
}
