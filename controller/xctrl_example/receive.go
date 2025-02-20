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

package xctrl_example

import (
	"bytes"
	"encoding/binary"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
)

type receiveHandler struct{}

func newReceiveHandler() channel.TypedReceiveHandler {
	logtrace.LogWithFunctionName()
	return &receiveHandler{}
}

func (h *receiveHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return contentType
}

func (h *receiveHandler) HandleReceive(msg *channel.Message, _ channel.Channel) {
	logtrace.LogWithFunctionName()
	if len(msg.Body) == 4 {
		buf := bytes.NewBuffer(msg.Body)
		var count int32
		if err := binary.Read(buf, binary.LittleEndian, &count); err == nil {
			pfxlog.Logger().Infof("received [%d]", count)
		} else {
			pfxlog.Logger().Errorf("unexpected error (%s)", err)
		}
	} else {
		pfxlog.Logger().Errorf("unexpected body length [%d]", len(msg.Body))
	}
}
