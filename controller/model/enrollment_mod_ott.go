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

package model

import (
	"encoding/pem"
	"time"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/common/cert"
	"ztna-core/ztna/common/eid"
	"ztna-core/ztna/controller/apierror"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"
)

type EnrollModuleOtt struct {
	env                  Env
	method               string
	fingerprintGenerator cert.FingerprintGenerator
}

func NewEnrollModuleOtt(env Env) *EnrollModuleOtt {
	logtrace.LogWithFunctionName()
	return &EnrollModuleOtt{
		env:                  env,
		method:               db.MethodEnrollOtt,
		fingerprintGenerator: cert.NewFingerprintGenerator(),
	}
}

func (module *EnrollModuleOtt) CanHandle(method string) bool {
	logtrace.LogWithFunctionName()
	return method == module.method
}

func (module *EnrollModuleOtt) Process(ctx EnrollmentContext) (*EnrollmentResult, error) {
	logtrace.LogWithFunctionName()
	enrollment, err := module.env.GetManagers().Enrollment.ReadByToken(ctx.GetToken())
	if err != nil {
		return nil, err
	}

	if enrollment == nil || enrollment.IdentityId == nil {
		return nil, apierror.NewInvalidEnrollmentToken()
	}

	if enrollment.ExpiresAt == nil || enrollment.ExpiresAt.IsZero() || enrollment.ExpiresAt.Before(time.Now()) {
		return nil, apierror.NewEnrollmentExpired()
	}

	identity, err := module.env.GetManagers().Identity.Read(*enrollment.IdentityId)

	if err != nil {
		return nil, err
	}

	if identity == nil {
		return nil, apierror.NewInvalidEnrollmentToken()
	}

	ctx.GetChangeContext().
		SetChangeAuthorType(change.AuthorTypeIdentity).
		SetChangeAuthorId(identity.Id).
		SetChangeAuthorName(identity.Name)

	csrPem := ctx.GetDataAsByteArray()

	csr, err := cert.ParseCsrPem(csrPem)

	if err != nil {
		apiErr := apierror.NewCouldNotProcessCsr()
		apiErr.Cause = err
		apiErr.AppendCause = true
		return nil, apiErr
	}

	certRaw, err := module.env.GetApiClientCsrSigner().SignCsr(csr, &cert.SigningOpts{})

	if err != nil {
		apiErr := apierror.NewCouldNotProcessCsr()
		apiErr.Cause = err
		apiErr.AppendCause = true
		return nil, apiErr
	}

	fp := module.fingerprintGenerator.FromRaw(certRaw)

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certRaw,
	})

	clientChainPem, err := module.env.GetManagers().Enrollment.GetCertChainPem(certRaw)
	if err != nil {
		return nil, err
	}

	newAuthenticator := &Authenticator{
		BaseEntity: models.BaseEntity{
			Id: eid.New(),
		},
		Method:     db.MethodAuthenticatorCert,
		IdentityId: *enrollment.IdentityId,
		SubType: &AuthenticatorCert{
			Fingerprint:       fp,
			Pem:               string(certPem),
			IsIssuedByNetwork: true,
		},
	}

	err = module.env.GetManagers().Enrollment.ReplaceWithAuthenticator(enrollment.Id, newAuthenticator, ctx.GetChangeContext())

	if err != nil {
		return nil, err
	}

	content := &rest_model.EnrollmentCerts{
		Cert: clientChainPem,
	}

	return &EnrollmentResult{
		Identity:      identity,
		Authenticator: newAuthenticator,
		Content:       content,
		TextContent:   certPem,
		Status:        200,
	}, nil

}
