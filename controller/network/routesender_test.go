package network

import (
	"testing"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/logtrace"

	"ztna-core/ztna/common/ctrl_msg"
	"ztna-core/ztna/controller/xt"
	"ztna-core/ztna/controller/xt_smartrouting"

	"github.com/michaelquigley/pfxlog"
)

func TestRouteSender_DestroysTerminatorWhenInvalidOnHandleRouteSendAndWeControl(t *testing.T) {
	logtrace.LogWithFunctionName()
	ctx := model.NewTestContext(t)
	defer ctx.Cleanup()

	config := newTestConfig(ctx)
	defer close(config.closeNotify)

	network, err := NewNetwork(config, ctx)
	ctx.NoError(err)

	entityHelper := newTestEntityHelper(ctx, network)
	logger := pfxlog.ChannelLogger("test")

	router1 := entityHelper.addTestRouter()
	router2 := entityHelper.addTestRouter()
	path := &model.Path{
		Nodes: []*model.Router{router1, router2},
	}

	svc := entityHelper.addTestService("svc")

	instanceId := "instanceId"

	term := entityHelper.addTestTerminator(svc.Id, router1.Id, instanceId, true)
	term.Binding = "edge"

	errCode := byte(ctrl_msg.ErrorTypeInvalidTerminator)

	rs := routeSender{
		serviceCounters: network,
		terminators:     network.Terminator,
		attendance:      make(map[string]bool),
	}

	status := &RouteStatus{
		Router:    router1,
		ErrorCode: &errCode,
		Success:   false,
		Attempt:   1,
		Err:       "THIS IS A TEST",
	}

	peerData, cleanup, err := rs.handleRouteSend(1, path, xt_smartrouting.NewFactory().NewStrategy(), status, term, logger)
	ctx.Error(err)
	ctx.ErrorContains(err, status.Err)
	ctx.Nil(peerData)
	ctx.Empty(cleanup)

	newTerm, err := network.Terminator.Read(term.Id)
	ctx.Error(err)
	ctx.Nil(newTerm)
}

func TestRouteSender_SetPrecidenceToNilTerminatorWhenInvalidOnHandleRouteSendAndWeDontControl(t *testing.T) {
	logtrace.LogWithFunctionName()
	ctx := model.NewTestContext(t)
	defer ctx.Cleanup()

	config := newTestConfig(ctx)
	defer close(config.closeNotify)

	network, err := NewNetwork(config, ctx)
	ctx.NoError(err)

	entityHelper := newTestEntityHelper(ctx, network)
	logger := pfxlog.ChannelLogger("test")

	router1 := entityHelper.addTestRouter()
	router2 := entityHelper.addTestRouter()
	path := &model.Path{
		Nodes: []*model.Router{router1, router2},
	}

	svc := entityHelper.addTestService("svc")

	identity := "identity"

	term := entityHelper.addTestTerminator(svc.Id, router1.Id, identity, true)
	term.Binding = "DNE"

	errCode := byte(ctrl_msg.ErrorTypeInvalidTerminator)

	rs := routeSender{
		serviceCounters: network,
		terminators:     network.Terminator,
		attendance:      make(map[string]bool),
	}

	status := &RouteStatus{
		Router:    router1,
		ErrorCode: &errCode,
		Success:   false,
		Attempt:   1,
		Err:       "THIS IS A TEST",
	}

	peerData, cleanup, err := rs.handleRouteSend(1, path, xt_smartrouting.NewFactory().NewStrategy(), status, term, logger)
	ctx.Error(err)
	ctx.ErrorContains(err, status.Err)
	ctx.Nil(peerData)
	ctx.Empty(cleanup)

	newTerm, err := network.Terminator.Read(term.Id)
	ctx.NoError(err)
	ctx.NotNil(newTerm)

	ctx.Equal(xt.Precedences.Failed, newTerm.GetPrecedence())
}
