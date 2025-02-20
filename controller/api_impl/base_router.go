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
	"fmt"
	"net/http"
	"reflect"
	"ztna-core/ztna/controller/api"
	"ztna-core/ztna/controller/apierror"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/controller/network"
	"ztna-core/ztna/controller/rest_model"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/errorz"
	"github.com/openziti/storage/ast"
	"github.com/openziti/storage/boltz"
)

const (
	EntityNameSelf = "self"
)

type ModelToApiMapper[T models.Entity] interface {
	ToApi(*network.Network, api.RequestContext, T) (interface{}, error)
}

func modelToApi[T models.Entity](network *network.Network, rc api.RequestContext, mapper ModelToApiMapper[T], es []T) ([]interface{}, error) {
	logtrace.LogWithFunctionName()
	apiEntities := make([]interface{}, 0)

	for _, e := range es {
		al, err := mapper.ToApi(network, rc, e)

		if err != nil {
			return nil, err
		}

		apiEntities = append(apiEntities, al)
	}

	return apiEntities, nil
}

func ListWithHandler[T models.Entity](n *network.Network, rc api.RequestContext, lister models.EntityRetriever[T], mapper ModelToApiMapper[T]) {
	logtrace.LogWithFunctionName()
	ListWithQueryF(n, rc, lister, mapper, lister.BasePreparedList)
}

func ListWithQueryF[T models.Entity](n *network.Network, rc api.RequestContext, lister models.EntityRetriever[T], mapper ModelToApiMapper[T], qf func(query ast.Query) (*models.EntityListResult[T], error)) {
	logtrace.LogWithFunctionName()
	ListWithQueryFAndCollector(n, rc, lister, mapper, defaultToListEnvelope, qf)
}

func defaultToListEnvelope(data []interface{}, meta *rest_model.Meta) interface{} {
	logtrace.LogWithFunctionName()
	return rest_model.Empty{
		Data: data,
		Meta: meta,
	}
}

type ApiListEnvelopeFactory func(data []interface{}, meta *rest_model.Meta) interface{}
type ApiEntityEnvelopeFactory func(data interface{}, meta *rest_model.Meta) interface{}

func ListWithQueryFAndCollector[T models.Entity](n *network.Network, rc api.RequestContext, lister models.EntityRetriever[T], mapper ModelToApiMapper[T], toEnvelope ApiListEnvelopeFactory, qf func(query ast.Query) (*models.EntityListResult[T], error)) {
	logtrace.LogWithFunctionName()
	ListWithEnvelopeFactory(rc, toEnvelope, func(rc api.RequestContext, queryOptions *PublicQueryOptions) (*QueryResult, error) {
		// validate that the submitted query is only using public symbols. The query options may contain an final
		// query which has been modified with additional filters
		query, err := queryOptions.getFullQuery(lister.GetListStore())
		if err != nil {
			return nil, err
		}

		result, err := qf(query)
		if err != nil {
			return nil, err
		}

		apiEntities, err := modelToApi(n, rc, mapper, result.GetEntities())
		if err != nil {
			return nil, err
		}

		return NewQueryResult(apiEntities, result.GetMetaData()), nil
	})
}

type modelListF func(rc api.RequestContext, queryOptions *PublicQueryOptions) (*QueryResult, error)

func ListWithEnvelopeFactory(rc api.RequestContext, toEnvelope ApiListEnvelopeFactory, f modelListF) {
	logtrace.LogWithFunctionName()
	qo, err := GetModelQueryOptionsFromRequest(rc.GetRequest())

	if err != nil {
		log := pfxlog.Logger()
		log.WithField("cause", err).Error("could not build query options")
		rc.RespondWithError(err)
		return
	}

	result, err := f(rc, qo)

	if err != nil {
		log := pfxlog.Logger()
		log.WithField("cause", err).Error("could not convert list")
		rc.RespondWithError(err)
		return
	}

	if result.Result == nil {
		result.Result = []interface{}{}
	}

	meta := &rest_model.Meta{
		Pagination: &rest_model.Pagination{
			Limit:      &result.Limit,
			Offset:     &result.Offset,
			TotalCount: &result.Count,
		},
		FilterableFields: result.FilterableFields,
	}

	switch reflect.TypeOf(result.Result).Kind() {
	case reflect.Slice:
		slice := reflect.ValueOf(result.Result)

		//noinspection GoPreferNilSlice
		elements := []interface{}{}
		for i := 0; i < slice.Len(); i++ {
			elem := slice.Index(i)
			elements = append(elements, elem.Interface())
		}

		envelope := toEnvelope(elements, meta)
		rc.Respond(envelope, http.StatusOK)
	default:
		envelope := toEnvelope([]interface{}{result.Result}, meta)
		rc.Respond(envelope, http.StatusOK)
	}
}

type ModelCreateF func() (string, error)

func Create(rc api.RequestContext, linkFactory CreateLinkFactory, creator ModelCreateF) {
	logtrace.LogWithFunctionName()
	CreateWithResponder(rc, linkFactory, creator)
}

func CreateWithResponder(rsp api.Responder, linkFactory CreateLinkFactory, creator ModelCreateF) {
	logtrace.LogWithFunctionName()
	id, err := creator()
	if err != nil {
		if boltz.IsErrNotFoundErr(err) {
			rsp.RespondWithNotFoundWithCause(err)
			return
		}

		if fe, ok := err.(*errorz.FieldError); ok {
			rsp.RespondWithFieldError(fe)
			return
		}

		if sve, ok := err.(*apierror.ValidationErrors); ok {
			rsp.RespondWithValidationErrors(sve)
			return
		}

		rsp.RespondWithError(err)
		return
	}

	RespondWithCreatedId(rsp, id, linkFactory.SelfLinkFromId(id))
}

func DetailWithHandler[T models.Entity](network *network.Network, rc api.RequestContext, loader models.EntityRetriever[T], mapper ModelToApiMapper[T]) {
	logtrace.LogWithFunctionName()
	Detail(rc, func(rc api.RequestContext, id string) (interface{}, error) {
		entity, err := loader.BaseLoad(id)
		if err != nil {
			return nil, err
		}
		return mapper.ToApi(network, rc, entity)
	})
}

type ModelDetailF func(rc api.RequestContext, id string) (interface{}, error)

func Detail(rc api.RequestContext, f ModelDetailF) {
	logtrace.LogWithFunctionName()
	id, err := rc.GetEntityId()

	if err != nil {
		pfxlog.Logger().Error(err)
		rc.RespondWithError(err)
		return
	}

	apiEntity, err := f(rc, id)

	if err != nil {
		if boltz.IsErrNotFoundErr(err) {
			rc.RespondWithNotFoundWithCause(err)
			return
		}

		pfxlog.Logger().WithField("id", id).WithError(err).Error("could not load entity by id")
		rc.RespondWithError(err)
		return
	}

	RespondWithOk(rc, apiEntity, &rest_model.Meta{})
}

type ModelDeleteF func(rc api.RequestContext, id string) error

type DeleteHandler interface {
	Delete(id string, ctx *change.Context) error
}

type DeleteHandlerF func(id string, ctx *change.Context) error

func (self DeleteHandlerF) Delete(id string, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return self(id, ctx)
}

func DeleteWithHandler(rc api.RequestContext, deleteHandler DeleteHandler) {
	logtrace.LogWithFunctionName()
	Delete(rc, func(rc api.RequestContext, id string) error {
		return deleteHandler.Delete(id, rc.NewChangeContext())
	})
}

func Delete(rc api.RequestContext, deleteF ModelDeleteF) {
	logtrace.LogWithFunctionName()
	id, err := rc.GetEntityId()

	if err != nil {
		log := pfxlog.Logger()
		log.Error(err)
		rc.RespondWithError(err)
		return
	}

	err = deleteF(rc, id)

	if err != nil {
		if boltz.IsErrNotFoundErr(err) {
			rc.RespondWithNotFoundWithCause(err)
		} else {
			rc.RespondWithError(err)
		}
		return
	}

	rc.RespondWithEmptyOk()
}

type ModelUpdateF func(id string) error

func Update(rc api.RequestContext, updateF ModelUpdateF) {
	logtrace.LogWithFunctionName()
	UpdateAllowEmptyBody(rc, updateF)
}

func UpdateAllowEmptyBody(rc api.RequestContext, updateF ModelUpdateF) {
	logtrace.LogWithFunctionName()
	id, err := rc.GetEntityId()

	if err != nil {
		log := pfxlog.Logger()
		log.Error(err)
		rc.RespondWithError(fmt.Errorf("error during update, retrieving id: %v", err))
		return
	}

	if err = updateF(id); err != nil {
		if boltz.IsErrNotFoundErr(err) {
			rc.RespondWithNotFoundWithCause(err)
			return
		}

		if fe, ok := err.(*errorz.FieldError); ok {
			rc.RespondWithFieldError(fe)
			return
		}

		if sve, ok := err.(*apierror.ValidationErrors); ok {
			rc.RespondWithValidationErrors(sve)
			return
		}

		rc.RespondWithError(err)
		return
	}

	rc.RespondWithEmptyOk()
}

type ModelPatchF func(id string, fields fields.UpdatedFields) error

func Patch(rc api.RequestContext, patchF ModelPatchF) {
	logtrace.LogWithFunctionName()
	id, err := rc.GetEntityId()

	if err != nil {
		log := pfxlog.Logger()
		log.Error(err)
		rc.RespondWithError(fmt.Errorf("error during patch, retrieving id: %v", err))
		return
	}

	updatedFields, err := api.GetFields(rc.GetBody())
	if err != nil {
		rc.RespondWithCouldNotParseBody(err)
	}

	err = patchF(id, updatedFields)
	if err != nil {
		if boltz.IsErrNotFoundErr(err) {
			rc.RespondWithNotFoundWithCause(err)
			return
		}

		if fe, ok := err.(*errorz.FieldError); ok {
			rc.RespondWithFieldError(fe)
			return
		}

		if sve, ok := err.(*apierror.ValidationErrors); ok {
			rc.RespondWithValidationErrors(sve)
			return
		}

		rc.RespondWithError(err)
		return
	}

	rc.RespondWithEmptyOk()
}

// type ListAssocF func(string, func(models.Entity)) error
type listAssocF func(rc api.RequestContext, id string, queryOptions *PublicQueryOptions) (*QueryResult, error)

func ListAssociationWithHandler[T models.Entity, A models.Entity](n *network.Network, rc api.RequestContext, sourceR models.EntityRetriever[T], associatedR models.EntityRetriever[A], mapper ModelToApiMapper[A]) {
	logtrace.LogWithFunctionName()
	ListAssociations(rc, func(rc api.RequestContext, id string, queryOptions *PublicQueryOptions) (*QueryResult, error) {
		// validate that the submitted query is only using public symbols. The query options may contain a final
		// query which has been modified with additional filters
		query, err := queryOptions.getFullQuery(sourceR.GetListStore())
		if err != nil {
			return nil, err
		}

		result := models.EntityListResult[A]{
			Loader: associatedR,
		}
		err = sourceR.PreparedListAssociatedWithHandler(id, associatedR.GetListStore().GetEntityType(), query, result.Collect)
		if err != nil {
			return nil, err
		}

		apiEntities, err := modelToApi(n, rc, mapper, result.GetEntities())
		if err != nil {
			return nil, err
		}

		return NewQueryResult(apiEntities, result.GetMetaData()), nil
	})
}

func ListAssociations(rc api.RequestContext, listF listAssocF) {
	logtrace.LogWithFunctionName()
	id, err := rc.GetEntityId()

	if err != nil {
		log := pfxlog.Logger()
		logErr := fmt.Errorf("could not find id property: %v", api.IdPropertyName)
		log.WithField("property", api.IdPropertyName).Error(logErr)
		rc.RespondWithError(err)
		return
	}

	queryOptions, err := GetModelQueryOptionsFromRequest(rc.GetRequest())

	if err != nil {
		rc.RespondWithError(err)
	}

	result, err := listF(rc, id, queryOptions)

	if err != nil {
		if boltz.IsErrNotFoundErr(err) {
			rc.RespondWithNotFoundWithCause(err)
			return
		}

		log := pfxlog.Logger()
		log.WithField("cause", err).Error("could not convert list")
		rc.RespondWithError(err)
		return
	}

	if result.Result == nil {
		result.Result = []interface{}{}
	}

	meta := &rest_model.Meta{
		Pagination: &rest_model.Pagination{
			Limit:      &result.Limit,
			Offset:     &result.Offset,
			TotalCount: &result.Count,
		},
		FilterableFields: result.FilterableFields,
	}

	RespondWithOk(rc, result.Result, meta)
}
