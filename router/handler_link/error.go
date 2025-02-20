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
	"ztna-core/ztna/router/env"
	"ztna-core/ztna/router/xlink"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
)

type errorHandler struct {
	link xlink.Xlink
	ctrl env.NetworkControllers
}

func newErrorHandler(link xlink.Xlink, ctrl env.NetworkControllers) *errorHandler {
	logtrace.LogWithFunctionName()
	return &errorHandler{link: link, ctrl: ctrl}
}

func (self *errorHandler) HandleError(err error, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label()).
		WithField("linkId", self.link.Id()).
		WithField("routerId", self.link.DestinationId())

	log.WithError(err).Error("link error, closing")
	if err := self.link.Close(); err != nil { // this will trigger the link close handler, which will send the fault
		log.WithError(err).Error("error while closing link")
	}
}
