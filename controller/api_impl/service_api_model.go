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

package api_impl

import (
	"ztna-core/ztna/controller/api"
	"ztna-core/ztna/controller/idgen"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/logtrace"

	"ztna-core/ztna/controller/rest_model"

	"ztna-core/ztna/controller/models"

	"github.com/openziti/foundation/v2/stringz"
)

const EntityNameService = "services"

var ServiceLinkFactory = NewServiceLinkFactory()

type ServiceLinkFactoryIml struct {
	BasicLinkFactory
}

func NewServiceLinkFactory() *ServiceLinkFactoryIml {
	logtrace.LogWithFunctionName()
	return &ServiceLinkFactoryIml{
		BasicLinkFactory: *NewBasicLinkFactory(EntityNameService),
	}
}

func (factory *ServiceLinkFactoryIml) Links(entity LinkEntity) rest_model.Links {
	logtrace.LogWithFunctionName()
	links := factory.BasicLinkFactory.Links(entity)
	links[EntityNameTerminator] = factory.NewNestedLink(entity, EntityNameTerminator)
	return links
}

func MapCreateServiceToModel(service *rest_model.ServiceCreate) *model.Service {
	logtrace.LogWithFunctionName()
	ret := &model.Service{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(service.Tags),
		},
		Name:               stringz.OrEmpty(service.Name),
		TerminatorStrategy: service.TerminatorStrategy,
	}

	if ret.Id == "" {
		ret.Id = idgen.New()
	}

	if ret.Id != "" && ret.Name == "" {
		ret.Name = ret.Id
	}

	return ret
}

func MapUpdateServiceToModel(id string, service *rest_model.ServiceUpdate) *model.Service {
	logtrace.LogWithFunctionName()
	ret := &model.Service{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(service.Tags),
			Id:   id,
		},
		Name:               stringz.OrEmpty(service.Name),
		TerminatorStrategy: service.TerminatorStrategy,
	}

	return ret
}

func MapPatchServiceToModel(id string, service *rest_model.ServicePatch) *model.Service {
	logtrace.LogWithFunctionName()
	ret := &model.Service{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(service.Tags),
			Id:   id,
		},
		Name:               service.Name,
		TerminatorStrategy: service.TerminatorStrategy,
	}

	return ret
}

type ServiceModelMapper struct{}

func (ServiceModelMapper) ToApi(_ *network.Network, _ api.RequestContext, service *model.Service) (interface{}, error) {
	logtrace.LogWithFunctionName()
	return &rest_model.ServiceDetail{
		BaseEntity:         BaseEntityToRestModel(service, ServiceLinkFactory),
		Name:               &service.Name,
		TerminatorStrategy: &service.TerminatorStrategy,
	}, nil
}
