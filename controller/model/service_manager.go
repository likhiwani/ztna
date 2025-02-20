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
	"time"
	"ztna-core/ztna/common/pb/cmd_pb"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/command"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/storage/boltz"
	cmap "github.com/orcaman/concurrent-map/v2"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

func newServiceManager(env Env) *ServiceManager {
	logtrace.LogWithFunctionName()
	result := &ServiceManager{
		baseEntityManager: newBaseEntityManager[*Service, *db.Service](env, env.GetStores().Service),
		cache:             cmap.New[*Service](),
	}
	result.impl = result

	env.GetStores().Service.AddEntityIdListener(result.RemoveFromCache, boltz.EntityUpdated, boltz.EntityDeleted)

	RegisterManagerDecoder[*Service](env, result)

	return result
}

type ServiceManager struct {
	baseEntityManager[*Service, *db.Service]
	cache cmap.ConcurrentMap[string, *Service]
}

func (self *ServiceManager) newModelEntity() *Service {
	logtrace.LogWithFunctionName()
	return &Service{}
}

func (self *ServiceManager) NotifyTerminatorChanged(terminator *db.Terminator) *db.Terminator {
	logtrace.LogWithFunctionName()
	// patched entities may not have all fields, if service is blank, load terminator
	serviceId := terminator.Service
	if serviceId == "" {
		err := self.GetDb().View(func(tx *bbolt.Tx) error {
			t, _, err := self.env.GetStores().Terminator.FindById(tx, terminator.Id)
			if t != nil {
				terminator = t
			}
			return err
		})
		if err != nil {
			self.clearCache()
			return terminator
		}
		serviceId = terminator.Service
	}
	pfxlog.Logger().Debugf("clearing service from cache: %v", serviceId)
	self.RemoveFromCache(serviceId)
	return terminator
}

func (self *ServiceManager) Create(entity *Service, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchCreate[*Service](self, entity, ctx)
}

func (self *ServiceManager) ApplyCreate(cmd *command.CreateEntityCommand[*Service], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	_, err := self.createEntity(cmd.Entity, ctx)
	return err
}

func (self *ServiceManager) Update(entity *Service, updatedFields fields.UpdatedFields, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchUpdate[*Service](self, entity, updatedFields, ctx)
}

func (self *ServiceManager) ApplyUpdate(cmd *command.UpdateEntityCommand[*Service], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	if err := self.updateEntity(cmd.Entity, cmd.UpdatedFields, ctx); err != nil {
		return err
	}
	self.RemoveFromCache(cmd.Entity.Id)
	return nil
}

func (self *ServiceManager) Read(id string) (entity *Service, err error) {
	logtrace.LogWithFunctionName()
	err = self.GetDb().View(func(tx *bbolt.Tx) error {
		entity, err = self.readInTx(tx, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return entity, err
}

func (self *ServiceManager) GetIdForName(id string) (string, error) {
	logtrace.LogWithFunctionName()
	var result []byte
	err := self.GetDb().View(func(tx *bbolt.Tx) error {
		result = self.env.GetStores().Service.GetNameIndex().Read(tx, []byte(id))
		return nil
	})
	return string(result), err
}

func (self *ServiceManager) readInTx(tx *bbolt.Tx, id string) (*Service, error) {
	logtrace.LogWithFunctionName()
	if service, _ := self.cache.Get(id); service != nil {
		return service, nil
	}

	entity := &Service{}
	if err := self.readEntityInTx(tx, id, entity); err != nil {
		return nil, err
	}

	self.cacheService(entity)
	return entity, nil
}

func (self *ServiceManager) cacheService(service *Service) {
	logtrace.LogWithFunctionName()
	pfxlog.Logger().Tracef("updated service cache: %v", service.Id)
	self.cache.Set(service.Id, service)
}

func (self *ServiceManager) RemoveFromCache(id string) {
	logtrace.LogWithFunctionName()
	pfxlog.Logger().Debugf("removed service from cache: %v", id)
	self.cache.Remove(id)
}

func (self *ServiceManager) clearCache() {
	logtrace.LogWithFunctionName()
	pfxlog.Logger().Debugf("clearing all services from cache")
	for _, key := range self.cache.Keys() {
		self.cache.Remove(key)
	}
}

func (self *ServiceManager) Marshall(entity *Service) ([]byte, error) {
	logtrace.LogWithFunctionName()
	tags, err := cmd_pb.EncodeTags(entity.Tags)
	if err != nil {
		return nil, err
	}

	msg := &cmd_pb.Service{
		Id:                 entity.Id,
		Name:               entity.Name,
		MaxIdleTime:        int64(entity.MaxIdleTime),
		TerminatorStrategy: entity.TerminatorStrategy,
		Tags:               tags,
	}

	return proto.Marshal(msg)
}

func (self *ServiceManager) Unmarshall(bytes []byte) (*Service, error) {
	logtrace.LogWithFunctionName()
	msg := &cmd_pb.Service{}
	if err := proto.Unmarshal(bytes, msg); err != nil {
		return nil, err
	}

	return &Service{
		BaseEntity: models.BaseEntity{
			Id:   msg.Id,
			Tags: cmd_pb.DecodeTags(msg.Tags),
		},
		Name:               msg.Name,
		MaxIdleTime:        time.Duration(msg.MaxIdleTime),
		TerminatorStrategy: msg.TerminatorStrategy,
	}, nil
}
