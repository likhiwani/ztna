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
	"fmt"
	"strings"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/foundation/v2/errorz"
	"github.com/openziti/foundation/v2/stringz"
	"github.com/openziti/storage/boltz"
	"github.com/pkg/errors"
)

const (
	RolePrefix   = "#"
	EntityPrefix = "@"
	AllRole      = "#all"
)

func validateRolesAndIds(field string, values []string) error {
	logtrace.LogWithFunctionName()
	if len(values) > 1 && stringz.Contains(values, AllRole) {
		return errorz.NewFieldError(fmt.Sprintf("if using %v, it should be the only role specified", AllRole), field, values)
	}

	var invalidKeys []string
	for _, entry := range values {
		if !strings.HasPrefix(entry, RolePrefix) && !strings.HasPrefix(entry, EntityPrefix) {
			invalidKeys = append(invalidKeys, entry)
		}
	}
	if len(invalidKeys) > 0 {
		return errorz.NewFieldError("role entries must prefixed with # (to indicate role attributes) or @ (to indicate a name or id)", field, invalidKeys)
	}
	return nil
}

func splitRolesAndIds(values []string) ([]string, []string, error) {
	logtrace.LogWithFunctionName()
	var roles []string
	var ids []string
	for _, entry := range values {
		if strings.HasPrefix(entry, RolePrefix) {
			entry = strings.TrimPrefix(entry, RolePrefix)
			roles = append(roles, entry)
		} else if strings.HasPrefix(entry, EntityPrefix) {
			entry = strings.TrimPrefix(entry, EntityPrefix)
			ids = append(ids, entry)
		} else {
			return nil, nil, errors.Errorf("'%v' is neither role attribute (prefixed with %v) or an entity id or name (prefixed with %v)",
				entry, RolePrefix, EntityPrefix)
		}
	}
	return roles, ids, nil
}

func FieldValuesToIds(new []boltz.FieldTypeAndValue) []string {
	logtrace.LogWithFunctionName()
	var entityRoles []string
	for _, fv := range new {
		entityRoles = append(entityRoles, string(fv.Value))
	}
	return entityRoles
}

func roleRef(val string) string {
	logtrace.LogWithFunctionName()
	return RolePrefix + val
}

func entityRef(val string) string {
	logtrace.LogWithFunctionName()
	return EntityPrefix + val
}
