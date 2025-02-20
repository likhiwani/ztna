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
	"ztna-core/ztna/controller/rest_server/operations/service"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime/middleware"
)

func init() {
	logtrace.LogWithFunctionName()
	r := NewServiceRouter()
	AddRouter(r)
}

type ServiceRouter struct {
	BasePath string
}

func NewServiceRouter() *ServiceRouter {
	logtrace.LogWithFunctionName()
	return &ServiceRouter{
		BasePath: "/" + EntityNameService,
	}
}

func (r *ServiceRouter) Register(fabricApi *operations.ZitiFabricAPI, wrapper RequestWrapper) {
	logtrace.LogWithFunctionName()
	fabricApi.ServiceDeleteServiceHandler = service.DeleteServiceHandlerFunc(func(params service.DeleteServiceParams) middleware.Responder {
		return wrapper.WrapRequest(r.Delete, params.HTTPRequest, params.ID, "")
	})

	fabricApi.ServiceDetailServiceHandler = service.DetailServiceHandlerFunc(func(params service.DetailServiceParams) middleware.Responder {
		return wrapper.WrapRequest(r.Detail, params.HTTPRequest, params.ID, "")
	})

	fabricApi.ServiceListServicesHandler = service.ListServicesHandlerFunc(func(params service.ListServicesParams) middleware.Responder {
		return wrapper.WrapRequest(r.ListServices, params.HTTPRequest, "", "")
	})

	fabricApi.ServiceUpdateServiceHandler = service.UpdateServiceHandlerFunc(func(params service.UpdateServiceParams) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.Update(n, rc, params) }, params.HTTPRequest, params.ID, "")
	})

	fabricApi.ServiceCreateServiceHandler = service.CreateServiceHandlerFunc(func(params service.CreateServiceParams) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.Create(n, rc, params) }, params.HTTPRequest, "", "")
	})

	fabricApi.ServicePatchServiceHandler = service.PatchServiceHandlerFunc(func(params service.PatchServiceParams) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.Patch(n, rc, params) }, params.HTTPRequest, params.ID, "")
	})

	fabricApi.ServiceListServiceTerminatorsHandler = service.ListServiceTerminatorsHandlerFunc(func(params service.ListServiceTerminatorsParams) middleware.Responder {
		return wrapper.WrapRequest(r.listManagementTerminators, params.HTTPRequest, params.ID, "")
	})
}

func (r *ServiceRouter) ListServices(n *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	ListWithHandler[*model.Service](n, rc, n.Managers.Service, ServiceModelMapper{})
}

func (r *ServiceRouter) Detail(n *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	DetailWithHandler[*model.Service](n, rc, n.Managers.Service, ServiceModelMapper{})
}

func (r *ServiceRouter) Create(n *network.Network, rc api.RequestContext, params service.CreateServiceParams) {
	logtrace.LogWithFunctionName()
	Create(rc, ServiceLinkFactory, func() (string, error) {
		svc := MapCreateServiceToModel(params.Service)
		err := n.Service.Create(svc, rc.NewChangeContext())
		if err != nil {
			return "", err
		}
		return svc.Id, nil
	})
}

func (r *ServiceRouter) Delete(network *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	DeleteWithHandler(rc, network.Managers.Service)
}

func (r *ServiceRouter) Update(n *network.Network, rc api.RequestContext, params service.UpdateServiceParams) {
	logtrace.LogWithFunctionName()
	Update(rc, func(id string) error {
		return n.Managers.Service.Update(MapUpdateServiceToModel(params.ID, params.Service), nil, rc.NewChangeContext())
	})
}

func (r *ServiceRouter) Patch(n *network.Network, rc api.RequestContext, params service.PatchServiceParams) {
	logtrace.LogWithFunctionName()
	Patch(rc, func(id string, fields fields.UpdatedFields) error {
		return n.Managers.Service.Update(MapPatchServiceToModel(params.ID, params.Service), fields.FilterMaps("tags"), rc.NewChangeContext())
	})
}

func (r *ServiceRouter) listManagementTerminators(n *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	ListAssociationWithHandler[*model.Service, *model.Terminator](n, rc, n.Managers.Service, n.Managers.Terminator, TerminatorModelMapper{})
}
