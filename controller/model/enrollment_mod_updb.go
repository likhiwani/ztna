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
	"encoding/base64"
	"errors"
	"time"
	"ztna-core/ztna/common/cert"
	"ztna-core/ztna/common/eid"
	"ztna-core/ztna/controller/apierror"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/foundation/v2/errorz"
)

type EnrollModuleUpdb struct {
	env                  Env
	method               string
	fingerprintGenerator cert.FingerprintGenerator
}

func NewEnrollModuleUpdb(env Env) *EnrollModuleUpdb {
	logtrace.LogWithFunctionName()
	return &EnrollModuleUpdb{
		env:                  env,
		method:               db.MethodEnrollUpdb,
		fingerprintGenerator: cert.NewFingerprintGenerator(),
	}
}

func (module *EnrollModuleUpdb) CanHandle(method string) bool {
	logtrace.LogWithFunctionName()
	return method == module.method
}

func (module *EnrollModuleUpdb) Process(ctx EnrollmentContext) (*EnrollmentResult, error) {
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

	data := ctx.GetDataAsMap()

	password := ""

	val, ok := data["password"]
	if !ok {
		return nil, errorz.NewUnhandled(errors.New("password expected for updb enrollment"))
	}
	password = val.(string)

	hash := Hash(password)

	encodedPassword := base64.StdEncoding.EncodeToString(hash.Hash)
	encodedSalt := base64.StdEncoding.EncodeToString(hash.Salt)

	newAuthenticator := &Authenticator{
		BaseEntity: models.BaseEntity{
			Id: eid.New(),
		},
		Method:     db.MethodAuthenticatorUpdb,
		IdentityId: *enrollment.IdentityId,
		SubType: &AuthenticatorUpdb{
			Username: *enrollment.Username,
			Password: encodedPassword,
			Salt:     encodedSalt,
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
		Status:        200,
	}, nil

}
