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
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
)

type baseHandler struct {
	router  *model.Router
	network *network.Network
}

func (self *baseHandler) newChangeContext(ch channel.Channel, method string) *change.Context {
	logtrace.LogWithFunctionName()
	return change.NewControlChannelChange(self.router.Id, self.router.Name, method, ch)
}
