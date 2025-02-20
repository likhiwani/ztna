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
	controllersClient "ztna-core/edge-api/rest_client_api_server/operations/controllers"
	controllersMan "ztna-core/edge-api/rest_management_api_server/operations/controllers"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/internal/permissions"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime/middleware"
)

func init() {
	logtrace.LogWithFunctionName()
	r := NewControllerRouter()
	env.AddRouter(r)
}

type ControllerRouter struct {
	BasePath string
}

func NewControllerRouter() *ControllerRouter {
	logtrace.LogWithFunctionName()
	return &ControllerRouter{
		BasePath: "/" + EntityNameController,
	}
}

func (r *ControllerRouter) Register(ae *env.AppEnv) {
	logtrace.LogWithFunctionName()
	ae.ManagementApi.ControllersListControllersHandler = controllersMan.ListControllersHandlerFunc(func(params controllersMan.ListControllersParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.ListManagement, params.HTTPRequest, "", "", permissions.IsAuthenticated())
	})

	ae.ClientApi.ControllersListControllersHandler = controllersClient.ListControllersHandlerFunc(func(params controllersClient.ListControllersParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.ListClient, params.HTTPRequest, "", "", permissions.IsAuthenticated())
	})
}

func (r *ControllerRouter) ListManagement(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	ListWithHandler[*model.Controller](ae, rc, ae.Managers.Controller, MapControllerToManagementRestEntity)
}

func (r *ControllerRouter) ListClient(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	ListWithHandler[*model.Controller](ae, rc, ae.Managers.Controller, MapControllerToClientRestEntity)
}
