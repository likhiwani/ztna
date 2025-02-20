package api_impl

import (
	"net/http"
	"ztna-core/ztna/controller/api"
	"ztna-core/ztna/controller/rest_model"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/foundation/v2/errorz"
)

func RespondWithCreatedId(responder api.Responder, id string, link rest_model.Link) {
	logtrace.LogWithFunctionName()
	createEnvelope := &rest_model.CreateEnvelope{
		Data: &rest_model.CreateLocation{
			Links: rest_model.Links{
				"self": link,
			},
			ID: id,
		},
		Meta: &rest_model.Meta{},
	}

	responder.Respond(createEnvelope, http.StatusCreated)
}

func RespondWithOk(responder api.Responder, data interface{}, meta *rest_model.Meta) {
	logtrace.LogWithFunctionName()
	responder.Respond(&rest_model.Empty{
		Data: data,
		Meta: meta,
	}, http.StatusOK)
}

type FabricResponseMapper struct{}

func (self FabricResponseMapper) EmptyOkData() interface{} {
	logtrace.LogWithFunctionName()
	return &rest_model.Empty{
		Data: map[string]interface{}{},
		Meta: &rest_model.Meta{},
	}
}

func (self FabricResponseMapper) MapApiError(requestId string, apiError *errorz.ApiError) interface{} {
	logtrace.LogWithFunctionName()
	return &rest_model.APIErrorEnvelope{
		Error: ToRestModel(apiError, requestId),
		Meta: &rest_model.Meta{
			APIEnrollmentVersion: ApiVersion,
			APIVersion:           ApiVersion,
		},
	}
}
