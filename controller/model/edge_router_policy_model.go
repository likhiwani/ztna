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

type EdgeRouterPolicy struct {
	models.BaseEntity
	Name            string
	Semantic        string
	IdentityRoles   []string
	EdgeRouterRoles []string
}

func (entity *EdgeRouterPolicy) toBoltEntity() (*db.EdgeRouterPolicy, error) {
	logtrace.LogWithFunctionName()
	return &db.EdgeRouterPolicy{
		BaseExtEntity:   *boltz.NewExtEntity(entity.Id, entity.Tags),
		Name:            entity.Name,
		Semantic:        entity.Semantic,
		IdentityRoles:   entity.IdentityRoles,
		EdgeRouterRoles: entity.EdgeRouterRoles,
	}, nil
}

func (entity *EdgeRouterPolicy) toBoltEntityForCreate(*bbolt.Tx, Env) (*db.EdgeRouterPolicy, error) {
	logtrace.LogWithFunctionName()
	return entity.toBoltEntity()
}

func (entity *EdgeRouterPolicy) toBoltEntityForUpdate(*bbolt.Tx, Env, boltz.FieldChecker) (*db.EdgeRouterPolicy, error) {
	logtrace.LogWithFunctionName()
	return entity.toBoltEntity()
}

func (entity *EdgeRouterPolicy) fillFrom(_ Env, _ *bbolt.Tx, boltEdgeRouterPolicy *db.EdgeRouterPolicy) error {
	logtrace.LogWithFunctionName()
	entity.FillCommon(boltEdgeRouterPolicy)
	entity.Name = boltEdgeRouterPolicy.Name
	entity.Semantic = boltEdgeRouterPolicy.Semantic
	entity.EdgeRouterRoles = boltEdgeRouterPolicy.EdgeRouterRoles
	entity.IdentityRoles = boltEdgeRouterPolicy.IdentityRoles
	return nil
}
