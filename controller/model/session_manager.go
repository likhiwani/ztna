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
	"time"
	"ztna-core/ztna/common"
	"ztna-core/ztna/controller/apierror"
	fabricApiError "ztna-core/ztna/controller/apierror"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lucsky/cuid"
	"github.com/openziti/foundation/v2/errorz"
	"github.com/openziti/foundation/v2/stringz"
	"github.com/openziti/storage/ast"
	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
)

func NewSessionManager(env Env) *SessionManager {
	logtrace.LogWithFunctionName()
	manager := &SessionManager{
		baseEntityManager: newBaseEntityManager[*Session, *db.Session](env, env.GetStores().Session),
	}
	manager.impl = manager
	return manager
}

type SessionManager struct {
	baseEntityManager[*Session, *db.Session]
}

func (self *SessionManager) newModelEntity() *Session {
	logtrace.LogWithFunctionName()
	return &Session{}
}

type SessionPostureResult struct {
	Passed           bool
	Failure          *PostureSessionRequestFailure
	PassingPolicyIds []string
	Cause            *fabricApiError.GenericCauseError
}

func (self *SessionManager) EvaluatePostureForService(identityId, apiSessionId, sessionType, serviceId, serviceName string) *SessionPostureResult {
	logtrace.LogWithFunctionName()

	failureByPostureCheckId := map[string]*PostureCheckFailure{} //cache individual check status
	validPosture := false
	hasMatchingPolicies := false

	policyPostureCheckMap := self.GetEnv().GetManagers().EdgeService.GetPolicyPostureChecks(identityId, serviceId)

	failedPolicies := map[string][]*PostureCheckFailure{}
	failedPoliciesIdToName := map[string]string{}

	var failedPolicyIds []string
	var successPolicyIds []string

	for policyId, policyPostureCheck := range policyPostureCheckMap {

		if policyPostureCheck.PolicyType.String() != sessionType {
			continue
		}
		hasMatchingPolicies = true
		var failedChecks []*PostureCheckFailure

		for _, postureCheck := range policyPostureCheck.PostureChecks {

			found := false

			if _, found = failureByPostureCheckId[postureCheck.Id]; !found {
				_, failureByPostureCheckId[postureCheck.Id] = self.GetEnv().GetManagers().PostureResponse.Evaluate(identityId, apiSessionId, postureCheck)
			}

			if failureByPostureCheckId[postureCheck.Id] != nil {
				failedChecks = append(failedChecks, failureByPostureCheckId[postureCheck.Id])
			}
		}

		if len(failedChecks) == 0 {
			validPosture = true
			successPolicyIds = append(successPolicyIds, policyId)
		} else {
			//save for error output
			failedPolicies[policyId] = failedChecks
			failedPoliciesIdToName[policyId] = policyPostureCheck.PolicyName
			failedPolicyIds = append(failedPolicyIds, policyId)
		}
	}

	if hasMatchingPolicies && !validPosture {
		failureMap := map[string]interface{}{}

		sessionFailure := &PostureSessionRequestFailure{
			When:           time.Now(),
			ServiceId:      serviceId,
			ServiceName:    serviceName,
			ApiSessionId:   apiSessionId,
			SessionType:    sessionType,
			PolicyFailures: []*PosturePolicyFailure{},
		}

		for policyId, failures := range failedPolicies {
			policyFailure := &PosturePolicyFailure{
				PolicyId:   policyId,
				PolicyName: failedPoliciesIdToName[policyId],
				Checks:     failures,
			}

			var outFailures []interface{}

			for _, failure := range failures {
				outFailures = append(outFailures, failure.ToClientErrorData())
			}
			failureMap[policyId] = outFailures

			sessionFailure.PolicyFailures = append(sessionFailure.PolicyFailures, policyFailure)
		}

		cause := &fabricApiError.GenericCauseError{
			Message: fmt.Sprintf("Failed to pass posture checks for service policies: %v", failedPolicyIds),
			DataMap: failureMap,
		}

		return &SessionPostureResult{
			Passed:           false,
			Cause:            cause,
			PassingPolicyIds: nil,
			Failure:          sessionFailure,
		}
	}

	return &SessionPostureResult{
		Passed:           true,
		Cause:            nil,
		PassingPolicyIds: successPolicyIds,
		Failure:          nil,
	}
}

func (self *SessionManager) CreateJwt(entity *Session, ctx *change.Context) (string, error) {
	logtrace.LogWithFunctionName()
	entity.Id = uuid.New().String()

	service, err := self.GetEnv().GetManagers().EdgeService.ReadForIdentity(entity.ServiceId, entity.IdentityId, nil)
	if err != nil {
		return "", err
	}

	if entity.Type == "" {
		entity.Type = db.SessionTypeDial
	}

	if db.SessionTypeDial == entity.Type && !stringz.Contains(service.Permissions, db.PolicyTypeDialName) {
		return "", errorz.NewFieldError("service not found", "ServiceId", entity.ServiceId)
	}

	if db.SessionTypeBind == entity.Type && !stringz.Contains(service.Permissions, db.PolicyTypeBindName) {
		return "", errorz.NewFieldError("service not found", "ServiceId", entity.ServiceId)
	}

	policyResult := self.EvaluatePostureForService(entity.IdentityId, entity.ApiSessionId, entity.Type, service.Id, service.Name)

	if !policyResult.Passed {
		self.env.GetManagers().PostureResponse.postureCache.AddSessionRequestFailure(entity.IdentityId, policyResult.Failure)
		return "", apierror.NewInvalidPosture(policyResult.Cause)
	}

	edgeRouterAvailable, err := self.GetEnv().GetManagers().EdgeRouter.IsSharedEdgeRouterPresent(entity.IdentityId, entity.ServiceId)
	if err != nil {
		return "", err
	}

	if !edgeRouterAvailable {
		return "", apierror.NewNoEdgeRoutersAvailable()
	}
	entity.ServicePolicies = policyResult.PassingPolicyIds

	claims := common.ServiceAccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    self.env.RootIssuer(),
			Subject:   entity.ServiceId,
			Audience:  jwt.ClaimStrings{common.ClaimAudienceOpenZiti},
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
			ID:        entity.Id,
			ExpiresAt: &jwt.NumericDate{Time: time.Now().AddDate(1, 0, 0)}, //bound by API Session
		},
		ApiSessionId: entity.ApiSessionId,
		IdentityId:   entity.IdentityId,
		Type:         entity.Type,
		TokenType:    common.TokenTypeServiceAccess,
	}

	return self.env.GetServerJwtSigner().Generate(claims)
}

func (self *SessionManager) Create(entity *Session, ctx *change.Context) (string, error) {
	logtrace.LogWithFunctionName()
	if self.getExistingSessionEntity(entity) {
		return entity.Id, nil
	}

	entity.Id = cuid.New() //use cuids which are longer than shortids but are monotonic

	apiSession, err := self.GetEnv().GetManagers().ApiSession.Read(entity.ApiSessionId)
	if err != nil {
		return "", err
	}
	if apiSession == nil {
		return "", errorz.NewFieldError("api session not found", "ApiSessionId", entity.ApiSessionId)
	}

	service, err := self.GetEnv().GetManagers().EdgeService.ReadForIdentity(entity.ServiceId, apiSession.IdentityId, nil)
	if err != nil {
		return "", err
	}

	if entity.Type == "" {
		entity.Type = db.SessionTypeDial
	}

	if db.SessionTypeDial == entity.Type && !stringz.Contains(service.Permissions, db.PolicyTypeDialName) {
		return "", errorz.NewFieldError("service not found", "ServiceId", entity.ServiceId)
	}

	if db.SessionTypeBind == entity.Type && !stringz.Contains(service.Permissions, db.PolicyTypeBindName) {
		return "", errorz.NewFieldError("service not found", "ServiceId", entity.ServiceId)
	}

	policyResult := self.EvaluatePostureForService(apiSession.IdentityId, apiSession.Id, entity.Type, service.Id, service.Name)

	if !policyResult.Passed {
		self.env.GetManagers().PostureResponse.postureCache.AddSessionRequestFailure(apiSession.IdentityId, policyResult.Failure)
		return "", apierror.NewInvalidPosture(policyResult.Cause)
	}

	edgeRouterAvailable, err := self.GetEnv().GetManagers().EdgeRouter.IsSharedEdgeRouterPresent(apiSession.IdentityId, entity.ServiceId)
	if err != nil {
		return "", err
	}

	if !edgeRouterAvailable {
		return "", apierror.NewNoEdgeRoutersAvailable()
	}

	entity.ServicePolicies = policyResult.PassingPolicyIds

	return self.createSessionEntity(entity, ctx)
}

func (self *SessionManager) createSessionEntity(session *Session, ctx *change.Context) (string, error) {
	logtrace.LogWithFunctionName()
	var id string
	err := self.GetDb().Update(ctx.NewMutateContext(), func(ctx boltz.MutateContext) error {
		if self.getExistingSessionEntityInTx(ctx.Tx(), session) {
			id = session.Id
			return nil
		}

		var err error
		id, err = self.createEntityInTx(ctx, session)
		return err
	})
	if err != nil {
		return "", err
	}
	return id, nil

}

func (self *SessionManager) getExistingSessionEntity(session *Session) bool {
	logtrace.LogWithFunctionName()
	var result bool
	_ = self.GetDb().View(func(tx *bbolt.Tx) error {
		result = self.getExistingSessionEntityInTx(tx, session)
		return nil
	})
	return result
}

func (self *SessionManager) getExistingSessionEntityInTx(tx *bbolt.Tx, session *Session) bool {
	logtrace.LogWithFunctionName()
	sessionId := self.env.GetStores().ApiSession.GetCachedSessionId(tx, session.ApiSessionId, session.Type, session.ServiceId)
	if sessionId != nil {
		if existingSession, _ := self.readInTx(tx, *sessionId); existingSession != nil {
			session.Id = existingSession.Id
			session.Token = existingSession.Token
			return true
		}
	}
	return false
}

func (self *SessionManager) ReadByToken(token string) (*Session, error) {
	logtrace.LogWithFunctionName()
	modelSession := &Session{}
	tokenIndex := self.env.GetStores().Session.GetTokenIndex()
	if err := self.readEntityWithIndex("token", []byte(token), tokenIndex, modelSession); err != nil {
		return nil, err
	}
	return modelSession, nil
}

func (self *SessionManager) ReadForIdentity(id string, identityId string) (*Session, error) {
	logtrace.LogWithFunctionName()
	identity, err := self.GetEnv().GetManagers().Identity.Read(identityId)

	if err != nil {
		return nil, err
	}
	if identity.IsAdmin {
		return self.Read(id)
	}

	query := fmt.Sprintf(`id = "%v" and apiSession.identity = "%v"`, id, identityId)
	result, err := self.Query(query)
	if err != nil {
		return nil, err
	}
	if len(result.Sessions) == 0 {
		return nil, boltz.NewNotFoundError(self.GetStore().GetSingularEntityType(), "id", id)
	}
	return result.Sessions[0], nil
}

func (self *SessionManager) Read(id string) (*Session, error) {
	logtrace.LogWithFunctionName()
	entity := &Session{}
	if err := self.readEntity(id, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (self *SessionManager) DeleteForIdentity(id, identityId string, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	session, err := self.ReadForIdentity(id, identityId)
	if err != nil {
		return err
	}
	if session == nil {
		return boltz.NewNotFoundError(self.GetStore().GetSingularEntityType(), "id", id)
	}
	return self.deleteEntity(id, ctx)
}

func (self *SessionManager) Delete(id string, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return self.deleteEntity(id, ctx)
}

func (self *SessionManager) PublicQueryForIdentity(sessionIdentity *Identity, query ast.Query) (*SessionListResult, error) {
	logtrace.LogWithFunctionName()
	if sessionIdentity.IsAdmin {
		return self.querySessions(query)
	}
	identityFilterString := fmt.Sprintf(`apiSession.identity = "%v"`, sessionIdentity.Id)
	identityFilter, err := ast.Parse(self.Store, identityFilterString)
	if err != nil {
		return nil, err
	}
	query.SetPredicate(ast.NewAndExprNode(query.GetPredicate(), identityFilter))
	return self.querySessions(query)
}

func (self *SessionManager) Query(query string) (*SessionListResult, error) {
	logtrace.LogWithFunctionName()
	result := &SessionListResult{manager: self}
	err := self.ListWithHandler(query, result.collect)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (self *SessionManager) querySessions(query ast.Query) (*SessionListResult, error) {
	logtrace.LogWithFunctionName()
	result := &SessionListResult{manager: self}
	err := self.PreparedListWithHandler(query, result.collect)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (self *SessionManager) ListSessionsForEdgeRouter(edgeRouterId string) (*SessionListResult, error) {
	logtrace.LogWithFunctionName()
	result := &SessionListResult{manager: self}
	query := fmt.Sprintf(`anyOf(apiSession.identity.edgeRouterPolicies.routers) = "%v" and `+
		`anyOf(service.serviceEdgeRouterPolicies.routers) = "%v"`, edgeRouterId, edgeRouterId)
	err := self.ListWithHandler(query, result.collect)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type SessionListResult struct {
	manager  *SessionManager
	Sessions []*Session
	models.QueryMetaData
}

func (result *SessionListResult) collect(tx *bbolt.Tx, ids []string, queryMetaData *models.QueryMetaData) error {
	logtrace.LogWithFunctionName()
	result.QueryMetaData = *queryMetaData
	for _, key := range ids {
		entity, err := result.manager.readInTx(tx, key)
		if err != nil {
			return err
		}
		result.Sessions = append(result.Sessions, entity)
	}
	return nil
}
