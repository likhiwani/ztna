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
	"time"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/controller/raft"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/channel/v3/latency"
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/openziti/metrics"
	"github.com/sirupsen/logrus"
)

func NewBindHandler(n *network.Network, raftCtrl *raft.Controller, heartbeatOptions *channel.HeartbeatOptions) channel.BindHandler {
	logtrace.LogWithFunctionName()
	bindHandler := func(binding channel.Binding) error {
		binding.AddTypedReceiveHandler(newCommandHandler(raftCtrl))
		binding.AddTypedReceiveHandler(newAddPeerHandler(raftCtrl))
		binding.AddTypedReceiveHandler(newRemovePeerHandler(raftCtrl))
		binding.AddTypedReceiveHandler(newTransferLeadershipHandler(raftCtrl))
		binding.AddTypedReceiveHandler(newInspectHandler(n))

		roundTripHistogram := n.GetMetricsRegistry().Histogram("peer.latency:" + binding.GetChannel().Id())
		queueTimeHistogram := n.GetMetricsRegistry().Histogram("peer.queue_time:" + binding.GetChannel().Id())
		binding.AddCloseHandler(channel.CloseHandlerF(func(ch channel.Channel) {
			roundTripHistogram.Dispose()
			queueTimeHistogram.Dispose()
		}))

		cb := &heartbeatCallback{
			latencyMetric:            roundTripHistogram,
			queueTimeMetric:          queueTimeHistogram,
			ch:                       binding.GetChannel(),
			latencySemaphore:         concurrenz.NewSemaphore(2),
			closeUnresponsiveTimeout: heartbeatOptions.CloseUnresponsiveTimeout,
			lastResponse:             time.Now().Add(heartbeatOptions.CloseUnresponsiveTimeout * 2).UnixMilli(), // wait at least 2x timeout before closing
		}

		channel.ConfigureHeartbeat(binding, heartbeatOptions.SendInterval, heartbeatOptions.CheckInterval, cb)
		return nil
	}

	return channel.BindHandlerF(bindHandler)
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
		log.Error("heartbeat not received in time, closing link")
		if err := self.ch.Close(); err != nil {
			log.WithError(err).Error("error while closing link")
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
