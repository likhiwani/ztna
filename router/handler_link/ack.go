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

package handler_link

import (
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/forwarder"
	"ztna-core/ztna/router/xgress"
	"ztna-core/ztna/router/xlink"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
)

type ackHandler struct {
	link      xlink.Xlink
	forwarder *forwarder.Forwarder
}

func newAckHandler(link xlink.Xlink, forwarder *forwarder.Forwarder) *ackHandler {
	logtrace.LogWithFunctionName()
	return &ackHandler{
		link:      link,
		forwarder: forwarder,
	}
}

func (self *ackHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return xgress.ContentTypeAcknowledgementType
}

func (self *ackHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	ack, err := xgress.UnmarshallAcknowledgement(msg)
	if err != nil {
		pfxlog.ContextLogger(ch.Label()).
			WithField("linkId", self.link.Id()).
			WithField("routerId", self.link.DestinationId()).
			WithError(err).Error("error unmarshalling ack")
		return
	}

	if err = self.forwarder.ForwardAcknowledgement(xgress.Address(self.link.Id()), ack); err != nil {
		pfxlog.ContextLogger(ch.Label()).
			WithField("linkId", self.link.Id()).
			WithField("routerId", self.link.DestinationId()).
			WithError(err).Debug("unable to forward acknowledgement")
	}
}
