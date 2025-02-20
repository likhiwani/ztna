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
	"slices"
	"strings"
	"ztna-core/edge-api/rest_management_api_client/posture_checks"
	"ztna-core/edge-api/rest_model"
	"ztna-core/edge-api/rest_util"
	"ztna-core/ztna/internal"
	"ztna-core/ztna/internal/rest/mgmt"

	"github.com/Jeffail/gabs/v2"
)

func (importer *Importer) IsPostureCheckImportRequired(args []string) bool {
	logtrace.LogWithFunctionName()
	return slices.Contains(args, "all") || len(args) == 0 || // explicit all or nothing specified
		slices.Contains(args, "posture-check")
}

func (importer *Importer) ProcessPostureChecks(input map[string][]interface{}) (map[string]string, error) {
	logtrace.LogWithFunctionName()

	var result = map[string]string{}
	for _, data := range input["postureChecks"] {

		// convert to a json doc so we can query inside the data
		jsonData, _ := json.Marshal(data)
		doc, jsonParseError := gabs.ParseJSON(jsonData)
		if jsonParseError != nil {
			log.WithError(jsonParseError).Error("Unable to parse json")
			return nil, jsonParseError
		}
		typeValue := doc.Path("typeId").Data().(string)

		var create rest_model.PostureCheckCreate
		switch strings.ToUpper(typeValue) {
		case string(rest_model.PostureCheckTypeDOMAIN):
			create = FromMap(data, rest_model.PostureCheckDomainCreate{})
		case string(rest_model.PostureCheckTypeMAC):
			create = FromMap(data, rest_model.PostureCheckMacAddressCreate{})
		case string(rest_model.PostureCheckTypeMFA):
			create = FromMap(data, rest_model.PostureCheckMfaCreate{})
		case string(rest_model.PostureCheckTypeOS):
			create = FromMap(data, rest_model.PostureCheckOperatingSystemCreate{})
		case string(rest_model.PostureCheckTypePROCESS):
			create = FromMap(data, rest_model.PostureCheckProcessCreate{})
		case string(rest_model.PostureCheckTypePROCESSMULTI):
			create = FromMap(data, rest_model.PostureCheckProcessMultiCreate{})
		default:
			log.WithFields(map[string]interface{}{
				"name":   *create.Name(),
				"typeId": create.TypeID,
			}).
				Error("Unknown PostureCheck type")
		}

		// see if the posture check already exists
		existing := mgmt.PostureCheckFromFilter(importer.client, mgmt.NameFilter(*create.Name()))
		if existing != nil {
			if importer.loginOpts.Verbose {
				log.WithFields(map[string]interface{}{
					"name":           *create.Name(),
					"postureCheckId": (*existing).ID(),
					"typeId":         create.TypeID(),
				}).
					Info("Found existing PostureCheck, skipping create")
			}
			_, _ = internal.FPrintfReusingLine(importer.loginOpts.Err, "Skipping PostureCheck %s\r", *create.Name())
			continue
		}

		// do the actual create since it doesn't exist
		_, _ = internal.FPrintfReusingLine(importer.loginOpts.Err, "Creating PostureCheck %s\r", *create.Name())
		if importer.loginOpts.Verbose {
			log.WithFields(map[string]interface{}{
				"name":   *create.Name(),
				"typeId": create.TypeID(),
			}).
				Debug("Creating PostureCheck")
		}
		created, createErr := importer.client.PostureChecks.CreatePostureCheck(&posture_checks.CreatePostureCheckParams{PostureCheck: create}, nil)
		if createErr != nil {
			if payloadErr, ok := createErr.(rest_util.ApiErrorPayload); ok {
				log.WithFields(map[string]interface{}{
					"field":  payloadErr.GetPayload().Error.Cause.APIFieldError.Field,
					"reason": payloadErr.GetPayload().Error.Cause.APIFieldError.Reason,
				}).
					Error("Unable to create PostureCheck")
			} else {
				log.WithError(createErr).Error("Unable to ")
				return nil, createErr
			}
		}
		if importer.loginOpts.Verbose {
			log.WithFields(map[string]interface{}{
				"name":           *create.Name(),
				"postureCheckId": created.Payload.Data.ID,
				"typeId":         create.TypeID(),
			}).
				Info("Created PostureCheck")
		}

		result[*create.Name()] = created.Payload.Data.ID
	}

	return result, nil
}
