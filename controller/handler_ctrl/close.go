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
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
)

type closeHandler struct {
	r       *model.Router
	network *network.Network
}

func newCloseHandler(r *model.Router, network *network.Network) *closeHandler {
	logtrace.LogWithFunctionName()
	return &closeHandler{r: r, network: network}
}

func (h *closeHandler) HandleClose(channel.Channel) {
	logtrace.LogWithFunctionName()
	pfxlog.Logger().WithField("routerId", h.r.Id).Warn("disconnected")
	h.network.DisconnectRouter(h.r)
}

type xctrlCloseHandler struct {
	done chan struct{}
}

func newXctrlCloseHandler(done chan struct{}) channel.CloseHandler {
	logtrace.LogWithFunctionName()
	return &xctrlCloseHandler{done: done}
}

func (h *xctrlCloseHandler) HandleClose(ch channel.Channel) {
	logtrace.LogWithFunctionName()
	pfxlog.ContextLogger(ch.Label()).Info("closing Xctrl instances")
	close(h.done)
}
