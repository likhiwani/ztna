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
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"ztna-core/ztna/controller/apierror"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/runtime"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/errorz"
)

func NewResponder(rc RequestContext, mapper ResponseMapper) *ResponderImpl {
	logtrace.LogWithFunctionName()
	return &ResponderImpl{
		rc:       rc,
		mapper:   mapper,
		producer: runtime.JSONProducer(),
	}
}

type ResponderImpl struct {
	rc       RequestContext
	mapper   ResponseMapper
	producer runtime.Producer
}

func (responder *ResponderImpl) SetProducer(producer runtime.Producer) {
	logtrace.LogWithFunctionName()
	responder.producer = producer
}

func (responder *ResponderImpl) GetProducer() runtime.Producer {
	logtrace.LogWithFunctionName()
	return responder.producer
}

func (responder *ResponderImpl) RespondWithCouldNotReadBody(err error) {
	logtrace.LogWithFunctionName()
	responder.RespondWithApiError(apierror.NewCouldNotReadBody(err))
}

func (responder *ResponderImpl) RespondWithCouldNotParseBody(err error) {
	logtrace.LogWithFunctionName()
	responder.RespondWithApiError(apierror.NewCouldNotParseBody(err))
}

func (responder *ResponderImpl) RespondWithValidationErrors(errors *apierror.ValidationErrors) {
	logtrace.LogWithFunctionName()
	responder.RespondWithApiError(errorz.NewCouldNotValidate(errors))
}

func (responder *ResponderImpl) RespondWithNotFound() {
	logtrace.LogWithFunctionName()
	responder.RespondWithApiError(errorz.NewNotFound())
}

func (responder *ResponderImpl) RespondWithNotFoundWithCause(cause error) {
	logtrace.LogWithFunctionName()
	apiErr := errorz.NewNotFound()
	apiErr.Cause = cause
	responder.RespondWithApiError(apiErr)
}

func (responder *ResponderImpl) RespondWithFieldError(fe *errorz.FieldError) {
	logtrace.LogWithFunctionName()
	responder.RespondWithApiError(errorz.NewFieldApiError(fe))
}

func (responder *ResponderImpl) RespondWithEmptyOk() {
	logtrace.LogWithFunctionName()
	responder.Respond(responder.mapper.EmptyOkData(), http.StatusOK)
}

func (responder *ResponderImpl) Respond(data interface{}, httpStatus int) {
	logtrace.LogWithFunctionName()
	responder.RespondWithProducer(responder.GetProducer(), data, httpStatus)
}

func (responder *ResponderImpl) RespondWithProducer(producer runtime.Producer, data interface{}, httpStatus int) bool {
	logtrace.LogWithFunctionName()
	w := responder.rc.GetResponseWriter()
	buff := &bytes.Buffer{}
	err := producer.Produce(buff, data)

	if err != nil {
		pfxlog.Logger().WithError(err).
			WithField("requestId", responder.rc.GetId()).
			WithField("path", responder.rc.GetRequest().URL.Path).
			WithError(err).
			Error("could not respond, producer errored")

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte(fmt.Errorf("could not respond, producer errored: %v", err).Error()))

		if err != nil {
			pfxlog.Logger().WithError(err).
				WithField("requestId", responder.rc.GetId()).
				WithField("path", responder.rc.GetRequest().URL.Path).
				WithError(err).
				Error("could not respond with producer error")
			return false
		}

		return true
	}

	w.Header().Set("Content-Length", strconv.Itoa(buff.Len()))
	w.WriteHeader(httpStatus)

	_, err = w.Write(buff.Bytes())

	if err != nil {
		pfxlog.Logger().WithError(err).
			WithField("requestId", responder.rc.GetId()).
			WithField("path", responder.rc.GetRequest().URL.Path).
			WithError(err).
			Error("could not respond, writing to response failed")
	}
	return err == nil
}

func (responder *ResponderImpl) RespondWithError(err error) {
	logtrace.LogWithFunctionName()
	var apiError *errorz.ApiError
	var ok bool

	if apiError, ok = err.(*errorz.ApiError); !ok {
		pfxlog.Logger().WithField("uri", responder.rc.GetRequest().RequestURI).WithError(err).Error("unhandled error returned to REST API")
		apiError = errorz.NewUnhandled(err)
	}

	responder.RespondWithApiError(apiError)
}

func (responder *ResponderImpl) RespondWithApiError(apiError *errorz.ApiError) {
	logtrace.LogWithFunctionName()
	data := responder.mapper.MapApiError(responder.rc.GetId(), apiError)

	producer := responder.rc.GetProducer()
	w := responder.rc.GetResponseWriter()

	if canRespondWithJson(responder.rc.GetRequest()) {
		producer = runtime.JSONProducer()
		w.Header().Set("content-type", "application/json")
	}

	w.WriteHeader(apiError.Status)
	err := producer.Produce(w, data)

	if err != nil {
		pfxlog.Logger().WithError(err).WithField("requestId", responder.rc.GetId()).Error("could not respond with error, producer errored")
	}
}

func canRespondWithJson(request *http.Request) bool {
	logtrace.LogWithFunctionName()
	//if we can return JSON for errors we should as they provide the most
	//information

	canReturnJson := false

	acceptHeaders := request.Header.Values("accept")
	if len(acceptHeaders) == 0 {
		//no accept == "*/*"
		canReturnJson = true
	} else {
		for _, acceptHeader := range acceptHeaders { //look at all headers values
			if canReturnJson {
				break
			}

			for _, value := range strings.Split(acceptHeader, ",") { //each header can have multiple mimeTypes
				mimeType := strings.Split(value, ";")[0] //remove quotients
				mimeType = strings.TrimSpace(mimeType)

				if mimeType == "*/*" || mimeType == "application/json" {
					canReturnJson = true
					break
				}
			}
		}
	}
	return canReturnJson
}
