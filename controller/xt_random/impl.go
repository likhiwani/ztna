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

package xt_random

import (
	"math/rand"
	"ztna-core/ztna/controller/xt"
	logtrace "ztna-core/ztna/logtrace"
)

/**
The random strategy uses a random selection from available terminators. It only picks from terminators which
match the precedence of the first terminator, which is presumably of the highest available precedence.
*/

func NewFactory() xt.Factory {
	logtrace.LogWithFunctionName()
	return &factory{}
}

type factory struct{}

func (self *factory) GetStrategyName() string {
	logtrace.LogWithFunctionName()
	return "random"
}

func (self *factory) NewStrategy() xt.Strategy {
	logtrace.LogWithFunctionName()
	return &strategy{}
}

type strategy struct{}

func (self *strategy) Select(_ xt.CreateCircuitParams, terminators []xt.CostedTerminator) (xt.CostedTerminator, xt.PeerData, error) {
	logtrace.LogWithFunctionName()
	terminators = xt.GetRelatedTerminators(terminators)
	count := len(terminators)
	if count == 1 {
		return terminators[0], nil, nil
	}
	selected := rand.Intn(count)
	return terminators[selected], nil, nil
}

func (self *strategy) NotifyEvent(xt.TerminatorEvent) {
	logtrace.LogWithFunctionName()
}

func (self *strategy) HandleTerminatorChange(xt.StrategyChangeEvent) error {
	logtrace.LogWithFunctionName()
	return nil
}
