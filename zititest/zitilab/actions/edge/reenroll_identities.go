package edge

import (
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/zititest/zitilab"

	"github.com/openziti/fablab/kernel/lib/actions/component"
	"github.com/openziti/fablab/kernel/model"
)

func ReEnrollIdentities(componentSpec string, concurrency int) model.Action {
	logtrace.LogWithFunctionName()
	return &reEnrollIdentitiesAction{
		componentSpec: componentSpec,
		concurrency:   concurrency,
	}
}

func (action *reEnrollIdentitiesAction) Execute(run model.Run) error {
	logtrace.LogWithFunctionName()
	return component.ExecInParallel(action.componentSpec, action.concurrency, zitilab.ZitiTunnelActionsReEnroll).Execute(run)
}

type reEnrollIdentitiesAction struct {
	componentSpec string
	concurrency   int
}
