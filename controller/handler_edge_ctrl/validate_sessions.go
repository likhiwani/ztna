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
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

type validateSessionsHandler struct {
	baseRequestHandler
}

func NewValidateSessionsHandler(appEnv *env.AppEnv, ch channel.Channel) channel.TypedReceiveHandler {
	logtrace.LogWithFunctionName()
	return &validateSessionsHandler{
		baseRequestHandler{
			ch:     ch,
			appEnv: appEnv,
		},
	}
}

func (self *validateSessionsHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(edge_ctrl_pb.ContentType_ValidateSessionsRequestType)
}

func (self *validateSessionsHandler) Label() string {
	logtrace.LogWithFunctionName()
	return "validate.sessions"
}

func (self *validateSessionsHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	req := &edge_ctrl_pb.ValidateSessionsRequest{}
	if err := proto.Unmarshal(msg.Body, req); err != nil {
		pfxlog.ContextLogger(ch.Label()).WithError(err).Error("could not unmarshal ValidateSessionsRequest")
		return
	}

	go self.validateSessions(req)
}

func (self *validateSessionsHandler) validateSessions(req *edge_ctrl_pb.ValidateSessionsRequest) {
	logtrace.LogWithFunctionName()
	sessionStore := self.getAppEnv().GetStores().Session
	tokenIndex := sessionStore.GetTokenIndex()

	var invalidTokens []string

	err := self.getAppEnv().GetDb().View(func(tx *bbolt.Tx) error {
		for _, token := range req.SessionTokens {
			if tokenIndex.Read(tx, []byte(token)) == nil {
				invalidTokens = append(invalidTokens, token)
			}
		}
		return nil
	})

	if err != nil {
		pfxlog.ContextLogger(self.ch.Label()).WithError(err).Errorf("failure while validating session tokens")
	}

	if len(invalidTokens) > 0 {
		sessionsRemoved := &edge_ctrl_pb.SessionRemoved{
			Tokens: invalidTokens,
		}

		body, err := proto.Marshal(sessionsRemoved)
		if err != nil {
			pfxlog.ContextLogger(self.ch.Label()).WithError(err).Error("failed to marshal sessions removed")
			return
		}

		msg := channel.NewMessage(int32(edge_ctrl_pb.ContentType_SessionRemovedType), body)
		if err := self.ch.Send(msg); err != nil {
			pfxlog.ContextLogger(self.ch.Label()).WithError(err).Error("failed to send validate sessions request")
			return
		}
	}
}
