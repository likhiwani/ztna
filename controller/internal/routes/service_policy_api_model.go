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

const EntityNameServicePolicy = "service-policies"

var ServicePolicyLinkFactory = NewServicePolicyLinkFactory()

type ServicePolicyLinkFactoryImpl struct {
	BasicLinkFactory
}

func NewServicePolicyLinkFactory() *ServicePolicyLinkFactoryImpl {
	logtrace.LogWithFunctionName()
	return &ServicePolicyLinkFactoryImpl{
		BasicLinkFactory: *NewBasicLinkFactory(EntityNameServicePolicy),
	}
}

func (factory *ServicePolicyLinkFactoryImpl) Links(entity models.Entity) rest_model.Links {
	logtrace.LogWithFunctionName()
	links := factory.BasicLinkFactory.Links(entity)
	links[EntityNameService] = factory.NewNestedLink(entity, EntityNameService)
	links[EntityNameIdentity] = factory.NewNestedLink(entity, EntityNameIdentity)
	links[EntityNamePostureCheck] = factory.NewNestedLink(entity, EntityNamePostureCheck)

	return links
}

func MapCreateServicePolicyToModel(policy *rest_model.ServicePolicyCreate) *model.ServicePolicy {
	logtrace.LogWithFunctionName()
	semantic := ""
	if policy.Semantic != nil {
		semantic = string(*policy.Semantic)
	}

	ret := &model.ServicePolicy{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(policy.Tags),
		},
		Name:              stringz.OrEmpty(policy.Name),
		PolicyType:        string(*policy.Type),
		Semantic:          semantic,
		ServiceRoles:      policy.ServiceRoles,
		IdentityRoles:     policy.IdentityRoles,
		PostureCheckRoles: policy.PostureCheckRoles,
	}

	return ret
}

func MapUpdateServicePolicyToModel(id string, policy *rest_model.ServicePolicyUpdate) *model.ServicePolicy {
	logtrace.LogWithFunctionName()
	semantic := ""
	if policy.Semantic != nil {
		semantic = string(*policy.Semantic)
	}

	ret := &model.ServicePolicy{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(policy.Tags),
			Id:   id,
		},
		Name:              stringz.OrEmpty(policy.Name),
		PolicyType:        string(*policy.Type),
		Semantic:          semantic,
		ServiceRoles:      policy.ServiceRoles,
		IdentityRoles:     policy.IdentityRoles,
		PostureCheckRoles: policy.PostureCheckRoles,
	}

	return ret
}

func MapPatchServicePolicyToModel(id string, policy *rest_model.ServicePolicyPatch) *model.ServicePolicy {
	logtrace.LogWithFunctionName()
	ret := &model.ServicePolicy{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(policy.Tags),
			Id:   id,
		},
		Name:              policy.Name,
		PolicyType:        string(policy.Type),
		Semantic:          string(policy.Semantic),
		ServiceRoles:      policy.ServiceRoles,
		IdentityRoles:     policy.IdentityRoles,
		PostureCheckRoles: policy.PostureCheckRoles,
	}

	return ret
}

func MapServicePolicyToRestEntity(ae *env.AppEnv, _ *response.RequestContext, policy *model.ServicePolicy) (interface{}, error) {
	logtrace.LogWithFunctionName()
	return MapServicePolicyToRestModel(ae, policy), nil
}

func MapServicePolicyToRestModel(ae *env.AppEnv, policy *model.ServicePolicy) *rest_model.ServicePolicyDetail {
	logtrace.LogWithFunctionName()
	semantic := rest_model.Semantic(policy.Semantic)
	dialBindType := rest_model.DialBind(policy.PolicyType)

	ret := &rest_model.ServicePolicyDetail{
		BaseEntity:               BaseEntityToRestModel(policy, ServicePolicyLinkFactory),
		IdentityRoles:            policy.IdentityRoles,
		IdentityRolesDisplay:     GetNamedIdentityRoles(ae.GetManagers().Identity, policy.IdentityRoles),
		Name:                     &policy.Name,
		Semantic:                 &semantic,
		ServiceRoles:             policy.ServiceRoles,
		ServiceRolesDisplay:      GetNamedServiceRoles(ae.GetManagers().EdgeService, policy.ServiceRoles),
		Type:                     &dialBindType,
		PostureCheckRoles:        policy.PostureCheckRoles,
		PostureCheckRolesDisplay: GetNamedPostureCheckRoles(ae.GetManagers().PostureCheck, policy.PostureCheckRoles),
	}

	return ret
}
