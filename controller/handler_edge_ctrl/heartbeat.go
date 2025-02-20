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

package handler_edge_ctrl

import (
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"google.golang.org/protobuf/proto"
)

type sessionHeartbeatHandler struct {
	appEnv *env.AppEnv
}

func NewSessionHeartbeatHandler(appEnv *env.AppEnv) *sessionHeartbeatHandler {
	logtrace.LogWithFunctionName()
	return &sessionHeartbeatHandler{
		appEnv: appEnv,
	}
}

func (h *sessionHeartbeatHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return env.ApiSessionHeartbeatType
}

func (h *sessionHeartbeatHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	go func() {
		req := &edge_ctrl_pb.ApiSessionHeartbeat{}
		routerId := ch.Id()
		if err := proto.Unmarshal(msg.Body, req); err == nil {

			identityIds, notFoundTokens, err := h.appEnv.GetManagers().ApiSession.MarkLastActivityByTokens(req.Tokens...)

			if err != nil {
				pfxlog.Logger().
					WithError(err).
					WithField("routerId", routerId).
					Errorf("unable to set activity for heartbeat: %v", err)
			}

			for _, identityId := range identityIds {
				h.appEnv.GetManagers().Identity.SetHasErConnection(identityId)
			}

			if len(notFoundTokens) > 0 {
				pfxlog.Logger().
					WithField("routerId", routerId).
					Debugf("api session tokens not found during heartbeat, sending delete: %v", notFoundTokens)

				msgStruct := &edge_ctrl_pb.ApiSessionRemoved{
					Tokens: notFoundTokens,
				}

				msgBytes, _ := proto.Marshal(msgStruct)
				msg := channel.NewMessage(env.ApiSessionRemovedType, msgBytes)

				_ = ch.Send(msg)
			}

		} else {
			pfxlog.Logger().Error("could not convert message as session heartbeat")
		}
	}()
}
