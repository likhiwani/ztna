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
	"slices"
	"ztna-core/edge-api/rest_management_api_client/external_jwt_signer"
	"ztna-core/edge-api/rest_model"
	"ztna-core/edge-api/rest_util"
	"ztna-core/ztna/internal"
	"ztna-core/ztna/internal/rest/mgmt"
)

func (importer *Importer) IsExtJwtSignerImportRequired(args []string) bool {
	logtrace.LogWithFunctionName()
	return slices.Contains(args, "all") || len(args) == 0 || // explicit all or nothing specified
		slices.Contains(args, "ext-jwt-signer") ||
		slices.Contains(args, "external-jwt-signer") ||
		slices.Contains(args, "auth-policy") ||
		slices.Contains(args, "identity")
}

func (importer *Importer) ProcessExternalJwtSigners(input map[string][]interface{}) (map[string]string, error) {
	logtrace.LogWithFunctionName()

	var result = map[string]string{}
	for _, data := range input["externalJwtSigners"] {
		create := FromMap(data, rest_model.ExternalJWTSignerCreate{})

		// see if the signer already exists
		existing := mgmt.ExternalJWTSignerFromFilter(importer.client, mgmt.NameFilter(*create.Name))
		if existing != nil {
			if importer.loginOpts.Verbose {
				log.WithFields(map[string]interface{}{
					"name":                *create.Name,
					"externalJwtSignerId": *existing.ID,
				}).
					Info("Found existing ExtJWTSigner, skipping create")
			}
			_, _ = internal.FPrintfReusingLine(importer.loginOpts.Err, "Skipping ExtJWTSigner %s\r", *create.Name)
			continue
		}

		// do the actual create since it doesn't exist
		_, _ = internal.FPrintfReusingLine(importer.loginOpts.Err, "Creating ExtJWTSigner %s\r", *create.Name)
		if importer.loginOpts.Verbose {
			log.WithField("name", *create.Name).Debug("Creating ExtJWTSigner")
		}
		created, createErr := importer.client.ExternalJWTSigner.CreateExternalJWTSigner(&external_jwt_signer.CreateExternalJWTSignerParams{ExternalJWTSigner: create}, nil)
		if createErr != nil {
			if payloadErr, ok := createErr.(rest_util.ApiErrorPayload); ok {
				log.WithFields(map[string]interface{}{
					"field":  payloadErr.GetPayload().Error.Cause.APIFieldError.Field,
					"reason": payloadErr.GetPayload().Error.Cause.APIFieldError.Reason,
					"err":    payloadErr,
				}).
					Error("Unable to create ExtJWTSigner")
				return nil, createErr
			} else {
				log.Error("Unable to create ExtJWTSigner")
				return nil, createErr
			}
		}
		if importer.loginOpts.Verbose {
			log.WithFields(map[string]interface{}{
				"name":                *create.Name,
				"externalJwtSignerId": created.Payload.Data.ID,
			}).
				Info("Created ExtJWTSigner")
		}

		result[*create.Name] = created.Payload.Data.ID
	}

	return result, nil
}
