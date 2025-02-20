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

package db

import (
	"ztna-core/ztna/logtrace"

	"github.com/openziti/storage/boltz"
)

const (
	FieldPostureCheckTypeOperatingSystems = "operatingSystems"
)

type PostureCheckType struct {
	boltz.BaseExtEntity
	Name             string            `json:"name"`
	OperatingSystems []OperatingSystem `json:"operatingSystems"`
}

func (entity *PostureCheckType) GetName() string {
	logtrace.LogWithFunctionName()
	return entity.Name
}

func (entity *PostureCheckType) GetEntityType() string {
	logtrace.LogWithFunctionName()
	return EntityTypePostureCheckTypes
}

var _ PostureCheckTypeStore = (*postureCheckTypeStoreImpl)(nil)

type PostureCheckTypeStore interface {
	NameIndexed
	Store[*PostureCheckType]
}

func newPostureCheckTypeStore(stores *stores) *postureCheckTypeStoreImpl {
	logtrace.LogWithFunctionName()
	store := &postureCheckTypeStoreImpl{}
	store.baseStore = newBaseStore[*PostureCheckType](stores, store)
	store.InitImpl(store)
	return store
}

type postureCheckTypeStoreImpl struct {
	*baseStore[*PostureCheckType]
	indexName boltz.ReadIndex
}

func (store *postureCheckTypeStoreImpl) initializeLocal() {
	logtrace.LogWithFunctionName()
	store.AddExtEntitySymbols()
	store.indexName = store.addUniqueNameField()
}

func (store *postureCheckTypeStoreImpl) initializeLinked() {
	logtrace.LogWithFunctionName()
	// no links
}

func (store *postureCheckTypeStoreImpl) GetNameIndex() boltz.ReadIndex {
	logtrace.LogWithFunctionName()
	return store.indexName
}

func (*postureCheckTypeStoreImpl) NewEntity() *PostureCheckType {
	logtrace.LogWithFunctionName()
	return &PostureCheckType{}
}

func (*postureCheckTypeStoreImpl) FillEntity(entity *PostureCheckType, bucket *boltz.TypedBucket) {
	logtrace.LogWithFunctionName()
	entity.LoadBaseValues(bucket)
	entity.Name = bucket.GetStringOrError(FieldName)

	osBucket := bucket.GetOrCreateBucket(FieldPostureCheckTypeOperatingSystems)
	cursor := osBucket.Cursor()

	for key, _ := cursor.First(); key != nil; key, _ = cursor.Next() {
		curOs := osBucket.GetBucket(string(key))

		if curOs == nil {
			continue
		}

		newOsMatch := OperatingSystem{
			OsType: curOs.GetStringOrError(FieldPostureCheckOsType),
		}

		newOsMatch.OsVersions = append(newOsMatch.OsVersions, curOs.GetStringList(FieldPostureCheckOsVersions)...)
		entity.OperatingSystems = append(entity.OperatingSystems, newOsMatch)
	}
}

func (*postureCheckTypeStoreImpl) PersistEntity(entity *PostureCheckType, ctx *boltz.PersistContext) {
	logtrace.LogWithFunctionName()
	entity.SetBaseValues(ctx)
	ctx.SetString(FieldName, entity.Name)

	osMap := map[string]OperatingSystem{}

	for _, os := range entity.OperatingSystems {
		osMap[os.OsType] = os
	}

	osBucket := ctx.Bucket.GetOrCreateBucket(FieldPostureCheckTypeOperatingSystems)
	cursor := osBucket.Cursor()

	for key, _ := cursor.First(); key != nil; key, _ = cursor.Next() {
		if _, found := osMap[string(key)]; !found {
			_ = osBucket.Delete(key)
		}
	}

	for _, os := range entity.OperatingSystems {
		existing := osBucket.GetOrCreateBucket(os.OsType)
		existing.SetString(FieldPostureCheckOsType, os.OsType, ctx.FieldChecker)
		existing.SetStringList(FieldPostureCheckOsVersions, os.OsVersions, ctx.FieldChecker)
	}
}
