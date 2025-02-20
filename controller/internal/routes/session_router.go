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
	"net/http"
	"time"
	"ztna-core/edge-api/rest_client_api_server/operations/session"
	clientSession "ztna-core/edge-api/rest_client_api_server/operations/session"
	managementSession "ztna-core/edge-api/rest_management_api_server/operations/session"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/internal/permissions"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/metrics"
)

func init() {
	logtrace.LogWithFunctionName()
	r := NewSessionRouter()
	env.AddRouter(r)
}

type SessionRouter struct {
	BasePath    string
	createTimer metrics.Timer
}

func NewSessionRouter() *SessionRouter {
	logtrace.LogWithFunctionName()
	return &SessionRouter{
		BasePath: "/" + EntityNameSession,
	}
}

func (r *SessionRouter) Register(ae *env.AppEnv) {
	logtrace.LogWithFunctionName()
	r.createTimer = ae.GetHostController().GetNetwork().GetMetricsRegistry().Timer("session.create")

	//Management
	ae.ManagementApi.SessionDeleteSessionHandler = managementSession.DeleteSessionHandlerFunc(func(params managementSession.DeleteSessionParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.Delete, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})

	ae.ManagementApi.SessionDetailSessionHandler = managementSession.DetailSessionHandlerFunc(func(params managementSession.DetailSessionParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.Detail, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})

	ae.ManagementApi.SessionListSessionsHandler = managementSession.ListSessionsHandlerFunc(func(params managementSession.ListSessionsParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.List, params.HTTPRequest, "", "", permissions.IsAuthenticated())
	})

	ae.ManagementApi.SessionDetailSessionRoutePathHandler = managementSession.DetailSessionRoutePathHandlerFunc(func(params managementSession.DetailSessionRoutePathParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.DetailRoutePath(ae, rc, params) }, params.HTTPRequest, "", "", permissions.IsAuthenticated())
	})

	//Client
	ae.ClientApi.SessionDeleteSessionHandler = clientSession.DeleteSessionHandlerFunc(func(params clientSession.DeleteSessionParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.Delete, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})

	ae.ClientApi.SessionDetailSessionHandler = clientSession.DetailSessionHandlerFunc(func(params clientSession.DetailSessionParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.Detail, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})

	ae.ClientApi.SessionListSessionsHandler = clientSession.ListSessionsHandlerFunc(func(params clientSession.ListSessionsParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.List, params.HTTPRequest, "", "", permissions.IsAuthenticated())
	})

	ae.ClientApi.SessionCreateSessionHandler = clientSession.CreateSessionHandlerFunc(func(params clientSession.CreateSessionParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.Create(ae, rc, params) }, params.HTTPRequest, "", "", permissions.IsAuthenticated())
	})

}

func (r *SessionRouter) List(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	// ListWithHandler won't do search limiting by logged in user
	List(rc, func(rc *response.RequestContext, queryOptions *PublicQueryOptions) (*QueryResult, error) {
		query, err := queryOptions.getFullQuery(ae.Managers.Session.GetStore())
		if err != nil {
			return nil, err
		}

		result, err := ae.Managers.Session.PublicQueryForIdentity(rc.Identity, query)
		if err != nil {
			return nil, err
		}
		sessions, err := MapSessionsToRestEntities(ae, rc, result.Sessions)
		if err != nil {
			return nil, err
		}
		return NewQueryResult(sessions, &result.QueryMetaData), nil
	})
}

func (r *SessionRouter) Detail(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	// DetailWithHandler won't do search limiting by logged in user
	Detail(rc, func(rc *response.RequestContext, id string) (interface{}, error) {
		service, err := ae.Managers.Session.ReadForIdentity(id, rc.ApiSession.IdentityId)
		if err != nil {
			return nil, err
		}
		return MapSessionToRestEntity(ae, rc, service)
	})
}

func (r *SessionRouter) Delete(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	Delete(rc, func(rc *response.RequestContext, id string) error {
		return ae.Managers.Session.DeleteForIdentity(id, rc.ApiSession.IdentityId, rc.NewChangeContext())
	})
}

func (r *SessionRouter) Create(ae *env.AppEnv, rc *response.RequestContext, params session.CreateSessionParams) {
	logtrace.LogWithFunctionName()
	start := time.Now()
	if rc.Claims != nil {
		request := MapCreateSessionToModel(rc.Claims.Subject, rc.Claims.ApiSessionId, params.Session)
		token, err := ae.Managers.Session.CreateJwt(request, rc.NewChangeContext())

		if err != nil {
			rc.RespondWithError(err)
			return
		}

		edgeRouters, err := getSessionEdgeRouters(ae, request)
		if err != nil {
			pfxlog.Logger().Errorf("could not render edge routers for jwt session: %s", err)
		}

		sessionType := rest_model.DialBind(request.Type)

		newSessionEnvelope := &rest_model.SessionCreateEnvelope{
			Data: &rest_model.SessionDetail{
				BaseEntity: rest_model.BaseEntity{
					ID: &request.Id,
				},
				APISessionID: &rc.ApiSession.Id,
				EdgeRouters:  edgeRouters,
				IdentityID:   &rc.Identity.Id,
				ServiceID:    &request.ServiceId,
				Token:        &token,
				Type:         &sessionType,
			},
			Meta: &rest_model.Meta{},
		}

		rc.Respond(newSessionEnvelope, http.StatusCreated)
	} else {
		responder := &SessionRequestResponder{ae: ae, Responder: rc}
		CreateWithResponder(rc, responder, SessionLinkFactory, func() (string, error) {
			return ae.Managers.Session.Create(MapCreateSessionToModel(rc.Identity.Id, rc.ApiSession.Id, params.Session), rc.NewChangeContext())
		})
	}

	r.createTimer.UpdateSince(start)
}

func (r *SessionRouter) DetailRoutePath(ae *env.AppEnv, rc *response.RequestContext, params managementSession.DetailSessionRoutePathParams) {
	logtrace.LogWithFunctionName()
	path := []string{} //must be non null

	for _, circuit := range ae.HostController.GetNetwork().GetAllCircuits() {
		if circuit.ClientId == params.ID {
			if circuit.Path != nil {
				for _, pathSeg := range circuit.Path.Nodes {
					path = append(path, pathSeg.Id)
				}
				break
			}
		}
	}

	routePath := &rest_model.SessionRoutePathDetail{
		RoutePath: path,
	}

	rc.RespondWithOk(routePath, &rest_model.Meta{})
}
