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

type ServiceEdgeRouterPolicy struct {
	models.BaseEntity
	Name            string
	Semantic        string
	ServiceRoles    []string
	EdgeRouterRoles []string
}

func (entity *ServiceEdgeRouterPolicy) toBoltEntity() (*db.ServiceEdgeRouterPolicy, error) {
	logtrace.LogWithFunctionName()
	return &db.ServiceEdgeRouterPolicy{
		BaseExtEntity:   *boltz.NewExtEntity(entity.Id, entity.Tags),
		Name:            entity.Name,
		Semantic:        entity.Semantic,
		ServiceRoles:    entity.ServiceRoles,
		EdgeRouterRoles: entity.EdgeRouterRoles,
	}, nil
}

func (entity *ServiceEdgeRouterPolicy) toBoltEntityForCreate(*bbolt.Tx, Env) (*db.ServiceEdgeRouterPolicy, error) {
	logtrace.LogWithFunctionName()
	return entity.toBoltEntity()
}

func (entity *ServiceEdgeRouterPolicy) toBoltEntityForUpdate(*bbolt.Tx, Env, boltz.FieldChecker) (*db.ServiceEdgeRouterPolicy, error) {
	logtrace.LogWithFunctionName()
	return entity.toBoltEntity()
}

func (entity *ServiceEdgeRouterPolicy) fillFrom(_ Env, _ *bbolt.Tx, boltServiceEdgeRouterPolicy *db.ServiceEdgeRouterPolicy) error {
	logtrace.LogWithFunctionName()
	entity.FillCommon(boltServiceEdgeRouterPolicy)
	entity.Name = boltServiceEdgeRouterPolicy.Name
	entity.Semantic = boltServiceEdgeRouterPolicy.Semantic
	entity.EdgeRouterRoles = boltServiceEdgeRouterPolicy.EdgeRouterRoles
	entity.ServiceRoles = boltServiceEdgeRouterPolicy.ServiceRoles
	return nil
}
