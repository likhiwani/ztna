package edge

import (
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/zititest/zitilab"

	"github.com/openziti/fablab/kernel/lib/actions/component"
	"github.com/openziti/fablab/kernel/model"
)

func ReEnrollEdgeRouters(componentSpec string, concurrency int) model.Action {
	logtrace.LogWithFunctionName()
	return &reEnrollEdgeRoutersAction{
		componentSpec: componentSpec,
		concurrency:   concurrency,
	}
}

func (action *reEnrollEdgeRoutersAction) Execute(run model.Run) error {
	logtrace.LogWithFunctionName()
	return component.ExecInParallel(action.componentSpec, action.concurrency, zitilab.RouterActionsReEnroll).Execute(run)
}

type reEnrollEdgeRoutersAction struct {
	componentSpec string
	concurrency   int
}
