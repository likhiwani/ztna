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

package routes

import (
	"ztna-core/edge-api/rest_management_api_server/operations/terminator"
	"ztna-core/ztna/controller/api_impl"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/internal/permissions"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime/middleware"
)

func init() {
	logtrace.LogWithFunctionName()
	r := NewTerminatorRouter()
	env.AddRouter(r)
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

func (r *TerminatorRouter) Register(ae *env.AppEnv) {
	logtrace.LogWithFunctionName()
	ae.ManagementApi.TerminatorDeleteTerminatorHandler = terminator.DeleteTerminatorHandlerFunc(func(params terminator.DeleteTerminatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.Delete, params.HTTPRequest, params.ID, "", permissions.IsAdmin())
	})

	ae.ManagementApi.TerminatorDetailTerminatorHandler = terminator.DetailTerminatorHandlerFunc(func(params terminator.DetailTerminatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.Detail, params.HTTPRequest, params.ID, "", permissions.IsAdmin())
	})

	ae.ManagementApi.TerminatorListTerminatorsHandler = terminator.ListTerminatorsHandlerFunc(func(params terminator.ListTerminatorsParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.List, params.HTTPRequest, "", "", permissions.IsAdmin())
	})

	ae.ManagementApi.TerminatorUpdateTerminatorHandler = terminator.UpdateTerminatorHandlerFunc(func(params terminator.UpdateTerminatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.Update(ae, rc, params) }, params.HTTPRequest, params.ID, "", permissions.IsAdmin())
	})

	ae.ManagementApi.TerminatorCreateTerminatorHandler = terminator.CreateTerminatorHandlerFunc(func(params terminator.CreateTerminatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.Create(ae, rc, params) }, params.HTTPRequest, "", "", permissions.IsAdmin())
	})

	ae.ManagementApi.TerminatorPatchTerminatorHandler = terminator.PatchTerminatorHandlerFunc(func(params terminator.PatchTerminatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.Patch(ae, rc, params) }, params.HTTPRequest, params.ID, "", permissions.IsAdmin())
	})
}

func (r *TerminatorRouter) List(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	api_impl.ListWithHandler[*model.Terminator](ae.GetHostController().GetNetwork(), rc, ae.Managers.Terminator, TerminatorModelMapper{})
}

func (r *TerminatorRouter) Detail(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	api_impl.DetailWithHandler[*model.Terminator](ae.GetHostController().GetNetwork(), rc, ae.Managers.Terminator, TerminatorModelMapper{})
}

func (r *TerminatorRouter) Create(ae *env.AppEnv, rc *response.RequestContext, params terminator.CreateTerminatorParams) {
	logtrace.LogWithFunctionName()
	Create(rc, rc, TerminatorLinkFactory, func() (string, error) {
		entity := MapCreateTerminatorToModel(params.Terminator)
		err := ae.Managers.Terminator.Create(entity, rc.NewChangeContext())
		if err != nil {
			return "", err
		}
		return entity.Id, nil
	})
}

func (r *TerminatorRouter) Delete(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	DeleteWithHandler(rc, ae.Managers.Terminator)
}

func (r *TerminatorRouter) Update(ae *env.AppEnv, rc *response.RequestContext, params terminator.UpdateTerminatorParams) {
	logtrace.LogWithFunctionName()
	Update(rc, func(id string) error {
		return ae.Managers.Terminator.Update(MapUpdateTerminatorToModel(params.ID, params.Terminator), nil, rc.NewChangeContext())
	})
}

func (r *TerminatorRouter) Patch(ae *env.AppEnv, rc *response.RequestContext, params terminator.PatchTerminatorParams) {
	logtrace.LogWithFunctionName()
	Patch(rc, func(id string, fields fields.UpdatedFields) error {
		return ae.Managers.Terminator.Update(MapPatchTerminatorToModel(params.ID, params.Terminator), fields.FilterMaps("tags"), rc.NewChangeContext())
	})
}
