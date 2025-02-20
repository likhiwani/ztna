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
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/controller/rest_model"
	"ztna-core/ztna/controller/rest_server/operations"
	"ztna-core/ztna/controller/rest_server/operations/link"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/storage/boltz"
)

func init() {
	logtrace.LogWithFunctionName()
	r := NewLinkRouter()
	AddRouter(r)
}

type LinkRouter struct {
	BasePath string
}

func NewLinkRouter() *LinkRouter {
	logtrace.LogWithFunctionName()
	return &LinkRouter{
		BasePath: "/" + EntityNameLink,
	}
}

func (r *LinkRouter) Register(fabricApi *operations.ZitiFabricAPI, wrapper RequestWrapper) {
	logtrace.LogWithFunctionName()
	fabricApi.LinkDetailLinkHandler = link.DetailLinkHandlerFunc(func(params link.DetailLinkParams) middleware.Responder {
		return wrapper.WrapRequest(r.Detail, params.HTTPRequest, params.ID, "")
	})

	fabricApi.LinkListLinksHandler = link.ListLinksHandlerFunc(func(params link.ListLinksParams) middleware.Responder {
		return wrapper.WrapRequest(r.ListLinks, params.HTTPRequest, "", "")
	})

	fabricApi.LinkPatchLinkHandler = link.PatchLinkHandlerFunc(func(params link.PatchLinkParams) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.Patch(n, rc, params) }, params.HTTPRequest, params.ID, "")
	})

	fabricApi.LinkDeleteLinkHandler = link.DeleteLinkHandlerFunc(func(params link.DeleteLinkParams) middleware.Responder {
		return wrapper.WrapRequest(func(n *network.Network, rc api.RequestContext) { r.Delete(n, rc) }, params.HTTPRequest, params.ID, "")
	})
}

func (r *LinkRouter) ListLinks(n *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	ListWithEnvelopeFactory(rc, defaultToListEnvelope, func(rc api.RequestContext, queryOptions *PublicQueryOptions) (*QueryResult, error) {
		query, err := queryOptions.getFullQuery(n.GetLinkStore())
		if err != nil {
			return nil, err
		}

		if query.GetLimit() == nil {
			query.SetLimit(10)
		}

		links, count, err := n.GetLinkStore().QueryEntitiesC(query)
		if err != nil {
			return nil, err
		}

		apiLinks := make([]*rest_model.LinkDetail, 0, len(links))
		for _, modelLink := range links {
			apiLink, err := MapLinkToRestModel(n, rc, modelLink)
			if err != nil {
				return nil, err
			}
			apiLinks = append(apiLinks, apiLink)
		}
		result := &QueryResult{
			Result:           apiLinks,
			Count:            count,
			Limit:            *query.GetLimit(),
			Offset:           *query.GetSkip(),
			FilterableFields: nil,
		}
		return result, nil
	})
}

func (r *LinkRouter) Detail(n *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	Detail(rc, func(rc api.RequestContext, id string) (interface{}, error) {
		l, found := n.GetLink(id)
		if !found {
			return nil, boltz.NewNotFoundError("link", "id", id)
		}
		apiLink, err := MapLinkToRestModel(n, rc, l)
		if err != nil {
			return nil, err
		}
		return apiLink, nil
	})
}

func (r *LinkRouter) Patch(n *network.Network, rc api.RequestContext, params link.PatchLinkParams) {
	logtrace.LogWithFunctionName()
	Patch(rc, func(id string, fields fields.UpdatedFields) error {
		l, found := n.GetLink(id)
		if !found {
			return boltz.NewNotFoundError("link", "id", id)
		}
		if fields.IsUpdated("staticCost") {
			l.SetStaticCost(int32(params.Link.StaticCost))
		}
		if fields.IsUpdated("down") {
			l.SetDown(params.Link.Down)
		}
		n.RerouteLink(l)
		return nil
	})
}

func (r *LinkRouter) Delete(network *network.Network, rc api.RequestContext) {
	logtrace.LogWithFunctionName()
	DeleteWithHandler(rc, DeleteHandlerF(func(id string, _ *change.Context) error {
		network.RemoveLink(id)
		return nil
	}))
}
