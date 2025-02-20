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

package db

import (
	"ztna-core/ztna/logtrace"

	"github.com/openziti/storage/boltz"
)

const (
	FieldPostureCheckMfaTimeoutSeconds        = "timeoutSeconds"
	FieldPostureCheckMfaPromptOnWake          = "promptOnWake"
	FieldPostureCheckMfaPromptOnUnlock        = "promptOnUnlock"
	FieldPostureCheckMfaIgnoreLegacyEndpoints = "ignoreLegacyEndpoints"
)

type PostureCheckMfa struct {
	TimeoutSeconds        int64 `json:"timeoutSeconds"`
	PromptOnWake          bool  `json:"promptOnWake"`
	PromptOnUnlock        bool  `json:"promptOnUnlock"`
	IgnoreLegacyEndpoints bool  `json:"ignoreLegacyEndpoints"`
}

func newPostureCheckMfa() PostureCheckSubType {
	logtrace.LogWithFunctionName()
	return &PostureCheckMfa{
		TimeoutSeconds:        0,
		PromptOnWake:          false,
		PromptOnUnlock:        false,
		IgnoreLegacyEndpoints: false,
	}
}

func (entity *PostureCheckMfa) LoadValues(bucket *boltz.TypedBucket) {
	logtrace.LogWithFunctionName()
	entity.TimeoutSeconds = bucket.GetInt64WithDefault(FieldPostureCheckMfaTimeoutSeconds, -1)

	if entity.TimeoutSeconds <= 0 {
		entity.TimeoutSeconds = -1
	}

	entity.PromptOnWake = bucket.GetBoolWithDefault(FieldPostureCheckMfaPromptOnWake, false)
	entity.PromptOnUnlock = bucket.GetBoolWithDefault(FieldPostureCheckMfaPromptOnUnlock, false)
	entity.IgnoreLegacyEndpoints = bucket.GetBoolWithDefault(FieldPostureCheckMfaIgnoreLegacyEndpoints, false)
}

func (entity *PostureCheckMfa) SetValues(ctx *boltz.PersistContext, bucket *boltz.TypedBucket) {
	logtrace.LogWithFunctionName()
	if entity.TimeoutSeconds <= 0 {
		entity.TimeoutSeconds = -1
	}

	bucket.SetInt64(FieldPostureCheckMfaTimeoutSeconds, entity.TimeoutSeconds, ctx.FieldChecker)
	bucket.SetBool(FieldPostureCheckMfaPromptOnWake, entity.PromptOnWake, ctx.FieldChecker)
	bucket.SetBool(FieldPostureCheckMfaPromptOnUnlock, entity.PromptOnUnlock, ctx.FieldChecker)
	bucket.SetBool(FieldPostureCheckMfaIgnoreLegacyEndpoints, entity.IgnoreLegacyEndpoints, ctx.FieldChecker)
}
