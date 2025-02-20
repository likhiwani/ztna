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
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/xlink"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"google.golang.org/protobuf/proto"
)

type faultHandler struct {
	xlinkRegistry xlink.Registry
}

func newFaultHandler(registry xlink.Registry) *faultHandler {
	logtrace.LogWithFunctionName()
	return &faultHandler{xlinkRegistry: registry}
}

func (self *faultHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(ctrl_pb.ContentType_FaultType)
}

func (self *faultHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label())

	fault := &ctrl_pb.Fault{}
	if err := proto.Unmarshal(msg.Body, fault); err != nil {
		log.WithError(err).Error("failed to unmarshal fault message")
		return
	}

	go self.handleFault(msg, ch, fault)
}

func (self *faultHandler) handleFault(_ *channel.Message, ch channel.Channel, fault *ctrl_pb.Fault) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label()).Entry

	switch fault.Subject {
	case ctrl_pb.FaultSubject_LinkFault:
		linkId := fault.Id
		log = log.WithField("linkId", linkId)
		if link, _ := self.xlinkRegistry.GetLinkById(linkId); link != nil {
			if fault.Iteration > 0 && fault.Iteration < link.Iteration() {
				log.WithField("fault.iteration", fault.Iteration).
					WithField("link.iteration", link.Iteration()).
					Info("link fault reported, but fault iteration < link iteration, ignoring")
				return
			}
			log.Info("link fault reported, closing")
			if err := link.CloseNotified(); err != nil {
				log.WithError(err).Error("failure closing link")
			}
		} else {
			log.Info("link fault reported, link already closed or unknown")
		}

	default:
		log.WithField("subject", fault.Subject.String()).Error("unhandled fault subject")
	}
}
