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

package response

import (
	"errors"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/common"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/logtrace"

	"github.com/golang-jwt/jwt/v5"

	"net/http"
	"time"
)

const (
	IdPropertyName    = "id"
	SubIdPropertyName = "subId"
)

type RequestContext struct {
	Responder

	// The unique id of the current request
	Id string

	// An opaque session token
	SessionToken string
	Jwt          *jwt.Token
	Claims       *common.AccessClaims
	ApiSession   *model.ApiSession

	Identity          *model.Identity
	AuthPolicy        *model.AuthPolicy
	AuthQueries       rest_model.AuthQueryList
	ActivePermissions []string
	ResponseWriter    http.ResponseWriter
	Request           *http.Request

	entityId    string
	entitySubId string
	Body        []byte
	StartTime   time.Time
	IsJwtToken  bool
}

func (rc *RequestContext) GetId() string {
	logtrace.LogWithFunctionName()
	return rc.Id
}

func (rc *RequestContext) GetBody() []byte {
	logtrace.LogWithFunctionName()
	return rc.Body
}

func (rc *RequestContext) GetRequest() *http.Request {
	logtrace.LogWithFunctionName()
	return rc.Request
}

func (rc *RequestContext) GetResponseWriter() http.ResponseWriter {
	logtrace.LogWithFunctionName()
	return rc.ResponseWriter
}

type EventLogger interface {
	Log(actorType, actorId, eventType, entityType, entityId, formatString string, formatData []string, data map[interface{}]interface{})
}

func (rc *RequestContext) SetEntityId(id string) {
	logtrace.LogWithFunctionName()
	rc.entityId = id
}

func (rc *RequestContext) SetEntitySubId(id string) {
	logtrace.LogWithFunctionName()
	rc.entitySubId = id
}

func (rc *RequestContext) GetEntityId() (string, error) {
	logtrace.LogWithFunctionName()
	if rc.entityId == "" {
		return "", errors.New("id not found")
	}
	return rc.entityId, nil
}

func (rc *RequestContext) GetEntitySubId() (string, error) {
	logtrace.LogWithFunctionName()
	if rc.entitySubId == "" {
		return "", errors.New("subId not found")
	}

	return rc.entitySubId, nil
}

func (rc *RequestContext) NewChangeContext() *change.Context {
	logtrace.LogWithFunctionName()
	changeCtx := change.New().SetSourceType(change.SourceTypeRest).
		SetSourceAuth("edge").
		SetSourceMethod(rc.GetRequest().Method).
		SetSourceLocal(rc.GetRequest().Host).
		SetSourceRemote(rc.GetRequest().RemoteAddr)

	if rc.Identity != nil {
		changeCtx.SetChangeAuthorType(change.AuthorTypeIdentity).
			SetChangeAuthorId(rc.Identity.Id).
			SetChangeAuthorName(rc.Identity.Name)
	} else {
		changeCtx.SetChangeAuthorType(change.AuthorTypeUnattributed)
	}

	if rc.Request.Form.Has("traceId") {
		changeCtx.SetTraceId(rc.Request.Form.Get("traceId"))
	}
	return changeCtx
}
