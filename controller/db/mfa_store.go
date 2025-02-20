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
	"ztna-core/ztna/common/eid"
	"ztna-core/ztna/logtrace"

	"github.com/google/uuid"
	"github.com/openziti/storage/boltz"
)

const (
	FieldMfaIdentity      = "identity"
	FieldMfaIsVerified    = "isVerified"
	FieldMfaRecoveryCodes = "recoveryCodes"
	FieldMfaSecret        = "secret"
	FieldMfaSalt          = "salt"
)

type Mfa struct {
	boltz.BaseExtEntity
	IdentityId    string   `json:"identityId"`
	IsVerified    bool     `json:"isVerified"`
	Secret        string   `json:"secret"`
	Salt          string   `json:"salt"`
	RecoveryCodes []string `json:"recoveryCodes"`
}

func NewMfa(identityId string) *Mfa {
	logtrace.LogWithFunctionName()
	return &Mfa{
		BaseExtEntity: boltz.BaseExtEntity{Id: eid.New()},
		IdentityId:    identityId,
		Salt:          uuid.New().String(),
		IsVerified:    false,
	}
}

func (entity *Mfa) GetEntityType() string {
	logtrace.LogWithFunctionName()
	return EntityTypeMfas
}

var _ MfaStore = (*MfaStoreImpl)(nil)

type MfaStore interface {
	Store[*Mfa]
}

func newMfaStore(stores *stores) *MfaStoreImpl {
	logtrace.LogWithFunctionName()
	store := &MfaStoreImpl{}
	store.baseStore = newBaseStore[*Mfa](stores, store)
	store.InitImpl(store)
	return store
}

type SecretStore interface {
	GetSecret() []byte
}

type MfaStoreImpl struct {
	*baseStore[*Mfa]
	symbolIdentity boltz.EntitySymbol
}

func (store *MfaStoreImpl) initializeLocal() {
	logtrace.LogWithFunctionName()
	store.AddExtEntitySymbols()
	store.symbolIdentity = store.AddFkSymbol(FieldMfaIdentity, store.stores.identity)

	store.AddFkConstraint(store.symbolIdentity, false, boltz.CascadeDelete)
}

func (store *MfaStoreImpl) initializeLinked() {
	logtrace.LogWithFunctionName()
}

func (store *MfaStoreImpl) NewEntity() *Mfa {
	logtrace.LogWithFunctionName()
	return &Mfa{}
}

func (store *MfaStoreImpl) FillEntity(entity *Mfa, bucket *boltz.TypedBucket) {
	logtrace.LogWithFunctionName()
	entity.LoadBaseValues(bucket)
	entity.IdentityId = bucket.GetStringOrError(FieldMfaIdentity)
	entity.IsVerified = bucket.GetBoolWithDefault(FieldMfaIsVerified, false)
	entity.RecoveryCodes = bucket.GetStringList(FieldMfaRecoveryCodes)
	entity.Salt = bucket.GetStringOrError(FieldMfaSalt)
	entity.Secret = bucket.GetStringWithDefault(FieldMfaSecret, "")
}

func (store *MfaStoreImpl) PersistEntity(entity *Mfa, ctx *boltz.PersistContext) {
	logtrace.LogWithFunctionName()
	entity.SetBaseValues(ctx)
	ctx.SetString(FieldMfaIdentity, entity.IdentityId)
	ctx.SetBool(FieldMfaIsVerified, entity.IsVerified)
	ctx.SetStringList(FieldMfaRecoveryCodes, entity.RecoveryCodes)
	ctx.SetString(FieldMfaSalt, entity.Salt)
	ctx.SetString(FieldMfaSecret, entity.Secret)
}
