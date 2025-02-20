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

package exporter

import (
	logtrace "ztna-core/ztna/logtrace"
	"slices"
	"ztna-core/edge-api/rest_management_api_client/config"
	"ztna-core/edge-api/rest_model"
)

func (exporter Exporter) IsConfigTypeExportRequired(args []string) bool {
	logtrace.LogWithFunctionName()
	return slices.Contains(args, "all") || len(args) == 0 || // explicit all or nothing specified
		slices.Contains(args, "config-type")
}

func (exporter Exporter) GetConfigTypes() ([]map[string]interface{}, error) {
	logtrace.LogWithFunctionName()

	return exporter.getEntities(
		"ConfigTypes",

		func() (int64, error) {
			limit := int64(1)
			resp, err := exporter.client.Config.ListConfigTypes(&config.ListConfigTypesParams{Limit: &limit}, nil)
			if err != nil {
				return -1, err
			}
			return *resp.GetPayload().Meta.Pagination.TotalCount, nil
		},

		func(offset *int64, limit *int64) ([]interface{}, error) {
			resp, _ := exporter.client.Config.ListConfigTypes(&config.ListConfigTypesParams{Limit: limit, Offset: offset}, nil)
			entities := make([]interface{}, len(resp.GetPayload().Data))
			for i, c := range resp.GetPayload().Data {
				entities[i] = interface{}(c)
			}
			return entities, nil
		},

		func(entity interface{}) (map[string]interface{}, error) {

			item := entity.(*rest_model.ConfigTypeDetail)
			wellknownTypes := []string{"intercept.v1", "host.v1", "host.v2", "ziti-tunneler-server.v1", "ziti-tunneler-client.v1"}

			// don't include the wellknown types, they already exist when a network is created
			if slices.Contains(wellknownTypes, *item.Name) {
				return nil, nil
			}

			// convert to a map of values
			m, err := exporter.ToMap(item)
			if err != nil {
				log.WithError(err).Error("error converting ConfigType to map")
			}
			exporter.Filter(m, []string{"id", "_links", "createdAt", "updatedAt"})

			return m, nil
		})
}
