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

package sync_strats

import (
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"go.etcd.io/bbolt"
)

func apiSessionToProto(ae *env.AppEnv, token, identityId, apiSessionId string) (*edge_ctrl_pb.ApiSession, error) {
	logtrace.LogWithFunctionName()
	var result *edge_ctrl_pb.ApiSession
	err := ae.GetDb().View(func(tx *bbolt.Tx) error {
		var err error
		result, err = apiSessionToProtoWithTx(tx, ae, token, identityId, apiSessionId)
		return err
	})
	return result, err
}

func apiSessionToProtoWithTx(tx *bbolt.Tx, ae *env.AppEnv, token, identityId, apiSessionId string) (*edge_ctrl_pb.ApiSession, error) {
	logtrace.LogWithFunctionName()
	fingerprints, err := getFingerprints(tx, ae, identityId, apiSessionId)
	if err != nil {
		return nil, err
	}

	return &edge_ctrl_pb.ApiSession{
		Token:            token,
		CertFingerprints: fingerprints,
		Id:               apiSessionId,
		IdentityId:       identityId,
	}, nil
}

func getFingerprints(tx *bbolt.Tx, ae *env.AppEnv, identityId, apiSessionId string) ([]string, error) {
	logtrace.LogWithFunctionName()
	prints := map[string]struct{}{}
	err := ae.Managers.ApiSession.VisitFingerprintsForApiSession(tx, identityId, apiSessionId, func(fingerprint string) bool {
		prints[fingerprint] = struct{}{}
		return false
	})

	if err != nil {
		return nil, err
	}
	var result []string
	for k := range prints {
		result = append(result, k)
	}

	pfxlog.Logger().WithField("apiSessionId", apiSessionId).WithField("fingerprints", result).Debug("resolving fingerprints for apiSession")

	return result, nil
}
