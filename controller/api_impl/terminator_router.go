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
	"ztna-core/ztna/controller/rest_server/operations/terminator"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime/middleware"
)

func init() {
	logtrace.LogWithFunctionName()
	r := NewTerminatorRouter()
	AddRouter(r)
}

type TerminatorRouter struct {
	BasePath string
}

func NewTerminatorRouter() *TerminatorRouter {
	logtrace.LogWithFunctionName()
	return &TerminatorRouter{
		BasePath: "/" + EntityNameTerminator,
	}
}

func (r *TerminatorRouter) Register(fabricApi *operations.ZitiFabricAPI, wrapper RequestWrapper) {
	logtrace.LogWithFunctionName()
	fabricApi.TerminatorDeleteTerminatorHandler = terminator.DeleteTerminatorHandlerFunc(func(params terminator.DeleteTerminatorParams) middleware.Responder {
		return wrapper.WrapRequest(r.Delete, params.HTTPRequest, params.ID, "")
	})

	fabricApi.TerminatorDetailTerminatorHandler = terminator.DetailTerminatorHandlerFunc(func(params terminator.DetailTerminatorParams) middleware.Responder {
		return wrapper.WrapRequest(r.Detail, params.HTTPRequest, params.ID, "")
	})

	fabricApi.TerminatorListTerminatorsHandler = terminator.ListTerminatorsHandlerFunc(func(params terminator.ListTerminatorsParams) middleware.Responder {
		return wrapper.WrapRequest(r.List, params.HTTPRequest, "", "")
	})

	fabricApi.TerminatorUpdateTerminatorHandler = terminator.UpdateTerminatorHandlerFunc(func(params terminator.UpdateTerminatorParams) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.Update(n, rc, params) }, params.HTTPRequest, params.ID, "")
	})

	fabricApi.TerminatorCreateTerminatorHandler = terminator.CreateTerminatorHandlerFunc(func(params terminator.CreateTerminatorParams) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.Create(n, rc, params) }, params.HTTPRequest, "", "")
	})

	fabricApi.TerminatorPatchTerminatorHandler = terminator.PatchTerminatorHandlerFunc(func(params terminator.PatchTerminatorParams) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.Patch(n, rc, params) }, params.HTTPRequest, params.ID, "")
	})
}

func (r *TerminatorRouter) List(n *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	ListWithHandler[*model.Terminator](n, rc, n.Managers.Terminator, TerminatorModelMapper{})
}

func (r *TerminatorRouter) Detail(n *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	DetailWithHandler[*model.Terminator](n, rc, n.Managers.Terminator, TerminatorModelMapper{})
}

func (r *TerminatorRouter) Create(n *network.Network, rc api.RequestContext, params terminator.CreateTerminatorParams) {
	logtrace.LogWithFunctionName()
	Create(rc, TerminatorLinkFactory, func() (string, error) {
		entity := MapCreateTerminatorToModel(params.Terminator)
		err := n.Terminator.Create(entity, rc.NewChangeContext())
		if err != nil {
			return "", err
		}
		return entity.Id, nil
	})
}

func (r *TerminatorRouter) Delete(n *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	DeleteWithHandler(rc, n.Managers.Terminator)
}

func (r *TerminatorRouter) Update(n *network.Network, rc api.RequestContext, params terminator.UpdateTerminatorParams) {
	logtrace.LogWithFunctionName()
	Update(rc, func(id string) error {
		return n.Managers.Terminator.Update(MapUpdateTerminatorToModel(params.ID, params.Terminator), nil, rc.NewChangeContext())
	})
}

func (r *TerminatorRouter) Patch(n *network.Network, rc api.RequestContext, params terminator.PatchTerminatorParams) {
	logtrace.LogWithFunctionName()
	Patch(rc, func(id string, fields fields.UpdatedFields) error {
		return n.Managers.Terminator.Update(MapPatchTerminatorToModel(params.ID, params.Terminator), fields.FilterMaps("tags"), rc.NewChangeContext())
	})
}
