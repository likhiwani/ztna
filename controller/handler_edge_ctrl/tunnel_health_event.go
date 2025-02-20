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
	"ztna-core/ztna/common"
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
	"github.com/openziti/metrics"
)

type tunnelHealthEventHandler struct {
	baseRequestHandler
	serviceHealthCheckPassedCounter metrics.IntervalCounter
	serviceHealthCheckFailedCounter metrics.IntervalCounter
}

func NewTunnelHealthEventHandler(appEnv *env.AppEnv, ch channel.Channel) channel.TypedReceiveHandler {
	logtrace.LogWithFunctionName()
	serviceEventMetrics := appEnv.GetHostController().GetNetwork().GetServiceEventsMetricsRegistry()
	return &tunnelHealthEventHandler{
		baseRequestHandler: baseRequestHandler{
			ch:     ch,
			appEnv: appEnv,
		},
		serviceHealthCheckPassedCounter: serviceEventMetrics.IntervalCounter("service.health_check.passed", time.Minute),
		serviceHealthCheckFailedCounter: serviceEventMetrics.IntervalCounter("service.health_check.failed", time.Minute),
	}
}

func (self *tunnelHealthEventHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(edge_ctrl_pb.ContentType_TunnelHealthEventType)
}

func (self *tunnelHealthEventHandler) Label() string {
	logtrace.LogWithFunctionName()
	return "tunnel.health.event"
}

func (self *tunnelHealthEventHandler) HandleReceive(msg *channel.Message, _ channel.Channel) {
	logtrace.LogWithFunctionName()
	terminatorId, _ := msg.GetStringHeader(int32(edge_ctrl_pb.Header_TerminatorId))
	checkPassed, _ := msg.GetBoolHeader(int32(edge_ctrl_pb.Header_CheckPassed))

	ctx := &TunnelHealthEventRequestContext{
		baseSessionRequestContext: baseSessionRequestContext{handler: self, msg: msg, env: self.appEnv},
		terminatorId:              terminatorId,
		checkPassed:               checkPassed,
	}

	go self.handleHealthEvent(ctx)
}

func (self *tunnelHealthEventHandler) handleHealthEvent(ctx *TunnelHealthEventRequestContext) {
	logtrace.LogWithFunctionName()
	if !ctx.loadRouter() {
		return
	}

	terminator := ctx.verifyTerminator(ctx.terminatorId, common.TunnelBinding)
	if terminator != nil && ctx.err == nil {
		if ctx.checkPassed {
			self.serviceHealthCheckPassedCounter.Update(terminator.Service, time.Now(), 1)
		} else {
			self.serviceHealthCheckFailedCounter.Update(terminator.Service, time.Now(), 1)
		}
	}

	self.logResult(ctx, ctx.err)
}

type TunnelHealthEventRequestContext struct {
	baseSessionRequestContext
	terminatorId string
	checkPassed  bool
}
