package posture

import (
	"time"
	"ztna-core/sdk-golang/pb/edge_client_pb"
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/logtrace"
)

type Cache struct {
	Os          *edge_client_pb.PostureResponse_Os
	Domain      *edge_client_pb.PostureResponse_Domain
	Macs        *edge_client_pb.PostureResponse_Macs
	Unlocked    *edge_client_pb.PostureResponse_Unlocked
	Woken       *edge_client_pb.PostureResponse_Woken
	ProcessList *edge_client_pb.PostureResponse_ProcessList
	PassedMfaAt *time.Time
}

type Check interface {
	Evaluate(state *Cache) *CheckError
}

func CtrlCheckToLogic(postureCheck *edge_ctrl_pb.DataState_PostureCheck) Check {
	logtrace.LogWithFunctionName()
	switch subCheck := postureCheck.Subtype.(type) {
	case *edge_ctrl_pb.DataState_PostureCheck_Mac_:
		return &MacCheck{
			DataState_PostureCheck:     postureCheck,
			DataState_PostureCheck_Mac: subCheck.Mac,
		}
	case *edge_ctrl_pb.DataState_PostureCheck_OsList_:
		return &OsCheck{
			DataState_PostureCheck:        postureCheck,
			DataState_PostureCheck_OsList: subCheck.OsList,
		}
	case *edge_ctrl_pb.DataState_PostureCheck_Process_:
		return &ProcessCheck{
			DataState_PostureCheck: postureCheck,
			DataState_PostureCheck_ProcessMulti: &edge_ctrl_pb.DataState_PostureCheck_ProcessMulti{
				Semantic: "AllOf",
				Processes: []*edge_ctrl_pb.DataState_PostureCheck_Process{
					{
						OsType:       subCheck.Process.OsType,
						Path:         subCheck.Process.Path,
						Hashes:       subCheck.Process.Hashes,
						Fingerprints: subCheck.Process.Fingerprints,
					},
				},
			},
		}
	case *edge_ctrl_pb.DataState_PostureCheck_ProcessMulti_:
		return &ProcessCheck{
			DataState_PostureCheck:              postureCheck,
			DataState_PostureCheck_ProcessMulti: subCheck.ProcessMulti,
		}
	case *edge_ctrl_pb.DataState_PostureCheck_Domains_:
		return &DomainCheck{
			DataState_PostureCheck:         postureCheck,
			DataState_PostureCheck_Domains: subCheck.Domains,
		}
	case *edge_ctrl_pb.DataState_PostureCheck_Mfa_:
		return &MfaCheck{
			DataState_PostureCheck:     postureCheck,
			DataState_PostureCheck_Mfa: subCheck.Mfa,
		}
	}

	return nil
}
