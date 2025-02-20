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
	"time"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/storage/boltz"
)

func (m *Migrations) addSystemAuthPolicies(step *boltz.MigrationStep) {
	logtrace.LogWithFunctionName()
	defaultPolicy := &AuthPolicy{
		BaseExtEntity: boltz.BaseExtEntity{
			Id:        DefaultAuthPolicyId,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tags:      nil,
			IsSystem:  true,
		},
		Name: "Default",
		Primary: AuthPolicyPrimary{
			Cert: AuthPolicyCert{
				Allowed:           true,
				AllowExpiredCerts: true,
			},
			Updb: AuthPolicyUpdb{
				Allowed:            true,
				MinPasswordLength:  DefaultUpdbMinPasswordLength,
				RequireSpecialChar: false,
				RequireNumberChar:  false,
				RequireMixedCase:   false,
				MaxAttempts:        0,
			},
			ExtJwt: AuthPolicyExtJwt{
				Allowed:              true,
				AllowedExtJwtSigners: nil,
			},
		},
		Secondary: AuthPolicySecondary{
			RequireTotp:          false,
			RequiredExtJwtSigner: nil,
		},
	}

	step.SetError(m.stores.AuthPolicy.Create(step.Ctx, defaultPolicy))
}
