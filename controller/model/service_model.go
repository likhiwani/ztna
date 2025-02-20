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
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
)

type Service struct {
	models.BaseEntity
	Name               string
	TerminatorStrategy string
	Terminators        []*Terminator
	MaxIdleTime        time.Duration
}

func (entity *Service) GetName() string {
	logtrace.LogWithFunctionName()
	return entity.Name
}

func (entity *Service) toBoltEntityForUpdate(tx *bbolt.Tx, env Env, _ boltz.FieldChecker) (*db.Service, error) {
	logtrace.LogWithFunctionName()
	return entity.toBoltEntityForCreate(tx, env)
}

func (entity *Service) toBoltEntityForCreate(*bbolt.Tx, Env) (*db.Service, error) {
	logtrace.LogWithFunctionName()
	return &db.Service{
		BaseExtEntity:      *boltz.NewExtEntity(entity.Id, entity.Tags),
		Name:               entity.Name,
		MaxIdleTime:        entity.MaxIdleTime,
		TerminatorStrategy: entity.TerminatorStrategy,
	}, nil
}

func (entity *Service) fillFrom(env Env, tx *bbolt.Tx, boltService *db.Service) error {
	logtrace.LogWithFunctionName()
	entity.Name = boltService.Name
	entity.MaxIdleTime = boltService.MaxIdleTime
	entity.TerminatorStrategy = boltService.TerminatorStrategy
	entity.FillCommon(boltService)

	terminatorIds := env.GetStores().Service.GetRelatedEntitiesIdList(tx, entity.Id, db.EntityTypeTerminators)
	for _, terminatorId := range terminatorIds {
		if terminator, _ := env.GetManagers().Terminator.readInTx(tx, terminatorId); terminator != nil {
			entity.Terminators = append(entity.Terminators, terminator)
		}
	}

	return nil
}
