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

type controlHandler struct {
	link      xlink.Xlink
	forwarder *forwarder.Forwarder
}

func newControlHandler(link xlink.Xlink, forwarder *forwarder.Forwarder) *controlHandler {
	logtrace.LogWithFunctionName()
	result := &controlHandler{
		link:      link,
		forwarder: forwarder,
	}
	return result
}

func (self *controlHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return xgress.ContentTypeControlType
}

func (self *controlHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label())

	if control, err := xgress.UnmarshallControl(msg); err == nil {
		if err = self.forwarder.ForwardControl(xgress.Address(self.link.Id()), control); err != nil {
			log.WithError(err).Debug("unable to forward")
		}
	} else {
		log.WithError(err).Errorf("unexpected error marshalling control instance")
	}
}
