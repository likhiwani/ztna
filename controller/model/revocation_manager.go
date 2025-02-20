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
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/storage/boltz"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func NewRevocationManager(env Env) *RevocationManager {
	logtrace.LogWithFunctionName()
	manager := &RevocationManager{
		baseEntityManager: newBaseEntityManager[*Revocation, *db.Revocation](env, env.GetStores().Revocation),
	}
	manager.impl = manager

	RegisterManagerDecoder[*Revocation](env, manager)

	return manager
}

type RevocationManager struct {
	baseEntityManager[*Revocation, *db.Revocation]
}

func (self *RevocationManager) ApplyUpdate(_ *command.UpdateEntityCommand[*Revocation], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	return errors.New("unsupported")
}

func (self *RevocationManager) Create(entity *Revocation, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchCreate[*Revocation](self, entity, ctx)
}

func (self *RevocationManager) ApplyCreate(cmd *command.CreateEntityCommand[*Revocation], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	_, err := self.createEntity(cmd.Entity, ctx)
	return err
}

func (self *RevocationManager) newModelEntity() *Revocation {
	logtrace.LogWithFunctionName()
	return &Revocation{}
}

func (self *RevocationManager) Read(id string) (*Revocation, error) {
	logtrace.LogWithFunctionName()
	modelEntity := &Revocation{}
	if err := self.readEntity(id, modelEntity); err != nil {
		return nil, err
	}
	return modelEntity, nil
}

func (self *RevocationManager) Marshall(entity *Revocation) ([]byte, error) {
	logtrace.LogWithFunctionName()
	tags, err := edge_cmd_pb.EncodeTags(entity.Tags)
	if err != nil {
		return nil, err
	}

	msg := &edge_cmd_pb.Revocation{
		Id:        entity.Id,
		ExpiresAt: timePtrToPb(&entity.ExpiresAt),
		Tags:      tags,
	}

	return proto.Marshal(msg)
}

func (self *RevocationManager) Unmarshall(bytes []byte) (*Revocation, error) {
	logtrace.LogWithFunctionName()
	msg := &edge_cmd_pb.Revocation{}
	if err := proto.Unmarshal(bytes, msg); err != nil {
		return nil, err
	}

	if msg.ExpiresAt == nil {
		return nil, errors.Errorf("revocation msg for id '%v' has nil ExspiresAt", msg.Id)
	}

	return &Revocation{
		BaseEntity: models.BaseEntity{
			Id:   msg.Id,
			Tags: edge_cmd_pb.DecodeTags(msg.Tags),
		},
		ExpiresAt: *pbTimeToTimePtr(msg.ExpiresAt),
	}, nil
}
