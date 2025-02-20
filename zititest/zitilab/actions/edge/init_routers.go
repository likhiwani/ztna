package edge

import (
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/zititest/zitilab"
	zitilib_actions "ztna-core/ztna/zititest/zitilab/actions"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fablab/kernel/lib/actions/component"
	"github.com/openziti/fablab/kernel/model"
)

func InitEdgeRouters(componentSpec string, concurrency int) model.Action {
	logtrace.LogWithFunctionName()
	return &initEdgeRoutersAction{
		componentSpec: componentSpec,
		concurrency:   concurrency,
	}
}

func (action *initEdgeRoutersAction) Execute(run model.Run) error {
	logtrace.LogWithFunctionName()
	if err := zitilib_actions.EdgeExec(run.GetModel(), "delete", "edge-router", "where", "true"); err != nil {
		pfxlog.Logger().WithError(err).Warn("unable to delete routers")
	}

	return component.ExecInParallel(action.componentSpec, action.concurrency, zitilab.RouterActionsCreateAndEnroll).Execute(run)
}

type initEdgeRoutersAction struct {
	componentSpec string
	concurrency   int
}
