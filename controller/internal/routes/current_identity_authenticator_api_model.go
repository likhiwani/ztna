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
	"path"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"
)

var CurrentIdentityAuthenticatorLinkFactory LinksFactory = NewCurrentIdentityAuthenticatorLinkFactory()

type CurrentIdentityAuthenticatorLinkFactoryImpl struct {
	BasicLinkFactory
}

func NewCurrentIdentityAuthenticatorLinkFactory() *CurrentIdentityAuthenticatorLinkFactoryImpl {
	logtrace.LogWithFunctionName()
	return &CurrentIdentityAuthenticatorLinkFactoryImpl{
		BasicLinkFactory: *NewBasicLinkFactory(EntityNameAuthenticator),
	}
}

func (factory *CurrentIdentityAuthenticatorLinkFactoryImpl) SelfUrlString(id string) string {
	logtrace.LogWithFunctionName()
	return "./" + path.Join(EntityNameCurrentIdentity, factory.entityName, id)
}

func (factory CurrentIdentityAuthenticatorLinkFactoryImpl) NewNestedLink(entity models.Entity, elem ...string) rest_model.Link {
	logtrace.LogWithFunctionName()
	elem = append(elem, entity.GetId())
	elem = append([]string{EntityNameCurrentIdentity}, elem...)
	return NewLink("./" + path.Join(elem...))
}

func (factory *CurrentIdentityAuthenticatorLinkFactoryImpl) SelfLink(entity models.Entity) rest_model.Link {
	logtrace.LogWithFunctionName()
	return NewLink("./" + path.Join(EntityNameCurrentIdentity, factory.entityName, entity.GetId()))
}

func (factory *CurrentIdentityAuthenticatorLinkFactoryImpl) Links(entity models.Entity) rest_model.Links {
	logtrace.LogWithFunctionName()
	return rest_model.Links{
		EntityNameSelf: factory.SelfLink(entity),
	}
}

func MapUpdateAuthenticatorWithCurrentToModel(id, identityId string, authenticator *rest_model.AuthenticatorUpdateWithCurrent) *model.AuthenticatorSelf {
	logtrace.LogWithFunctionName()
	ret := &model.AuthenticatorSelf{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(authenticator.Tags),
			Id:   id,
		},
		CurrentPassword: string(*authenticator.CurrentPassword),
		NewPassword:     string(*authenticator.Password),
		IdentityId:      identityId,
		Username:        string(*authenticator.Username),
	}

	return ret
}

func MapPatchAuthenticatorWithCurrentToModel(id, identityId string, authenticator *rest_model.AuthenticatorPatchWithCurrent) *model.AuthenticatorSelf {
	logtrace.LogWithFunctionName()
	ret := &model.AuthenticatorSelf{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(authenticator.Tags),
			Id:   id,
		},
		CurrentPassword: string(*authenticator.CurrentPassword),
		IdentityId:      identityId,
	}

	if authenticator.Password != nil {
		ret.NewPassword = string(*authenticator.Password)
	}

	if authenticator.Username != nil {
		ret.Username = string(*authenticator.Username)
	}

	return ret
}
