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

package env

import (
	"sync/atomic"
	"time"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/foundation/v2/versions"

	"github.com/openziti/channel/v3"
)

type NetworkController interface {
	Channel() channel.Channel
	Address() string
	Latency() time.Duration
	HeartbeatCallback() channel.HeartbeatCallback
	IsUnresponsive() bool
	isMoreResponsive(other NetworkController) bool
	GetVersion() *versions.VersionInfo
	TimeSinceLastContact() time.Duration
	IsConnected() bool
	GetLastReportedDataModelIndex() uint64
	updateDataModelIndex(index uint64)
}

func newNetworkCtrl(ch channel.Channel, address string, heartbeatOptions *HeartbeatOptions) *networkCtrl {
	logtrace.LogWithFunctionName()
	result := &networkCtrl{
		ch:               ch,
		address:          address,
		heartbeatOptions: heartbeatOptions,
	}
	result.lastContact.Store(time.Now().UnixMilli())
	return result
}

type networkCtrl struct {
	ch               channel.Channel
	address          string
	heartbeatOptions *HeartbeatOptions
	lastTx           int64
	lastRx           int64
	latency          atomic.Int64
	unresponsive     atomic.Bool
	versionInfo      *versions.VersionInfo
	lastContact      atomic.Int64
	currentIndex     atomic.Uint64
}

func (self *networkCtrl) TimeSinceLastContact() time.Duration {
	logtrace.LogWithFunctionName()
	return time.Millisecond * time.Duration(time.Now().UnixMilli()-self.lastContact.Load())
}

func (self *networkCtrl) HeartbeatCallback() channel.HeartbeatCallback {
	logtrace.LogWithFunctionName()
	return self
}

func (self *networkCtrl) Channel() channel.Channel {
	logtrace.LogWithFunctionName()
	return self.ch
}

func (self *networkCtrl) GetVersion() *versions.VersionInfo {
	logtrace.LogWithFunctionName()
	return self.versionInfo
}

func (self *networkCtrl) Address() string {
	logtrace.LogWithFunctionName()
	return self.address
}

func (self *networkCtrl) GetLastReportedDataModelIndex() uint64 {
	logtrace.LogWithFunctionName()
	return self.currentIndex.Load()
}

func (self *networkCtrl) updateDataModelIndex(index uint64) {
	logtrace.LogWithFunctionName()
	self.currentIndex.Store(index)
}

func (self *networkCtrl) Latency() time.Duration {
	logtrace.LogWithFunctionName()
	return time.Duration(self.latency.Load())
}

func (self *networkCtrl) IsUnresponsive() bool {
	logtrace.LogWithFunctionName()
	return self.unresponsive.Load()
}

func (self *networkCtrl) isMoreResponsive(other NetworkController) bool {
	logtrace.LogWithFunctionName()
	if self.IsConnected() && !other.IsConnected() {
		return true
	} else if other.IsConnected() && !self.IsConnected() {
		return false
	} else if self.IsUnresponsive() {
		if !other.IsUnresponsive() {
			return false
		}
	} else if other.IsUnresponsive() {
		return true
	}
	return self.Latency() < other.Latency()
}

func (self *networkCtrl) HeartbeatTx(int64) {
	logtrace.LogWithFunctionName()
	self.lastTx = time.Now().UnixMilli()
	self.lastContact.Store(self.lastTx)
}

func (self *networkCtrl) HeartbeatRx(int64) {
	logtrace.LogWithFunctionName()
}

func (self *networkCtrl) HeartbeatRespTx(int64) {
	logtrace.LogWithFunctionName()
}

func (self *networkCtrl) HeartbeatRespRx(ts int64) {
	logtrace.LogWithFunctionName()
	now := time.Now()
	self.lastRx = now.UnixMilli()
	self.latency.Store(now.UnixNano() - ts)
	self.lastContact.Store(self.lastRx)
}

func (self *networkCtrl) CheckHeartBeat() {
	logtrace.LogWithFunctionName()
	if time.Duration(self.latency.Load()) > self.heartbeatOptions.UnresponsiveAfter {
		// if latency is greater than 5 seconds, consider this channel unresponsive
		self.unresponsive.Store(true)
	} else if self.lastTx > 0 && self.lastRx < self.lastTx && (time.Now().UnixMilli()-self.lastTx) > 5000 {
		// if we've sent a heartbeat and not gotten a response in over 5s, consider ourselves unresponsive
		self.unresponsive.Store(true)
	} else if !self.IsConnected() {
		self.unresponsive.Store(true)
	} else {
		self.unresponsive.Store(false)
	}
}

func (self *networkCtrl) IsConnected() bool {
	logtrace.LogWithFunctionName()
	connectable, ok := self.ch.Underlay().(interface{ IsConnected() bool })
	return ok && connectable.IsConnected()
}

func NewDefaultHeartbeatOptions() *HeartbeatOptions {
	logtrace.LogWithFunctionName()
	return &HeartbeatOptions{
		HeartbeatOptions:  *channel.DefaultHeartbeatOptions(),
		UnresponsiveAfter: 5 * time.Second,
	}
}

func NewHeartbeatOptions(options *channel.HeartbeatOptions) (*HeartbeatOptions, error) {
	logtrace.LogWithFunctionName()
	unresponsiveAfter, err := options.GetDuration("unresponsiveAfter")
	if err != nil {
		return nil, err
	}
	result := NewDefaultHeartbeatOptions()
	result.HeartbeatOptions = *options
	if unresponsiveAfter != nil {
		result.UnresponsiveAfter = *unresponsiveAfter
	}
	return result, nil
}

type HeartbeatOptions struct {
	channel.HeartbeatOptions
	UnresponsiveAfter time.Duration
}
