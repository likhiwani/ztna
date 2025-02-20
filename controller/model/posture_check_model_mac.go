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

	"ztna-core/ztna/common/pb/edge_cmd_pb"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/logtrace"

	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

var _ PostureCheckSubType = &PostureCheckMacAddresses{}

type PostureCheckMacAddresses struct {
	MacAddresses []string
}

func (p *PostureCheckMacAddresses) TypeId() string {
	logtrace.LogWithFunctionName()
	return db.PostureCheckTypeMAC
}

func (p *PostureCheckMacAddresses) fillProtobuf(msg *edge_cmd_pb.PostureCheck) {
	logtrace.LogWithFunctionName()
	msg.Subtype = &edge_cmd_pb.PostureCheck_Mac_{
		Mac: &edge_cmd_pb.PostureCheck_Mac{
			MacAddresses: p.MacAddresses,
		},
	}
}

func (p *PostureCheckMacAddresses) fillFromProtobuf(msg *edge_cmd_pb.PostureCheck) error {
	logtrace.LogWithFunctionName()
	if mac, ok := msg.Subtype.(*edge_cmd_pb.PostureCheck_Mac_); ok {
		if mac.Mac != nil {
			p.MacAddresses = mac.Mac.MacAddresses
		}
	} else {
		return errors.Errorf("expected posture check sub type data of mac address, but got %T", msg.Subtype)
	}
	return nil
}

func (p *PostureCheckMacAddresses) LastUpdatedAt(apiSessionId string, pd *PostureData) *time.Time {
	logtrace.LogWithFunctionName()
	return nil
}

func (p *PostureCheckMacAddresses) GetTimeoutRemainingSeconds(_ string, _ *PostureData) int64 {
	logtrace.LogWithFunctionName()
	return PostureCheckNoTimeout
}

func (p *PostureCheckMacAddresses) GetTimeoutSeconds() int64 {
	logtrace.LogWithFunctionName()
	return PostureCheckNoTimeout
}

func (p *PostureCheckMacAddresses) FailureValues(_ string, pd *PostureData) PostureCheckFailureValues {
	logtrace.LogWithFunctionName()
	return &PostureCheckFailureValuesMac{
		ActualValue:   pd.Mac.Addresses,
		ExpectedValue: p.MacAddresses,
	}
}

func (p *PostureCheckMacAddresses) Evaluate(_ string, pd *PostureData) bool {
	logtrace.LogWithFunctionName()
	if pd.Mac.TimedOut {
		return false
	}

	validAddresses := map[string]struct{}{}
	for _, address := range p.MacAddresses {
		validAddresses[address] = struct{}{}
	}

	for _, address := range pd.Mac.Addresses {
		if _, found := validAddresses[address]; found {
			return true
		}
	}

	return false
}

func newPostureCheckMacAddresses() PostureCheckSubType {
	logtrace.LogWithFunctionName()
	return &PostureCheckMacAddresses{}
}

func (p *PostureCheckMacAddresses) fillFrom(_ Env, tx *bbolt.Tx, check *db.PostureCheck, subType db.PostureCheckSubType) error {
	logtrace.LogWithFunctionName()
	subCheck := subType.(*db.PostureCheckMacAddresses)

	if subCheck == nil {
		return fmt.Errorf("could not convert mac address check to bolt type")
	}

	p.MacAddresses = subCheck.MacAddresses
	return nil
}

func (p *PostureCheckMacAddresses) toBoltEntityForCreate(*bbolt.Tx, Env) (db.PostureCheckSubType, error) {
	logtrace.LogWithFunctionName()
	return &db.PostureCheckMacAddresses{
		MacAddresses: p.MacAddresses,
	}, nil
}

type PostureCheckFailureValuesMac struct {
	ActualValue   []string
	ExpectedValue []string
}

func (p PostureCheckFailureValuesMac) Expected() interface{} {
	logtrace.LogWithFunctionName()
	return p.ExpectedValue
}

func (p PostureCheckFailureValuesMac) Actual() interface{} {
	logtrace.LogWithFunctionName()
	return p.ActualValue
}
