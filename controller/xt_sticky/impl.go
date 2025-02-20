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

package xt_sticky

import (
	"time"
	"ztna-core/ztna/common/ctrl_msg"
	"ztna-core/ztna/controller/xt"
	"ztna-core/ztna/controller/xt_common"
	logtrace "ztna-core/ztna/logtrace"
)

const (
	Name = "sticky"
)

/**
The sticky strategy uses the smart routing strategy to select an initial terminator for a client. It also
returns a token to the client which can be passed back in on subsequent dials. If the token is passed
back in, then strategy will try to use the same terminator. If it's not available a different terminator
will be selected and a different token will be returned.
*/

func NewFactory() xt.Factory {
	logtrace.LogWithFunctionName()
	return &factory{}
}

type factory struct{}

func (self *factory) GetStrategyName() string {
	logtrace.LogWithFunctionName()
	return Name
}

func (self *factory) NewStrategy() xt.Strategy {
	logtrace.LogWithFunctionName()
	strategy := strategy{
		CostVisitor: *xt_common.NewCostVisitor(2, 20, 2),
	}
	strategy.CostVisitor.CreditOverTime(5, time.Minute)
	return &strategy
}

type strategy struct {
	xt_common.CostVisitor
}

func (self *strategy) Select(params xt.CreateCircuitParams, terminators []xt.CostedTerminator) (xt.CostedTerminator, xt.PeerData, error) {
	logtrace.LogWithFunctionName()
	id := params.GetClientId()
	var result xt.CostedTerminator

	terminators = xt.GetRelatedTerminators(terminators)

	if id != nil {
		if terminatorId, ok := id.Data[ctrl_msg.XtStickinessToken]; ok {
			strId := string(terminatorId)
			for _, terminator := range terminators {
				if terminator.GetId() == strId {
					result = terminator
					break
				}
			}
		}
	}

	if result == nil {
		result = terminators[0]
	}

	return result, xt.PeerData{
		ctrl_msg.XtStickinessToken: []byte(result.GetId()),
	}, nil
}
