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
	"ztna-core/edge-api/rest_management_api_server/operations/auth_policy"
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
	r := NewAuthPolicyRouter()
	env.AddRouter(r)
}

type AuthPolicyRouter struct {
	BasePath string
}

func NewAuthPolicyRouter() *AuthPolicyRouter {
	logtrace.LogWithFunctionName()
	return &AuthPolicyRouter{
		BasePath: "/" + EntityNameAuthPolicy,
	}
}

func (r *AuthPolicyRouter) Register(ae *env.AppEnv) {
	logtrace.LogWithFunctionName()
	ae.ManagementApi.AuthPolicyDeleteAuthPolicyHandler = auth_policy.DeleteAuthPolicyHandlerFunc(func(params auth_policy.DeleteAuthPolicyParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.Delete, params.HTTPRequest, params.ID, "", permissions.IsAdmin())
	})

	ae.ManagementApi.AuthPolicyDetailAuthPolicyHandler = auth_policy.DetailAuthPolicyHandlerFunc(func(params auth_policy.DetailAuthPolicyParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.Detail, params.HTTPRequest, params.ID, "", permissions.IsAdmin())
	})

	ae.ManagementApi.AuthPolicyListAuthPoliciesHandler = auth_policy.ListAuthPoliciesHandlerFunc(func(params auth_policy.ListAuthPoliciesParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.List, params.HTTPRequest, "", "", permissions.IsAdmin())
	})

	ae.ManagementApi.AuthPolicyUpdateAuthPolicyHandler = auth_policy.UpdateAuthPolicyHandlerFunc(func(params auth_policy.UpdateAuthPolicyParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.Update(ae, rc, params) }, params.HTTPRequest, params.ID, "", permissions.IsAdmin())
	})

	ae.ManagementApi.AuthPolicyCreateAuthPolicyHandler = auth_policy.CreateAuthPolicyHandlerFunc(func(params auth_policy.CreateAuthPolicyParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.Create(ae, rc, params) }, params.HTTPRequest, "", "", permissions.IsAdmin())
	})

	ae.ManagementApi.AuthPolicyPatchAuthPolicyHandler = auth_policy.PatchAuthPolicyHandlerFunc(func(params auth_policy.PatchAuthPolicyParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.Patch(ae, rc, params) }, params.HTTPRequest, params.ID, "", permissions.IsAdmin())
	})
}

func (r *AuthPolicyRouter) List(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	ListWithHandler[*model.AuthPolicy](ae, rc, ae.Managers.AuthPolicy, MapAuthPolicyToRestEntity)
}

func (r *AuthPolicyRouter) Detail(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	DetailWithHandler[*model.AuthPolicy](ae, rc, ae.Managers.AuthPolicy, MapAuthPolicyToRestEntity)
}

func (r *AuthPolicyRouter) Create(ae *env.AppEnv, rc *response.RequestContext, params auth_policy.CreateAuthPolicyParams) {
	logtrace.LogWithFunctionName()
	Create(rc, rc, AuthPolicyLinkFactory, func() (string, error) {
		return MapCreate(ae.Managers.AuthPolicy.Create, MapCreateAuthPolicyToModel(params.AuthPolicy), rc)
	})
}

func (r *AuthPolicyRouter) Delete(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	DeleteWithHandler(rc, ae.Managers.AuthPolicy)
}

func (r *AuthPolicyRouter) Update(ae *env.AppEnv, rc *response.RequestContext, params auth_policy.UpdateAuthPolicyParams) {
	logtrace.LogWithFunctionName()
	Update(rc, func(id string) error {
		return ae.Managers.AuthPolicy.Update(MapUpdateAuthPolicyToModel(params.ID, params.AuthPolicy), nil, rc.NewChangeContext())
	})
}

func (r *AuthPolicyRouter) Patch(ae *env.AppEnv, rc *response.RequestContext, params auth_policy.PatchAuthPolicyParams) {
	logtrace.LogWithFunctionName()
	Patch(rc, func(id string, fields fields.UpdatedFields) error {
		return ae.Managers.AuthPolicy.Update(MapPatchAuthPolicyToModel(params.ID, params.AuthPolicy), fields.FilterMaps("tags"), rc.NewChangeContext())
	})
}
