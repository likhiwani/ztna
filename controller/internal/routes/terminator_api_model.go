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
	"fmt"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/controller/api"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/controller/xt"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/stringz"
)

const EntityNameTerminator = "terminators"

var TerminatorLinkFactory = NewBasicLinkFactory(EntityNameTerminator)

func MapCreateTerminatorToModel(terminator *rest_model.TerminatorCreate) *model.Terminator {
	logtrace.LogWithFunctionName()
	ret := &model.Terminator{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(terminator.Tags),
		},
		Service:        stringz.OrEmpty(terminator.Service),
		Router:         stringz.OrEmpty(terminator.Router),
		Binding:        stringz.OrEmpty(terminator.Binding),
		Address:        stringz.OrEmpty(terminator.Address),
		InstanceId:     terminator.Identity,
		InstanceSecret: terminator.IdentitySecret,
		Precedence:     xt.GetPrecedenceForName(string(terminator.Precedence)),
	}

	if terminator.Cost != nil {
		ret.Cost = uint16(*terminator.Cost)
	}

	return ret
}

func MapUpdateTerminatorToModel(id string, terminator *rest_model.TerminatorUpdate) *model.Terminator {
	logtrace.LogWithFunctionName()
	ret := &model.Terminator{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(terminator.Tags),
			Id:   id,
		},
		Service:    stringz.OrEmpty(terminator.Service),
		Router:     stringz.OrEmpty(terminator.Router),
		Binding:    stringz.OrEmpty(terminator.Binding),
		Address:    stringz.OrEmpty(terminator.Address),
		Precedence: xt.GetPrecedenceForName(string(terminator.Precedence)),
	}

	if terminator.Cost != nil {
		ret.Cost = uint16(*terminator.Cost)
	}

	return ret
}

func MapPatchTerminatorToModel(id string, terminator *rest_model.TerminatorPatch) *model.Terminator {
	logtrace.LogWithFunctionName()
	ret := &model.Terminator{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(terminator.Tags),
			Id:   id,
		},
		Service:    terminator.Service,
		Router:     terminator.Router,
		Binding:    terminator.Binding,
		Address:    terminator.Address,
		Precedence: xt.GetPrecedenceForName(string(terminator.Precedence)),
	}

	if terminator.Cost != nil {
		ret.Cost = uint16(*terminator.Cost)
	}

	return ret
}

type TerminatorModelMapper struct{}

func (TerminatorModelMapper) ToApi(n *network.Network, _ api.RequestContext, terminator *model.Terminator) (interface{}, error) {
	logtrace.LogWithFunctionName()
	restModel, err := MapTerminatorToRestModel(n, terminator)

	if err != nil {
		err := fmt.Errorf("could not convert to API entity \"%s\": %s", terminator.GetId(), err)
		log := pfxlog.Logger()
		log.Error(err)
		return nil, err
	}
	return restModel, nil
}

func MapTerminatorToRestEntity(ae *env.AppEnv, _ *response.RequestContext, terminator *model.Terminator) (interface{}, error) {
	logtrace.LogWithFunctionName()
	return MapTerminatorToRestModel(ae.GetHostController().GetNetwork(), terminator)
}

func MapTerminatorToRestModel(n *network.Network, terminator *model.Terminator) (*rest_model.TerminatorDetail, error) {
	logtrace.LogWithFunctionName()

	service, err := n.Managers.Service.Read(terminator.Service)
	if err != nil {
		return nil, err
	}

	router, err := n.Managers.Router.Read(terminator.Router)
	if err != nil {
		return nil, err
	}

	cost := rest_model.TerminatorCost(int64(terminator.Cost))
	dynamicCost := rest_model.TerminatorCost(xt.GlobalCosts().GetDynamicCost(terminator.Id))

	ret := &rest_model.TerminatorDetail{
		BaseEntity:  BaseEntityToRestModel(terminator, TerminatorLinkFactory),
		ServiceID:   &terminator.Service,
		Service:     ToEntityRef(service.Name, service, ServiceLinkFactory),
		RouterID:    &terminator.Router,
		Router:      ToEntityRef(router.Name, router, TransitRouterLinkFactory),
		Binding:     &terminator.Binding,
		Address:     &terminator.Address,
		Identity:    &terminator.InstanceId,
		Cost:        &cost,
		DynamicCost: &dynamicCost,
	}

	precedence := terminator.Precedence

	resultPrecedence := rest_model.TerminatorPrecedenceDefault

	if precedence.IsRequired() {
		resultPrecedence = rest_model.TerminatorPrecedenceRequired
	} else if precedence.IsFailed() {
		resultPrecedence = rest_model.TerminatorPrecedenceFailed
	}

	ret.Precedence = &resultPrecedence

	return ret, nil
}

func MapClientTerminatorToRestEntity(ae *env.AppEnv, _ *response.RequestContext, terminator *model.Terminator) (interface{}, error) {
	logtrace.LogWithFunctionName()
	return MapLimitedTerminatorToRestModel(ae, terminator)
}

func MapLimitedTerminatorToRestModel(ae *env.AppEnv, terminator *model.Terminator) (*rest_model.TerminatorClientDetail, error) {
	logtrace.LogWithFunctionName()
	service, err := ae.Managers.EdgeService.Read(terminator.Service)
	if err != nil {
		return nil, err
	}

	ret := &rest_model.TerminatorClientDetail{
		BaseEntity: BaseEntityToRestModel(terminator, TerminatorLinkFactory),
		ServiceID:  &terminator.Service,
		Service:    ToEntityRef(service.Name, service, ServiceLinkFactory),
		RouterID:   &terminator.Router,
		Identity:   &terminator.InstanceId,
	}

	return ret, nil
}
