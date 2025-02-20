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
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"
)

const EntityNameIdentityType = "identity-types"

var IdentityTypeLinkFactory = NewBasicLinkFactory(EntityNameIdentityType)

func MapIdentityTypeToRestEntity(_ *env.AppEnv, _ *response.RequestContext, identityType *model.IdentityType) (interface{}, error) {
	logtrace.LogWithFunctionName()
	return MapIdentityTypeToRestModel(identityType), nil
}

func MapIdentityTypeToRestModel(identityType *model.IdentityType) *rest_model.IdentityTypeDetail {
	logtrace.LogWithFunctionName()
	ret := &rest_model.IdentityTypeDetail{
		BaseEntity: BaseEntityToRestModel(identityType, IdentityTypeLinkFactory),
		Name:       identityType.Name,
	}
	return ret
}
