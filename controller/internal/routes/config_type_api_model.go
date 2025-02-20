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
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/foundation/v2/stringz"
)

const EntityNameConfigType = "config-types"

var ConfigTypeLinkFactory = NewBasicLinkFactory(EntityNameConfigType)

func MapCreateConfigTypeToModel(configType *rest_model.ConfigTypeCreate) *model.ConfigType {
	logtrace.LogWithFunctionName()
	ret := &model.ConfigType{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(configType.Tags),
		},
		Name: stringz.OrEmpty(configType.Name),
	}

	if schemaMap, ok := configType.Schema.(map[string]interface{}); ok {
		ret.Schema = schemaMap
	}

	return ret
}

func MapUpdateConfigTypeToModel(id string, configType *rest_model.ConfigTypeUpdate) *model.ConfigType {
	logtrace.LogWithFunctionName()
	ret := &model.ConfigType{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(configType.Tags),
			Id:   id,
		},
		Name: stringz.OrEmpty(configType.Name),
	}

	if schemaMap, ok := configType.Schema.(map[string]interface{}); ok {
		ret.Schema = schemaMap
	}

	return ret
}

func MapPatchConfigTypeToModel(id string, configType *rest_model.ConfigTypePatch) *model.ConfigType {
	logtrace.LogWithFunctionName()
	ret := &model.ConfigType{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(configType.Tags),
			Id:   id,
		},
		Name: configType.Name,
	}

	if schemaMap, ok := configType.Schema.(map[string]interface{}); ok {
		ret.Schema = schemaMap
	}

	return ret
}

func MapConfigTypeToRestEntity(_ *env.AppEnv, _ *response.RequestContext, configType *model.ConfigType) (interface{}, error) {
	logtrace.LogWithFunctionName()
	return MapConfigTypeToRestModel(configType)
}

func MapConfigTypeToRestModel(configType *model.ConfigType) (*rest_model.ConfigTypeDetail, error) {
	logtrace.LogWithFunctionName()
	ret := &rest_model.ConfigTypeDetail{
		BaseEntity: BaseEntityToRestModel(configType, ConfigTypeLinkFactory),
		Name:       &configType.Name,
		Schema:     configType.Schema,
	}

	return ret, nil
}
