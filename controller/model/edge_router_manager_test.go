package model

import (
	"testing"
	"ztna-core/ztna/common/eid"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/logtrace"

	"go.etcd.io/bbolt"
)

func TestEdgeRouterManager(t *testing.T) {
	logtrace.LogWithFunctionName()
	ctx := NewTestContext(t)
	defer ctx.Cleanup()
	ctx.Init()

	t.Run("test get edge routers for service and identity", ctx.testGetEdgeRoutersForServiceAndIdentity)
}

func (ctx *TestContext) testGetEdgeRoutersForServiceAndIdentity(*testing.T) {
	logtrace.LogWithFunctionName()
	edgeRouter := ctx.requireNewEdgeRouter()
	edgeRouter2 := ctx.requireNewEdgeRouter()
	identity := ctx.requireNewIdentity(false)
	service := ctx.requireNewService()
	service.RoleAttributes = []string{eid.New()}
	ctx.NoError(ctx.managers.EdgeService.Update(service, nil, change.New()))

	ctx.requireNewEdgeRouterPolicy(ss("#all"), ss("#all"))

	// test default case, with no limits on service
	ctx.False(ctx.isEdgeRouterAccessible(edgeRouter.Id, identity.Id, service.Id))
	ctx.False(ctx.isEdgeRouterAccessible(edgeRouter2.Id, identity.Id, service.Id))
	ctx.False(ctx.managers.EdgeRouter.IsSharedEdgeRouterPresent(identity.Id, service.Id))

	serp := ctx.requireNewServiceNewEdgeRouterPolicy(ss("@"+service.Id), ss("#"+eid.New()))

	// should not be accessible if we limit to a role no one has
	ctx.False(ctx.isEdgeRouterAccessible(edgeRouter.Id, identity.Id, service.Id))
	ctx.False(ctx.isEdgeRouterAccessible(edgeRouter2.Id, identity.Id, service.Id))
	ctx.False(ctx.managers.EdgeRouter.IsSharedEdgeRouterPresent(identity.Id, service.Id))

	serp.EdgeRouterRoles = []string{"@" + edgeRouter.Id}
	ctx.NoError(ctx.managers.ServiceEdgeRouterPolicy.Update(serp, nil, change.New()))

	// should be accessible if we limit to our specific router
	ctx.True(ctx.isEdgeRouterAccessible(edgeRouter.Id, identity.Id, service.Id))
	ctx.False(ctx.isEdgeRouterAccessible(edgeRouter2.Id, identity.Id, service.Id))
	ctx.True(ctx.managers.EdgeRouter.IsSharedEdgeRouterPresent(identity.Id, service.Id))

}

func (ctx *TestContext) isEdgeRouterAccessible(edgeRouterId, identityId, serviceId string) bool {
	logtrace.LogWithFunctionName()
	found := false
	err := ctx.GetDb().View(func(tx *bbolt.Tx) error {
		result, err := ctx.managers.EdgeRouter.ListForIdentityAndServiceWithTx(tx, identityId, serviceId)
		if err != nil {
			return err
		}
		for _, er := range result.EdgeRouters {
			if er.Id == edgeRouterId {
				found = true
				break
			}
		}
		return nil
	})
	ctx.NoError(err)

	accessAllowed, err := ctx.managers.EdgeRouter.IsAccessToEdgeRouterAllowed(identityId, serviceId, edgeRouterId)
	ctx.NoError(err)
	ctx.Equal(found, accessAllowed)

	return found
}
