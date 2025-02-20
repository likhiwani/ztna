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

package api_impl

import (
	"bytes"
	"io"
	"net/http"
	"ztna-core/ztna/controller/api"
	"ztna-core/ztna/controller/idgen"
	"ztna-core/ztna/logtrace"
)

func NewRequestContext(rw http.ResponseWriter, r *http.Request) api.RequestContext {
	logtrace.LogWithFunctionName()
	rid := idgen.New()

	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(body))

	requestContext := &api.RequestContextImpl{
		Id:             rid,
		Body:           body,
		ResponseWriter: rw,
		Request:        r,
	}

	requestContext.Responder = api.NewResponder(requestContext, FabricResponseMapper{})

	return requestContext
}
