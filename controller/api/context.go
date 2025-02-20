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

package api

import (
	"context"
	"fmt"
	"net/http"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/logtrace"

	"github.com/pkg/errors"
)

type RequestContextImpl struct {
	Responder
	Id             string
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	entityId       string
	entitySubId    string
	Body           []byte
}

const (
	IdPropertyName    = "id"
	SubIdPropertyName = "subId"
)

func (rc *RequestContextImpl) GetId() string {
	logtrace.LogWithFunctionName()
	return rc.Id
}

func (rc *RequestContextImpl) GetBody() []byte {
	logtrace.LogWithFunctionName()
	return rc.Body
}

func (rc *RequestContextImpl) GetRequest() *http.Request {
	logtrace.LogWithFunctionName()
	return rc.Request
}

func (rc *RequestContextImpl) GetResponseWriter() http.ResponseWriter {
	logtrace.LogWithFunctionName()
	return rc.ResponseWriter
}

func (rc *RequestContextImpl) SetEntityId(id string) {
	logtrace.LogWithFunctionName()
	rc.entityId = id
}

func (rc *RequestContextImpl) SetEntitySubId(id string) {
	logtrace.LogWithFunctionName()
	rc.entitySubId = id
}

func (rc *RequestContextImpl) GetEntityId() (string, error) {
	logtrace.LogWithFunctionName()
	if rc.entityId == "" {
		return "", errors.New("id not found")
	}
	return rc.entityId, nil
}

func (rc *RequestContextImpl) GetEntitySubId() (string, error) {
	logtrace.LogWithFunctionName()
	if rc.entitySubId == "" {
		return "", errors.New("subId not found")
	}

	return rc.entitySubId, nil
}

func (rc *RequestContextImpl) NewChangeContext() *change.Context {
	logtrace.LogWithFunctionName()
	changeCtx := change.New().SetSourceType(change.SourceTypeRest).
		SetSourceAuth("fabric").
		SetSourceMethod(rc.GetRequest().Method).
		SetSourceLocal(rc.GetRequest().Host).
		SetSourceRemote(rc.GetRequest().RemoteAddr)

	changeCtx.SetChangeAuthorType(change.AuthorTypeCert)

	if rc.Request.TLS != nil {
		for _, cert := range rc.Request.TLS.PeerCertificates {
			if !cert.IsCA {
				changeCtx.SetChangeAuthorId(cert.Subject.CommonName)
			}
		}
	}

	return changeCtx
}

// ContextKey is used a custom type to avoid accidental context key collisions
type ContextKey string

const ZitiContextKey = ContextKey("context")

func AddRequestContextToHttpContext(r *http.Request, rc RequestContext) {
	logtrace.LogWithFunctionName()
	ctx := context.WithValue(r.Context(), ZitiContextKey, rc)
	*r = *r.WithContext(ctx)
}

func GetRequestContextFromHttpContext(r *http.Request) (RequestContext, error) {
	logtrace.LogWithFunctionName()
	val := r.Context().Value(ZitiContextKey)
	if val == nil {
		return nil, fmt.Errorf("value for key %s no found in context", ZitiContextKey)
	}

	requestContext := val.(RequestContext)

	if requestContext == nil {
		return nil, fmt.Errorf("value for key %s is not a request context", ZitiContextKey)
	}

	return requestContext, nil
}
