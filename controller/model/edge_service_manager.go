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
	"time"
	"ztna-core/ztna/common/pb/edge_cmd_pb"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/command"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/storage/ast"
	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

func NewEdgeServiceManager(env Env) *EdgeServiceManager {
	logtrace.LogWithFunctionName()
	manager := &EdgeServiceManager{
		baseEntityManager: newBaseEntityManager[*EdgeService, *db.EdgeService](env, env.GetStores().EdgeService),
		detailLister:      &ServiceDetailLister{},
	}
	manager.impl = manager
	manager.detailLister.manager = manager

	RegisterManagerDecoder[*EdgeService](env, manager)

	return manager
}

type EdgeServiceManager struct {
	baseEntityManager[*EdgeService, *db.EdgeService]
	detailLister *ServiceDetailLister
}

func (self *EdgeServiceManager) GetDetailLister() *ServiceDetailLister {
	logtrace.LogWithFunctionName()
	return self.detailLister
}

func (self *EdgeServiceManager) GetEntityTypeId() string {
	logtrace.LogWithFunctionName()
	return "edgeServices"
}

func (self *EdgeServiceManager) newModelEntity() *EdgeService {
	logtrace.LogWithFunctionName()
	return &EdgeService{}
}

func (self *EdgeServiceManager) Create(entity *EdgeService, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchCreate[*EdgeService](self, entity, ctx)
}

func (self *EdgeServiceManager) ApplyCreate(cmd *command.CreateEntityCommand[*EdgeService], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	_, err := self.createEntity(cmd.Entity, ctx)
	return err
}

func (self *EdgeServiceManager) Update(entity *EdgeService, checker fields.UpdatedFields, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	if checker != nil {
		checker = checker.RemoveFields("encryptionRequired")
	}
	return DispatchUpdate[*EdgeService](self, entity, checker, ctx)
}

func (self *EdgeServiceManager) ApplyUpdate(cmd *command.UpdateEntityCommand[*EdgeService], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	return self.updateEntity(cmd.Entity, cmd.UpdatedFields, ctx)
}

func (self *EdgeServiceManager) ReadByName(name string) (*EdgeService, error) {
	logtrace.LogWithFunctionName()
	entity := &EdgeService{}
	nameIndex := self.env.GetStores().EdgeService.GetNameIndex()
	if err := self.readEntityWithIndex("name", []byte(name), nameIndex, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (self *EdgeServiceManager) readInTx(tx *bbolt.Tx, id string) (*ServiceDetail, error) {
	logtrace.LogWithFunctionName()
	entity := &ServiceDetail{}
	boltEntity := self.GetStore().GetEntityStrategy().NewEntity()
	found, err := self.GetStore().LoadEntity(tx, id, boltEntity)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, boltz.NewNotFoundError(self.GetStore().GetSingularEntityType(), "id", id)
	}

	if err = entity.fillFrom(self.env, tx, boltEntity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (self *EdgeServiceManager) ReadForIdentity(id string, identityId string, configTypes map[string]struct{}) (*ServiceDetail, error) {
	logtrace.LogWithFunctionName()
	var service *ServiceDetail
	err := self.GetDb().View(func(tx *bbolt.Tx) error {
		var err error
		service, err = self.ReadForIdentityInTx(tx, id, identityId, configTypes)
		return err
	})
	return service, err
}

func (self *EdgeServiceManager) IsDialableByIdentity(id string, identityId string) (bool, error) {
	logtrace.LogWithFunctionName()
	result := false
	err := self.GetDb().View(func(tx *bbolt.Tx) error {
		var err error
		result = self.env.GetStores().EdgeService.IsBindableByIdentity(tx, id, identityId)
		return err
	})
	return result, err
}

func (self *EdgeServiceManager) IsBindableByIdentity(id string, identityId string) (bool, error) {
	logtrace.LogWithFunctionName()
	result := false
	err := self.GetDb().View(func(tx *bbolt.Tx) error {
		var err error
		result = self.env.GetStores().EdgeService.IsBindableByIdentity(tx, id, identityId)
		return err
	})
	return result, err
}

func (self *EdgeServiceManager) ReadForIdentityInTx(tx *bbolt.Tx, id string, identityId string, configTypes map[string]struct{}) (*ServiceDetail, error) {
	logtrace.LogWithFunctionName()
	edgeServiceStore := self.env.GetStores().EdgeService
	identity, err := self.GetEnv().GetManagers().Identity.readInTx(tx, identityId)
	if err != nil {
		return nil, err
	}
	isBindable := edgeServiceStore.IsBindableByIdentity(tx, id, identityId)
	isDialable := edgeServiceStore.IsDialableByIdentity(tx, id, identityId)

	if !isBindable && !isDialable && !identity.IsAdmin { // admin can view services even if policies don't permit bind/dial {
		return nil, boltz.NewNotFoundError(self.GetStore().GetSingularEntityType(), "id", id)
	}

	result, err := self.readInTx(tx, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, boltz.NewNotFoundError(self.GetStore().GetSingularEntityType(), "id", id)
	}
	if isBindable {
		result.Permissions = append(result.Permissions, db.PolicyTypeBindName)
	}
	if isDialable {
		result.Permissions = append(result.Permissions, db.PolicyTypeDialName)
	}
	if result.Permissions == nil {
		// don't return results with no permissions, since some SDKs assume non-nil permissions
		result.Permissions = []string{db.PolicyTypeInvalidName}
	}

	if len(configTypes) > 0 {
		identityServiceConfigs := self.env.GetStores().Identity.LoadServiceConfigsByServiceAndType(tx, identityId, configTypes)
		self.mergeConfigs(tx, configTypes, result, identityServiceConfigs)
	}

	return result, err
}

func (self *EdgeServiceManager) PublicQueryForIdentity(sessionIdentity *Identity, configTypes map[string]struct{}, query ast.Query) (*ServiceListResult, error) {
	logtrace.LogWithFunctionName()
	if sessionIdentity.IsAdmin {
		return self.queryServices(query, sessionIdentity.Id, configTypes, true)
	}
	return self.QueryForIdentity(sessionIdentity.Id, configTypes, query)
}

func (self *EdgeServiceManager) QueryForIdentity(identityId string, configTypes map[string]struct{}, query ast.Query) (*ServiceListResult, error) {
	logtrace.LogWithFunctionName()
	return self.queryServices(query, identityId, configTypes, false)
}

func (self *EdgeServiceManager) queryServices(query ast.Query, identityId string, configTypes map[string]struct{}, isAdmin bool) (*ServiceListResult, error) {
	logtrace.LogWithFunctionName()
	result := &ServiceListResult{
		manager:     self,
		identityId:  identityId,
		configTypes: configTypes,
		isAdmin:     isAdmin,
	}
	if isAdmin {
		if err := self.PreparedListWithHandler(query, result.collect); err != nil {
			return nil, err
		}
	} else {
		cursorProvider := self.env.GetStores().Identity.GetIdentityServicesCursorProvider(identityId)
		if err := self.PreparedListIndexed(cursorProvider, query, result.collect); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (self *EdgeServiceManager) QueryRoleAttributes(queryString string) ([]string, *models.QueryMetaData, error) {
	logtrace.LogWithFunctionName()
	index := self.env.GetStores().EdgeService.GetRoleAttributesIndex()
	return self.queryRoleAttributes(index, queryString)
}

func (self *EdgeServiceManager) Marshall(entity *EdgeService) ([]byte, error) {
	logtrace.LogWithFunctionName()
	tags, err := edge_cmd_pb.EncodeTags(entity.Tags)
	if err != nil {
		return nil, err
	}

	msg := &edge_cmd_pb.Service{
		Id:                 entity.Id,
		Name:               entity.Name,
		MaxIdleTime:        int64(entity.MaxIdleTime),
		Tags:               tags,
		TerminatorStrategy: entity.TerminatorStrategy,
		RoleAttributes:     entity.RoleAttributes,
		Configs:            entity.Configs,
		EncryptionRequired: entity.EncryptionRequired,
	}

	return proto.Marshal(msg)
}

func (self *EdgeServiceManager) Unmarshall(bytes []byte) (*EdgeService, error) {
	logtrace.LogWithFunctionName()
	msg := &edge_cmd_pb.Service{}
	if err := proto.Unmarshal(bytes, msg); err != nil {
		return nil, err
	}

	return &EdgeService{
		BaseEntity: models.BaseEntity{
			Id:   msg.Id,
			Tags: edge_cmd_pb.DecodeTags(msg.Tags),
		},
		Name:               msg.Name,
		MaxIdleTime:        time.Duration(msg.MaxIdleTime),
		TerminatorStrategy: msg.TerminatorStrategy,
		RoleAttributes:     msg.RoleAttributes,
		Configs:            msg.Configs,
		EncryptionRequired: msg.EncryptionRequired,
	}, nil
}

type ServiceListResult struct {
	manager     *EdgeServiceManager
	Services    []*ServiceDetail
	identityId  string
	configTypes map[string]struct{}
	isAdmin     bool
	models.QueryMetaData
}

func (result *ServiceListResult) collect(tx *bbolt.Tx, ids []string, queryMetaData *models.QueryMetaData) error {
	logtrace.LogWithFunctionName()
	result.QueryMetaData = *queryMetaData
	var service *ServiceDetail
	var err error

	identityServiceConfigs := result.manager.env.GetStores().Identity.LoadServiceConfigsByServiceAndType(tx, result.identityId, result.configTypes)

	for _, key := range ids {
		// service permissions for admin & non-admin identities will be set according to policies
		service, err = result.manager.ReadForIdentityInTx(tx, key, result.identityId, result.configTypes)
		if err != nil {
			return err
		}
		result.manager.mergeConfigs(tx, result.configTypes, service, identityServiceConfigs)
		result.Services = append(result.Services, service)
	}
	return nil
}

func (self *EdgeServiceManager) mergeConfigs(tx *bbolt.Tx, configTypes map[string]struct{}, service *ServiceDetail,
	identityServiceConfigs map[string]map[string]map[string]interface{}) {
	logtrace.LogWithFunctionName()
	service.Config = map[string]map[string]interface{}{}

	_, wantsAll := configTypes["all"]

	configTypeStore := self.env.GetStores().ConfigType

	if len(configTypes) > 0 && len(service.Configs) > 0 {
		configStore := self.env.GetStores().Config
		for _, configId := range service.Configs {
			config, _ := configStore.LoadById(tx, configId)
			if config != nil {
				_, wantsConfig := configTypes[config.Type]
				if wantsAll || wantsConfig {
					service.Config[config.Type] = config.Data
				}
			}
		}
	}

	// inject overrides
	if serviceMap, ok := identityServiceConfigs[service.Id]; ok {
		for configTypeId, config := range serviceMap {
			wantsConfig := wantsAll
			if !wantsConfig {
				_, wantsConfig = configTypes[configTypeId]
			}
			if wantsConfig {
				service.Config[configTypeId] = config
			}
		}
	}

	for configTypeId, config := range service.Config {
		configTypeName := configTypeStore.GetName(tx, configTypeId)
		if configTypeName != nil {
			delete(service.Config, configTypeId)
			service.Config[*configTypeName] = config
		} else {
			pfxlog.Logger().Errorf("name for config type %v not found!", configTypeId)
		}
	}
}

type PolicyPostureChecks struct {
	PostureChecks []*PostureCheck
	PolicyType    db.PolicyType
	PolicyName    string
}

func (self *EdgeServiceManager) GetPolicyPostureChecks(identityId, serviceId string) map[string]*PolicyPostureChecks {
	logtrace.LogWithFunctionName()
	policyIdToChecks := map[string]*PolicyPostureChecks{}

	postureCheckCache := map[string]*PostureCheck{}

	servicePolicyStore := self.env.GetStores().ServicePolicy
	postureCheckLinks := servicePolicyStore.GetLinkCollection(db.EntityTypePostureChecks)
	serviceLinks := servicePolicyStore.GetLinkCollection(db.EntityTypeServices)

	policyNameSymbol := self.env.GetStores().ServicePolicy.GetSymbol(db.FieldName)
	policyTypeSymbol := self.env.GetStores().ServicePolicy.GetSymbol(db.FieldServicePolicyType)

	_ = self.GetDb().View(func(tx *bbolt.Tx) error {
		if !self.env.GetStores().PostureCheck.IterateIds(tx, ast.BoolNodeTrue).IsValid() {
			return nil
		}

		policyCursor := self.env.GetStores().Identity.GetRelatedEntitiesCursor(tx, identityId, db.EntityTypeServicePolicies, true)
		policyCursor = ast.NewFilteredCursor(policyCursor, func(policyId []byte) bool {
			return serviceLinks.IsLinked(tx, policyId, []byte(serviceId))
		})

		for policyCursor.IsValid() {
			policyIdBytes := policyCursor.Current()
			policyIdStr := string(policyIdBytes)
			policyCursor.Next()

			policyName := boltz.FieldToString(policyNameSymbol.Eval(tx, policyIdBytes))
			policyType := db.PolicyTypeDial
			if fieldType, policyTypeValue := policyTypeSymbol.Eval(tx, policyIdBytes); fieldType == boltz.TypeInt32 {
				policyType = db.GetPolicyTypeForId(*boltz.BytesToInt32(policyTypeValue))
			}

			//required to provide an entry for policies w/ no checks
			policyIdToChecks[policyIdStr] = &PolicyPostureChecks{
				PostureChecks: []*PostureCheck{},
				PolicyType:    policyType,
				PolicyName:    *policyName,
			}

			cursor := postureCheckLinks.IterateLinks(tx, policyIdBytes)
			for cursor.IsValid() {
				checkId := string(cursor.Current())
				if postureCheck, found := postureCheckCache[checkId]; !found {
					postureCheck, _ := self.env.GetManagers().PostureCheck.readInTx(tx, checkId)
					postureCheckCache[checkId] = postureCheck
					policyIdToChecks[policyIdStr].PostureChecks = append(policyIdToChecks[policyIdStr].PostureChecks, postureCheck)
				} else {
					policyIdToChecks[policyIdStr].PostureChecks = append(policyIdToChecks[policyIdStr].PostureChecks, postureCheck)
				}
				cursor.Next()
			}
		}
		return nil
	})

	return policyIdToChecks
}

type ServiceDetailLister struct {
	manager *EdgeServiceManager
}

func (self *ServiceDetailLister) GetListStore() boltz.Store {
	logtrace.LogWithFunctionName()
	return self.manager.GetListStore()
}

func (self *ServiceDetailLister) BaseLoadInTx(tx *bbolt.Tx, id string) (*ServiceDetail, error) {
	logtrace.LogWithFunctionName()
	return self.manager.readInTx(tx, id)
}

func (self *ServiceDetailLister) BasePreparedList(query ast.Query) (*models.EntityListResult[*ServiceDetail], error) {
	logtrace.LogWithFunctionName()
	result := &models.EntityListResult[*ServiceDetail]{
		Loader: self,
	}

	if err := self.manager.PreparedListWithHandler(query, result.Collect); err != nil {
		return nil, err
	}

	return result, nil
}

func (self *ServiceDetailLister) BasePreparedListIndexed(cursorProvider ast.SetCursorProvider, query ast.Query) (*models.EntityListResult[*ServiceDetail], error) {
	logtrace.LogWithFunctionName()
	result := &models.EntityListResult[*ServiceDetail]{
		Loader: self,
	}

	if err := self.manager.PreparedListIndexed(cursorProvider, query, result.Collect); err != nil {
		return nil, err
	}

	return result, nil
}
