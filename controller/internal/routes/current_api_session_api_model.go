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
	"time"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/strfmt"
	"github.com/michaelquigley/pfxlog"
)

const EntityNameCurrentSession = "current-api-session"
const EntityNameCurrentSessionCertificates = "certificates"

var CurrentApiSessionCertificateLinkFactory *BasicLinkFactory = NewBasicLinkFactory(EntityNameCurrentSession + "/" + EntityNameCurrentSessionCertificates)

var CurrentApiSessionLinkFactory LinksFactory = NewCurrentApiSessionLinkFactory()

type CurrentApiSessionLinkFactoryImpl struct {
	BasicLinkFactory
}

func NewCurrentApiSessionLinkFactory() *CurrentApiSessionLinkFactoryImpl {
	logtrace.LogWithFunctionName()
	return &CurrentApiSessionLinkFactoryImpl{
		BasicLinkFactory: *NewBasicLinkFactory(EntityNameCurrentSession),
	}
}

func (factory *CurrentApiSessionLinkFactoryImpl) SelfLink(entity models.Entity) rest_model.Link {
	logtrace.LogWithFunctionName()
	return NewLink("./" + EntityNameCurrentSession)
}

func (factory *CurrentApiSessionLinkFactoryImpl) Links(entity models.Entity) rest_model.Links {
	logtrace.LogWithFunctionName()
	return rest_model.Links{
		EntityNameSelf:            factory.SelfLink(entity),
		EntityNameCurrentIdentity: CurrentIdentityLinkFactory.SelfLink(entity),
	}
}

func MapToCurrentApiSessionRestModel(ae *env.AppEnv, rc *response.RequestContext, sessionTimeout time.Duration) *rest_model.CurrentAPISessionDetail {
	logtrace.LogWithFunctionName()

	detail, err := MapApiSessionToRestModel(ae, rc.ApiSession)

	MapApiSessionAuthQueriesToRestEntity(ae, rc, detail)

	if err != nil {
		pfxlog.Logger().Errorf("error could not convert apiSession to rest model: %v", err)
	}

	if detail == nil {
		detail = &rest_model.APISessionDetail{}
	}
	expiresAt := strfmt.DateTime(time.Time(detail.LastActivityAt).Add(sessionTimeout))
	expirationSeconds := int64(rc.ApiSession.ExpirationDuration.Seconds())

	ret := &rest_model.CurrentAPISessionDetail{
		APISessionDetail:  *detail,
		ExpiresAt:         &expiresAt,
		ExpirationSeconds: &expirationSeconds,
	}

	return ret
}

func MapApiSessionAuthQueriesToRestEntity(ae *env.AppEnv, rc *response.RequestContext, detail *rest_model.APISessionDetail) {
	logtrace.LogWithFunctionName()
	for _, authQuery := range rc.AuthQueries {
		detail.AuthQueries = append(detail.AuthQueries, &rest_model.AuthQueryDetail{
			Format:     authQuery.Format,
			HTTPMethod: authQuery.HTTPMethod,
			HTTPURL:    authQuery.HTTPURL,
			MaxLength:  authQuery.MaxLength,
			MinLength:  authQuery.MinLength,
			Provider:   authQuery.Provider,
			TypeID:     authQuery.TypeID,
			ClientID:   authQuery.ClientID,
			Scopes:     authQuery.Scopes,
			ID:         authQuery.ID,
		})
	}
}

func MapApiSessionCertificateToRestEntity(appEnv *env.AppEnv, context *response.RequestContext, cert *model.ApiSessionCertificate) (interface{}, error) {
	logtrace.LogWithFunctionName()
	return MapApiSessionCertificateToRestModel(cert)
}

func MapApiSessionCertificateToRestModel(apiSessionCert *model.ApiSessionCertificate) (*rest_model.CurrentAPISessionCertificateDetail, error) {
	logtrace.LogWithFunctionName()

	validFrom := strfmt.DateTime(*apiSessionCert.ValidAfter)
	validTo := strfmt.DateTime(*apiSessionCert.ValidBefore)

	ret := &rest_model.CurrentAPISessionCertificateDetail{
		BaseEntity:  BaseEntityToRestModel(apiSessionCert, CurrentApiSessionCertificateLinkFactory),
		Fingerprint: &apiSessionCert.Fingerprint,
		Subject:     &apiSessionCert.Subject,
		ValidFrom:   &validFrom,
		ValidTo:     &validTo,
		Certificate: &apiSessionCert.PEM,
	}

	return ret, nil
}
