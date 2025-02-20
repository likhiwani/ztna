package env

import (
	"bytes"
	"io"
	"net/http"
	"ztna-core/ztna/common/eid"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"
)

func NewRequestContext(rw http.ResponseWriter, r *http.Request) *response.RequestContext {
	logtrace.LogWithFunctionName()
	rid := eid.New()

	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(body))

	requestContext := &response.RequestContext{
		Id:                rid,
		ResponseWriter:    rw,
		Request:           r,
		Body:              body,
		Identity:          nil,
		ApiSession:        nil,
		ActivePermissions: []string{},
	}

	requestContext.Responder = response.NewResponder(requestContext)

	return requestContext
}
