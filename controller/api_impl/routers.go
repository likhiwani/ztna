package api_impl

import (
	"ztna-core/ztna/controller/rest_server/operations"
	"ztna-core/ztna/logtrace"
)

var Routers []Router

func AddRouter(router Router) {
	logtrace.LogWithFunctionName()
	Routers = append(Routers, router)
}

type Router interface {
	Register(fabricApi *operations.ZitiFabricAPI, wrapper RequestWrapper)
}
