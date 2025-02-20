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
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/logtrace"

	"ztna-core/ztna/controller/rest_model"
)

const EntityNameCircuit = "circuits"

var CircuitLinkFactory = NewCircuitLinkFactory()

type CircuitLinkFactoryIml struct {
	BasicLinkFactory
}

func NewCircuitLinkFactory() *CircuitLinkFactoryIml {
	logtrace.LogWithFunctionName()
	return &CircuitLinkFactoryIml{
		BasicLinkFactory: *NewBasicLinkFactory(EntityNameCircuit),
	}
}

func (factory *CircuitLinkFactoryIml) Links(entity LinkEntity) rest_model.Links {
	logtrace.LogWithFunctionName()
	links := factory.BasicLinkFactory.Links(entity)
	links[EntityNameTerminator] = factory.NewNestedLink(entity, EntityNameTerminator)
	links[EntityNameService] = factory.NewNestedLink(entity, EntityNameService)
	return links
}

func MapCircuitToRestModel(n *network.Network, _ api.RequestContext, circuit *model.Circuit) (*rest_model.CircuitDetail, error) {
	logtrace.LogWithFunctionName()
	path := &rest_model.Path{}
	for _, node := range circuit.Path.Nodes {
		path.Nodes = append(path.Nodes, ToEntityRef(node.Name, node, RouterLinkFactory))
	}
	for _, link := range circuit.Path.Links {
		path.Links = append(path.Links, ToEntityRef(link.Id, link, LinkLinkFactory))
	}

	var svcEntityRef *rest_model.EntityRef
	if svc, _ := n.Service.Read(circuit.ServiceId); svc != nil {
		svcEntityRef = ToEntityRef(svc.Name, svc, ServiceLinkFactory)
	} else {
		svcEntityRef = ToEntityRef("<deleted>", deletedEntity(circuit.ServiceId), ServiceLinkFactory)
	}

	ret := &rest_model.CircuitDetail{
		BaseEntity: BaseEntityToRestModel(circuit, CircuitLinkFactory),
		ClientID:   circuit.ClientId,
		Path:       path,
		Service:    svcEntityRef,
		Terminator: ToEntityRef(circuit.Terminator.GetId(), circuit.Terminator, TerminatorLinkFactory),
	}

	return ret, nil
}

type deletedEntity string

func (self deletedEntity) GetId() string {
	logtrace.LogWithFunctionName()
	return string(self)
}
