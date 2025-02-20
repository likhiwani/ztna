/*
	(c) Copyright NetFoundry Inc.

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

package actions

import (
	"time"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/zititest/zitilab/actions/edge"

	zitilib_actions "ztna-core/ztna/zititest/zitilab/actions"
	"ztna-core/ztna/zititest/zitilab/models"

	"github.com/openziti/fablab/kernel/lib/actions"
	"github.com/openziti/fablab/kernel/lib/actions/component"
	"github.com/openziti/fablab/kernel/lib/actions/semaphore"
	"github.com/openziti/fablab/kernel/model"
)

func NewStartAction() model.ActionBinder {
	logtrace.LogWithFunctionName()
	action := &startAction{}
	return action.bind
}

func (a *startAction) bind(m *model.Model) model.Action {
	logtrace.LogWithFunctionName()
	workflow := actions.Workflow()
	workflow.AddAction(component.Start(".ctrl"))
	workflow.AddAction(edge.ControllerAvailable("#ctrl", 30*time.Second))
	workflow.AddAction(component.StartInParallel(models.EdgeRouterTag, 25))

	workflow.AddAction(semaphore.Sleep(5 * time.Second))
	workflow.AddAction(edge.Login("#ctrl"))
	workflow.AddAction(zitilib_actions.Edge("list", "edge-routers", "limit none"))
	workflow.AddAction(zitilib_actions.Edge("list", "terminators", "limit none"))

	return workflow
}

type startAction struct{}
