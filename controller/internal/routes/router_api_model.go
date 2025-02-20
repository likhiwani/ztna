package routes

import (
	"fmt"

	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/controller/env"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/controller/response"
	"ztna-core/ztna/logtrace"

	"github.com/go-openapi/strfmt"
	"github.com/openziti/foundation/v2/stringz"
)

const EntityNameTransitRouter = "transit-routers"

var TransitRouterLinkFactory = NewBasicLinkFactory(EntityNameTransitRouter)

func MapCreateRouterToModel(router *rest_model.RouterCreate) *model.TransitRouter {
	logtrace.LogWithFunctionName()
	ret := &model.TransitRouter{
		BaseEntity:  models.BaseEntity{},
		Name:        stringz.OrEmpty(router.Name),
		Cost:        uint16(Int64OrDefault(router.Cost)),
		NoTraversal: BoolOrDefault(router.NoTraversal),
	}

	return ret
}

func MapUpdateTransitRouterToModel(id string, router *rest_model.RouterUpdate) *model.TransitRouter {
	logtrace.LogWithFunctionName()
	ret := &model.TransitRouter{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(router.Tags),
			Id:   id,
		},
		Name:        stringz.OrEmpty(router.Name),
		Cost:        uint16(Int64OrDefault(router.Cost)),
		NoTraversal: BoolOrDefault(router.NoTraversal),
		Disabled:    BoolOrDefault(router.Disabled),
	}

	return ret
}

func MapPatchTransitRouterToModel(id string, router *rest_model.RouterPatch) *model.TransitRouter {
	logtrace.LogWithFunctionName()
	ret := &model.TransitRouter{
		BaseEntity: models.BaseEntity{
			Tags: TagsOrDefault(router.Tags),
			Id:   id,
		},
		Name:        router.Name,
		Cost:        uint16(Int64OrDefault(router.Cost)),
		NoTraversal: BoolOrDefault(router.NoTraversal),
		Disabled:    BoolOrDefault(router.Disabled),
	}

	return ret
}

func MapTransitRouterToRestEntity(ae *env.AppEnv, _ *response.RequestContext, router *model.TransitRouter) (interface{}, error) {
	logtrace.LogWithFunctionName()
	return MapTransitRouterToRestModel(ae, router)
}

func MapTransitRouterToRestModel(ae *env.AppEnv, router *model.TransitRouter) (*rest_model.RouterDetail, error) {
	logtrace.LogWithFunctionName()
	isConnected := ae.GetManagers().Router.IsConnected(router.GetId())
	cost := int64(router.Cost)
	ret := &rest_model.RouterDetail{
		BaseEntity:            BaseEntityToRestModel(router, TransitRouterLinkFactory),
		Fingerprint:           router.Fingerprint,
		IsOnline:              &isConnected,
		IsVerified:            &router.IsVerified,
		Name:                  &router.Name,
		UnverifiedFingerprint: router.UnverifiedFingerprint,
		UnverifiedCertPem:     router.UnverifiedCertPem,
		Cost:                  &cost,
		NoTraversal:           &router.NoTraversal,
		Disabled:              &router.Disabled,
	}

	if !router.IsBase && !router.IsVerified {
		var enrollments []*model.Enrollment

		err := ae.GetManagers().TransitRouter.CollectEnrollments(router.Id, func(entity *model.Enrollment) error {
			enrollments = append(enrollments, entity)
			return nil
		})

		if err != nil {
			return nil, err
		}

		if len(enrollments) != 1 {
			return nil, fmt.Errorf("expected enrollment not found for unverified transit router %s", router.Id)
		}
		enrollment := enrollments[0]

		expiresAt := strfmt.DateTime(*enrollment.ExpiresAt)
		createdAt := strfmt.DateTime(*enrollment.IssuedAt)

		ret.EnrollmentExpiresAt = &expiresAt
		ret.EnrollmentCreatedAt = &createdAt
		ret.EnrollmentJWT = &enrollment.Jwt
		ret.EnrollmentToken = &enrollment.Token
	}

	return ret, nil
}
