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
	"time"
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/metrics"
	"google.golang.org/protobuf/proto"
)

type healthEventHandler struct {
	baseRequestHandler
	serviceHealthCheckPassedCounter metrics.IntervalCounter
	serviceHealthCheckFailedCounter metrics.IntervalCounter
}

func NewHealthEventHandler(appEnv *env.AppEnv, ch channel.Channel) channel.TypedReceiveHandler {
	logtrace.LogWithFunctionName()
	serviceEventMetrics := appEnv.GetHostController().GetNetwork().GetServiceEventsMetricsRegistry()
	return &healthEventHandler{
		baseRequestHandler: baseRequestHandler{
			ch:     ch,
			appEnv: appEnv,
		},
		serviceHealthCheckPassedCounter: serviceEventMetrics.IntervalCounter("service.health_check.passed", time.Minute),
		serviceHealthCheckFailedCounter: serviceEventMetrics.IntervalCounter("service.health_check.failed", time.Minute),
	}
}

func (self *healthEventHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(edge_ctrl_pb.ContentType_HealthEventType)
}

func (self *healthEventHandler) Label() string {
	logtrace.LogWithFunctionName()
	return "health.event"
}

func (self *healthEventHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	req := &edge_ctrl_pb.HealthEventRequest{}
	if err := proto.Unmarshal(msg.Body, req); err != nil {
		pfxlog.ContextLogger(ch.Label()).WithError(err).Error("could not unmarshal health event")
		return
	}

	ctx := &HealthEventRequestContext{
		baseSessionRequestContext: baseSessionRequestContext{handler: self, msg: msg, env: self.appEnv},
		req:                       req,
	}

	go self.handleHealthEvent(ctx)
}

func (self *healthEventHandler) handleHealthEvent(ctx *HealthEventRequestContext) {
	logtrace.LogWithFunctionName()
	if !ctx.loadRouter() {
		return
	}

	ctx.loadSession(ctx.req.SessionToken, ctx.req.ApiSessionToken)
	ctx.checkSessionType(db.SessionTypeBind)
	ctx.checkSessionFingerprints(ctx.req.Fingerprints)

	if ctx.err == nil {
		if ctx.req.CheckPassed {
			self.serviceHealthCheckPassedCounter.Update(ctx.session.ServiceId, time.Now(), 1)
		} else {
			self.serviceHealthCheckFailedCounter.Update(ctx.session.ServiceId, time.Now(), 1)
		}
	}

	self.logResult(ctx, ctx.err)
}

type HealthEventRequestContext struct {
	baseSessionRequestContext
	req *edge_ctrl_pb.HealthEventRequest
}

func (self *HealthEventRequestContext) GetSessionToken() string {
	logtrace.LogWithFunctionName()
	return self.req.SessionToken
}
