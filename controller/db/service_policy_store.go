package db

import (
	"fmt"
	"sort"
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/foundation/v2/errorz"
	"github.com/openziti/foundation/v2/stringz"
	"github.com/openziti/storage/ast"
	"github.com/openziti/storage/boltz"
)

type PolicyType string

func (self PolicyType) String() string {
	logtrace.LogWithFunctionName()
	return string(self)
}

func (self PolicyType) Id() int32 {
	logtrace.LogWithFunctionName()
	if self == PolicyTypeDial {
		return 1
	}
	if self == PolicyTypeBind {
		return 2
	}
	return 0
}

func (self PolicyType) IsDial() bool {
	logtrace.LogWithFunctionName()
	return self == PolicyTypeDial
}

func (self PolicyType) IsBind() bool {
	logtrace.LogWithFunctionName()
	return self == PolicyTypeBind
}

func GetPolicyTypeForId(policyTypeId int32) PolicyType {
	logtrace.LogWithFunctionName()
	policyType := PolicyTypeInvalid
	if policyTypeId == PolicyTypeDial.Id() {
		policyType = PolicyTypeDial
	} else if policyTypeId == PolicyTypeBind.Id() {
		policyType = PolicyTypeBind
	}
	return policyType
}

const (
	FieldServicePolicyType = "type"

	PolicyTypeInvalidName = "Invalid"
	PolicyTypeDialName    = "Dial"
	PolicyTypeBindName    = "Bind"

	PolicyTypeInvalid PolicyType = PolicyTypeInvalidName
	PolicyTypeDial    PolicyType = PolicyTypeDialName
	PolicyTypeBind    PolicyType = PolicyTypeBindName
)

type ServicePolicy struct {
	boltz.BaseExtEntity
	PolicyType        PolicyType `json:"policyType"`
	Name              string     `json:"name"`
	Semantic          string     `json:"semantic"`
	IdentityRoles     []string   `json:"identityRoles"`
	ServiceRoles      []string   `json:"serviceRoles"`
	PostureCheckRoles []string   `json:"postureCheckRoles"`
}

func (entity *ServicePolicy) GetName() string {
	logtrace.LogWithFunctionName()
	return entity.Name
}

func (entity *ServicePolicy) GetSemantic() string {
	logtrace.LogWithFunctionName()
	return entity.Semantic
}

func (entity *ServicePolicy) GetEntityType() string {
	logtrace.LogWithFunctionName()
	return EntityTypeServicePolicies
}

type ServicePolicyChangeEventListener func(event *edge_ctrl_pb.DataState_ServicePolicyChange)

var _ ServicePolicyStore = (*servicePolicyStoreImpl)(nil)

type ServicePolicyStore interface {
	NameIndexed
	Store[*ServicePolicy]
}

func newServicePolicyStore(stores *stores) *servicePolicyStoreImpl {
	logtrace.LogWithFunctionName()
	store := &servicePolicyStoreImpl{}
	store.baseStore = newBaseStore[*ServicePolicy](stores, store)
	store.InitImpl(store)
	return store
}

type servicePolicyStoreImpl struct {
	*baseStore[*ServicePolicy]

	indexName        boltz.ReadIndex
	symbolPolicyType boltz.EntitySymbol
	symbolSemantic   boltz.EntitySymbol

	symbolIdentityRoles     boltz.EntitySetSymbol
	symbolServiceRoles      boltz.EntitySetSymbol
	symbolPostureCheckRoles boltz.EntitySetSymbol

	symbolIdentities    boltz.EntitySetSymbol
	symbolServices      boltz.EntitySetSymbol
	symbolPostureChecks boltz.EntitySetSymbol

	identityCollection     boltz.LinkCollection
	serviceCollection      boltz.LinkCollection
	postureCheckCollection boltz.LinkCollection
}

func (store *servicePolicyStoreImpl) GetNameIndex() boltz.ReadIndex {
	logtrace.LogWithFunctionName()
	return store.indexName
}

func (store *servicePolicyStoreImpl) NewEntity() *ServicePolicy {
	logtrace.LogWithFunctionName()
	return &ServicePolicy{}
}

func (store *servicePolicyStoreImpl) initializeLocal() {
	logtrace.LogWithFunctionName()
	store.AddExtEntitySymbols()

	store.indexName = store.addUniqueNameField()
	store.symbolPolicyType = store.AddSymbol(FieldServicePolicyType, ast.NodeTypeInt64)
	store.symbolSemantic = store.AddSymbol(FieldSemantic, ast.NodeTypeString)

	store.symbolIdentityRoles = store.AddPublicSetSymbol(FieldIdentityRoles, ast.NodeTypeString)
	store.symbolServiceRoles = store.AddPublicSetSymbol(FieldServiceRoles, ast.NodeTypeString)
	store.symbolPostureCheckRoles = store.AddPublicSetSymbol(FieldPostureCheckRoles, ast.NodeTypeString)

	store.symbolIdentities = store.AddFkSetSymbol(EntityTypeIdentities, store.stores.identity)
	store.symbolServices = store.AddFkSetSymbol(EntityTypeServices, store.stores.edgeService)
	store.symbolPostureChecks = store.AddFkSetSymbol(EntityTypePostureChecks, store.stores.postureCheck)

	store.MakeSymbolPublic(EntityTypeIdentities)
	store.MakeSymbolPublic(EntityTypeServices)
}

func (store *servicePolicyStoreImpl) initializeLinked() {
	logtrace.LogWithFunctionName()
	store.serviceCollection = store.AddLinkCollection(store.symbolServices, store.stores.edgeService.symbolServicePolicies)
	store.identityCollection = store.AddLinkCollection(store.symbolIdentities, store.stores.identity.symbolServicePolicies)
	store.postureCheckCollection = store.AddLinkCollection(store.symbolPostureChecks, store.stores.postureCheck.symbolServicePolicies)
}

func (store *servicePolicyStoreImpl) FillEntity(entity *ServicePolicy, bucket *boltz.TypedBucket) {
	logtrace.LogWithFunctionName()
	entity.LoadBaseValues(bucket)
	entity.Name = bucket.GetStringOrError(FieldName)
	entity.PolicyType = GetPolicyTypeForId(bucket.GetInt32WithDefault(FieldServicePolicyType, PolicyTypeDial.Id()))
	entity.Semantic = bucket.GetStringWithDefault(FieldSemantic, SemanticAllOf)
	entity.IdentityRoles = bucket.GetStringList(FieldIdentityRoles)
	entity.ServiceRoles = bucket.GetStringList(FieldServiceRoles)
	entity.PostureCheckRoles = bucket.GetStringList(FieldPostureCheckRoles)
}

func (store *servicePolicyStoreImpl) PersistEntity(entity *ServicePolicy, ctx *boltz.PersistContext) {
	logtrace.LogWithFunctionName()
	policyTypeChanged := false

	currentPolicyType := GetPolicyTypeForId(ctx.Bucket.GetInt32WithDefault(FieldServicePolicyType, PolicyTypeDial.Id()))
	if ctx.ProceedWithSet(FieldServicePolicyType) {
		if entity.PolicyType != PolicyTypeBind && entity.PolicyType != PolicyTypeDial {
			ctx.Bucket.SetError(errorz.NewFieldError("invalid policy type", FieldServicePolicyType, entity.PolicyType))
			return
		}
		policyTypeChanged = entity.PolicyType != currentPolicyType
	} else {
		// PolicyType needs to be correct in the entity as we use it later
		// TODO: Add test for this
		entity.PolicyType = currentPolicyType
	}

	if err := validateRolesAndIds(FieldIdentityRoles, entity.IdentityRoles); err != nil {
		ctx.Bucket.SetError(err)
	}

	if err := validateRolesAndIds(FieldServiceRoles, entity.ServiceRoles); err != nil {
		ctx.Bucket.SetError(err)
	}

	if err := validateRolesAndIds(FieldPostureCheckRoles, entity.PostureCheckRoles); err != nil {
		ctx.Bucket.SetError(err)
	}

	if ctx.ProceedWithSet(FieldSemantic) && !isSemanticValid(entity.Semantic) {
		ctx.Bucket.SetError(errorz.NewFieldError("invalid semantic", FieldSemantic, entity.Semantic))
		return
	}

	entity.SetBaseValues(ctx)
	ctx.SetRequiredString(FieldName, entity.Name)
	ctx.SetInt32(FieldServicePolicyType, entity.PolicyType.Id())
	ctx.SetRequiredString(FieldSemantic, entity.Semantic)
	servicePolicyStore := ctx.Store.(*servicePolicyStoreImpl)

	sort.Strings(entity.ServiceRoles)
	sort.Strings(entity.IdentityRoles)
	sort.Strings(entity.PostureCheckRoles)

	if policyTypeChanged && !ctx.IsCreate {
		// if the policy type has changed, we need to remove all roles for the old policy type and then all the roles
		// for the new policy type

		updatedFields := ctx.FieldChecker
		if updatedFields == nil {
			updatedFields = FieldCheckerF(func(s string) bool {
				return true
			})
		}

		ctx.FieldChecker = FieldCheckerF(func(s string) bool {
			return s == FieldIdentityRoles || s == FieldServiceRoles || s == FieldPostureCheckRoles || updatedFields.IsUpdated(s)
		})

		newIdentityRoles := entity.IdentityRoles
		newServiceRoles := entity.ServiceRoles
		newPostureCheckRoles := entity.PostureCheckRoles

		entity.IdentityRoles = nil
		entity.ServiceRoles = nil
		entity.PostureCheckRoles = nil

		currentIdentityRoles, _ := ctx.GetAndSetStringList(FieldIdentityRoles, entity.IdentityRoles)
		currentServiceRoles, _ := ctx.GetAndSetStringList(FieldServiceRoles, entity.ServiceRoles)
		currentPostureCheckRoles, _ := ctx.GetAndSetStringList(FieldPostureCheckRoles, entity.PostureCheckRoles)

		if !updatedFields.IsUpdated(FieldIdentityRoles) {
			newIdentityRoles = currentIdentityRoles
		}

		if !updatedFields.IsUpdated(FieldServiceRoles) {
			newServiceRoles = currentServiceRoles
		}

		if !updatedFields.IsUpdated(FieldPostureCheckRoles) {
			newPostureCheckRoles = currentPostureCheckRoles
		}

		newPolicyType := entity.PolicyType
		entity.PolicyType = currentPolicyType

		servicePolicyStore.identityRolesUpdated(ctx, entity)
		servicePolicyStore.serviceRolesUpdated(ctx, entity)
		servicePolicyStore.postureCheckRolesUpdated(ctx, entity)

		entity.PolicyType = newPolicyType
		entity.IdentityRoles = newIdentityRoles
		entity.ServiceRoles = newServiceRoles
		entity.PostureCheckRoles = newPostureCheckRoles

		_, _ = ctx.GetAndSetStringList(FieldIdentityRoles, entity.IdentityRoles)
		_, _ = ctx.GetAndSetStringList(FieldServiceRoles, entity.ServiceRoles)

		servicePolicyStore.identityRolesUpdated(ctx, entity)
		servicePolicyStore.serviceRolesUpdated(ctx, entity)
		servicePolicyStore.postureCheckRolesUpdated(ctx, entity)
	} else {
		currentIdentityRoles, identityRolesSet := ctx.GetAndSetStringList(FieldIdentityRoles, entity.IdentityRoles)
		currentServiceRoles, serviceRolesSet := ctx.GetAndSetStringList(FieldServiceRoles, entity.ServiceRoles)
		currentPostureCheckRoles, postureCheckRolesSet := ctx.GetAndSetStringList(FieldPostureCheckRoles, entity.PostureCheckRoles)

		if identityRolesSet && !stringz.EqualSlices(currentIdentityRoles, entity.IdentityRoles) {
			servicePolicyStore.identityRolesUpdated(ctx, entity)
		}

		if serviceRolesSet && !stringz.EqualSlices(currentServiceRoles, entity.ServiceRoles) {
			servicePolicyStore.serviceRolesUpdated(ctx, entity)
		}
		if postureCheckRolesSet && !stringz.EqualSlices(currentPostureCheckRoles, entity.PostureCheckRoles) {
			servicePolicyStore.postureCheckRolesUpdated(ctx, entity)
		}
	}
}

/*
Optimizations
 1. When changing policies if only ids have changed, only add/remove ids from groups as needed
 2. When related entities added/changed, only evaluate policies against that one entity (identity/edge router/service),
    and just add/remove/ignore
 3. Related entity deletes should be handled automatically by FK Indexes on those entities (need to verify the reverse as well/deleting policy)
*/
func (store *servicePolicyStoreImpl) serviceRolesUpdated(persistCtx *boltz.PersistContext, policy *ServicePolicy) {
	logtrace.LogWithFunctionName()
	ctx := &roleAttributeChangeContext{
		mutateCtx:             persistCtx.MutateContext,
		rolesSymbol:           store.symbolServiceRoles,
		linkCollection:        store.serviceCollection,
		relatedLinkCollection: store.identityCollection,
		ErrorHolder:           persistCtx.Bucket,
	}
	if policy.PolicyType == PolicyTypeDial {
		ctx.denormLinkCollection = store.stores.edgeService.dialIdentitiesCollection
		ctx.denormChangeHandler = func(fromId, toId []byte, add bool) {
			ctx.addServicePolicyEvent(toId, fromId, PolicyTypeDial, add)
		}
	} else {
		ctx.denormLinkCollection = store.stores.edgeService.bindIdentitiesCollection
		ctx.denormChangeHandler = func(fromId, toId []byte, add bool) {
			ctx.addServicePolicyEvent(toId, fromId, PolicyTypeBind, add)
		}
	}

	ctx.changeHandler = func(policyId []byte, relatedId []byte, add bool) {
		ctx.notifyOfPolicyChangeEvent(policyId, relatedId, edge_ctrl_pb.ServicePolicyRelatedEntityType_RelatedService, add)
	}

	EvaluatePolicy(ctx, policy, store.stores.edgeService.symbolRoleAttributes)
}

func (store *servicePolicyStoreImpl) identityRolesUpdated(persistCtx *boltz.PersistContext, policy *ServicePolicy) {
	logtrace.LogWithFunctionName()
	ctx := &roleAttributeChangeContext{
		mutateCtx:             persistCtx.MutateContext,
		rolesSymbol:           store.symbolIdentityRoles,
		linkCollection:        store.identityCollection,
		relatedLinkCollection: store.serviceCollection,
		ErrorHolder:           persistCtx.Bucket,
	}

	if policy.PolicyType == PolicyTypeDial {
		ctx.denormLinkCollection = store.stores.identity.dialServicesCollection
		ctx.denormChangeHandler = func(fromId, toId []byte, add bool) {
			ctx.addServicePolicyEvent(fromId, toId, PolicyTypeDial, add)
		}
	} else {
		ctx.denormLinkCollection = store.stores.identity.bindServicesCollection
		ctx.denormChangeHandler = func(fromId, toId []byte, add bool) {
			ctx.addServicePolicyEvent(fromId, toId, PolicyTypeBind, add)
		}
	}

	ctx.changeHandler = func(policyId []byte, relatedId []byte, add bool) {
		ctx.notifyOfPolicyChangeEvent(policyId, relatedId, edge_ctrl_pb.ServicePolicyRelatedEntityType_RelatedIdentity, add)
	}

	EvaluatePolicy(ctx, policy, store.stores.identity.symbolRoleAttributes)
}

func (store *servicePolicyStoreImpl) postureCheckRolesUpdated(persistCtx *boltz.PersistContext, policy *ServicePolicy) {
	logtrace.LogWithFunctionName()
	ctx := &roleAttributeChangeContext{
		mutateCtx:             persistCtx.MutateContext,
		rolesSymbol:           store.symbolPostureCheckRoles,
		linkCollection:        store.postureCheckCollection,
		relatedLinkCollection: store.serviceCollection,
		ErrorHolder:           persistCtx.Bucket,
	}

	ctx.denormChangeHandler = func(fromId, toId []byte, add bool) {
		ctx.addServiceUpdatedEvent(store.stores, ctx.tx(), toId)
	}

	if policy.PolicyType == PolicyTypeDial {
		ctx.denormLinkCollection = store.stores.postureCheck.dialServicesCollection
	} else {
		ctx.denormLinkCollection = store.stores.postureCheck.bindServicesCollection
	}

	ctx.changeHandler = func(policyId []byte, relatedId []byte, add bool) {
		ctx.notifyOfPolicyChangeEvent(policyId, relatedId, edge_ctrl_pb.ServicePolicyRelatedEntityType_RelatedPostureCheck, add)
	}

	EvaluatePolicy(ctx, policy, store.stores.postureCheck.symbolRoleAttributes)
}

func (store *servicePolicyStoreImpl) DeleteById(ctx boltz.MutateContext, id string) error {
	logtrace.LogWithFunctionName()
	policy, err := store.LoadById(ctx.Tx(), id)
	if err != nil {
		return err
	}
	policy.IdentityRoles = nil
	policy.ServiceRoles = nil
	policy.PostureCheckRoles = nil

	err = store.Update(ctx, policy, nil)
	if err != nil {
		return fmt.Errorf("failure while clearing policy before delete: %w", err)
	}
	return store.BaseStore.DeleteById(ctx, id)
}

func (store *servicePolicyStoreImpl) CheckIntegrity(mutateCtx boltz.MutateContext, fix bool, errorSink func(err error, fixed bool)) error {
	logtrace.LogWithFunctionName()
	ctx := &denormCheckCtx{
		name:                   "service-policies/bind",
		mutateCtx:              mutateCtx,
		sourceStore:            store.stores.identity,
		targetStore:            store.stores.edgeService,
		policyStore:            store,
		sourceCollection:       store.identityCollection,
		targetCollection:       store.serviceCollection,
		targetDenormCollection: store.stores.identity.bindServicesCollection,
		errorSink:              errorSink,
		repair:                 fix,
		policyFilter: func(policyId []byte) bool {
			policyType := PolicyTypeInvalid
			if result := boltz.FieldToInt32(store.symbolPolicyType.Eval(mutateCtx.Tx(), policyId)); result != nil {
				policyType = GetPolicyTypeForId(*result)
			}
			return policyType == PolicyTypeBind
		},
	}
	if err := validatePolicyDenormalization(ctx); err != nil {
		return err
	}

	ctx = &denormCheckCtx{
		name:                   "service-policies/dial",
		mutateCtx:              mutateCtx,
		sourceStore:            store.stores.identity,
		targetStore:            store.stores.edgeService,
		policyStore:            store,
		sourceCollection:       store.identityCollection,
		targetCollection:       store.serviceCollection,
		targetDenormCollection: store.stores.identity.dialServicesCollection,
		errorSink:              errorSink,
		repair:                 fix,
		policyFilter: func(policyId []byte) bool {
			policyType := PolicyTypeInvalid
			if result := boltz.FieldToInt32(store.symbolPolicyType.Eval(mutateCtx.Tx(), policyId)); result != nil {
				policyType = GetPolicyTypeForId(*result)
			}
			return policyType == PolicyTypeDial
		},
	}

	if err := validatePolicyDenormalization(ctx); err != nil {
		return err
	}

	return store.BaseStore.CheckIntegrity(mutateCtx, fix, errorSink)
}

type FieldCheckerF func(string) bool

func (f FieldCheckerF) IsUpdated(s string) bool {
	logtrace.LogWithFunctionName()
	return f(s)
}
