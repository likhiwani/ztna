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

type TransitRouter struct {
	models.BaseEntity
	Name                  string
	Fingerprint           *string
	IsVerified            bool
	IsBase                bool
	UnverifiedFingerprint *string
	UnverifiedCertPem     *string
	Cost                  uint16
	NoTraversal           bool
	Disabled              bool
}

func (self *TransitRouter) GetName() string {
	logtrace.LogWithFunctionName()
	return self.Name
}

func (entity *TransitRouter) toBoltEntityForCreate(*bbolt.Tx, Env) (*db.TransitRouter, error) {
	logtrace.LogWithFunctionName()
	boltEntity := &db.TransitRouter{
		Router: db.Router{
			BaseExtEntity: *boltz.NewExtEntity(entity.Id, entity.Tags),
			Name:          entity.Name,
			Fingerprint:   entity.Fingerprint,
			Cost:          entity.Cost,
			NoTraversal:   entity.NoTraversal,
			Disabled:      entity.Disabled,
		},
		IsVerified: false,
	}

	return boltEntity, nil
}

func (entity *TransitRouter) toBoltEntityForUpdate(*bbolt.Tx, Env, boltz.FieldChecker) (*db.TransitRouter, error) {
	logtrace.LogWithFunctionName()
	ret := &db.TransitRouter{
		Router: db.Router{
			BaseExtEntity: *boltz.NewExtEntity(entity.Id, entity.Tags),
			Name:          entity.Name,
			Fingerprint:   entity.Fingerprint,
			Cost:          entity.Cost,
			NoTraversal:   entity.NoTraversal,
			Disabled:      entity.Disabled,
		},
		IsVerified:            entity.IsVerified,
		UnverifiedFingerprint: entity.UnverifiedFingerprint,
		UnverifiedCertPem:     entity.UnverifiedCertPem,
	}

	return ret, nil
}

func (entity *TransitRouter) fillFrom(_ Env, _ *bbolt.Tx, boltTransitRouter *db.TransitRouter) error {
	logtrace.LogWithFunctionName()
	entity.FillCommon(boltTransitRouter)
	entity.Name = boltTransitRouter.Name
	entity.IsVerified = boltTransitRouter.IsVerified
	entity.IsBase = boltTransitRouter.IsBase
	entity.Fingerprint = boltTransitRouter.Fingerprint
	entity.UnverifiedFingerprint = boltTransitRouter.UnverifiedFingerprint
	entity.UnverifiedCertPem = boltTransitRouter.UnverifiedCertPem
	entity.Cost = boltTransitRouter.Cost
	entity.NoTraversal = boltTransitRouter.NoTraversal
	entity.Disabled = boltTransitRouter.Disabled

	return nil
}
