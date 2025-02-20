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

package xt_smartrouting

import (
	"time"
	"ztna-core/ztna/controller/xt"
	"ztna-core/ztna/controller/xt_common"
	logtrace "ztna-core/ztna/logtrace"
)

const (
	Name = "smartrouting"
)

/**
The smart routing strategy relies purely on manipulating costs and lets the smart routing algorithm pick the terminator.
It increases costs by a small amount when a new circuit uses the terminator and drops it back down when the circuit
closes. It also increases the cost whenever a dial fails and decreases it whenever a dial succeeds. Dial successes
will only reduce costs by the amount that failures have previously increased it.
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
	strategy.CreditOverTime(5, time.Minute)
	return &strategy
}

type strategy struct {
	xt_common.CostVisitor
}

func (self *strategy) Select(_ xt.CreateCircuitParams, terminators []xt.CostedTerminator) (xt.CostedTerminator, xt.PeerData, error) {
	logtrace.LogWithFunctionName()
	return terminators[0], nil, nil
}
