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
	"crypto/x509"
	clientCurrentApiSession "ztna-core/edge-api/rest_client_api_server/operations/current_api_session"
	managementCurrentApiSession "ztna-core/edge-api/rest_management_api_server/operations/current_api_session"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/internal/permissions"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime/middleware"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/errorz"
	"github.com/openziti/storage/boltz"
)

func init() {
	logtrace.LogWithFunctionName()
	r := NewCurrentIdentityAuthenticatorRouter()
	env.AddRouter(r)
}

type CurrentIdentityAuthenticatorRouter struct {
	BasePath string
}

func NewCurrentIdentityAuthenticatorRouter() *CurrentIdentityAuthenticatorRouter {
	logtrace.LogWithFunctionName()
	return &CurrentIdentityAuthenticatorRouter{
		BasePath: "/" + EntityNameAuthenticator,
	}
}

func (r *CurrentIdentityAuthenticatorRouter) Register(ae *env.AppEnv) {
	logtrace.LogWithFunctionName()
	//Client

	ae.ClientApi.CurrentAPISessionDetailCurrentIdentityAuthenticatorHandler = clientCurrentApiSession.DetailCurrentIdentityAuthenticatorHandlerFunc(func(params clientCurrentApiSession.DetailCurrentIdentityAuthenticatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.Detail, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})

	ae.ClientApi.CurrentAPISessionListCurrentIdentityAuthenticatorsHandler = clientCurrentApiSession.ListCurrentIdentityAuthenticatorsHandlerFunc(func(params clientCurrentApiSession.ListCurrentIdentityAuthenticatorsParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.List, params.HTTPRequest, "", "", permissions.IsAuthenticated())
	})

	ae.ClientApi.CurrentAPISessionUpdateCurrentIdentityAuthenticatorHandler = clientCurrentApiSession.UpdateCurrentIdentityAuthenticatorHandlerFunc(func(params clientCurrentApiSession.UpdateCurrentIdentityAuthenticatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.Update(ae, rc, params.Authenticator) }, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})

	ae.ClientApi.CurrentAPISessionPatchCurrentIdentityAuthenticatorHandler = clientCurrentApiSession.PatchCurrentIdentityAuthenticatorHandlerFunc(func(params clientCurrentApiSession.PatchCurrentIdentityAuthenticatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.Patch(ae, rc, params.Authenticator) }, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})

	ae.ClientApi.CurrentAPISessionExtendCurrentIdentityAuthenticatorHandler = clientCurrentApiSession.ExtendCurrentIdentityAuthenticatorHandlerFunc(func(params clientCurrentApiSession.ExtendCurrentIdentityAuthenticatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.Extend(ae, rc, params.Extend) }, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})

	ae.ClientApi.CurrentAPISessionExtendVerifyCurrentIdentityAuthenticatorHandler = clientCurrentApiSession.ExtendVerifyCurrentIdentityAuthenticatorHandlerFunc(func(params clientCurrentApiSession.ExtendVerifyCurrentIdentityAuthenticatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.ExtendVerify(ae, rc, params.Extend) }, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})

	//Management

	ae.ManagementApi.CurrentAPISessionDetailCurrentIdentityAuthenticatorHandler = managementCurrentApiSession.DetailCurrentIdentityAuthenticatorHandlerFunc(func(params managementCurrentApiSession.DetailCurrentIdentityAuthenticatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.Detail, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})

	ae.ManagementApi.CurrentAPISessionListCurrentIdentityAuthenticatorsHandler = managementCurrentApiSession.ListCurrentIdentityAuthenticatorsHandlerFunc(func(params managementCurrentApiSession.ListCurrentIdentityAuthenticatorsParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(r.List, params.HTTPRequest, "", "", permissions.IsAuthenticated())
	})

	ae.ManagementApi.CurrentAPISessionUpdateCurrentIdentityAuthenticatorHandler = managementCurrentApiSession.UpdateCurrentIdentityAuthenticatorHandlerFunc(func(params managementCurrentApiSession.UpdateCurrentIdentityAuthenticatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.Update(ae, rc, params.Authenticator) }, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})

	ae.ManagementApi.CurrentAPISessionPatchCurrentIdentityAuthenticatorHandler = managementCurrentApiSession.PatchCurrentIdentityAuthenticatorHandlerFunc(func(params managementCurrentApiSession.PatchCurrentIdentityAuthenticatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.Patch(ae, rc, params.Authenticator) }, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})

	ae.ManagementApi.CurrentAPISessionExtendCurrentIdentityAuthenticatorHandler = managementCurrentApiSession.ExtendCurrentIdentityAuthenticatorHandlerFunc(func(params managementCurrentApiSession.ExtendCurrentIdentityAuthenticatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.Extend(ae, rc, params.Extend) }, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})

	ae.ManagementApi.CurrentAPISessionExtendVerifyCurrentIdentityAuthenticatorHandler = managementCurrentApiSession.ExtendVerifyCurrentIdentityAuthenticatorHandlerFunc(func(params managementCurrentApiSession.ExtendVerifyCurrentIdentityAuthenticatorParams, _ interface{}) middleware.Responder {
		return ae.IsAllowed(func(ae *env.AppEnv, rc *response.RequestContext) { r.ExtendVerify(ae, rc, params.Extend) }, params.HTTPRequest, params.ID, "", permissions.IsAuthenticated())
	})
}

func (r *CurrentIdentityAuthenticatorRouter) List(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	List(rc, func(rc *response.RequestContext, queryOptions *PublicQueryOptions) (*QueryResult, error) {
		query, err := queryOptions.getFullQuery(ae.Managers.Authenticator.GetStore())
		if err != nil {
			return nil, err
		}

		result, err := ae.Managers.Authenticator.ListForIdentity(rc.Identity.Id, query)
		if err != nil {
			pfxlog.Logger().Errorf("error executing list query: %+v", err)
			return nil, err
		}

		apiAuthenticators, err := MapAuthenticatorsToRestEntities(ae, rc, result.Entities)
		if err != nil {
			return nil, err
		}
		for i, authenticator := range apiAuthenticators {
			authenticator.Links = CurrentIdentityAuthenticatorLinkFactory.Links(result.Entities[i])
		}

		return NewQueryResult(apiAuthenticators, result.GetMetaData()), nil
	})
}

func (r *CurrentIdentityAuthenticatorRouter) Detail(ae *env.AppEnv, rc *response.RequestContext) {
	logtrace.LogWithFunctionName()
	Detail(rc, func(rc *response.RequestContext, id string) (entity interface{}, err error) {
		authenticator, err := ae.GetManagers().Authenticator.ReadForIdentity(rc.Identity.Id, id)
		if err != nil {
			return nil, err
		}

		if authenticator == nil {
			return nil, boltz.NewNotFoundError(ae.GetManagers().Authenticator.GetStore().GetSingularEntityType(), "id", id)
		}

		apiAuthenticator, err := MapAuthenticatorToRestModel(ae, authenticator)

		if err != nil {
			return nil, err
		}

		apiAuthenticator.Links = CurrentIdentityAuthenticatorLinkFactory.Links(authenticator)

		return apiAuthenticator, nil
	})
}

func (r *CurrentIdentityAuthenticatorRouter) Update(ae *env.AppEnv, rc *response.RequestContext, authenticator *rest_model.AuthenticatorUpdateWithCurrent) {
	logtrace.LogWithFunctionName()
	Update(rc, func(id string) error {
		return ae.Managers.Authenticator.UpdateSelf(MapUpdateAuthenticatorWithCurrentToModel(id, rc.Identity.Id, authenticator), rc.NewChangeContext())
	})
}

func (r *CurrentIdentityAuthenticatorRouter) Patch(ae *env.AppEnv, rc *response.RequestContext, authenticator *rest_model.AuthenticatorPatchWithCurrent) {
	logtrace.LogWithFunctionName()
	Patch(rc, func(id string, fields fields.UpdatedFields) error {
		return ae.Managers.Authenticator.PatchSelf(MapPatchAuthenticatorWithCurrentToModel(id, rc.Identity.Id, authenticator), fields.FilterMaps("tags"), rc.NewChangeContext())
	})
}

func (r *CurrentIdentityAuthenticatorRouter) Extend(ae *env.AppEnv, rc *response.RequestContext, extend *rest_model.IdentityExtendEnrollmentRequest) {
	logtrace.LogWithFunctionName()
	peerCerts := rc.Request.TLS.PeerCertificates

	if len(peerCerts) == 0 {
		rc.RespondWithApiError(errorz.NewUnauthorized())
		return
	}

	var cert *x509.Certificate
	for _, peerCert := range peerCerts {
		if !peerCert.IsCA {
			cert = peerCert
		}
	}

	if cert == nil {
		rc.RespondWithApiError(errorz.NewUnauthorized())
		return
	}

	fingerprint := ae.GetFingerprintGenerator().FromCert(cert)

	if fingerprint == "" {
		rc.RespondWithApiError(errorz.NewUnauthorized())
		return
	}

	if extend.ClientCertCsr == nil {
		rc.RespondWithError(errorz.NewFieldApiError(&errorz.FieldError{
			Reason:     "client CSR is required",
			FieldName:  "certCsr",
			FieldValue: extend.ClientCertCsr,
		}))
		return
	}
	authId, err := rc.GetEntityId()

	if err != nil {
		rc.RespondWithError(errorz.NewFieldApiError(&errorz.FieldError{
			Reason:     "id is required",
			FieldName:  "id",
			FieldValue: "",
		}))
		return
	}

	certPem, err := ae.Managers.Authenticator.ExtendCertForIdentity(rc.Identity.Id, authId, peerCerts, *extend.ClientCertCsr, rc.NewChangeContext())

	if err != nil {
		rc.RespondWithError(err)
		return
	}

	rc.RespondWithOk(&rest_model.IdentityExtendCerts{
		Ca:         string(ae.GetConfig().Edge.CaPems()),
		ClientCert: string(certPem),
	}, &rest_model.Meta{})
}

func (r *CurrentIdentityAuthenticatorRouter) ExtendVerify(ae *env.AppEnv, rc *response.RequestContext, extend *rest_model.IdentityExtendValidateEnrollmentRequest) {
	logtrace.LogWithFunctionName()
	authId, err := rc.GetEntityId()
	if err != nil {
		rc.RespondWithError(err)
		return
	}

	err = ae.Managers.Authenticator.VerifyExtendCertForIdentity(rc.ApiSession.Id, rc.Identity.Id, authId, *extend.ClientCert, rc.NewChangeContext())

	if err != nil {
		rc.RespondWithError(err)
		return
	}

	rc.RespondWithEmptyOk()
}
