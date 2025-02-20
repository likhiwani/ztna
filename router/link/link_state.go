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

package link

import (
	"container/heap"
	"math/rand"
	"sync/atomic"
	"time"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/xlink"

	"github.com/michaelquigley/pfxlog"
)

const (
	StatusPending     linkStatus = "pending"
	StatusDialing     linkStatus = "dialing"
	StatusQueueFailed linkStatus = "queueFailed"
	StatusDialFailed  linkStatus = "dialFailed"
	StatusLinkFailed  linkStatus = "linkFailed"
	StatusDestRemoved linkStatus = "destRemoved"
	StatusEstablished linkStatus = "established"
)

type linkStatus string

func (self linkStatus) String() string {
	logtrace.LogWithFunctionName()
	return string(self)
}

func newLinkDest(destId string) *linkDest {
	logtrace.LogWithFunctionName()
	return &linkDest{
		id:          destId,
		healthy:     true,
		unhealthyAt: time.Time{},
		linkMap:     map[string]*linkState{},
	}
}

type linkDest struct {
	id          string
	version     string
	healthy     bool
	unhealthyAt time.Time
	linkMap     map[string]*linkState
}

func (self *linkDest) update(update *linkDestUpdate) {
	logtrace.LogWithFunctionName()
	if self.healthy && !update.healthy {
		self.unhealthyAt = time.Now()
	}

	self.healthy = update.healthy

	if update.healthy {
		self.version = update.version
	}
}

type linkFault struct {
	linkId    string
	iteration uint32
}

type linkState struct {
	linkKey        string
	linkId         string
	status         linkStatus
	dialAttempts   atomic.Uint64
	connectedCount uint64
	retryDelay     time.Duration
	nextDial       time.Time
	dest           *linkDest
	listener       *ctrl_pb.Listener
	dialer         xlink.Dialer
	allowedDials   int64
	ctrlsNotified  bool
	linkFaults     []linkFault
	dialActive     atomic.Bool
	link           xlink.Xlink
}

func (self *linkState) updateStatus(status linkStatus) {
	logtrace.LogWithFunctionName()
	if self.status != status {
		log := pfxlog.Logger().
			WithField("key", self.linkKey).
			WithField("oldState", self.status).
			WithField("newState", status).
			WithField("linkId", self.linkId).
			WithField("iteration", self.dialAttempts.Load())
		self.status = status
		log.Info("status updated")
		if self.status != StatusEstablished {
			self.link = nil
		}
	}
}

func (self *linkState) GetLinkKey() string {
	logtrace.LogWithFunctionName()
	return self.linkKey
}

func (self *linkState) GetLinkId() string {
	logtrace.LogWithFunctionName()
	return self.linkId
}

func (self *linkState) GetRouterId() string {
	logtrace.LogWithFunctionName()
	return self.dest.id
}

func (self *linkState) GetAddress() string {
	logtrace.LogWithFunctionName()
	return self.listener.Address
}

func (self *linkState) GetLinkProtocol() string {
	logtrace.LogWithFunctionName()
	return self.listener.Protocol
}

func (self *linkState) GetRouterVersion() string {
	logtrace.LogWithFunctionName()
	return self.dest.version
}

func (self *linkState) GetIteration() uint32 {
	logtrace.LogWithFunctionName()
	return uint32(self.dialAttempts.Load())
}

func (self *linkState) addPendingLinkFault(linkId string, iteration uint32) {
	logtrace.LogWithFunctionName()
	for _, fault := range self.linkFaults {
		if fault.linkId == linkId {
			if fault.iteration < iteration {
				fault.iteration = iteration
			}
			return
		}
	}
	self.linkFaults = append(self.linkFaults, linkFault{
		linkId:    linkId,
		iteration: iteration,
	})
}

func (self *linkState) clearFaultsForLinkId(linkId string) {
	logtrace.LogWithFunctionName()
	faults := self.linkFaults
	self.linkFaults = nil

	for _, fault := range faults {
		if fault.linkId != linkId {
			self.linkFaults = append(self.linkFaults, fault)
		}
	}
}

func (self *linkState) clearFault(toClear linkFault) {
	logtrace.LogWithFunctionName()
	faults := self.linkFaults
	self.linkFaults = nil

	for _, fault := range faults {
		if fault.linkId != toClear.linkId || fault.iteration > toClear.iteration {
			self.linkFaults = append(self.linkFaults, fault)
		}
	}
}

func (self *linkState) dialFailed(registry *linkRegistryImpl) {
	logtrace.LogWithFunctionName()
	if self.allowedDials > 0 {
		self.allowedDials--
	}

	if self.allowedDials == 0 {
		delete(self.dest.linkMap, self.linkKey)
		return
	}

	backoffConfig := self.dialer.GetHealthyBackoffConfig()
	if !self.dest.healthy {
		backoffConfig = self.dialer.GetUnhealthyBackoffConfig()
	}

	factor := backoffConfig.GetRetryBackoffFactor() + (rand.Float64() - 0.5)
	if factor < 1 {
		factor = 1
	}

	self.retryDelay = time.Duration(float64(self.retryDelay) * factor)
	if self.retryDelay < backoffConfig.GetMinRetryInterval() {
		self.retryDelay = backoffConfig.GetMinRetryInterval()
	}

	if self.retryDelay > backoffConfig.GetMaxRetryInterval() {
		self.retryDelay = backoffConfig.GetMaxRetryInterval()
	}

	self.nextDial = time.Now().Add(self.retryDelay)

	heap.Push(registry.linkStateQueue, self)
}
