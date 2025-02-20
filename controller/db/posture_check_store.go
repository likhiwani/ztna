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

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/errorz"
	"github.com/openziti/storage/ast"
	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
)

const (
	//Fields
	FieldPostureCheckTypeId       = "typeId"
	FieldPostureCheckVersion      = "version"
	FieldPostureCheckBindServices = "bindServices"
	FieldPostureCheckDialServices = "dialServices"
)

const (
	PostureCheckTypeOs           = "OS"
	PostureCheckTypeDomain       = "DOMAIN"
	PostureCheckTypeProcess      = "PROCESS"
	PostureCheckTypeProcessMulti = "PROCESS_MULTI"
	PostureCheckTypeMAC          = "MAC"
	PostureCheckTypeMFA          = "MFA"
)

var postureCheckSubTypeMap = map[string]newPostureCheckSubType{
	PostureCheckTypeOs:           newPostureCheckOperatingSystem,
	PostureCheckTypeDomain:       newPostureCheckWindowsDomain,
	PostureCheckTypeProcess:      newPostureCheckProcess,
	PostureCheckTypeProcessMulti: newPostureCheckProcessMulti,
	PostureCheckTypeMAC:          newPostureCheckMacAddresses,
	PostureCheckTypeMFA:          newPostureCheckMfa,
}

type newPostureCheckSubType func() PostureCheckSubType

type PostureCheckSubType interface {
	LoadValues(bucket *boltz.TypedBucket)
	SetValues(ctx *boltz.PersistContext, bucket *boltz.TypedBucket)
}

func newPostureCheck(typeId string) PostureCheckSubType {
	logtrace.LogWithFunctionName()
	if newChild, found := postureCheckSubTypeMap[typeId]; found {
		return newChild()
	}
	return nil
}

type PostureCheck struct {
	boltz.BaseExtEntity
	Name           string              `json:"name"`
	TypeId         string              `json:"typeId"`
	Version        int64               `json:"version"`
	RoleAttributes []string            `json:"roleAttributes"`
	SubType        PostureCheckSubType `json:"subType"`
}

func (entity *PostureCheck) GetName() string {
	logtrace.LogWithFunctionName()
	return entity.Name
}

func (entity *PostureCheck) GetEntityType() string {
	logtrace.LogWithFunctionName()
	return EntityTypePostureChecks
}

type PostureCheckStore interface {
	Store[*PostureCheck]
	GetRoleAttributesIndex() boltz.SetReadIndex
	GetRoleAttributesCursorProvider(filters []string, semantic string) (ast.SetCursorProvider, error)
}

func newPostureCheckStore(stores *stores) *postureCheckStoreImpl {
	logtrace.LogWithFunctionName()
	store := &postureCheckStoreImpl{}
	store.baseStore = newBaseStore[*PostureCheck](stores, store)
	store.InitImpl(store)
	return store
}

type postureCheckStoreImpl struct {
	*baseStore[*PostureCheck]
	indexName           boltz.ReadIndex
	indexRoleAttributes boltz.SetReadIndex

	symbolServicePolicies boltz.EntitySetSymbol
	symbolRoleAttributes  boltz.EntitySetSymbol
	symbolBindServices    boltz.EntitySetSymbol
	symbolDialServices    boltz.EntitySetSymbol

	bindServicesCollection boltz.RefCountedLinkCollection
	dialServicesCollection boltz.RefCountedLinkCollection
}

func (store *postureCheckStoreImpl) NewEntity() *PostureCheck {
	logtrace.LogWithFunctionName()
	return &PostureCheck{}
}

func (store *postureCheckStoreImpl) FillEntity(entity *PostureCheck, bucket *boltz.TypedBucket) {
	logtrace.LogWithFunctionName()
	entity.LoadBaseValues(bucket)
	entity.Name = bucket.GetStringOrError(FieldName)
	entity.TypeId = bucket.GetStringOrError(FieldPostureCheckTypeId)
	entity.Version = bucket.GetInt64WithDefault(FieldPostureCheckVersion, 0)
	entity.RoleAttributes = bucket.GetStringList(FieldRoleAttributes)

	entity.SubType = newPostureCheck(entity.TypeId)
	if entity.SubType == nil {
		pfxlog.Logger().Panicf("cannot load unsupported posture check type [%v]", entity.TypeId)
	}

	childBucket := bucket.GetOrCreateBucket(entity.TypeId)

	entity.SubType.LoadValues(childBucket)
}

func (store *postureCheckStoreImpl) PersistEntity(entity *PostureCheck, ctx *boltz.PersistContext) {
	logtrace.LogWithFunctionName()
	entity.SetBaseValues(ctx)
	ctx.SetString(FieldName, entity.Name)
	ctx.SetString(FieldPostureCheckTypeId, entity.TypeId)
	ctx.SetInt64(FieldPostureCheckVersion, entity.Version)
	ctx.SetStringList(FieldRoleAttributes, entity.RoleAttributes)

	childBucket := ctx.Bucket.GetOrCreateBucket(entity.TypeId)

	entity.SubType.SetValues(ctx, childBucket)

	// index change won't fire if we don't have any roles on create, but we need to evaluate if we match any #all roles
	if ctx.IsCreate && len(entity.RoleAttributes) == 0 {
		store.rolesChanged(ctx.MutateContext, []byte(entity.Id), nil, nil, ctx.Bucket)
	}
}

func (store *postureCheckStoreImpl) GetRoleAttributesIndex() boltz.SetReadIndex {
	logtrace.LogWithFunctionName()
	return store.indexRoleAttributes
}

func (store *postureCheckStoreImpl) initializeLocal() {
	logtrace.LogWithFunctionName()
	store.AddExtEntitySymbols()
	store.indexName = store.addUniqueNameField()
	store.AddSymbol(FieldPostureCheckMfaPromptOnUnlock, ast.NodeTypeBool, PostureCheckTypeMFA)
	store.AddSymbol(FieldPostureCheckMfaPromptOnWake, ast.NodeTypeBool, PostureCheckTypeMFA)

	store.symbolRoleAttributes = store.AddSetSymbol(FieldRoleAttributes, ast.NodeTypeString)
	store.indexRoleAttributes = store.AddSetIndex(store.symbolRoleAttributes)

	store.symbolBindServices = store.AddFkSetSymbol(FieldPostureCheckBindServices, store.stores.edgeService)
	store.symbolDialServices = store.AddFkSetSymbol(FieldPostureCheckDialServices, store.stores.edgeService)

	store.symbolServicePolicies = store.AddFkSetSymbol(EntityTypeServicePolicies, store.stores.servicePolicy)

	store.indexRoleAttributes.AddListener(store.rolesChanged)
}

func (store *postureCheckStoreImpl) initializeLinked() {
	logtrace.LogWithFunctionName()
	store.AddLinkCollection(store.symbolServicePolicies, store.stores.servicePolicy.symbolPostureChecks)

	store.bindServicesCollection = store.AddRefCountedLinkCollection(store.symbolBindServices, store.stores.edgeService.symbolBindIdentities)
	store.dialServicesCollection = store.AddRefCountedLinkCollection(store.symbolDialServices, store.stores.edgeService.symbolDialIdentities)
}

func (store *postureCheckStoreImpl) GetNameIndex() boltz.ReadIndex {
	logtrace.LogWithFunctionName()
	return store.indexName
}

func (store *postureCheckStoreImpl) DeleteById(ctx boltz.MutateContext, id string) error {
	logtrace.LogWithFunctionName()
	if entity, _ := store.LoadById(ctx.Tx(), id); entity != nil {
		// Remove entity from PostureCheckRoles in service policies
		if err := store.deleteEntityReferences(ctx.Tx(), entity, store.stores.servicePolicy.symbolPostureCheckRoles); err != nil {
			return err
		}
	}

	store.createServiceChangeEvents(ctx.Tx(), id)
	return store.baseStore.DeleteById(ctx, id)
}

func (store *postureCheckStoreImpl) Update(ctx boltz.MutateContext, entity *PostureCheck, checker boltz.FieldChecker) error {
	logtrace.LogWithFunctionName()
	store.createServiceChangeEvents(ctx.Tx(), entity.GetId())
	return store.baseStore.Update(ctx, entity, checker)
}

func (store *postureCheckStoreImpl) createServiceChangeEvents(tx *bbolt.Tx, id string) {
	logtrace.LogWithFunctionName()
	eh := &serviceEventHandler{}

	cursor := store.bindServicesCollection.IterateLinks(tx, []byte(id), true)
	for cursor.IsValid() {
		eh.addServiceUpdatedEvent(store.stores, tx, cursor.Current())
		cursor.Next()
	}

	cursor = store.dialServicesCollection.IterateLinks(tx, []byte(id), true)
	for cursor.IsValid() {
		eh.addServiceUpdatedEvent(store.stores, tx, cursor.Current())
		cursor.Next()
	}
}

func (store *postureCheckStoreImpl) rolesChanged(mutateCtx boltz.MutateContext, rowId []byte, _ []boltz.FieldTypeAndValue, new []boltz.FieldTypeAndValue, holder errorz.ErrorHolder) {
	logtrace.LogWithFunctionName()
	ctx := &roleAttributeChangeContext{
		mutateCtx:             mutateCtx,
		rolesSymbol:           store.stores.servicePolicy.symbolPostureCheckRoles,
		linkCollection:        store.stores.servicePolicy.postureCheckCollection,
		relatedLinkCollection: store.stores.servicePolicy.serviceCollection,
		ErrorHolder:           holder,
	}
	store.updateServicePolicyRelatedRoles(ctx, rowId, new)
}

func (store *postureCheckStoreImpl) GetRoleAttributesCursorProvider(values []string, semantic string) (ast.SetCursorProvider, error) {
	logtrace.LogWithFunctionName()
	return store.getRoleAttributesCursorProvider(store.indexRoleAttributes, values, semantic)
}
