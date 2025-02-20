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

package xt_weighted

import (
	"math/rand"
	"time"
	"ztna-core/ztna/controller/xt"
	"ztna-core/ztna/controller/xt_common"
	logtrace "ztna-core/ztna/logtrace"
)

/**
The weighted strategy does random selection of available strategies in proportion to the terminator costs. So if a
given terminator has twice the fully evaluated cost as another terminator it should ideally be selected roughly half
as often.
*/

func NewFactory() xt.Factory {
	logtrace.LogWithFunctionName()
	return &factory{}
}

type factory struct{}

func (self *factory) GetStrategyName() string {
	logtrace.LogWithFunctionName()
	return "weighted"
}

func (self *factory) NewStrategy() xt.Strategy {
	logtrace.LogWithFunctionName()
	strategy := &strategy{
		CostVisitor: *xt_common.NewCostVisitor(2, 20, 2),
	}
	strategy.CostVisitor.CreditOverTime(5, time.Minute)
	return strategy
}

type strategy struct {
	xt_common.CostVisitor
}

func (self *strategy) Select(_ xt.CreateCircuitParams, terminators []xt.CostedTerminator) (xt.CostedTerminator, xt.PeerData, error) {
	logtrace.LogWithFunctionName()
	terminators = xt.GetRelatedTerminators(terminators)
	if len(terminators) == 1 {
		return terminators[0], nil, nil
	}

	var costIdx []float32
	totalCost := float32(0)
	for _, t := range terminators {
		unbiasedCost := float32(t.GetPrecedence().Unbias(t.GetRouteCost()))
		if unbiasedCost == 0 {
			unbiasedCost = 1
		}
		costIdx = append(costIdx, unbiasedCost)
		totalCost += unbiasedCost
	}

	total := float32(0)
	for idx, cost := range costIdx {
		total += 1 - (cost / totalCost)
		costIdx[idx] = total
	}

	selected := rand.Float32()
	for idx, cost := range costIdx {
		if selected < cost {
			return terminators[idx], nil, nil
		}
	}

	return terminators[0], nil, nil
}
