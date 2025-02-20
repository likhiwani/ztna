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
	"ztna-core/edge-api/rest_management_api_server/operations/role_attributes"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/internal/permissions"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime/middleware"
)

func init() {
	logtrace.LogWithFunctionName()
	r := NewRoleAttributesRouter()
	env.AddRouter(r)
}

type RoleAttributesRouter struct{}

func NewRoleAttributesRouter() *RoleAttributesRouter {
	logtrace.LogWithFunctionName()
	return &RoleAttributesRouter{}
}

func (r *RoleAttributesRouter) Register(ae *env.AppEnv) {
	logtrace.LogWithFunctionName()
	ae.ManagementApi.RoleAttributesListEdgeRouterRoleAttributesHandler = role_attributes.ListEdgeRouterRoleAttributesHandlerFunc(func(params role_attributes.ListEdgeRouterRoleAttributesParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.listEdgeRouterRoleAttributes, params.HTTPRequest, "", "", permissions.IsAdmin())
	})

	ae.ManagementApi.RoleAttributesListIdentityRoleAttributesHandler = role_attributes.ListIdentityRoleAttributesHandlerFunc(func(params role_attributes.ListIdentityRoleAttributesParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.listIdentityRoleAttributes, params.HTTPRequest, "", "", permissions.IsAdmin())
	})

	ae.ManagementApi.RoleAttributesListServiceRoleAttributesHandler = role_attributes.ListServiceRoleAttributesHandlerFunc(func(params role_attributes.ListServiceRoleAttributesParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.listServiceRoleAttributes, params.HTTPRequest, "", "", permissions.IsAdmin())
	})

	ae.ManagementApi.RoleAttributesListPostureCheckRoleAttributesHandler = role_attributes.ListPostureCheckRoleAttributesHandlerFunc(func(params role_attributes.ListPostureCheckRoleAttributesParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.listPostureCheckAttributes, params.HTTPRequest, "", "", permissions.IsAdmin())
	})
}

func (r *RoleAttributesRouter) listEdgeRouterRoleAttributes(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	r.listRoleAttributes(rc, ae.Managers.EdgeRouter)
}

func (r *RoleAttributesRouter) listIdentityRoleAttributes(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	r.listRoleAttributes(rc, ae.Managers.Identity)
}

func (r *RoleAttributesRouter) listServiceRoleAttributes(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	r.listRoleAttributes(rc, ae.Managers.EdgeService)
}

func (r *RoleAttributesRouter) listPostureCheckAttributes(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	r.listRoleAttributes(rc, ae.Managers.PostureCheck)
}

func (r *RoleAttributesRouter) listRoleAttributes(rc *response.RequestContext, queryable roleAttributeQueryable) {
	logtrace.LogWithFunctionName()
	List(rc, func(rc *response.RequestContext, queryOptions *PublicQueryOptions) (*QueryResult, error) {
		results, qmd, err := queryable.QueryRoleAttributes(queryOptions.Predicate)
		if err != nil {
			return nil, err
		}

		var list rest_model.RoleAttributesList

		for _, result := range results {
			list = append(list, result)
		}

		return NewQueryResult(list, qmd), nil
	})
}

type roleAttributeQueryable interface {
	QueryRoleAttributes(queryString string) ([]string, *models.QueryMetaData, error)
}
