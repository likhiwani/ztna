package debugops

import (
	logtrace "ztna-core/ztna/logtrace"
	"ztna-core/ztna/router"
)

const (
	DumpApiSessions byte = 128
)

func RegisterEdgeRouterAgentOps(router *router.Router, debugEnabled bool) {
	logtrace.LogWithFunctionName()
	if sm := router.GetStateManager(); sm != nil {
		router.RegisterAgentOp(DumpApiSessions, sm.DumpApiSessions)
	}
}
