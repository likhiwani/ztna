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

package state

import (
	"time"
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/common/runner"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/logtrace"
	routerEnv "ztna-core/ztna/router/env"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"google.golang.org/protobuf/proto"
)

const maxTokensPerMessage = 10000

type heartbeatOperation struct {
	env routerEnv.RouterEnv
	*runner.BaseOperation
	tokenProvider TokenProvider
}

type TokenProvider interface {
	ActiveApiSessionTokens() []string
	flushRecentlyRemoved()
}

func newHeartbeatOperation(env routerEnv.RouterEnv, frequency time.Duration, tokenProvider TokenProvider) *heartbeatOperation {
	logtrace.LogWithFunctionName()
	return &heartbeatOperation{
		env:           env,
		tokenProvider: tokenProvider,
		BaseOperation: runner.NewBaseOperation("Heartbeat Operation", frequency)}
}

func (operation *heartbeatOperation) Run() error {
	logtrace.LogWithFunctionName()
	tokens := operation.tokenProvider.ActiveApiSessionTokens()

	operation.beat(tokens)
	operation.tokenProvider.flushRecentlyRemoved()

	return nil
}

func (operation *heartbeatOperation) beat(tokens []string) {
	logtrace.LogWithFunctionName()
	var msgs []*channel.Message

	pfxlog.Logger().Tracef("heartbeat tokens: %d", len(tokens))

	for len(tokens) > 0 {

		if maxTokensPerMessage >= len(tokens) {
			msg := &edge_ctrl_pb.ApiSessionHeartbeat{
				Tokens: tokens,
			}
			bodyBytes, err := proto.Marshal(msg)

			if err != nil {
				pfxlog.Logger().Panic("could not marshal SessionHeartbeat type (1)")
			}

			msgs = append(msgs, channel.NewMessage(env.ApiSessionHeartbeatType, bodyBytes))

			tokens = nil
		} else {
			msg := &edge_ctrl_pb.ApiSessionHeartbeat{
				Tokens: tokens[:maxTokensPerMessage],
			}

			bodyBytes, err := proto.Marshal(msg)

			if err != nil {
				pfxlog.Logger().Panic("could not marshal SessionHeartbeat type (2)")
			}

			tokens = tokens[maxTokensPerMessage:]
			msgs = append(msgs, channel.NewMessage(env.ApiSessionHeartbeatType, bodyBytes))
		}
	}

	operation.env.GetNetworkControllers().ForEach(func(ctrlId string, ch channel.Channel) {
		for _, msg := range msgs {
			if err := ch.Send(msg); err != nil {
				pfxlog.Logger().WithError(err).Error("could not send heartbeats on control channel")
			}

		}
	})

}
