package posture

import (
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/logtrace"
)

type DomainCheck struct {
	*edge_ctrl_pb.DataState_PostureCheck
	*edge_ctrl_pb.DataState_PostureCheck_Domains
}

func (m *DomainCheck) Evaluate(state *Cache) *CheckError {
	logtrace.LogWithFunctionName()
	if state == nil {
		return &CheckError{
			Id:    m.Id,
			Name:  m.Name,
			Cause: NilStateError,
		}
	}

	notInListErr := &AnyInListError[Str]{
		GivenValues: []Str{Str(state.Domain.Name)},
	}

	for _, domain := range m.Domains {
		if state.Domain.Name != domain {
			notInListErr.FailedValues = append(notInListErr.FailedValues, FailedValueError[Str]{
				ExpectedValue: Str(domain),
				GivenValue:    Str(state.Domain.Name),
				Reason:        NotEqualError,
			})
		} else {
			return nil
		}
	}

	return &CheckError{
		Id:    m.Id,
		Name:  m.Name,
		Cause: notInListErr,
	}
}
