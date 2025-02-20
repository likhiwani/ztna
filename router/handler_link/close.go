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
	"ztna-core/ztna/router/xlink"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
)

type closeHandler struct {
	link          xlink.Xlink
	forwarder     *forwarder.Forwarder
	xlinkRegistry xlink.Registry
}

func newCloseHandler(link xlink.Xlink, forwarder *forwarder.Forwarder, registry xlink.Registry) *closeHandler {
	logtrace.LogWithFunctionName()
	return &closeHandler{
		link:          link,
		forwarder:     forwarder,
		xlinkRegistry: registry,
	}
}

func (self *closeHandler) HandleClose(ch channel.Channel) {
	logtrace.LogWithFunctionName()
	self.link.CloseOnce(func() {
		log := pfxlog.ContextLogger(ch.Label()).
			WithField("linkId", self.link.Id()).
			WithField("routerId", self.link.DestinationId()).
			WithField("iteration", self.link.Iteration())

		self.forwarder.UnregisterLink(self.link)

		// ensure that both parts of a split link are closed, if one side closes
		go func() {
			_ = self.link.Close()
			// Close can be called from the link registry, so we can't call back into it from the same go-routine
			self.xlinkRegistry.LinkClosed(self.link)
		}()

		log.Info("link closed")
	})
}
