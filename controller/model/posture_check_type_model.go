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

type PostureCheckType struct {
	models.BaseEntity
	Name             string
	OperatingSystems []OperatingSystem
}

func (entity *PostureCheckType) toBoltEntity() (*db.PostureCheckType, error) {
	logtrace.LogWithFunctionName()
	var operatingSystems []db.OperatingSystem

	for _, os := range entity.OperatingSystems {
		operatingSystems = append(operatingSystems, db.OperatingSystem{
			OsType:     os.OsType,
			OsVersions: os.OsVersions,
		})
	}

	return &db.PostureCheckType{
		Name:             entity.Name,
		OperatingSystems: operatingSystems,
		BaseExtEntity:    *boltz.NewExtEntity(entity.Id, entity.Tags),
	}, nil
}

func (entity *PostureCheckType) toBoltEntityForCreate(*bbolt.Tx, Env) (*db.PostureCheckType, error) {
	logtrace.LogWithFunctionName()
	return entity.toBoltEntity()
}

func (entity *PostureCheckType) toBoltEntityForUpdate(*bbolt.Tx, Env, boltz.FieldChecker) (*db.PostureCheckType, error) {
	logtrace.LogWithFunctionName()
	return entity.toBoltEntity()
}

func (entity *PostureCheckType) fillFrom(_ Env, _ *bbolt.Tx, boltPostureCheckType *db.PostureCheckType) error {
	logtrace.LogWithFunctionName()
	var operatingSystems []OperatingSystem

	for _, os := range boltPostureCheckType.OperatingSystems {
		operatingSystems = append(operatingSystems, OperatingSystem{
			OsType:     os.OsType,
			OsVersions: os.OsVersions,
		})
	}

	entity.FillCommon(boltPostureCheckType)
	entity.Name = boltPostureCheckType.Name
	entity.OperatingSystems = operatingSystems
	return nil
}
