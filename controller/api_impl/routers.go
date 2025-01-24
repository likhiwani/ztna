package api_impl

import (
	"github.com/cosmic-cloak/ztna/controller/rest_server/operations"
)

var Routers []Router

func AddRouter(router Router) {
	Routers = append(Routers, router)
}

type Router interface {
	Register(fabricApi *operations.ZitiFabricAPI, wrapper RequestWrapper)
}
