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
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/controller/api"
	"ztna-core/ztna/controller/apierror"
	edgeApiError "ztna-core/ztna/controller/apierror"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/errorz"
	"github.com/openziti/storage/ast"
	"github.com/openziti/storage/boltz"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

const (
	EntityNameSelf = "self"
)

func modelToApi[E models.Entity](ae *env.AppEnv, rc *response.RequestContext, mapper func(*env.AppEnv, *response.RequestContext, E) (interface{}, error), es []E) ([]interface{}, error) {
	logtrace.LogWithFunctionName()
	apiEntities := make([]interface{}, 0)

	for _, e := range es {
		al, err := mapper(ae, rc, e)

		if err != nil {
			pfxlog.Logger().
				WithError(err).
				WithField("entityId", e.GetId()).
				WithField("entityType", reflect.TypeOf(e).String()).
				Error("could not convert to API entity")
			err = errors.Wrapf(err, "could not convert %T with id [%v] to API entity", e, e.GetId())
			return nil, err
		}

		apiEntities = append(apiEntities, al)
	}

	return apiEntities, nil
}

func ListWithHandler[E models.Entity](ae *env.AppEnv, rc *response.RequestContext, lister models.EntityRetriever[E],
	mapper func(*env.AppEnv, *response.RequestContext, E) (interface{}, error)) {
	logtrace.LogWithFunctionName()
	ListWithQueryF(ae, rc, lister, mapper, lister.BasePreparedList)
}

func ListWithQueryF[E models.Entity](ae *env.AppEnv,
	rc *response.RequestContext,
	lister models.EntityRetriever[E],
	mapper func(*env.AppEnv, *response.RequestContext, E) (interface{}, error),
	qf func(query ast.Query) (*models.EntityListResult[E], error)) {
	logtrace.LogWithFunctionName()
	ListWithQueryFAndCollector(ae, rc, lister, mapper, defaultToListEnvelope, qf)
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

func ListWithQueryFAndCollector[E models.Entity](ae *env.AppEnv,
	rc *response.RequestContext,
	lister models.EntityRetriever[E],
	mapper func(*env.AppEnv, *response.RequestContext, E) (interface{}, error),
	toEnvelope ApiListEnvelopeFactory,
	qf func(query ast.Query) (*models.EntityListResult[E], error)) {
	logtrace.LogWithFunctionName()
	ListWithEnvelopeFactory(rc, toEnvelope, func(rc *response.RequestContext, queryOptions *PublicQueryOptions) (*QueryResult, error) {
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

		apiEntities, err := modelToApi(ae, rc, mapper, result.GetEntities())
		if err != nil {
			return nil, err
		}

		return NewQueryResult(apiEntities, result.GetMetaData()), nil
	})
}

type modelListF func(rc *response.RequestContext, queryOptions *PublicQueryOptions) (*QueryResult, error)

func List(rc *response.RequestContext, f modelListF) {
	logtrace.LogWithFunctionName()
	ListWithEnvelopeFactory(rc, defaultToListEnvelope, f)
}

func ListWithEnvelopeFactory(rc *response.RequestContext, toEnvelope ApiListEnvelopeFactory, f modelListF) {
	logtrace.LogWithFunctionName()
	qo, err := GetModelQueryOptionsFromRequest(rc.Request)

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

func Create(rc *response.RequestContext, _ response.Responder, linkFactory CreateLinkFactory, creator ModelCreateF) {
	logtrace.LogWithFunctionName()
	CreateWithResponder(rc, rc, linkFactory, creator)
}

func CreateWithResponder(rc *response.RequestContext, rsp response.Responder, linkFactory CreateLinkFactory, creator ModelCreateF) {
	logtrace.LogWithFunctionName()
	id, err := creator()
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

		if uie, ok := err.(*boltz.UniqueIndexDuplicateError); ok {
			rc.RespondWithFieldError(&errorz.FieldError{
				Reason:     uie.Error(),
				FieldName:  uie.Field,
				FieldValue: uie.Value,
			})
			return
		}

		rc.RespondWithError(err)
		return
	}

	rsp.RespondWithCreatedId(id, linkFactory.SelfLinkFromId(id))
}

func DetailWithHandler[E models.Entity](ae *env.AppEnv, rc *response.RequestContext, loader models.EntityRetriever[E],
	mapper func(*env.AppEnv, *response.RequestContext, E) (interface{}, error)) {
	logtrace.LogWithFunctionName()
	Detail(rc, func(rc *response.RequestContext, id string) (interface{}, error) {
		entity, err := loader.BaseLoad(id)
		if err != nil {
			return nil, err
		}
		return mapper(ae, rc, entity)
	})
}

type ModelDetailF func(rc *response.RequestContext, id string) (interface{}, error)

func Detail(rc *response.RequestContext, f ModelDetailF) {
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

	rc.RespondWithOk(apiEntity, &rest_model.Meta{})
}

type ModelDeleteF func(rc *response.RequestContext, id string) error

type DeleteHandler interface {
	Delete(id string, ctx *change.Context) error
}

func DeleteWithHandler(rc *response.RequestContext, deleteHandler DeleteHandler) {
	logtrace.LogWithFunctionName()
	Delete(rc, func(rc *response.RequestContext, id string) error {
		return deleteHandler.Delete(id, rc.NewChangeContext())
	})
}

func Delete(rc *response.RequestContext, deleteF ModelDeleteF) {
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
		} else if refErr, ok := err.(*boltz.ReferenceExistsError); ok {
			rc.RespondWithApiError(edgeApiError.NewCanNotDeleteReferencedEntity(refErr.LocalType, refErr.RemoteType, refErr.RemoteIds, refErr.RemoteField))
		} else {
			rc.RespondWithError(err)
		}
		return
	}

	rc.RespondWithEmptyOk()
}

type ModelUpdateF func(id string) error

func Update(rc *response.RequestContext, updateF ModelUpdateF) {
	logtrace.LogWithFunctionName()
	UpdateAllowEmptyBody(rc, updateF)
}

func UpdateAllowEmptyBody(rc *response.RequestContext, updateF ModelUpdateF) {
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

func Patch(rc *response.RequestContext, patchF ModelPatchF) {
	logtrace.LogWithFunctionName()
	id, err := rc.GetEntityId()

	if err != nil {
		log := pfxlog.Logger()
		log.Error(err)
		rc.RespondWithError(fmt.Errorf("error during patch, retrieving id: %v", err))
		return
	}

	jsonFields, err := api.GetFields(rc.Body)
	if err != nil {
		rc.RespondWithCouldNotParseBody(err)
	}

	err = patchF(id, jsonFields)
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

func listWithId(rc *response.RequestContext, f func(id string) ([]interface{}, error)) {
	logtrace.LogWithFunctionName()
	id, err := rc.GetEntityId()

	if err != nil {
		log := pfxlog.Logger()
		logErr := fmt.Errorf("could not find id property: %v", response.IdPropertyName)
		log.WithField("property", response.IdPropertyName).Error(logErr)
		rc.RespondWithError(err)
		return
	}

	results, err := f(id)

	if err != nil {
		if boltz.IsErrNotFoundErr(err) {
			rc.RespondWithNotFoundWithCause(err)
			return
		}

		log := pfxlog.Logger()
		log.WithField("id", id).WithError(err).Error("could not load associations by id")
		rc.RespondWithError(err)
		return
	}

	count := len(results)

	limit := int64(count)
	offset := int64(0)
	totalCount := int64(count)

	meta := &rest_model.Meta{
		FilterableFields: []string{},
		Pagination: &rest_model.Pagination{
			Limit:      &limit,
			Offset:     &offset,
			TotalCount: &totalCount,
		},
	}

	rc.RespondWithOk(results, meta)
}

// type ListAssocF func(string, func(models.Entity)) error
type listAssocF func(rc *response.RequestContext, id string, queryOptions *PublicQueryOptions) (*QueryResult, error)

type AssociationLister[E models.Entity] interface {
	GetListStore() boltz.Store
	BasePreparedList(query ast.Query) (*models.EntityListResult[E], error)
	BaseLoadInTx(tx *bbolt.Tx, id string) (E, error)
}

func ListAssociationsWithFilter[E models.Entity](ae *env.AppEnv,
	rc *response.RequestContext,
	filterTemplate string,
	entityController AssociationLister[E],
	mapper func(*env.AppEnv, *response.RequestContext, E) (interface{}, error)) {
	logtrace.LogWithFunctionName()
	ListAssociations(rc, func(rc *response.RequestContext, id string, queryOptions *PublicQueryOptions) (*QueryResult, error) {
		query, err := queryOptions.getFullQuery(entityController.GetListStore())
		if err != nil {
			return nil, err
		}

		filter := filterTemplate
		if strings.Contains(filterTemplate, "%v") || strings.Contains(filterTemplate, "%s") {
			filter = fmt.Sprintf(filterTemplate, id)
		}

		filterQuery, err := ast.Parse(entityController.GetListStore(), filter)
		if err != nil {
			return nil, err
		}

		query.SetPredicate(ast.NewAndExprNode(query.GetPredicate(), filterQuery.GetPredicate()))

		result, err := entityController.BasePreparedList(query)
		if err != nil {
			return nil, err
		}

		entities, err := modelToApi(ae, rc, mapper, result.GetEntities())
		if err != nil {
			return nil, err
		}

		return NewQueryResult(entities, &result.QueryMetaData), nil
	})
}

func ListAssociationWithHandler[E models.Entity, A models.Entity](ae *env.AppEnv,
	rc *response.RequestContext,
	lister models.EntityRetriever[E],
	associationLoader AssociationLister[A],
	mapper func(*env.AppEnv, *response.RequestContext, A) (interface{}, error)) {
	logtrace.LogWithFunctionName()
	ListAssociations(rc, func(rc *response.RequestContext, id string, queryOptions *PublicQueryOptions) (*QueryResult, error) {
		// validate that the submitted query is only using public symbols. The query options may contain an final
		// query which has been modified with additional filters
		query, err := queryOptions.getFullQuery(associationLoader.GetListStore())
		if err != nil {
			return nil, err
		}

		result := models.EntityListResult[A]{
			Loader: associationLoader,
		}
		err = lister.PreparedListAssociatedWithHandler(id, associationLoader.GetListStore().GetEntityType(), query, result.Collect)
		if err != nil {
			return nil, err
		}

		apiEntities, err := modelToApi(ae, rc, mapper, result.GetEntities())
		if err != nil {
			return nil, err
		}

		return NewQueryResult(apiEntities, result.GetMetaData()), nil
	})
}

func ListTerminatorAssociations(ae *env.AppEnv, rc *response.RequestContext,
	lister models.EntityRetriever[*model.EdgeService],
	associationLoader *model.TerminatorManager,
	mapper func(ae *env.AppEnv, _ *response.RequestContext, terminator *model.Terminator) (interface{}, error)) {
	logtrace.LogWithFunctionName()
	ListAssociations(rc, func(rc *response.RequestContext, id string, queryOptions *PublicQueryOptions) (*QueryResult, error) {
		// validate that the submitted query is only using public symbols. The query options may contain an final
		// query which has been modified with additional filters
		query, err := queryOptions.getFullQuery(associationLoader.GetStore())
		if err != nil {
			return nil, err
		}

		result := models.EntityListResult[*model.Terminator]{
			Loader: associationLoader,
		}
		err = lister.PreparedListAssociatedWithHandler(id, associationLoader.GetStore().GetEntityType(), query, result.Collect)
		if err != nil {
			return nil, err
		}

		apiEntities := make([]interface{}, 0)

		for _, e := range result.GetEntities() {
			al, err := mapper(ae, rc, e)

			if err != nil {
				return nil, err
			}

			apiEntities = append(apiEntities, al)
		}

		return NewQueryResult(apiEntities, result.GetMetaData()), nil
	})
}

func ListAssociations(rc *response.RequestContext, listF listAssocF) {
	logtrace.LogWithFunctionName()
	id, err := rc.GetEntityId()

	if err != nil {
		log := pfxlog.Logger()
		logErr := fmt.Errorf("could not find id property: %v", response.IdPropertyName)
		log.WithField("property", response.IdPropertyName).Error(logErr)
		rc.RespondWithError(err)
		return
	}

	queryOptions, err := GetModelQueryOptionsFromRequest(rc.Request)

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

	rc.RespondWithOk(result.Result, meta)
}

func MapCreate[T models.Entity](f func(T, *change.Context) error, entity T, rc *response.RequestContext) (string, error) {
	logtrace.LogWithFunctionName()
	err := f(entity, rc.NewChangeContext())
	if err != nil {
		return "", err
	}
	return entity.GetId(), nil
}
