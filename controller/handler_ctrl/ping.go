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
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
)

type pingHandler struct{}

func newPingHandler() *pingHandler {
	logtrace.LogWithFunctionName()
	return &pingHandler{}
}

func (h *pingHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return channel.ContentTypePingType
}

func (h *pingHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	go handler_common.SendResult(msg, ch, "ok", true)
}
