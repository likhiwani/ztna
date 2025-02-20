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
	"ztna-core/ztna/logtrace"

	"github.com/openziti/storage/boltz"
)

func NewIdentityTypeManager(env Env) *IdentityTypeManager {
	logtrace.LogWithFunctionName()
	manager := &IdentityTypeManager{
		baseEntityManager: newBaseEntityManager[*IdentityType, *db.IdentityType](env, env.GetStores().IdentityType),
	}
	manager.impl = manager

	return manager
}

type IdentityTypeManager struct {
	baseEntityManager[*IdentityType, *db.IdentityType]
}

func (self *IdentityTypeManager) newModelEntity() *IdentityType {
	logtrace.LogWithFunctionName()
	return &IdentityType{}
}

func (self *IdentityTypeManager) ReadByIdOrName(idOrName string) (*IdentityType, error) {
	logtrace.LogWithFunctionName()
	modelEntity := &IdentityType{}
	err := self.readEntity(idOrName, modelEntity)

	if err == nil {
		return modelEntity, nil
	}

	if !boltz.IsErrNotFoundErr(err) {
		return nil, err
	}

	if modelEntity.Id == "" {
		modelEntity, err = self.ReadByName(idOrName)
	}

	if err != nil {
		return nil, err
	}

	return modelEntity, nil
}

func (self *IdentityTypeManager) ReadByName(name string) (*IdentityType, error) {
	logtrace.LogWithFunctionName()
	modelIdentityType := &IdentityType{}
	nameIndex := self.env.GetStores().IdentityType.GetNameIndex()
	if err := self.readEntityWithIndex("name", []byte(name), nameIndex, modelIdentityType); err != nil {
		return nil, err
	}
	return modelIdentityType, nil
}
