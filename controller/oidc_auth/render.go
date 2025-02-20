package oidc_auth

import (
	"encoding"
	"errors"
	"net/http"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/errorz"
)

// render will attempt to send a responses on the provided http.ResponseWriter. All error output will be directed to the
// http.ResponseWriter. The provided http.ResponseWriter will have had its header sent after calling this function.
func render(w http.ResponseWriter, contentType string, status int, data encoding.BinaryMarshaler) {
	logtrace.LogWithFunctionName()
	payload, err := data.MarshalBinary()

	if err != nil {
		pfxlog.Logger().WithError(err).Error("could not marshal data payload, attempting to respond with a marshalling error")
		internalErr := &rest_model.APIError{
			Code:    errorz.UnhandledCode,
			Message: "could not marshal, see cause",
			Cause: &rest_model.APIErrorCause{
				APIError: rest_model.APIError{
					Code:    "UNHANDLED",
					Message: err.Error(),
				},
			},
		}

		// if there is an err here, we are giving up as we have already tried to recover.
		internalErrPayload, internalErrMarshalErr := internalErr.MarshalBinary()

		if err != nil {
			pfxlog.Logger().WithError(internalErrMarshalErr).Error("could not write marshaling error, failed to marshal internal error, writing '{}'")
		}

		if len(internalErrPayload) == 0 {
			internalErrPayload = []byte("{}")
		}

		w.Header().Set(ContentTypeHeader, contentType)
		w.WriteHeader(http.StatusInternalServerError)

		_, err = w.Write(internalErrPayload)

		if err != nil {
			pfxlog.Logger().WithError(err).WithField("internalErrPayload", internalErrPayload).Error("could not write the internal error payload, giving up")
		}

		return
	}

	w.Header().Set(ContentTypeHeader, contentType)
	w.WriteHeader(status)
	_, err = w.Write(payload)

	if err != nil {
		pfxlog.Logger().WithError(err).Error("error writing payload")
	}
}

// renderJson will attempt to render the provided data as JSON.
func renderJson(w http.ResponseWriter, status int, data encoding.BinaryMarshaler) {
	logtrace.LogWithFunctionName()
	render(w, JsonContentType, status, data)
}

func renderJsonError(w http.ResponseWriter, err error) {
	logtrace.LogWithFunctionName()
	restErr, status := errorToRestApiError(err)
	renderJson(w, status, restErr)
}

func renderJsonApiError(w http.ResponseWriter, err *errorz.ApiError) {
	logtrace.LogWithFunctionName()
	restErr, status := errorToRestApiError(err)
	renderJson(w, status, restErr)
}

func errorToRestApiError(err error) (*rest_model.APIError, int) {
	logtrace.LogWithFunctionName()
	var typedErr *errorz.ApiError
	switch {
	case errors.As(err, &typedErr):
		restErr := &rest_model.APIError{
			Code:    typedErr.Code,
			Message: typedErr.Message,
		}

		if typedErr.Cause != nil {
			causeErr, _ := errorToRestApiError(typedErr.Cause)
			restErr.Cause = &rest_model.APIErrorCause{
				APIError: *causeErr,
			}
		}

		return restErr, typedErr.Status
	default:
		return &rest_model.APIError{
			Code:    "UNHANDLED",
			Message: err.Error(),
		}, http.StatusInternalServerError
	}

}
