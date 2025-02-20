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
	"encoding/json"
	"ztna-core/ztna/common/pb/edge_cmd_pb"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/command"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/stringz"
	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

const (
	ConfigTypeAll = "all"
)

func NewConfigTypeManager(env Env) *ConfigTypeManager {
	logtrace.LogWithFunctionName()
	manager := &ConfigTypeManager{
		baseEntityManager: newBaseEntityManager[*ConfigType, *db.ConfigType](env, env.GetStores().ConfigType),
	}
	manager.impl = manager

	RegisterManagerDecoder[*ConfigType](env, manager)

	return manager
}

type ConfigTypeManager struct {
	baseEntityManager[*ConfigType, *db.ConfigType]
}

func (self *ConfigTypeManager) newModelEntity() *ConfigType {
	logtrace.LogWithFunctionName()
	return &ConfigType{}
}

func (self *ConfigTypeManager) Create(entity *ConfigType, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchCreate[*ConfigType](self, entity, ctx)
}

func (self *ConfigTypeManager) ApplyCreate(cmd *command.CreateEntityCommand[*ConfigType], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	_, err := self.createEntity(cmd.Entity, ctx)
	return err
}

func (self *ConfigTypeManager) Update(entity *ConfigType, checker fields.UpdatedFields, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchUpdate[*ConfigType](self, entity, checker, ctx)
}

func (self *ConfigTypeManager) ApplyUpdate(cmd *command.UpdateEntityCommand[*ConfigType], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	return self.updateEntity(cmd.Entity, cmd.UpdatedFields, ctx)
}

func (self *ConfigTypeManager) Read(id string) (*ConfigType, error) {
	logtrace.LogWithFunctionName()
	modelEntity := &ConfigType{}
	if err := self.readEntity(id, modelEntity); err != nil {
		return nil, err
	}
	return modelEntity, nil
}

func (self *ConfigTypeManager) readInTx(tx *bbolt.Tx, id string) (*ConfigType, error) {
	logtrace.LogWithFunctionName()
	modelEntity := &ConfigType{}
	if err := self.readEntityInTx(tx, id, modelEntity); err != nil {
		return nil, err
	}
	return modelEntity, nil
}

func (self *ConfigTypeManager) ReadByName(name string) (*ConfigType, error) {
	logtrace.LogWithFunctionName()
	modelEntity := &ConfigType{}
	nameIndex := self.env.GetStores().ConfigType.GetNameIndex()
	if err := self.readEntityWithIndex("name", []byte(name), nameIndex, modelEntity); err != nil {
		return nil, err
	}
	return modelEntity, nil
}

func (self *ConfigTypeManager) MapConfigTypeNamesToIds(values []string, identityId string) map[string]struct{} {
	logtrace.LogWithFunctionName()
	var result []string
	if stringz.Contains(values, "all") {
		result = []string{"all"}
	} else {
		for _, val := range values {
			if configType, _ := self.Read(val); configType != nil {
				result = append(result, val)
			} else if configType, _ := self.ReadByName(val); configType != nil {
				result = append(result, configType.Id)
			} else {
				pfxlog.Logger().Debugf("user %v submitted %v as a config type of interest, but no matching records found", identityId, val)
			}
		}
	}
	return stringz.SliceToSet(result)
}

func (self *ConfigTypeManager) Marshall(entity *ConfigType) ([]byte, error) {
	logtrace.LogWithFunctionName()
	tags, err := edge_cmd_pb.EncodeTags(entity.Tags)
	if err != nil {
		return nil, err
	}

	schema, err := json.Marshal(entity.Schema)
	if err != nil {
		return nil, err
	}

	msg := &edge_cmd_pb.ConfigType{
		Id:     entity.Id,
		Name:   entity.Name,
		Schema: schema,
		Tags:   tags,
	}

	return proto.Marshal(msg)
}

func (self *ConfigTypeManager) Unmarshall(bytes []byte) (*ConfigType, error) {
	logtrace.LogWithFunctionName()
	msg := &edge_cmd_pb.ConfigType{}
	if err := proto.Unmarshal(bytes, msg); err != nil {
		return nil, err
	}

	schema := map[string]interface{}{}
	if err := json.Unmarshal(msg.Schema, &schema); err != nil {
		return nil, err
	}

	return &ConfigType{
		BaseEntity: models.BaseEntity{
			Id:   msg.Id,
			Tags: edge_cmd_pb.DecodeTags(msg.Tags),
		},
		Name:   msg.Name,
		Schema: schema,
	}, nil
}
