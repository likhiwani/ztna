package posture

import (
	"fmt"
	"ztna-core/ztna/common"
	"ztna-core/ztna/common/eid"
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
)

func HasAccess(rdm *common.RouterDataModel, identityId string, serviceId string, cache *Cache, policyType edge_ctrl_pb.PolicyType) (*common.ServicePolicy, error) {
	logtrace.LogWithFunctionName()
	log := pfxlog.Logger().WithField("instance", eid.New()).WithField("identityId", identityId).WithField("serviceId", serviceId)

	accessPolicies, err := rdm.GetServiceAccessPolicies(identityId, serviceId, policyType)

	if err != nil {
		log.WithError(err).Debug("could not find access path for authorization checks")
		return nil, err
	}

	if accessPolicies == nil || len(accessPolicies.Policies) == 0 {
		return nil, &NoPoliciesError{}
	}

	grantingPolicy, errs := IsPassing(accessPolicies, cache)

	if errs != nil && len(*errs) > 0 {
		log.Debug("policies provided access but posture checks failed")
		return nil, errs
	}

	log.Debugf("access provided via policy %s [%s]", grantingPolicy.Name, grantingPolicy.Id)
	return grantingPolicy, nil
}

func IsPassing(accessPolicies *common.AccessPolicies, cache *Cache) (*common.ServicePolicy, *PolicyAccessErrors) {
	logtrace.LogWithFunctionName()
	errs := &PolicyAccessErrors{}

	for _, policy := range accessPolicies.Policies {
		policyErr := &PolicyAccessError{
			Id:     policy.Id,
			Name:   policy.Name,
			Errors: []error{},
		}

		policy.PostureChecks.IterCb(func(postureCheckId string, _ struct{}) {
			postureCheck, ok := accessPolicies.PostureChecks[postureCheckId]

			if !ok || postureCheck == nil {
				policyErr.Errors = append(policyErr.Errors, fmt.Errorf("posture check id %s not found model", postureCheckId))
				return
			}

			if err := EvaluatePostureCheck(postureCheck, cache); err != nil {
				policyErr.Errors = append(policyErr.Errors, err)
			}
		})

		if len(policyErr.Errors) == 0 {
			return policy, nil
		} else {
			*errs = append(*errs, policyErr)
		}
	}

	return nil, errs
}
