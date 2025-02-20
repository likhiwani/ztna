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

package importer

import (
	logtrace "ztna-core/ztna/logtrace"
	"encoding/json"
	"errors"
	"slices"
	"ztna-core/edge-api/rest_management_api_client/identity"
	"ztna-core/edge-api/rest_model"
	"ztna-core/edge-api/rest_util"
	"ztna-core/ztna/internal"
	"ztna-core/ztna/internal/ascode"
	"ztna-core/ztna/internal/rest/mgmt"

	"github.com/Jeffail/gabs/v2"
)

func (importer *Importer) IsIdentityImportRequired(args []string) bool {
	logtrace.LogWithFunctionName()
	return slices.Contains(args, "all") || len(args) == 0 || // explicit all or nothing specified
		slices.Contains(args, "identity")
}

func (importer *Importer) ProcessIdentities(input map[string][]interface{}) (map[string]string, error) {
	logtrace.LogWithFunctionName()

	var result = map[string]string{}

	for _, data := range input["identities"] {
		create := FromMap(data, rest_model.IdentityCreate{})

		existing := mgmt.IdentityFromFilter(importer.client, mgmt.NameFilter(*create.Name))
		if existing != nil {
			if importer.loginOpts.Verbose {
				log.WithFields(map[string]interface{}{
					"name":       *create.Name,
					"identityId": *existing.ID,
				}).
					Info("Found existing Identity, skipping create")
			}
			_, _ = internal.FPrintfReusingLine(importer.loginOpts.Err, "Skipping Identity %s\r", *create.Name)
			continue
		}

		// set the type because it is not in the input
		typ := rest_model.IdentityTypeDefault
		create.Type = &typ

		// convert to a json doc so we can query inside the data
		jsonData, _ := json.Marshal(data)
		doc, jsonParseError := gabs.ParseJSON(jsonData)
		if jsonParseError != nil {
			log.WithError(jsonParseError).Error("Unable to parse json")
			return nil, jsonParseError
		}
		policyName := doc.Path("authPolicy").Data().(string)[1:]

		// look up the auth policy id from the name and add to the create, omit if it's the "Default" policy
		policy, _ := ascode.GetItemFromCache(importer.authPolicyCache, policyName, func(name string) (interface{}, error) {
			return mgmt.AuthPolicyFromFilter(importer.client, mgmt.NameFilter(name)), nil
		})
		if policy == nil {
			return nil, errors.New("error reading Auth Policy: " + policyName)
		}
		if policy != "" && policy != "Default" {
			create.AuthPolicyID = policy.(*rest_model.AuthPolicyDetail).ID
		}

		// do the actual create since it doesn't exist
		_, _ = internal.FPrintfReusingLine(importer.loginOpts.Err, "Creating Identity %s\r", *create.Name)
		created, createErr := importer.client.Identity.CreateIdentity(&identity.CreateIdentityParams{Identity: create}, nil)
		if createErr != nil {
			if payloadErr, ok := createErr.(rest_util.ApiErrorPayload); ok {
				log.WithFields(map[string]interface{}{
					"field":  payloadErr.GetPayload().Error.Cause.APIFieldError.Field,
					"reason": payloadErr.GetPayload().Error.Cause.APIFieldError.Reason,
				}).
					Error("Unable to create Identity")
				return nil, createErr
			} else {
				log.WithError(createErr).Error("Unable to create Identity")
				return nil, createErr
			}
		}
		if importer.loginOpts.Verbose {
			log.WithFields(map[string]interface{}{
				"name":       *create.Name,
				"identityId": created.Payload.Data.ID,
			}).
				Info("Created identity")
		}

		result[*create.Name] = created.Payload.Data.ID
	}

	return result, nil
}

func (importer *Importer) lookupIdentities(roles []string) ([]string, error) {
	logtrace.LogWithFunctionName()
	identityRoles := []string{}
	for _, role := range roles {
		if role[0:1] == "@" {
			roleName := role[1:]
			value, lookupErr := ascode.GetItemFromCache(importer.identityCache, roleName, func(name string) (interface{}, error) {
				return mgmt.IdentityFromFilter(importer.client, mgmt.NameFilter(name)), nil
			})
			if lookupErr != nil {
				return nil, lookupErr
			}
			ident := value.(*rest_model.IdentityDetail)
			if ident == nil {
				return nil, errors.New("error reading Identity: " + roleName)
			}
			identityRoles = append(identityRoles, "@"+*ident.ID)
		} else {
			identityRoles = append(identityRoles, role)
		}
	}
	return identityRoles, nil
}
