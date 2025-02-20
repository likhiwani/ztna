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

type Revocation struct {
	models.BaseEntity
	ExpiresAt time.Time
}

func (entity *Revocation) toBoltEntityForUpdate(tx *bbolt.Tx, env Env, _ boltz.FieldChecker) (*db.Revocation, error) {
	logtrace.LogWithFunctionName()
	return entity.toBoltEntityForCreate(tx, env)
}

func (entity *Revocation) fillFrom(_ Env, _ *bbolt.Tx, boltRevocation *db.Revocation) error {
	logtrace.LogWithFunctionName()
	entity.FillCommon(boltRevocation)
	entity.ExpiresAt = boltRevocation.ExpiresAt

	return nil
}

func (entity *Revocation) toBoltEntityForCreate(*bbolt.Tx, Env) (*db.Revocation, error) {
	logtrace.LogWithFunctionName()
	boltEntity := &db.Revocation{
		BaseExtEntity: *boltz.NewExtEntity(entity.Id, entity.Tags),
		ExpiresAt:     entity.ExpiresAt,
	}

	return boltEntity, nil
}
