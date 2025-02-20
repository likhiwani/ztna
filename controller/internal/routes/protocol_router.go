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
	"ztna-core/edge-api/rest_client_api_server/operations/informational"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/internal/permissions"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime/middleware"
)

func init() {
	logtrace.LogWithFunctionName()
	r := NewProtocolRouter()
	env.AddRouter(r)
}

type ProtocolRouter struct {
	BasePath string
}

func NewProtocolRouter() *ProtocolRouter {
	logtrace.LogWithFunctionName()
	return &ProtocolRouter{
		BasePath: "/Protocol",
	}
}

func (router *ProtocolRouter) Register(ae *env.AppEnv) {
	logtrace.LogWithFunctionName()
	ae.ClientApi.InformationalListProtocolsHandler = informational.ListProtocolsHandlerFunc(func(params informational.ListProtocolsParams) middleware.Responder {
		return ae.IsAllowed(router.List, params.HTTPRequest, "", "", permissions.Always())
	})
}

func (router *ProtocolRouter) List(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	data := rest_model.ListProtocols{
		"https": rest_model.Protocol{
			Address: &ae.GetConfig().Edge.Api.Address,
		},
	}
	rc.RespondWithOk(data, &rest_model.Meta{})
}
