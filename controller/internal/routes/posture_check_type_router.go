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
	"ztna-core/edge-api/rest_management_api_server/operations/posture_checks"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/internal/permissions"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime/middleware"
)

func init() {
	logtrace.LogWithFunctionName()
	r := NewPostureCheckTypeRouter()
	env.AddRouter(r)
}

type PostureCheckTypeRouter struct {
	BasePath string
}

func NewPostureCheckTypeRouter() *PostureCheckTypeRouter {
	logtrace.LogWithFunctionName()
	return &PostureCheckTypeRouter{
		BasePath: "/" + EntityNamePostureCheckType,
	}
}

func (r *PostureCheckTypeRouter) Register(ae *env.AppEnv) {
	logtrace.LogWithFunctionName()

	ae.ManagementApi.PostureChecksDetailPostureCheckTypeHandler = posture_checks.DetailPostureCheckTypeHandlerFunc(func(params posture_checks.DetailPostureCheckTypeParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.Detail, params.HTTPRequest, params.ID, "", permissions.IsAdmin())
	})

	ae.ManagementApi.PostureChecksListPostureCheckTypesHandler = posture_checks.ListPostureCheckTypesHandlerFunc(func(params posture_checks.ListPostureCheckTypesParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.List, params.HTTPRequest, "", "", permissions.IsAdmin())
	})

}

func (r *PostureCheckTypeRouter) List(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	ListWithHandler[*model.PostureCheckType](ae, rc, ae.Managers.PostureCheckType, MapPostureCheckTypeToRestEntity)
}

func (r *PostureCheckTypeRouter) Detail(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	DetailWithHandler[*model.PostureCheckType](ae, rc, ae.Managers.PostureCheckType, MapPostureCheckTypeToRestEntity)
}
