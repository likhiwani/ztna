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
	"ztna-core/ztna/zititest/zitilab/cli"

	"github.com/openziti/fablab/kernel/model"
)

func Edge(args ...string) model.Action {
	logtrace.LogWithFunctionName()
	return &edge{
		args: args,
	}
}

func (a *edge) Execute(run model.Run) error {
	logtrace.LogWithFunctionName()
	return EdgeExec(run.GetModel(), a.args...)
}

type edge struct {
	args []string
}

func EdgeExec(m *model.Model, args ...string) error {
	logtrace.LogWithFunctionName()
	_, err := EdgeExecWithOutput(m, args...)
	return err
}

func EdgeExecWithOutput(m *model.Model, args ...string) (string, error) {
	logtrace.LogWithFunctionName()
	return cli.Exec(m, append([]string{"edge", "-i", model.ActiveInstanceId()}, args...)...)
}
