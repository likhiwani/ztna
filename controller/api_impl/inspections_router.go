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
	"net/http"
	"ztna-core/ztna/controller/api"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/controller/rest_model"
	"ztna-core/ztna/controller/rest_server/operations"
	"ztna-core/ztna/controller/rest_server/operations/inspect"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/foundation/v2/stringz"
)

func init() {
	logtrace.LogWithFunctionName()
	r := NewInspectRouter()
	AddRouter(r)
}

type InspectRouter struct {
	BasePath string
}

func NewInspectRouter() *InspectRouter {
	logtrace.LogWithFunctionName()
	return &InspectRouter{
		BasePath: "/" + EntityNameInspect,
	}
}

func (r *InspectRouter) Register(fabricApi *operations.ZitiFabricAPI, wrapper RequestWrapper) {
	logtrace.LogWithFunctionName()
	fabricApi.InspectInspectHandler = inspect.InspectHandlerFunc(func(params inspect.InspectParams) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.Inspect(n, rc, params.Request) }, params.HTTPRequest, "", "")
	})
}

func (r *InspectRouter) Inspect(n *network.Network, rc api.RequestContext, request *rest_model.InspectRequest) {
	logtrace.LogWithFunctionName()
	result := n.Inspections.Inspect(stringz.OrEmpty(request.AppRegex), request.RequestedValues)
	resp := MapInspectResultToRestModel(n, result)
	rc.Respond(resp, http.StatusOK)
}
