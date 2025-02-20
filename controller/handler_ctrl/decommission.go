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
	"ztna-core/ztna/common/handler_common"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
)

type decommissionRouterHandler struct {
	baseHandler
}

func newDecommissionRouterHandler(router *model.Router, network *network.Network) *decommissionRouterHandler {
	logtrace.LogWithFunctionName()
	return &decommissionRouterHandler{
		baseHandler: baseHandler{
			router:  router,
			network: network,
		},
	}
}

func (self *decommissionRouterHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(ctrl_pb.ContentType_DecommissionRouterRequestType)
}

func (self *decommissionRouterHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label()).Entry
	log = log.WithField("routerId", self.router.Id)

	go func() {
		if err := self.network.Router.Delete(self.router.Id, self.newChangeContext(ch, "decommission.router")); err == nil {
			// we don't send success because deleting the router will close the router connection
			log.Debug("router decommission successful")
		} else {
			handler_common.SendFailure(msg, ch, err.Error())
			log.WithError(err).Error("router decommission failed")
		}
	}()
}
