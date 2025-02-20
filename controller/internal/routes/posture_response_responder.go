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
	"net/http"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"
)

type PostureResponseResponder struct {
	response.Responder
	ae       *env.AppEnv
	services []*rest_model.PostureResponseService
}

func (responder *PostureResponseResponder) RespondWithCreatedId(id string, _ rest_model.Link) {
	logtrace.LogWithFunctionName()
	data := &rest_model.PostureResponse{
		Services: responder.services,
	}
	newSessionEnvelope := &rest_model.PostureResponseEnvelope{
		Data: data,
		Meta: &rest_model.Meta{},
	}

	responder.Respond(newSessionEnvelope, http.StatusCreated)
}
