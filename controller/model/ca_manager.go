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
	"fmt"
	"strings"
	"ztna-core/ztna/common/pb/edge_cmd_pb"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/command"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/storage/ast"
	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

func NewCaManager(env Env) *CaManager {
	logtrace.LogWithFunctionName()
	manager := &CaManager{
		baseEntityManager: newBaseEntityManager[*Ca, *db.Ca](env, env.GetStores().Ca),
	}
	manager.impl = manager

	RegisterManagerDecoder[*Ca](env, manager)

	return manager
}

type CaManager struct {
	baseEntityManager[*Ca, *db.Ca]
}

func (self *CaManager) newModelEntity() *Ca {
	logtrace.LogWithFunctionName()
	return &Ca{}
}

func (self *CaManager) Create(entity *Ca, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchCreate[*Ca](self, entity, ctx)
}

func (self *CaManager) ApplyCreate(cmd *command.CreateEntityCommand[*Ca], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	_, err := self.createEntity(cmd.Entity, ctx)
	return err
}

func (self *CaManager) Update(entity *Ca, checker fields.UpdatedFields, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	if checker != nil {
		checker.RemoveFields(db.FieldCaIsVerified)
	}
	return DispatchUpdate[*Ca](self, entity, checker, ctx)
}

func (self *CaManager) ApplyUpdate(cmd *command.UpdateEntityCommand[*Ca], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	var checker boltz.FieldChecker = self

	// isVerified should only be set by the Verified method. We remove isVerified
	// from updated fields coming through Update method
	if cmd.UpdatedFields != nil {
		if cmd.UpdatedFields.IsUpdated(db.FieldCaIsVerified) {
			checker = cmd.UpdatedFields
		} else {
			checker = &AndFieldChecker{first: self, second: cmd.UpdatedFields}
		}
	}

	return self.updateEntity(cmd.Entity, checker, ctx)
}

func (self *CaManager) Read(id string) (*Ca, error) {
	logtrace.LogWithFunctionName()
	modelEntity := &Ca{}
	if err := self.readEntity(id, modelEntity); err != nil {
		return nil, err
	}
	return modelEntity, nil
}

func (self *CaManager) readInTx(tx *bbolt.Tx, id string) (*Ca, error) {
	logtrace.LogWithFunctionName()
	modelEntity := &Ca{}
	if err := self.readEntityInTx(tx, id, modelEntity); err != nil {
		return nil, err
	}
	return modelEntity, nil
}

func (self *CaManager) IsUpdated(field string) bool {
	logtrace.LogWithFunctionName()
	return strings.EqualFold(field, db.FieldName) ||
		strings.EqualFold(field, boltz.FieldTags) ||
		strings.EqualFold(field, db.FieldCaIsAutoCaEnrollmentEnabled) ||
		strings.EqualFold(field, db.FieldCaIsOttCaEnrollmentEnabled) ||
		strings.EqualFold(field, db.FieldCaIsAuthEnabled) ||
		strings.EqualFold(field, db.FieldIdentityRoles) ||
		strings.EqualFold(field, db.FieldCaIdentityNameFormat) ||
		strings.HasPrefix(field, db.FieldCaExternalIdClaim+".")
}

func (self *CaManager) Verified(ca *Ca, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	ca.IsVerified = true
	checker := &fields.UpdatedFieldsMap{
		db.FieldCaIsVerified: struct{}{},
	}
	return DispatchUpdate[*Ca](self, ca, checker, ctx)
}

func (self *CaManager) Query(query string) (*CaListResult, error) {
	logtrace.LogWithFunctionName()
	result := &CaListResult{manager: self}
	if err := self.ListWithHandler(query, result.collect); err != nil {
		return nil, err
	}
	return result, nil
}

func (self *CaManager) Stream(query string, collect func(*Ca, error) error) error {
	logtrace.LogWithFunctionName()
	filter, err := ast.Parse(self.Store, query)

	if err != nil {
		return fmt.Errorf("could not parse query for streaming cas: %v", err)
	}

	return self.env.GetDb().View(func(tx *bbolt.Tx) error {
		for cursor := self.Store.IterateIds(tx, filter); cursor.IsValid(); cursor.Next() {
			current := cursor.Current()

			ca, err := self.readInTx(tx, string(current))
			if err := collect(ca, err); err != nil {
				return err
			}
		}
		return collect(nil, nil)
	})
}

func (self *CaManager) Marshall(entity *Ca) ([]byte, error) {
	logtrace.LogWithFunctionName()
	tags, err := edge_cmd_pb.EncodeTags(entity.Tags)
	if err != nil {
		return nil, err
	}

	var externalIdClaim *edge_cmd_pb.Ca_ExternalIdClaim
	if entity.ExternalIdClaim != nil {
		externalIdClaim = &edge_cmd_pb.Ca_ExternalIdClaim{
			Location:        entity.ExternalIdClaim.Location,
			Matcher:         entity.ExternalIdClaim.Matcher,
			MatcherCriteria: entity.ExternalIdClaim.MatcherCriteria,
			Parser:          entity.ExternalIdClaim.Parser,
			ParserCriteria:  entity.ExternalIdClaim.ParserCriteria,
			Index:           entity.ExternalIdClaim.Index,
		}
	}

	msg := &edge_cmd_pb.Ca{
		Id:                        entity.Id,
		Name:                      entity.Name,
		Tags:                      tags,
		Fingerprint:               entity.Fingerprint,
		CertPem:                   entity.CertPem,
		IsVerified:                entity.IsVerified,
		VerificationToken:         entity.VerificationToken,
		IsAutoCaEnrollmentEnabled: entity.IsAutoCaEnrollmentEnabled,
		IsOttCaEnrollmentEnabled:  entity.IsOttCaEnrollmentEnabled,
		IsAuthEnabled:             entity.IsAuthEnabled,
		IdentityRoles:             entity.IdentityRoles,
		IdentityNameFormat:        entity.IdentityNameFormat,
		ExternalIdClaim:           externalIdClaim,
	}

	return proto.Marshal(msg)
}

func (self *CaManager) Unmarshall(bytes []byte) (*Ca, error) {
	logtrace.LogWithFunctionName()
	msg := &edge_cmd_pb.Ca{}
	if err := proto.Unmarshal(bytes, msg); err != nil {
		return nil, err
	}

	var externalIdClaim *ExternalIdClaim
	if msg.ExternalIdClaim != nil {
		externalIdClaim = &ExternalIdClaim{
			Location:        msg.ExternalIdClaim.Location,
			Matcher:         msg.ExternalIdClaim.Matcher,
			MatcherCriteria: msg.ExternalIdClaim.MatcherCriteria,
			Parser:          msg.ExternalIdClaim.Parser,
			ParserCriteria:  msg.ExternalIdClaim.ParserCriteria,
			Index:           msg.ExternalIdClaim.Index,
		}
	}

	return &Ca{
		BaseEntity: models.BaseEntity{
			Id:   msg.Id,
			Tags: edge_cmd_pb.DecodeTags(msg.Tags),
		},
		Name:                      msg.Name,
		Fingerprint:               msg.Fingerprint,
		CertPem:                   msg.CertPem,
		IsVerified:                msg.IsVerified,
		VerificationToken:         msg.VerificationToken,
		IsAutoCaEnrollmentEnabled: msg.IsAutoCaEnrollmentEnabled,
		IsOttCaEnrollmentEnabled:  msg.IsOttCaEnrollmentEnabled,
		IsAuthEnabled:             msg.IsAuthEnabled,
		IdentityRoles:             msg.IdentityRoles,
		IdentityNameFormat:        msg.IdentityNameFormat,
		ExternalIdClaim:           externalIdClaim,
	}, nil
}

type CaListResult struct {
	manager *CaManager
	Cas     []*Ca
	models.QueryMetaData
}

func (result *CaListResult) collect(tx *bbolt.Tx, ids []string, queryMetaData *models.QueryMetaData) error {
	logtrace.LogWithFunctionName()
	result.QueryMetaData = *queryMetaData
	for _, key := range ids {
		entity, err := result.manager.readInTx(tx, key)
		if err != nil {
			return err
		}
		result.Cas = append(result.Cas, entity)
	}
	return nil
}

const (
	FormatSentinelStart       = "["
	FormatSentinelEnd         = "]"
	FormatSymbolCaName        = "caName"
	FormatSymbolCaId          = "caId"
	FormatSymbolCommonName    = "commonName"
	FormatSymbolRequestedName = "requestedName"
	FormatSymbolIdentityId    = "identityId"

	// DefaultCaIdentityNameFormat = "[caName] - [commonName]"
	DefaultCaIdentityNameFormat = FormatSentinelStart + FormatSymbolCaName + FormatSentinelEnd + "-" + FormatSentinelStart + FormatSymbolCommonName + FormatSentinelEnd
)

type Formatter struct {
	symbolValues  map[string]string
	sentinelStart string
	sentinelEnd   string
}

func NewFormatter(symbols map[string]string) *Formatter {
	logtrace.LogWithFunctionName()
	return &Formatter{
		symbolValues:  symbols,
		sentinelStart: FormatSentinelStart,
		sentinelEnd:   FormatSentinelEnd,
	}
}

func (formatter *Formatter) Format(name string) string {
	logtrace.LogWithFunctionName()
	for symbol, value := range formatter.symbolValues {
		searchSymbol := formatter.sentinelStart + symbol + formatter.sentinelEnd
		name = strings.Replace(name, searchSymbol, value, -1)
	}

	return name
}
