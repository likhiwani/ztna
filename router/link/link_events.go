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
	"sync/atomic"
	"time"
	"ztna-core/ztna/common/inspect"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/controller/idgen"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/xlink"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/foundation/v2/stringz"
	"github.com/pkg/errors"
)

const (
	GroupDefault = "default"
)

type event interface {
	Handle(registry *linkRegistryImpl)
}

type removeLinkDest struct {
	id string
}

func (self *removeLinkDest) Handle(registry *linkRegistryImpl) {
	logtrace.LogWithFunctionName()
	dest := registry.destinations[self.id]
	delete(registry.destinations, self.id)
	if dest != nil {
		for _, state := range dest.linkMap {
			state.updateStatus(StatusDestRemoved)
			if link, _ := registry.GetLink(state.linkId); link != nil {
				if err := link.Close(); err != nil {
					pfxlog.Logger().
						WithField("linkKey", state.linkKey).
						WithField("linkId", link.Id()).
						WithError(err).
						Error("error closing link")
				}
			}
		}
	}
}

type linkDestUpdate struct {
	id        string
	version   string
	healthy   bool
	listeners []*ctrl_pb.Listener
}

func (self *linkDestUpdate) Handle(registry *linkRegistryImpl) {
	logtrace.LogWithFunctionName()
	dest := registry.destinations[self.id]

	becameHealthy := false

	if dest == nil {
		dest = newLinkDest(self.id)
		registry.destinations[self.id] = dest
	} else {
		if !dest.healthy && self.healthy {
			becameHealthy = true
		}
	}
	dest.update(self)

	if self.healthy {
		self.ApplyListenerChanges(registry, dest, becameHealthy)
	}
}

func (self *linkDestUpdate) ApplyListenerChanges(registry *linkRegistryImpl, dest *linkDest, becameHealthy bool) {
	logtrace.LogWithFunctionName()
	currentLinkKeys := map[string]struct{}{}

	for k := range dest.linkMap {
		currentLinkKeys[k] = struct{}{}
	}

	for _, listener := range self.listeners {
		for _, dialer := range registry.env.GetXlinkDialers() {
			if stringz.ContainsAny(listener.Groups, dialer.GetGroups()...) {
				linkKey := registry.GetLinkKey(dialer.GetBinding(), listener.Protocol, self.id, listener.GetLocalBinding())

				delete(currentLinkKeys, linkKey)

				log := pfxlog.Logger().WithField("routerId", self.id).
					WithField("address", listener.Address).
					WithField("linkKey", linkKey)

				existingLinkState, ok := dest.linkMap[linkKey]
				if !ok {
					newLinkState := &linkState{
						linkKey:      linkKey,
						linkId:       idgen.NewUUIDString(),
						status:       StatusPending,
						dest:         dest,
						listener:     listener,
						dialer:       dialer,
						allowedDials: -1,
					}
					dest.linkMap[linkKey] = newLinkState
					log.Info("new potential link")
					registry.evaluateLinkState(newLinkState)
				} else {
					log.Info("link already known")
					// if link isn't established, try establishing now
					if becameHealthy && existingLinkState.status != StatusEstablished {
						existingLinkState.retryDelay = time.Duration(0)
						existingLinkState.nextDial = time.Now()
						registry.evaluateLinkState(existingLinkState)
					}
				}
			}
		}
	}

	// anything left is an orphaned link entry
	for linkKey := range currentLinkKeys {
		// this will prevent the link from being recreated once closed
		delete(dest.linkMap, linkKey)
	}
}

type dialRequest struct {
	ctrlCh channel.Channel
	dial   *ctrl_pb.Dial
}

func (self *dialRequest) Handle(registry *linkRegistryImpl) {
	logtrace.LogWithFunctionName()
	dest := registry.destinations[self.dial.RouterId]

	if dest == nil {
		dest = newLinkDest(self.dial.RouterId)
		dest.version = self.dial.RouterVersion
		registry.destinations[self.dial.RouterId] = dest
	}

	for _, dialer := range registry.env.GetXlinkDialers() {
		if stringz.ContainsAny(dialer.GetGroups(), GroupDefault) {
			linkKey := registry.GetLinkKey(GroupDefault, self.dial.LinkProtocol, self.dial.RouterId, GroupDefault)

			log := pfxlog.Logger().WithField("routerId", self.dial.RouterId).
				WithField("address", self.dial.Address).
				WithField("linkKey", linkKey)

			if link, found := registry.GetLink(linkKey); found {
				registry.SendRouterLinkMessage(link, self.ctrlCh)
				continue
			}

			existingLinkState, ok := dest.linkMap[linkKey]
			if !ok {
				newLinkState := &linkState{
					linkKey: linkKey,
					linkId:  idgen.NewUUIDString(),
					status:  StatusPending,
					dest:    dest,
					listener: &ctrl_pb.Listener{
						Address:  self.dial.Address,
						Protocol: self.dial.LinkProtocol,
						Groups:   []string{GroupDefault},
					},
					dialer:       dialer,
					allowedDials: 1,
				}
				log = log.WithField("linkId", newLinkState.linkId)
				dest.linkMap[linkKey] = newLinkState
				log.Info("new potential link")
				registry.evaluateLinkState(newLinkState)
			} else if existingLinkState.status != StatusEstablished {
				log = log.WithField("linkId", existingLinkState.linkId)

				existingLinkState.retryDelay = time.Duration(0)
				existingLinkState.nextDial = time.Now()
				existingLinkState.allowedDials = 1

				log.Info("dial request received for existing link, re-evaluating")
				registry.evaluateLinkState(existingLinkState)
			}
		}
	}
}

type updateLinkStatusForLink struct {
	link   xlink.Xlink
	status linkStatus
}

func (self *updateLinkStatusForLink) Handle(registry *linkRegistryImpl) {
	logtrace.LogWithFunctionName()
	link := self.link
	log := pfxlog.Logger().WithField("linkKey", link.Key()).WithField("linkId", link.Id())
	dest, found := registry.destinations[link.DestinationId()]
	if !found {
		if link.IsDialed() { // if link was created by listener, rather than dialer we may not have an entry for it
			log.WithField("linkDest", link.DestinationId()).Warnf("unable to mark link as %s, link destination not present in registry", self.status)
		}
		return
	}

	state, found := dest.linkMap[link.Key()]
	if !found {
		if link.IsDialed() { // if link was created by listener, rather than dialer we may not have an entry for it
			log.WithField("linkDest", link.DestinationId()).Warnf("unable to mark link as %s, link state not present in registry", self.status)
		}
		return
	}

	if state.status == StatusDestRemoved {
		return
	}

	state.updateStatus(self.status)
	if state.status == StatusEstablished {
		state.connectedCount++
		state.retryDelay = time.Duration(0)
		state.ctrlsNotified = false
		state.link = self.link
		registry.triggerNotify()
	}

	if state.status == StatusLinkFailed {
		state.retryDelay = time.Duration(0)
		state.nextDial = time.Now()
		registry.evaluateLinkState(state)
		state.addPendingLinkFault(link.Id(), link.Iteration())
		state.link = nil
	}
}

type addLinkFaultForReplacedLink struct {
	link xlink.Xlink
}

func (self *addLinkFaultForReplacedLink) Handle(registry *linkRegistryImpl) {
	logtrace.LogWithFunctionName()
	link := self.link
	log := pfxlog.Logger().WithField("linkKey", link.Key()).WithField("linkId", link.Id())
	dest, found := registry.destinations[link.DestinationId()]
	if !found {
		if link.IsDialed() { // if link was created by listener, rather than dialer we may not have an entry for it
			log.WithField("linkDest", link.DestinationId()).Info("link destination not present in registry")
		}
		return
	}

	state, found := dest.linkMap[link.Key()]
	if !found {
		if link.IsDialed() { // if link was created by listener, rather than dialer we may not have an entry for it
			log.WithField("linkDest", link.DestinationId()).Info("link state not present in registry")
		}
		return
	}

	state.addPendingLinkFault(link.Id(), link.Iteration())
}

type updateLinkStatusToDialFailed struct {
	linkState *linkState
}

func (self *updateLinkStatusToDialFailed) Handle(registry *linkRegistryImpl) {
	logtrace.LogWithFunctionName()
	if self.linkState.status == StatusDialing {
		self.linkState.updateStatus(StatusDialFailed)
		self.linkState.dialFailed(registry)
	}
}

type inspectLinkStatesEvent struct {
	result atomic.Pointer[[]*inspect.LinkDest]
	done   chan struct{}
}

func (self *inspectLinkStatesEvent) Handle(registry *linkRegistryImpl) {
	logtrace.LogWithFunctionName()
	var result []*inspect.LinkDest
	for _, dest := range registry.destinations {
		inspectDest := &inspect.LinkDest{
			Id:      dest.id,
			Version: dest.version,
			Healthy: dest.healthy,
		}
		unhealthySince := dest.unhealthyAt
		if !dest.healthy {
			inspectDest.UnhealthySince = &unhealthySince
		}

		for _, state := range dest.linkMap {
			establishedLinkId := ""
			if link := state.link; link != nil {
				establishedLinkId = link.Id()
			}
			inspectLinkState := &inspect.LinkState{
				Id:                state.linkId,
				Key:               state.linkKey,
				Status:            state.status.String(),
				DialAttempts:      state.dialAttempts.Load(),
				ConnectedCount:    state.connectedCount,
				RetryDelay:        state.retryDelay.String(),
				NextDial:          state.nextDial.Format(time.RFC3339),
				TargetAddress:     state.listener.Address,
				TargetGroups:      state.listener.Groups,
				TargetBinding:     state.listener.LocalBinding,
				DialerGroups:      state.dialer.GetGroups(),
				DialerBinding:     state.dialer.GetBinding(),
				CtrlsNotified:     state.ctrlsNotified,
				EstablishedLinkId: establishedLinkId,
			}
			if inspectLinkState.TargetBinding == "" {
				inspectLinkState.TargetBinding = "default"
			}
			if inspectLinkState.DialerBinding == "" {
				inspectLinkState.DialerBinding = "default"
			}
			inspectDest.LinkStates = append(inspectDest.LinkStates, inspectLinkState)
		}

		result = append(result, inspectDest)
	}
	self.result.Store(&result)
	close(self.done)
}

func (self *inspectLinkStatesEvent) GetResults(timeout time.Duration) ([]*inspect.LinkDest, error) {
	logtrace.LogWithFunctionName()
	select {
	case <-self.done:
		return *self.result.Load(), nil
	case <-time.After(timeout):
		return nil, errors.New("timed out waiting for result")
	}
}

type markNewLinksNotified struct {
	links []stateAndLink
}

func (self *markNewLinksNotified) Handle(*linkRegistryImpl) {
	logtrace.LogWithFunctionName()
	for _, pair := range self.links {
		if pair.state.status == StatusEstablished && pair.link == pair.state.link {
			pair.state.ctrlsNotified = true
		}
	}
}

type markFaultedLinksNotified struct {
	successfullySent []stateAndFaults
}

func (self *markFaultedLinksNotified) Handle(*linkRegistryImpl) {
	logtrace.LogWithFunctionName()
	for _, pair := range self.successfullySent {
		state := pair.state
		for _, fault := range pair.faults {
			state.clearFault(fault)
		}
	}
}

type scanForLinkIdEvent struct {
	linkId  string
	resultC chan bool
}

func (self *scanForLinkIdEvent) Handle(r *linkRegistryImpl) {
	logtrace.LogWithFunctionName()
	for _, dest := range r.destinations {
		for _, state := range dest.linkMap {
			if state.linkId == self.linkId {
				self.resultC <- true
				return
			}
		}
	}
	self.resultC <- false
}
