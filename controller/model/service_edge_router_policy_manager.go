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
	"ztna-core/ztna/common/pb/edge_cmd_pb"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/command"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/storage/boltz"
	"google.golang.org/protobuf/proto"
)

func NewServiceEdgeRouterPolicyManager(env Env) *ServiceEdgeRouterPolicyManager {
	logtrace.LogWithFunctionName()
	manager := &ServiceEdgeRouterPolicyManager{
		baseEntityManager: newBaseEntityManager[*ServiceEdgeRouterPolicy, *db.ServiceEdgeRouterPolicy](env, env.GetStores().ServiceEdgeRouterPolicy),
	}
	manager.impl = manager

	RegisterManagerDecoder[*ServiceEdgeRouterPolicy](env, manager)

	return manager
}

type ServiceEdgeRouterPolicyManager struct {
	baseEntityManager[*ServiceEdgeRouterPolicy, *db.ServiceEdgeRouterPolicy]
}

func (self *ServiceEdgeRouterPolicyManager) newModelEntity() *ServiceEdgeRouterPolicy {
	logtrace.LogWithFunctionName()
	return &ServiceEdgeRouterPolicy{}
}

func (self *ServiceEdgeRouterPolicyManager) Create(entity *ServiceEdgeRouterPolicy, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchCreate[*ServiceEdgeRouterPolicy](self, entity, ctx)
}

func (self *ServiceEdgeRouterPolicyManager) ApplyCreate(cmd *command.CreateEntityCommand[*ServiceEdgeRouterPolicy], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	_, err := self.createEntity(cmd.Entity, ctx)
	return err
}

func (self *ServiceEdgeRouterPolicyManager) Update(entity *ServiceEdgeRouterPolicy, checker fields.UpdatedFields, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchUpdate[*ServiceEdgeRouterPolicy](self, entity, checker, ctx)
}

func (self *ServiceEdgeRouterPolicyManager) ApplyUpdate(cmd *command.UpdateEntityCommand[*ServiceEdgeRouterPolicy], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	return self.updateEntity(cmd.Entity, cmd.UpdatedFields, ctx)
}

func (self *ServiceEdgeRouterPolicyManager) Marshall(entity *ServiceEdgeRouterPolicy) ([]byte, error) {
	logtrace.LogWithFunctionName()
	tags, err := edge_cmd_pb.EncodeTags(entity.Tags)
	if err != nil {
		return nil, err
	}

	msg := &edge_cmd_pb.ServiceEdgeRouterPolicy{
		Id:              entity.Id,
		Name:            entity.Name,
		Tags:            tags,
		Semantic:        entity.Semantic,
		EdgeRouterRoles: entity.EdgeRouterRoles,
		ServiceRoles:    entity.ServiceRoles,
	}

	return proto.Marshal(msg)
}

func (self *ServiceEdgeRouterPolicyManager) Unmarshall(bytes []byte) (*ServiceEdgeRouterPolicy, error) {
	logtrace.LogWithFunctionName()
	msg := &edge_cmd_pb.ServiceEdgeRouterPolicy{}
	if err := proto.Unmarshal(bytes, msg); err != nil {
		return nil, err
	}

	return &ServiceEdgeRouterPolicy{
		BaseEntity: models.BaseEntity{
			Id:   msg.Id,
			Tags: edge_cmd_pb.DecodeTags(msg.Tags),
		},
		Name:            msg.Name,
		Semantic:        msg.Semantic,
		EdgeRouterRoles: msg.EdgeRouterRoles,
		ServiceRoles:    msg.ServiceRoles,
	}, nil
}
