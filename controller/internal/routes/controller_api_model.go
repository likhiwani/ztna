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

package routes

import (
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/response"
	edgeController "ztna-core/ztna/controller/webapis"
	"ztna-core/ztna/logtrace"
)

const EntityNameController = "controllers"

var ControllerLinkFactory = NewBasicLinkFactory(EntityNameController)

func MapControllerToManagementRestEntity(_ *env.AppEnv, _ *response.RequestContext, Controller *model.Controller) (interface{}, error) {
	logtrace.LogWithFunctionName()
	return MapControllerToManagementRestModel(Controller)
}

func MapControllerToClientRestEntity(_ *env.AppEnv, _ *response.RequestContext, Controller *model.Controller) (interface{}, error) {
	logtrace.LogWithFunctionName()
	return MapControllerToClientRestModel(Controller)
}

func MapControllerToManagementRestModel(controller *model.Controller) (*rest_model.ControllerDetail, error) {
	logtrace.LogWithFunctionName()
	ret := &rest_model.ControllerDetail{
		BaseEntity:   BaseEntityToRestModel(controller, ControllerLinkFactory),
		Name:         &controller.Name,
		CtrlAddress:  &controller.CtrlAddress,
		APIAddresses: rest_model.APIAddressList{},
		CertPem:      &controller.CertPem,
		Fingerprint:  &controller.Fingerprint,
		IsOnline:     &controller.IsOnline,
		LastJoinedAt: toStrFmtDateTimeP(controller.LastJoinedAt),
	}

	for apiKey, instances := range controller.ApiAddresses {
		ret.APIAddresses[apiKey] = rest_model.APIAddressArray{}
		for _, instance := range instances {
			ret.APIAddresses[apiKey] = append(ret.APIAddresses[apiKey], &rest_model.APIAddress{
				URL:     instance.Url,
				Version: instance.Version,
			})
		}
	}
	return ret, nil
}

func MapControllerToClientRestModel(controller *model.Controller) (*rest_model.ControllerDetail, error) {
	logtrace.LogWithFunctionName()
	ret, err := MapControllerToManagementRestModel(controller)

	if err != nil {
		return nil, err
	}

	ret.CtrlAddress = nil

	for apiKey := range ret.APIAddresses {
		if apiKey != edgeController.ClientApiBinding && apiKey != edgeController.OidcApiBinding {
			delete(ret.APIAddresses, apiKey)
		}
	}

	return ret, nil
}
