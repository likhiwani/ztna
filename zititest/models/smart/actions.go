/*
	Copyright 2020 NetFoundry Inc.

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

package main

import (
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/zititest/models/smart/actions"
	zitilab_actions "ztna-core/ztna/zititest/zitilab/actions"
	"ztna-core/ztna/zititest/zitilab/console"

	"github.com/openziti/fablab/kernel/model"
)

func newActionsFactory() model.Factory {
	logtrace.LogWithFunctionName()
	return &actionsFactory{}
}

func (_ *actionsFactory) Build(m *model.Model) error {
	logtrace.LogWithFunctionName()
	m.Actions = model.ActionBinders{
		"bootstrap": actions.NewBootstrapAction(),
		"start":     actions.NewStartAction(),
		"stop":      actions.NewStopAction(),
		"console":   func(m *model.Model) model.Action { return console.Console() },
		"logs":      func(m *model.Model) model.Action { return zitilab_actions.Logs() },
	}
	return nil
}

type actionsFactory struct{}
