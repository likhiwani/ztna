/*
	Copyright 2019 NetFoundry Inc.

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

package zitilib_actions

import (
	"ztna-core/ztna/logtrace"

	"github.com/openziti/fablab/kernel/model"
)

func StopAll(hostSpec string) model.Action {
	logtrace.LogWithFunctionName()
	return StopAllInParallel(hostSpec, 1)
}

func StopAllInParallel(hostSpec string, concurrency int) model.Action {
	logtrace.LogWithFunctionName()
	return &stopAll{
		hostSpec:    hostSpec,
		concurrency: concurrency,
	}
}

func (stop *stopAll) Execute(run model.Run) error {
	logtrace.LogWithFunctionName()
	return run.GetModel().ForEachHost(stop.hostSpec, stop.concurrency, func(c *model.Host) error {
		return nil
	})
}

type stopAll struct {
	hostSpec    string
	concurrency int
}
