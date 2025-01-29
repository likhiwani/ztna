package api_impl

import (
	"github.com/go-openapi/runtime/middleware"
	"ztna-core/ztna/controller/api"
	"ztna-core/ztna/controller/network"
	"net/http"
)

type RequestHandler func(network *network.Network, rc api.RequestContext)

type RequestWrapper interface {
	WrapRequest(handler RequestHandler, request *http.Request, entityId, entitySubId string) middleware.Responder
	WrapHttpHandler(handler http.Handler) http.Handler
	WrapWsHandler(handler http.Handler) http.Handler
}
