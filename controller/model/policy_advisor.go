package model

import (
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/foundation/v2/stringz"
	"go.etcd.io/bbolt"
)

func NewPolicyAdvisor(env Env) *PolicyAdvisor {
	logtrace.LogWithFunctionName()
	result := &PolicyAdvisor{
		env: env,
	}
	return result
}

type PolicyAdvisor struct {
	env Env
}

type AdvisorEdgeRouter struct {
	Router   *EdgeRouter
	IsOnline bool
}

type AdvisorServiceReachability struct {
	Identity            *Identity
	Service             *EdgeService
	IsBindAllowed       bool
	IsDialAllowed       bool
	IdentityRouterCount int
	ServiceRouterCount  int
	CommonRouters       []*AdvisorEdgeRouter
}

func (advisor *PolicyAdvisor) AnalyzeServiceReachability(identityId, serviceId string) (*AdvisorServiceReachability, error) {
	logtrace.LogWithFunctionName()
	identity, err := advisor.env.GetManagers().Identity.Read(identityId)
	if err != nil {
		return nil, err
	}

	service, err := advisor.env.GetManagers().EdgeService.Read(serviceId)
	if err != nil {
		return nil, err
	}

	permissions, err := advisor.getServicePermissions(identityId, serviceId)

	if err != nil {
		return nil, err
	}

	edgeRouters, err := advisor.getIdentityEdgeRouters(identityId)
	if err != nil {
		return nil, err
	}

	serviceEdgeRouters, err := advisor.getServiceEdgeRouters(serviceId)
	if err != nil {
		return nil, err
	}

	result := &AdvisorServiceReachability{
		Identity:            identity,
		Service:             service,
		IsBindAllowed:       stringz.Contains(permissions, db.PolicyTypeBindName),
		IsDialAllowed:       stringz.Contains(permissions, db.PolicyTypeDialName),
		IdentityRouterCount: len(edgeRouters),
		ServiceRouterCount:  len(serviceEdgeRouters),
	}

	for edgeRouterId := range serviceEdgeRouters {
		if edgeRouter, ok := edgeRouters[edgeRouterId]; ok {
			result.CommonRouters = append(result.CommonRouters, edgeRouter)
		}
	}

	return result, nil
}

func (advisor *PolicyAdvisor) getServicePermissions(identityId, serviceId string) ([]string, error) {
	logtrace.LogWithFunctionName()
	var permissions []string

	servicePolicyStore := advisor.env.GetStores().ServicePolicy
	servicePolicyIterator := func(tx *bbolt.Tx, servicePolicyId string) error {
		servicePolicy, err := servicePolicyStore.LoadById(tx, servicePolicyId)
		if err != nil {
			return err
		}
		if servicePolicyStore.IsEntityRelated(tx, servicePolicyId, db.EntityTypeServices, serviceId) {
			if !stringz.Contains(permissions, string(servicePolicy.PolicyType)) {
				permissions = append(permissions, string(servicePolicy.PolicyType))
			}
		}
		return nil
	}

	if err := advisor.env.GetManagers().Identity.iterateRelatedEntities(identityId, db.EntityTypeServicePolicies, servicePolicyIterator); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (advisor *PolicyAdvisor) getIdentityEdgeRouters(identityId string) (map[string]*AdvisorEdgeRouter, error) {
	logtrace.LogWithFunctionName()
	edgeRouters := map[string]*AdvisorEdgeRouter{}

	edgeRouterPolicyIterator := func(tx *bbolt.Tx, edgeRouterPolicyId string) error {
		edgeRouterIterator := func(tx *bbolt.Tx, edgeRouterId string) error {
			commonRouter := edgeRouters[edgeRouterId]
			if commonRouter == nil {
				edgeRouter, err := advisor.env.GetManagers().EdgeRouter.readInTx(tx, edgeRouterId)
				if err != nil {
					return err
				}
				commonRouter = &AdvisorEdgeRouter{
					Router:   edgeRouter,
					IsOnline: advisor.env.IsEdgeRouterOnline(edgeRouter.Id),
				}
				edgeRouters[edgeRouterId] = commonRouter
			}

			return nil
		}

		return advisor.env.GetManagers().EdgeRouterPolicy.iterateRelatedEntitiesInTx(tx, edgeRouterPolicyId, db.EntityTypeRouters, edgeRouterIterator)
	}
	if err := advisor.env.GetManagers().Identity.iterateRelatedEntities(identityId, db.EntityTypeEdgeRouterPolicies, edgeRouterPolicyIterator); err != nil {
		return nil, err
	}

	return edgeRouters, nil
}

func (advisor *PolicyAdvisor) getServiceEdgeRouters(serviceId string) (map[string]struct{}, error) {
	logtrace.LogWithFunctionName()
	edgeRouters := map[string]struct{}{}

	serviceEdgeRouterPolicyIterator := func(tx *bbolt.Tx, policyId string) error {
		edgeRouterIterator := func(tx *bbolt.Tx, edgeRouterId string) error {
			edgeRouters[edgeRouterId] = struct{}{}
			return nil
		}
		return advisor.env.GetManagers().ServiceEdgeRouterPolicy.iterateRelatedEntitiesInTx(tx, policyId, db.EntityTypeRouters, edgeRouterIterator)
	}

	if err := advisor.env.GetManagers().EdgeService.iterateRelatedEntities(serviceId, db.EntityTypeServiceEdgeRouterPolicies, serviceEdgeRouterPolicyIterator); err != nil {
		return nil, err
	}

	return edgeRouters, nil
}

type AdvisorIdentityEdgeRouterLinks struct {
	Identity   *Identity
	EdgeRouter *EdgeRouter
	Policies   []*EdgeRouterPolicy
}

func (advisor *PolicyAdvisor) InspectIdentityEdgeRouterLinks(identityId, edgeRouterId string) (*AdvisorIdentityEdgeRouterLinks, error) {
	logtrace.LogWithFunctionName()
	identity, err := advisor.env.GetManagers().Identity.Read(identityId)
	if err != nil {
		return nil, err
	}

	edgeRouter, err := advisor.env.GetManagers().EdgeRouter.Read(edgeRouterId)
	if err != nil {
		return nil, err
	}

	policies, err := advisor.getEdgeRouterPolicies(identityId, edgeRouterId)
	if err != nil {
		return nil, err
	}

	result := &AdvisorIdentityEdgeRouterLinks{
		Identity:   identity,
		EdgeRouter: edgeRouter,
		Policies:   policies,
	}

	return result, nil
}

func (advisor *PolicyAdvisor) getEdgeRouterPolicies(identityId, edgeRouterId string) ([]*EdgeRouterPolicy, error) {
	logtrace.LogWithFunctionName()
	var result []*EdgeRouterPolicy

	policyStore := advisor.env.GetStores().EdgeRouterPolicy
	policyIterator := func(tx *bbolt.Tx, policyId string) error {
		policy, err := advisor.env.GetManagers().EdgeRouterPolicy.readInTx(tx, policyId)
		if err != nil {
			return err
		}
		if policyStore.IsEntityRelated(tx, policyId, db.EntityTypeRouters, edgeRouterId) {
			result = append(result, policy)
		}
		return nil
	}

	if err := advisor.env.GetManagers().Identity.iterateRelatedEntities(identityId, db.EntityTypeEdgeRouterPolicies, policyIterator); err != nil {
		return nil, err
	}

	return result, nil
}

type AdvisorIdentityServiceLinks struct {
	Identity *Identity
	Service  *EdgeService
	Policies []*ServicePolicy
}

func (advisor *PolicyAdvisor) InspectIdentityServiceLinks(identityId, serviceId string) (*AdvisorIdentityServiceLinks, error) {
	logtrace.LogWithFunctionName()
	identity, err := advisor.env.GetManagers().Identity.Read(identityId)
	if err != nil {
		return nil, err
	}

	service, err := advisor.env.GetManagers().EdgeService.Read(serviceId)
	if err != nil {
		return nil, err
	}

	policies, err := advisor.getServicePolicies(identityId, serviceId)
	if err != nil {
		return nil, err
	}

	result := &AdvisorIdentityServiceLinks{
		Identity: identity,
		Service:  service,
		Policies: policies,
	}

	return result, nil
}

func (advisor *PolicyAdvisor) getServicePolicies(identityId, serviceId string) ([]*ServicePolicy, error) {
	logtrace.LogWithFunctionName()
	var result []*ServicePolicy

	policyStore := advisor.env.GetStores().ServicePolicy
	policyIterator := func(tx *bbolt.Tx, policyId string) error {
		policy, err := advisor.env.GetManagers().ServicePolicy.readInTx(tx, policyId)
		if err != nil {
			return err
		}
		if policyStore.IsEntityRelated(tx, policyId, db.EntityTypeServices, serviceId) {
			result = append(result, policy)
		}
		return nil
	}

	if err := advisor.env.GetManagers().Identity.iterateRelatedEntities(identityId, db.EntityTypeServicePolicies, policyIterator); err != nil {
		return nil, err
	}

	return result, nil
}

type AdvisorServiceEdgeRouterLinks struct {
	Service    *EdgeService
	EdgeRouter *EdgeRouter
	Policies   []*ServiceEdgeRouterPolicy
}

func (advisor *PolicyAdvisor) InspectServiceEdgeRouterLinks(serviceId, edgeRouterId string) (*AdvisorServiceEdgeRouterLinks, error) {
	logtrace.LogWithFunctionName()
	service, err := advisor.env.GetManagers().EdgeService.Read(serviceId)
	if err != nil {
		return nil, err
	}

	edgeRouter, err := advisor.env.GetManagers().EdgeRouter.Read(edgeRouterId)
	if err != nil {
		return nil, err
	}

	policies, err := advisor.getServiceEdgeRouterPolicies(serviceId, edgeRouterId)
	if err != nil {
		return nil, err
	}

	result := &AdvisorServiceEdgeRouterLinks{
		Service:    service,
		EdgeRouter: edgeRouter,
		Policies:   policies,
	}

	return result, nil
}

func (advisor *PolicyAdvisor) getServiceEdgeRouterPolicies(serviceId, edgeRouterId string) ([]*ServiceEdgeRouterPolicy, error) {
	logtrace.LogWithFunctionName()
	var result []*ServiceEdgeRouterPolicy

	policyStore := advisor.env.GetStores().ServiceEdgeRouterPolicy
	policyIterator := func(tx *bbolt.Tx, policyId string) error {
		policy, err := advisor.env.GetManagers().ServiceEdgeRouterPolicy.readInTx(tx, policyId)
		if err != nil {
			return err
		}
		if policyStore.IsEntityRelated(tx, policyId, db.EntityTypeRouters, edgeRouterId) {
			result = append(result, policy)
		}
		return nil
	}

	if err := advisor.env.GetManagers().EdgeService.iterateRelatedEntities(serviceId, db.EntityTypeServiceEdgeRouterPolicies, policyIterator); err != nil {
		return nil, err
	}

	return result, nil
}
