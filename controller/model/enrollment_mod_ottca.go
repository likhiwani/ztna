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
	"crypto/x509"
	"encoding/pem"
	"time"
	"ztna-core/ztna/common/cert"
	"ztna-core/ztna/controller/apierror"
	fabricApiError "ztna-core/ztna/controller/apierror"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"
)

type EnrollModuleOttCa struct {
	env                  Env
	method               string
	fingerprintGenerator cert.FingerprintGenerator
}

func NewEnrollModuleOttCa(env Env) *EnrollModuleOttCa {
	logtrace.LogWithFunctionName()
	return &EnrollModuleOttCa{
		env:                  env,
		method:               db.MethodEnrollOttCa,
		fingerprintGenerator: cert.NewFingerprintGenerator(),
	}
}

func (module *EnrollModuleOttCa) CanHandle(method string) bool {
	logtrace.LogWithFunctionName()
	return method == module.method
}

func (module *EnrollModuleOttCa) Process(ctx EnrollmentContext) (*EnrollmentResult, error) {
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

	if enrollment.CaId == nil {
		return nil, apierror.NewInvalidEnrollmentToken()
	}

	ca, err := module.env.GetManagers().Ca.Read(*enrollment.CaId)

	if err != nil {
		return nil, err
	}

	if ca == nil {
		return nil, apierror.NewInvalidEnrollmentToken()
	}

	if !ca.IsOttCaEnrollmentEnabled {
		return nil, apierror.NewEnrollmentCaNoLongValid()
	}

	cp := x509.NewCertPool()
	cp.AppendCertsFromPEM([]byte(ca.CertPem))

	vo := x509.VerifyOptions{
		Roots:     cp,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	var validCert *x509.Certificate = nil

	for _, c := range ctx.GetCerts() {
		vc, err := c.Verify(vo)

		if err == nil || vc != nil {
			validCert = c
			break
		}
	}

	if validCert == nil {
		return nil, apierror.NewCertFailedValidation()
	}

	fingerprint := module.fingerprintGenerator.FromCert(validCert)

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: validCert.Raw,
	})

	existing, _ := module.env.GetManagers().Authenticator.ReadByFingerprint(fingerprint)

	if existing != nil {
		apiError := apierror.NewCertInUse()
		apiError.Cause = &fabricApiError.GenericCauseError{
			DataMap: map[string]interface{}{
				"fingerprint": fingerprint,
			},
		}
		return nil, apiError
	}

	newAuthenticator := &Authenticator{
		BaseEntity: models.BaseEntity{},
		Method:     db.MethodAuthenticatorCert,
		IdentityId: identity.Id,
		SubType: &AuthenticatorCert{
			Fingerprint:       fingerprint,
			Pem:               string(certPem),
			IsIssuedByNetwork: false,
		},
	}

	err = module.env.GetManagers().Enrollment.ReplaceWithAuthenticator(enrollment.Id, newAuthenticator, ctx.GetChangeContext())

	if err != nil {
		return nil, err
	}

	return &EnrollmentResult{
		Identity:      identity,
		Authenticator: newAuthenticator,
		Content:       map[string]interface{}{},
		TextContent:   []byte(""),
		Status:        200,
	}, nil
}
