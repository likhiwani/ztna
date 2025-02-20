package posture

import (
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/logtrace"
)

type MacCheck struct {
	*edge_ctrl_pb.DataState_PostureCheck
	*edge_ctrl_pb.DataState_PostureCheck_Mac
}

func (m MacCheck) Evaluate(state *Cache) *CheckError {
	logtrace.LogWithFunctionName()
	if state == nil {
		return &CheckError{
			Id:    m.Id,
			Name:  m.Name,
			Cause: NilStateError,
		}
	}

	notInListErr := &OneInListError[Str]{}

	macMap := map[string]struct{}{}
	for _, macAddresses := range state.Macs.Addresses {
		macMap[macAddresses] = struct{}{}
	}

	for _, macAddress := range m.MacAddresses {
		if _, ok := macMap[macAddress]; ok {
			return nil
		}
	}

	for _, mac := range m.MacAddresses {
		notInListErr.ValidValues = append(notInListErr.ValidValues, Str(mac))
	}

	for _, mac := range state.Macs.Addresses {
		notInListErr.GivenValues = append(notInListErr.GivenValues, Str(mac))
	}

	return &CheckError{
		Id:    m.Id,
		Name:  m.Name,
		Cause: notInListErr,
	}
}
