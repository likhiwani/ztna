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

func NewEdgeRouterPolicyManager(env Env) *EdgeRouterPolicyManager {
	logtrace.LogWithFunctionName()
	manager := &EdgeRouterPolicyManager{
		baseEntityManager: newBaseEntityManager[*EdgeRouterPolicy, *db.EdgeRouterPolicy](env, env.GetStores().EdgeRouterPolicy),
	}
	manager.impl = manager

	RegisterManagerDecoder[*EdgeRouterPolicy](env, manager)

	return manager
}

type EdgeRouterPolicyManager struct {
	baseEntityManager[*EdgeRouterPolicy, *db.EdgeRouterPolicy]
}

func (self *EdgeRouterPolicyManager) newModelEntity() *EdgeRouterPolicy {
	logtrace.LogWithFunctionName()
	return &EdgeRouterPolicy{}
}

func (self *EdgeRouterPolicyManager) Create(entity *EdgeRouterPolicy, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchCreate[*EdgeRouterPolicy](self, entity, ctx)
}

func (self *EdgeRouterPolicyManager) ApplyCreate(cmd *command.CreateEntityCommand[*EdgeRouterPolicy], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	_, err := self.createEntity(cmd.Entity, ctx)
	return err
}

func (self *EdgeRouterPolicyManager) Update(entity *EdgeRouterPolicy, checker fields.UpdatedFields, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchUpdate[*EdgeRouterPolicy](self, entity, checker, ctx)
}

func (self *EdgeRouterPolicyManager) ApplyUpdate(cmd *command.UpdateEntityCommand[*EdgeRouterPolicy], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	return self.updateEntity(cmd.Entity, cmd.UpdatedFields, ctx)
}

func (self *EdgeRouterPolicyManager) Marshall(entity *EdgeRouterPolicy) ([]byte, error) {
	logtrace.LogWithFunctionName()
	tags, err := edge_cmd_pb.EncodeTags(entity.Tags)
	if err != nil {
		return nil, err
	}

	msg := &edge_cmd_pb.EdgeRouterPolicy{
		Id:              entity.Id,
		Name:            entity.Name,
		Tags:            tags,
		Semantic:        entity.Semantic,
		EdgeRouterRoles: entity.EdgeRouterRoles,
		IdentityRoles:   entity.IdentityRoles,
	}

	return proto.Marshal(msg)
}

func (self *EdgeRouterPolicyManager) Unmarshall(bytes []byte) (*EdgeRouterPolicy, error) {
	logtrace.LogWithFunctionName()
	msg := &edge_cmd_pb.EdgeRouterPolicy{}
	if err := proto.Unmarshal(bytes, msg); err != nil {
		return nil, err
	}

	return &EdgeRouterPolicy{
		BaseEntity: models.BaseEntity{
			Id:   msg.Id,
			Tags: edge_cmd_pb.DecodeTags(msg.Tags),
		},
		Name:            msg.Name,
		Semantic:        msg.Semantic,
		IdentityRoles:   msg.IdentityRoles,
		EdgeRouterRoles: msg.EdgeRouterRoles,
	}, nil
}
