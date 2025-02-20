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

	"github.com/openziti/foundation/v2/versions"
	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
)

type EdgeRouter struct {
	models.BaseEntity
	Name                  string
	RoleAttributes        []string
	IsVerified            bool
	Fingerprint           *string
	CertPem               *string
	Hostname              *string
	VersionInfo           *versions.VersionInfo
	IsTunnelerEnabled     bool
	AppData               map[string]interface{}
	UnverifiedFingerprint *string
	UnverifiedCertPem     *string
	Cost                  uint16
	NoTraversal           bool
	Disabled              bool
}

func (self *EdgeRouter) GetName() string {
	logtrace.LogWithFunctionName()
	return self.Name
}

func (entity *EdgeRouter) toBoltEntityForCreate(*bbolt.Tx, Env) (*db.EdgeRouter, error) {
	logtrace.LogWithFunctionName()
	boltEntity := &db.EdgeRouter{
		Router: db.Router{
			BaseExtEntity: *boltz.NewExtEntity(entity.Id, entity.Tags),
			Name:          entity.Name,
			Cost:          entity.Cost,
			NoTraversal:   entity.NoTraversal,
			Disabled:      entity.Disabled,
		},
		RoleAttributes:    entity.RoleAttributes,
		IsVerified:        false,
		IsTunnelerEnabled: entity.IsTunnelerEnabled,
		AppData:           entity.AppData,
	}

	return boltEntity, nil
}

func (entity *EdgeRouter) toBoltEntityForUpdate(*bbolt.Tx, Env, boltz.FieldChecker) (*db.EdgeRouter, error) {
	logtrace.LogWithFunctionName()
	return &db.EdgeRouter{
		Router: db.Router{
			BaseExtEntity: *boltz.NewExtEntity(entity.Id, entity.Tags),
			Name:          entity.Name,
			Fingerprint:   entity.Fingerprint,
			Cost:          entity.Cost,
			NoTraversal:   entity.NoTraversal,
			Disabled:      entity.Disabled,
		},
		RoleAttributes:        entity.RoleAttributes,
		IsVerified:            entity.IsVerified,
		CertPem:               entity.CertPem,
		IsTunnelerEnabled:     entity.IsTunnelerEnabled,
		AppData:               entity.AppData,
		UnverifiedFingerprint: entity.UnverifiedFingerprint,
		UnverifiedCertPem:     entity.UnverifiedCertPem,
	}, nil
}

func (entity *EdgeRouter) fillFrom(_ Env, _ *bbolt.Tx, boltEdgeRouter *db.EdgeRouter) error {
	logtrace.LogWithFunctionName()
	entity.FillCommon(boltEdgeRouter)
	entity.Name = boltEdgeRouter.Name
	entity.RoleAttributes = boltEdgeRouter.RoleAttributes
	entity.IsVerified = boltEdgeRouter.IsVerified
	entity.Fingerprint = boltEdgeRouter.Fingerprint
	entity.CertPem = boltEdgeRouter.CertPem
	entity.IsTunnelerEnabled = boltEdgeRouter.IsTunnelerEnabled
	entity.AppData = boltEdgeRouter.AppData
	entity.UnverifiedFingerprint = boltEdgeRouter.UnverifiedFingerprint
	entity.UnverifiedCertPem = boltEdgeRouter.UnverifiedCertPem
	entity.Cost = boltEdgeRouter.Cost
	entity.NoTraversal = boltEdgeRouter.NoTraversal
	entity.Disabled = boltEdgeRouter.Disabled

	return nil
}
