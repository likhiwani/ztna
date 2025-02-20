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

const EntityNameEdgeRouterPolicy = "edge-router-policies"

var EdgeRouterPolicyLinkFactory = NewEdgeRouterPolicyLinkFactory()

type EdgeRouterPolicyLinkFactoryImpl struct {
	BasicLinkFactory
}

func NewEdgeRouterPolicyLinkFactory() *EdgeRouterPolicyLinkFactoryImpl {
	logtrace.LogWithFunctionName()
	return &EdgeRouterPolicyLinkFactoryImpl{
		BasicLinkFactory: *NewBasicLinkFactory(EntityNameEdgeRouterPolicy),
	}
}

func (factory *EdgeRouterPolicyLinkFactoryImpl) Links(entity models.Entity) rest_model.Links {
	logtrace.LogWithFunctionName()
	links := factory.BasicLinkFactory.Links(entity)
	links[EntityNameEdgeRouter] = factory.NewNestedLink(entity, EntityNameEdgeRouter)
	links[EntityNameIdentity] = factory.NewNestedLink(entity, EntityNameIdentity)

	return links
}

func MapCreateEdgeRouterPolicyToModel(policy *rest_model.EdgeRouterPolicyCreate) *model.EdgeRouterPolicy {
	logtrace.LogWithFunctionName()
	semantic := ""
	if policy.Semantic != nil {
		semantic = string(*policy.Semantic)
	}
	ret := &model.EdgeRouterPolicy{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(policy.Tags),
		},
		Name:            stringz.OrEmpty(policy.Name),
		Semantic:        semantic,
		EdgeRouterRoles: policy.EdgeRouterRoles,
		IdentityRoles:   policy.IdentityRoles,
	}

	return ret
}

func MapUpdateEdgeRouterPolicyToModel(id string, policy *rest_model.EdgeRouterPolicyUpdate) *model.EdgeRouterPolicy {
	logtrace.LogWithFunctionName()
	semantic := ""
	if policy.Semantic != nil {
		semantic = string(*policy.Semantic)
	}

	ret := &model.EdgeRouterPolicy{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(policy.Tags),
			Id:   id,
		},
		Name:            stringz.OrEmpty(policy.Name),
		Semantic:        semantic,
		EdgeRouterRoles: policy.EdgeRouterRoles,
		IdentityRoles:   policy.IdentityRoles,
	}

	return ret
}

func MapPatchEdgeRouterPolicyToModel(id string, policy *rest_model.EdgeRouterPolicyPatch) *model.EdgeRouterPolicy {
	logtrace.LogWithFunctionName()
	ret := &model.EdgeRouterPolicy{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(policy.Tags),
			Id:   id,
		},
		Name:            policy.Name,
		Semantic:        string(policy.Semantic),
		EdgeRouterRoles: policy.EdgeRouterRoles,
		IdentityRoles:   policy.IdentityRoles,
	}

	return ret
}

func MapEdgeRouterPolicyToRestEntity(ae *env.AppEnv, _ *response.RequestContext, policy *model.EdgeRouterPolicy) (interface{}, error) {
	logtrace.LogWithFunctionName()
	return MapEdgeRouterPolicyToRestModel(ae, policy)
}

func MapEdgeRouterPolicyToRestModel(ae *env.AppEnv, policy *model.EdgeRouterPolicy) (*rest_model.EdgeRouterPolicyDetail, error) {
	logtrace.LogWithFunctionName()
	semantic := rest_model.Semantic(policy.Semantic)
	ret := &rest_model.EdgeRouterPolicyDetail{
		BaseEntity:             BaseEntityToRestModel(policy, EdgeRouterPolicyLinkFactory),
		EdgeRouterRoles:        policy.EdgeRouterRoles,
		EdgeRouterRolesDisplay: GetNamedEdgeRouterRoles(ae.GetManagers().EdgeRouter, policy.EdgeRouterRoles),
		IdentityRoles:          policy.IdentityRoles,
		IdentityRolesDisplay:   GetNamedIdentityRoles(ae.GetManagers().Identity, policy.IdentityRoles),
		Name:                   &policy.Name,
		Semantic:               &semantic,
		IsSystem:               &policy.IsSystem,
	}

	return ret, nil
}
