/*
	(c) Copyright NetFoundry Inc.

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

package xlink_transport

import (
	"sync/atomic"
	"time"
	"ztna-core/ztna/common/inspect"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/xgress"

	"github.com/openziti/channel/v3"
	"github.com/openziti/metrics"
)

type impl struct {
	id            string
	key           string
	ch            channel.Channel
	routerId      string
	routerVersion string
	linkProtocol  string
	dialAddress   string
	closed        atomic.Bool
	faultsSent    atomic.Bool
	dialed        bool
	iteration     uint32
	dupsRejected  uint32

	droppedMsgMeter    metrics.Meter
	droppedXgMsgMeter  metrics.Meter
	droppedRtxMsgMeter metrics.Meter
	droppedFwdMsgMeter metrics.Meter
}

func (self *impl) Id() string {
	logtrace.LogWithFunctionName()
	return self.id
}

func (self *impl) Key() string {
	logtrace.LogWithFunctionName()
	return self.key
}

func (self *impl) Iteration() uint32 {
	logtrace.LogWithFunctionName()
	return self.iteration
}

func (self *impl) Init(metricsRegistry metrics.Registry) error {
	logtrace.LogWithFunctionName()
	if self.droppedMsgMeter == nil {
		self.droppedMsgMeter = metricsRegistry.Meter("link.dropped_msgs:" + self.id)
		self.droppedXgMsgMeter = metricsRegistry.Meter("link.dropped_xg_msgs:" + self.id)
		self.droppedRtxMsgMeter = metricsRegistry.Meter("link.dropped_rtx_msgs:" + self.id)
		self.droppedFwdMsgMeter = metricsRegistry.Meter("link.dropped_fwd_msgs:" + self.id)
	}
	return nil
}

func (self *impl) SendPayload(msg *xgress.Payload, timeout time.Duration, payloadType xgress.PayloadType) error {
	logtrace.LogWithFunctionName()
	if timeout == 0 {
		sent, err := self.ch.TrySend(msg.Marshall())
		if err == nil && !sent {
			self.droppedMsgMeter.Mark(1)
			if payloadType == xgress.PayloadTypeXg {
				self.droppedXgMsgMeter.Mark(1)
			} else if payloadType == xgress.PayloadTypeRtx {
				self.droppedRtxMsgMeter.Mark(1)
			} else if payloadType == xgress.PayloadTypeFwd {
				self.droppedFwdMsgMeter.Mark(1)
			}
		}
		return err
	}

	return msg.Marshall().WithTimeout(timeout).Send(self.ch)
}

func (self *impl) SendAcknowledgement(msg *xgress.Acknowledgement) error {
	logtrace.LogWithFunctionName()
	sent, err := self.ch.TrySend(msg.Marshall())
	if err == nil && !sent {
		self.droppedMsgMeter.Mark(1)
	}
	return err
}

func (self *impl) SendControl(msg *xgress.Control) error {
	logtrace.LogWithFunctionName()
	sent, err := self.ch.TrySend(msg.Marshall())
	if err == nil && !sent {
		self.droppedMsgMeter.Mark(1)
	}
	return err
}

func (self *impl) Close() error {
	logtrace.LogWithFunctionName()
	self.droppedMsgMeter.Dispose()
	return self.ch.Close()
}

func (self *impl) CloseNotified() error {
	logtrace.LogWithFunctionName()
	self.faultsSent.Store(true)
	return self.Close()
}

func (self *impl) AreFaultsSent() bool {
	logtrace.LogWithFunctionName()
	return self.faultsSent.Load()
}

func (self *impl) DestinationId() string {
	logtrace.LogWithFunctionName()
	return self.routerId
}

func (self *impl) DestVersion() string {
	logtrace.LogWithFunctionName()
	return self.routerVersion
}

func (self *impl) LinkProtocol() string {
	logtrace.LogWithFunctionName()
	return self.linkProtocol
}

func (self *impl) DialAddress() string {
	logtrace.LogWithFunctionName()
	return self.dialAddress
}

func (self *impl) IsDialed() bool {
	logtrace.LogWithFunctionName()
	return self.dialed
}

func (self *impl) CloseOnce(f func()) {
	logtrace.LogWithFunctionName()
	if self.closed.CompareAndSwap(false, true) {
		f()
	}
}

func (self *impl) IsClosed() bool {
	logtrace.LogWithFunctionName()
	return self.ch.IsClosed()
}

func (self *impl) InspectCircuit(detail *inspect.CircuitInspectDetail) {
	logtrace.LogWithFunctionName()
	detail.LinkDetails[self.id] = self.InspectLink()
}

func (self *impl) InspectLink() *inspect.LinkInspectDetail {
	logtrace.LogWithFunctionName()
	return &inspect.LinkInspectDetail{
		Id:          self.Id(),
		Iteration:   self.Iteration(),
		Key:         self.key,
		Split:       false,
		Protocol:    self.LinkProtocol(),
		DialAddress: self.DialAddress(),
		Dest:        self.DestinationId(),
		DestVersion: self.DestVersion(),
		Dialed:      self.dialed,
	}
}

func (self *impl) GetAddresses() []*ctrl_pb.LinkConn {
	logtrace.LogWithFunctionName()
	localAddr := self.ch.Underlay().GetLocalAddr()
	remoteAddr := self.ch.Underlay().GetRemoteAddr()
	return []*ctrl_pb.LinkConn{
		{
			Id:         "single",
			LocalAddr:  localAddr.Network() + ":" + localAddr.String(),
			RemoteAddr: remoteAddr.Network() + ":" + remoteAddr.String(),
		},
	}
}

func (self *impl) DuplicatesRejected() uint32 {
	logtrace.LogWithFunctionName()
	return atomic.AddUint32(&self.dupsRejected, 1)
}
