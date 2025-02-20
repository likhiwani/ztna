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
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
)

type IdentityType struct {
	models.BaseEntity
	Name string `json:"name"`
}

func (entity *IdentityType) toBoltEntity() (*db.IdentityType, error) {
	logtrace.LogWithFunctionName()
	return &db.IdentityType{
		Name:          entity.Name,
		BaseExtEntity: *boltz.NewExtEntity(entity.Id, entity.Tags),
	}, nil
}

func (entity *IdentityType) toBoltEntityForCreate(*bbolt.Tx, Env) (*db.IdentityType, error) {
	logtrace.LogWithFunctionName()
	return entity.toBoltEntity()
}

func (entity *IdentityType) toBoltEntityForUpdate(*bbolt.Tx, Env, boltz.FieldChecker) (*db.IdentityType, error) {
	logtrace.LogWithFunctionName()
	return entity.toBoltEntity()
}

func (entity *IdentityType) fillFrom(_ Env, _ *bbolt.Tx, boltIdentityType *db.IdentityType) error {
	logtrace.LogWithFunctionName()
	entity.FillCommon(boltIdentityType)
	entity.Name = boltIdentityType.Name
	return nil
}
