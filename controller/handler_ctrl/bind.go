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
	"time"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/logtrace"

	"github.com/sirupsen/logrus"

	"ztna-core/ztna/common/trace"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/controller/xctrl"
	metrics2 "ztna-core/ztna/router/metrics"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/channel/v3/latency"
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/openziti/metrics"
)

type bindHandler struct {
	heartbeatOptions *channel.HeartbeatOptions
	router           *model.Router
	network          *network.Network
	xctrls           []xctrl.Xctrl
}

func newBindHandler(heartbeatOptions *channel.HeartbeatOptions, router *model.Router, network *network.Network, xctrls []xctrl.Xctrl) channel.BindHandler {
	logtrace.LogWithFunctionName()
	return &bindHandler{
		heartbeatOptions: heartbeatOptions,
		router:           router,
		network:          network,
		xctrls:           xctrls,
	}
}

func (self *bindHandler) BindChannel(binding channel.Binding) error {
	logtrace.LogWithFunctionName()
	log := pfxlog.Logger().WithFields(map[string]interface{}{
		"routerId":      self.router.Id,
		"routerVersion": self.router.VersionInfo.Version,
	})
	log.Debug("binding router channel")

	binding.AddTypedReceiveHandler(newCircuitRequestHandler(self.router, self.network))
	binding.AddTypedReceiveHandler(newRouteResultHandler(self.network, self.router))
	binding.AddTypedReceiveHandler(newCircuitConfirmationHandler(self.network, self.router))
	binding.AddTypedReceiveHandler(newCreateTerminatorHandler(self.network, self.router))
	binding.AddTypedReceiveHandler(newRemoveTerminatorHandler(self.network, self.router))
	binding.AddTypedReceiveHandler(newRemoveTerminatorsHandler(self.network, self.router))
	binding.AddTypedReceiveHandler(newUpdateTerminatorHandler(self.network, self.router))
	binding.AddTypedReceiveHandler(newLinkConnectedHandler(self.router, self.network))
	binding.AddTypedReceiveHandler(newRouterLinkHandler(self.router, self.network))
	binding.AddTypedReceiveHandler(newVerifyRouterHandler(self.router, self.network))
	binding.AddTypedReceiveHandler(newFaultHandler(self.router, self.network))
	binding.AddTypedReceiveHandler(newMetricsHandler(self.network))
	binding.AddTypedReceiveHandler(newTraceHandler(self.network.GetTraceController()))
	binding.AddTypedReceiveHandler(newInspectHandler(self.network))
	binding.AddTypedReceiveHandler(newQuiesceRouterHandler(self.router, self.network))
	binding.AddTypedReceiveHandler(newDequiesceRouterHandler(self.router, self.network))
	binding.AddTypedReceiveHandler(newDecommissionRouterHandler(self.router, self.network))
	binding.AddTypedReceiveHandler(newPingHandler())
	binding.AddTypedReceiveHandler(&channel.AsyncFunctionReceiveAdapter{
		Type:    int32(ctrl_pb.ContentType_ValidateTerminatorsV2ResponseType),
		Handler: self.network.RouterMessaging.NewValidationResponseHandler(self.network, self.router),
	})
	binding.AddPeekHandler(trace.NewChannelPeekHandler(self.network.GetAppId(), binding.GetChannel(), self.network.GetTraceController()))
	binding.AddPeekHandler(metrics2.NewCtrlChannelPeekHandler(self.router.Id, self.network.GetMetricsRegistry()))

	roundTripHistogram := self.network.GetMetricsRegistry().Histogram("ctrl.latency:" + self.router.Id)
	queueTimeHistogram := self.network.GetMetricsRegistry().Histogram("ctrl.queue_time:" + self.router.Id)
	binding.AddCloseHandler(channel.CloseHandlerF(func(ch channel.Channel) {
		roundTripHistogram.Dispose()
		queueTimeHistogram.Dispose()
	}))

	cb := &heartbeatCallback{
		latencyMetric:            roundTripHistogram,
		queueTimeMetric:          queueTimeHistogram,
		ch:                       binding.GetChannel(),
		latencySemaphore:         concurrenz.NewSemaphore(2),
		closeUnresponsiveTimeout: self.heartbeatOptions.CloseUnresponsiveTimeout,
		lastResponse:             time.Now().Add(self.heartbeatOptions.CloseUnresponsiveTimeout * 2).UnixMilli(), // wait at least 2x timeout before closing
	}
	channel.ConfigureHeartbeat(binding, self.heartbeatOptions.SendInterval, self.heartbeatOptions.CheckInterval, cb)

	xctrlDone := make(chan struct{})
	for _, x := range self.xctrls {
		if err := binding.Bind(x); err != nil {
			return err
		}
		if err := x.Run(binding.GetChannel(), self.network.GetDb(), xctrlDone); err != nil {
			return err
		}
	}
	if len(self.xctrls) > 0 {
		binding.AddCloseHandler(newXctrlCloseHandler(xctrlDone))
	}

	binding.AddCloseHandler(newCloseHandler(self.router, self.network))
	return nil
}

type heartbeatCallback struct {
	latencyMetric            metrics.Histogram
	queueTimeMetric          metrics.Histogram
	lastResponse             int64
	ch                       channel.Channel
	latencySemaphore         concurrenz.Semaphore
	closeUnresponsiveTimeout time.Duration
}

func (self *heartbeatCallback) HeartbeatTx(int64) {
	logtrace.LogWithFunctionName()
}

func (self *heartbeatCallback) HeartbeatRx(int64) {
	logtrace.LogWithFunctionName()
}

func (self *heartbeatCallback) HeartbeatRespTx(int64) {
	logtrace.LogWithFunctionName()
}

func (self *heartbeatCallback) HeartbeatRespRx(ts int64) {
	logtrace.LogWithFunctionName()
	now := time.Now()
	self.lastResponse = now.UnixMilli()
	self.latencyMetric.Update(now.UnixNano() - ts)
}

func (self *heartbeatCallback) timeSinceLastResponse(nowUnixMillis int64) time.Duration {
	logtrace.LogWithFunctionName()
	return time.Duration(nowUnixMillis-self.lastResponse) * time.Millisecond
}

func (self *heartbeatCallback) CheckHeartBeat() {
	logtrace.LogWithFunctionName()
	now := time.Now().UnixMilli()
	if self.timeSinceLastResponse(now) > self.closeUnresponsiveTimeout {
		log := self.logger()
		log.Error("heartbeat not received in time, closing control channel connection")
		if err := self.ch.Close(); err != nil {
			log.WithError(err).Error("error while closing control channel connection")
		}
	}
	go self.checkQueueTime()
}

func (self *heartbeatCallback) checkQueueTime() {
	logtrace.LogWithFunctionName()
	if !self.latencySemaphore.TryAcquire() {
		self.logger().Warn("unable to check queue time, too many check already running")
		return
	}

	defer self.latencySemaphore.Release()

	sendTracker := &latency.SendTimeTracker{
		Handler: func(latencyType latency.Type, latency time.Duration) {
			self.queueTimeMetric.Update(latency.Nanoseconds())
		},
		StartTime: time.Now(),
	}
	if err := self.ch.Send(sendTracker); err != nil && !self.ch.IsClosed() {
		self.logger().WithError(err).Error("unable to send queue time tracer")
	}
}

func (self *heartbeatCallback) logger() *logrus.Entry {
	logtrace.LogWithFunctionName()
	return pfxlog.Logger().WithField("channelType", "router").WithField("channelId", self.ch.Id())
}
