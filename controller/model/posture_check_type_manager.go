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
)

func NewPostureCheckTypeManager(env Env) *PostureCheckTypeManager {
	logtrace.LogWithFunctionName()
	manager := &PostureCheckTypeManager{
		baseEntityManager: newBaseEntityManager[*PostureCheckType, *db.PostureCheckType](env, env.GetStores().PostureCheckType),
	}
	manager.impl = manager
	return manager
}

type PostureCheckTypeManager struct {
	baseEntityManager[*PostureCheckType, *db.PostureCheckType]
}

func (self *PostureCheckTypeManager) newModelEntity() *PostureCheckType {
	logtrace.LogWithFunctionName()
	return &PostureCheckType{}
}
