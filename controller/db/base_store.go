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
	"strings"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/foundation/v2/errorz"
	"github.com/openziti/storage/ast"
	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
)

type initializableStore interface {
	boltz.Store
	initializeLocal()
	initializeLinked()
	initializeIndexes(tx *bbolt.Tx, errorHolder errorz.ErrorHolder)
}

type Store[E boltz.ExtEntity] interface {
	boltz.EntityStore[E]
	initializableStore
}

type baseStore[E boltz.ExtEntity] struct {
	stores *stores
	*boltz.BaseStore[E]
}

func (store *baseStore[E]) addUniqueNameField() boltz.ReadIndex {
	logtrace.LogWithFunctionName()
	symbolName := store.AddSymbol(FieldName, ast.NodeTypeString)
	return store.AddUniqueIndex(symbolName)
}

func (store *baseStore[E]) initializeIndexes(tx *bbolt.Tx, errorHolder errorz.ErrorHolder) {
	logtrace.LogWithFunctionName()
	store.InitializeIndexes(tx, errorHolder)
}

func (store *baseStore[E]) deleteEntityReferences(tx *bbolt.Tx, entity boltz.NamedExtEntity, rolesSymbol boltz.EntitySetSymbol) error {
	logtrace.LogWithFunctionName()
	idRef := entityRef(entity.GetId())

	for _, policyHolderId := range store.GetRelatedEntitiesIdList(tx, entity.GetId(), rolesSymbol.GetStore().GetEntityType()) {
		err := rolesSymbol.Map(tx, []byte(policyHolderId), func(ctx *boltz.MapContext) {
			if ctx.ValueS() == idRef {
				ctx.Delete()
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (store *baseStore[E]) getParentBucket(entity boltz.Entity, childBucket *boltz.TypedBucket) *boltz.TypedBucket {
	logtrace.LogWithFunctionName()
	parentBucket := store.GetParentStore().GetEntityBucket(childBucket.Tx(), []byte(entity.GetId()))
	parentBucket.ErrorHolderImpl = childBucket.ErrorHolderImpl
	return parentBucket
}

type NameIndexed interface {
	GetNameIndex() boltz.ReadIndex
}

func (store *baseStore[E]) GetName(tx *bbolt.Tx, id string) *string {
	logtrace.LogWithFunctionName()
	symbol := store.GetSymbol(FieldName)
	if symbol == nil {
		return nil
	}
	_, val := symbol.Eval(tx, []byte(id))
	if val != nil {
		result := string(val)
		return &result
	}
	return nil
}

func (store *baseStore[E]) getRoleAttributesCursorProvider(index boltz.SetReadIndex, values []string, semantic string) (ast.SetCursorProvider, error) {
	logtrace.LogWithFunctionName()
	if semantic == "" {
		semantic = SemanticAllOf
	}

	if !isSemanticValid(semantic) {
		return nil, errorz.NewFieldError("invalid semantic", FieldSemantic, semantic)
	}

	roles, ids, err := splitRolesAndIds(values)
	if err != nil {
		return nil, err
	}

	return func(tx *bbolt.Tx, forward bool) ast.SetCursor {
		validIds := ast.NewTreeSet(forward)
		for _, id := range ids {
			if store.IsEntityPresent(tx, id) {
				validIds.Add([]byte(id))
			}
		}

		var rolesCursor ast.SetCursor
		if strings.EqualFold(semantic, SemanticAllOf) {
			rolesCursor = store.IteratorMatchingAllOf(index, roles)(tx, forward)
		} else {
			rolesCursor = store.IteratorMatchingAnyOf(index, roles)(tx, forward)
		}
		if validIds.Size() == 0 {
			return rolesCursor
		}
		return ast.NewUnionSetCursor(rolesCursor, validIds.ToCursor(), forward)
	}, nil
}
