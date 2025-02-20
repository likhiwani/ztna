/*
	Copyright NetFoundry Inc.

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

package api_impl

import (
	"ztna-core/ztna/controller/api"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/controller/rest_server/operations"
	"ztna-core/ztna/controller/rest_server/operations/router"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime/middleware"
)

func init() {
	logtrace.LogWithFunctionName()
	r := NewRouterRouter()
	AddRouter(r)
}

type RouterRouter struct {
	BasePath string
}

func NewRouterRouter() *RouterRouter {
	logtrace.LogWithFunctionName()
	return &RouterRouter{
		BasePath: "/" + EntityNameRouter,
	}
}

func (r *RouterRouter) Register(fabricApi *operations.ZitiFabricAPI, wrapper RequestWrapper) {
	logtrace.LogWithFunctionName()
	fabricApi.RouterDeleteRouterHandler = router.DeleteRouterHandlerFunc(func(params router.DeleteRouterParams) middleware.Responder {
		return wrapper.WrapRequest(r.Delete, params.HTTPRequest, params.ID, "")
	})

	fabricApi.RouterDetailRouterHandler = router.DetailRouterHandlerFunc(func(params router.DetailRouterParams) middleware.Responder {
		return wrapper.WrapRequest(r.Detail, params.HTTPRequest, params.ID, "")
	})

	fabricApi.RouterListRoutersHandler = router.ListRoutersHandlerFunc(func(params router.ListRoutersParams) middleware.Responder {
		return wrapper.WrapRequest(r.ListRouters, params.HTTPRequest, "", "")
	})

	fabricApi.RouterCreateRouterHandler = router.CreateRouterHandlerFunc(func(params router.CreateRouterParams) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.Create(n, rc, params) }, params.HTTPRequest, "", "")
	})

	fabricApi.RouterUpdateRouterHandler = router.UpdateRouterHandlerFunc(func(params router.UpdateRouterParams) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.Update(n, rc, params) }, params.HTTPRequest, params.ID, "")
	})

	fabricApi.RouterPatchRouterHandler = router.PatchRouterHandlerFunc(func(params router.PatchRouterParams) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.Patch(n, rc, params) }, params.HTTPRequest, params.ID, "")
	})

	fabricApi.RouterListRouterTerminatorsHandler = router.ListRouterTerminatorsHandlerFunc(func(params router.ListRouterTerminatorsParams) middleware.Responder {
		return wrapper.WrapRequest(r.listManagementTerminators, params.HTTPRequest, params.ID, "")
	})
}

func (r *RouterRouter) ListRouters(n *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	ListWithHandler[*model.Router](n, rc, n.Managers.Router, RouterModelMapper{})
}

func (r *RouterRouter) Detail(n *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	DetailWithHandler[*model.Router](n, rc, n.Managers.Router, RouterModelMapper{})
}

func (r *RouterRouter) Create(n *network.Network, rc api.RequestContext, params router.CreateRouterParams) {
	logtrace.LogWithFunctionName()
	Create(rc, RouterLinkFactory, func() (string, error) {
		router := MapCreateRouterToModel(params.Router)
		err := n.Router.Create(router, rc.NewChangeContext())
		if err != nil {
			return "", err
		}
		return router.Id, nil
	})
}

func (r *RouterRouter) Delete(network *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	DeleteWithHandler(rc, network.Managers.Router)
}

func (r *RouterRouter) Update(n *network.Network, rc api.RequestContext, params router.UpdateRouterParams) {
	logtrace.LogWithFunctionName()
	Update(rc, func(id string) error {
		return n.Managers.Router.Update(MapUpdateRouterToModel(params.ID, params.Router), nil, rc.NewChangeContext())
	})
}

func (r *RouterRouter) Patch(n *network.Network, rc api.RequestContext, params router.PatchRouterParams) {
	logtrace.LogWithFunctionName()
	Patch(rc, func(id string, fields fields.UpdatedFields) error {
		return n.Managers.Router.Update(MapPatchRouterToModel(params.ID, params.Router), fields.FilterMaps("tags"), rc.NewChangeContext())
	})
}

func (r *RouterRouter) listManagementTerminators(n *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	ListAssociationWithHandler[*model.Router, *model.Terminator](n, rc, n.Managers.Router, n.Managers.Terminator, TerminatorModelMapper{})
}
