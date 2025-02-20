//go:build apitests

package tests

import (
	"io"
	"testing"
	"time"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/env"
	"ztna-core/ztna/router/link"
	"ztna-core/ztna/router/xgress"
	"ztna-core/ztna/router/xlink"
	"ztna-core/ztna/router/xlink_transport"
	"ztna-core/ztna/tests/testutil"

	"github.com/openziti/channel/v3"
	"github.com/openziti/channel/v3/protobufs"
	"github.com/openziti/foundation/v2/goroutines"
	id "github.com/openziti/identity"
	"github.com/openziti/metrics"
	"github.com/openziti/transport/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type testXlinkAcceptor struct {
	link xlink.Xlink
}

func (self *testXlinkAcceptor) Accept(l xlink.Xlink) error {
	logtrace.LogWithFunctionName()
	logrus.Infof("xlink accepted: %+v", l)
	self.link = l
	return nil
}

func (self *testXlinkAcceptor) getLink() xlink.Xlink {
	logtrace.LogWithFunctionName()
	return self.link
}

type testBindHandlerFactory struct{}

func (t testBindHandlerFactory) BindChannel(channel.Binding) error {
	logtrace.LogWithFunctionName()
	return nil
}

func (t testBindHandlerFactory) NewBindHandler(l xlink.Xlink, _ bool, _ bool) channel.BindHandler {
	logtrace.LogWithFunctionName()
	_ = l.Init(metrics.NewRegistry("test", nil))
	return t
}

type testRegistryEnv struct {
	ctrls           env.NetworkControllers
	closeNotify     chan struct{}
	metricsRegistry metrics.UsageRegistry
}

func (self *testRegistryEnv) GetRouterId() *id.TokenId {
	logtrace.LogWithFunctionName()
	return &id.TokenId{
		Token: "test-router",
	}
}

func (self *testRegistryEnv) GetNetworkControllers() env.NetworkControllers {
	logtrace.LogWithFunctionName()
	return self.ctrls
}

func (self *testRegistryEnv) GetXlinkDialers() []xlink.Dialer {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testRegistryEnv) GetCloseNotify() <-chan struct{} {
	logtrace.LogWithFunctionName()
	return self.closeNotify
}

func (self *testRegistryEnv) GetLinkDialerPool() goroutines.Pool {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testRegistryEnv) GetRateLimiterPool() goroutines.Pool {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testRegistryEnv) GetMetricsRegistry() metrics.UsageRegistry {
	logtrace.LogWithFunctionName()
	return self.metricsRegistry
}

type testDial struct {
	Key           string
	LinkId        string
	RouterId      string
	Address       string
	LinkProtocol  string
	RouterVersion string
}

func (self *testDial) GetLinkKey() string {
	logtrace.LogWithFunctionName()
	return self.Key
}

func (self *testDial) GetLinkId() string {
	logtrace.LogWithFunctionName()
	return self.LinkId
}

func (self *testDial) GetRouterId() string {
	logtrace.LogWithFunctionName()
	return self.RouterId
}

func (self *testDial) GetAddress() string {
	logtrace.LogWithFunctionName()
	return self.Address
}

func (self *testDial) GetLinkProtocol() string {
	logtrace.LogWithFunctionName()
	return self.LinkProtocol
}

func (self *testDial) GetRouterVersion() string {
	logtrace.LogWithFunctionName()
	return self.RouterVersion
}

func (self *testDial) GetIteration() uint32 {
	logtrace.LogWithFunctionName()
	return 1
}

func setupEnv() link.Env {
	logtrace.LogWithFunctionName()
	closeNotify := make(chan struct{})
	ctrls := env.NewNetworkControllers(time.Second, nil, env.NewDefaultHeartbeatOptions())
	metricsRegistry := metrics.NewUsageRegistry("test", nil, closeNotify)
	return &testRegistryEnv{
		ctrls:           ctrls,
		closeNotify:     closeNotify,
		metricsRegistry: metricsRegistry,
	}
}

func Test_LinkWithValidCertFromUnknownChain(t *testing.T) {
	logtrace.LogWithFunctionName()
	ctx := NewFabricTestContext(t)
	defer ctx.Teardown()
	ctx.StartServer()
	mgmtClient := ctx.createTestFabricRestClient()
	mgmtClient.EnrollRouter("001", "router-1", "testdata/router/001-client.cert.pem")
	ctx.startRouter(1)
	ctx.Req.NoError(ctx.waitForPort("127.0.0.1:6004", 2*time.Second))

	badId, err := id.LoadClientIdentity(
		"./testdata/invalid_client_cert/client.cert",
		"./testdata/invalid_client_cert/client.key",
		"./testdata/ca/intermediate/certs/ca-chain.cert.pem")
	ctx.Req.NoError(err)

	xla := &testXlinkAcceptor{}
	tcfg := transport.Configuration{
		"split": false,
	}
	metricsRegistery := metrics.NewRegistry("test", nil)
	factory := xlink_transport.NewFactory(xla, testBindHandlerFactory{}, tcfg, link.NewLinkRegistry(setupEnv()), metricsRegistery)
	dialer, err := factory.CreateDialer(badId, nil, tcfg)
	ctx.Req.NoError(err)
	dialReq := &testDial{
		Key:          "default->tls:router1->default",
		LinkId:       "testLinkId",
		Address:      "tls:127.0.0.1:6004",
		RouterId:     "002",
		LinkProtocol: "tls",
	}
	_, err = dialer.Dial(dialReq)
	ctx.Req.Error(err)
	ctx.Req.ErrorIs(err, io.EOF)
}

func Test_UnrequestedLinkFromValidRouter(t *testing.T) {
	logtrace.LogWithFunctionName()
	ctx := NewFabricTestContext(t)
	defer ctx.Teardown()
	ctx.StartServer()
	mgmtClient := ctx.createTestFabricRestClient()
	mgmtClient.EnrollRouter("001", "router-1", "testdata/router/001-client.cert.pem")
	mgmtClient.EnrollRouter("002", "router-2", "testdata/router/002-client.cert.pem")
	ctx.startRouter(1)
	ctx.Req.NoError(ctx.waitForPort("127.0.0.1:6004", 2*time.Second))

	router2Id, err := id.LoadClientIdentity(
		"./testdata/router/002-client.cert.pem",
		"./testdata/router/002.key.pem",
		"./testdata/ca/intermediate/certs/ca-chain.cert.pem")
	ctx.Req.NoError(err)

	xla := &testXlinkAcceptor{}
	tcfg := transport.Configuration{
		"split": false,
	}

	metricsRegistery := metrics.NewRegistry("test", nil)
	factory := xlink_transport.NewFactory(xla, testBindHandlerFactory{}, tcfg, link.NewLinkRegistry(setupEnv()), metricsRegistery)
	dialer, err := factory.CreateDialer(router2Id, nil, tcfg)
	ctx.Req.NoError(err)
	dialReq := &testDial{
		Key:          "default->tls:router1->default",
		LinkId:       "testLinkId",
		Address:      "tls:127.0.0.1:6004",
		RouterId:     "002",
		LinkProtocol: "tls",
	}
	_, err = dialer.Dial(dialReq)
	if err != nil {
		ctx.Req.ErrorIs(err, io.EOF, "unexpected error: %v", err)
	} else {
		for i := int32(0); i < 100 && err == nil; i++ {
			payload := &xgress.Payload{
				CircuitId: "hello",
				Sequence:  i,
				Headers:   nil,
				Data:      []byte{1, 2, 3, 4},
			}
			err = xla.getLink().SendPayload(payload, time.Second, xgress.PayloadTypeXg)
			ctx.Req.NoErrorf(err, "iteration %v", i)
		}
	}
}

func Test_DuplicateLinkWithLinkCloseDialer(t *testing.T) {
	logtrace.LogWithFunctionName()
	ctx := NewFabricTestContext(t)
	defer ctx.Teardown()
	ctx.StartServer()
	mgmtClient := ctx.createTestFabricRestClient()
	mgmtClient.EnrollRouter("001", "router-1", "testdata/router/001-client.cert.pem")
	mgmtClient.EnrollRouter("002", "router-2", "testdata/router/002-client.cert.pem")
	ctx.Teardown()

	ctrlListener := ctx.NewControlChannelListener()
	router1 := ctx.startRouter(1)

	linkChecker := testutil.NewLinkChecker(ctx.Req)
	router1cc := testutil.StartLinkTest(linkChecker, "router-1", ctrlListener, ctx.Req)

	router1Listeners := &ctrl_pb.Listeners{}
	if val, found := router1cc.Underlay().Headers()[int32(ctrl_pb.ControlHeaders_ListenersHeader)]; found {
		ctx.Req.NoError(proto.Unmarshal(val, router1Listeners))
	}

	router2 := ctx.startRouter(2)
	router2cc := testutil.StartLinkTest(linkChecker, "router-2", ctrlListener, ctx.Req)

	router2Listeners := &ctrl_pb.Listeners{}
	if val, found := router1cc.Underlay().Headers()[int32(ctrl_pb.ControlHeaders_ListenersHeader)]; found {
		ctx.Req.NoError(proto.Unmarshal(val, router1Listeners))
	}

	peerUpdates1 := &ctrl_pb.PeerStateChanges{
		Changes: []*ctrl_pb.PeerStateChange{
			{
				Id:        router1.GetRouterId().Token,
				Version:   "v0.0.0",
				State:     ctrl_pb.PeerState_Healthy,
				Listeners: router1Listeners.Listeners,
			},
		},
	}

	ctx.Req.NoError(protobufs.MarshalTyped(peerUpdates1).WithTimeout(time.Second).SendAndWaitForWire(router2cc))

	peerUpdates2 := &ctrl_pb.PeerStateChanges{
		Changes: []*ctrl_pb.PeerStateChange{
			{
				Id:        router2.GetRouterId().Token,
				Version:   "v0.0.0",
				State:     ctrl_pb.PeerState_Healthy,
				Listeners: router2Listeners.Listeners,
			},
		},
	}

	ctx.Req.NoError(protobufs.MarshalTyped(peerUpdates2).WithTimeout(time.Second).SendAndWaitForWire(router1cc))

	time.Sleep(time.Second)

	linkChecker.RequireNoErrors()
	link1 := linkChecker.RequireOneActiveLink()

	linkChecker.RequireNoErrors()
	link2 := linkChecker.RequireOneActiveLink()

	ctx.Req.Equal(link1.Id, link2.Id)

	// Test closing control ch to router 1. On reconnect the existing link should get reported
	ctx.Req.NoError(router1cc.Close())
	_ = testutil.StartLinkTest(linkChecker, "router-1", ctrlListener, ctx.Req)

	time.Sleep(time.Second)
	linkChecker.RequireNoErrors()
	link1 = linkChecker.RequireOneActiveLink()
	ctx.Req.Equal(link1.Id, link2.Id)

	// Test closing control ch to router 2. On reconnect the existing link should get reported
	ctx.Req.NoError(router2cc.Close())
	_ = testutil.StartLinkTest(linkChecker, "router-2", ctrlListener, ctx.Req)

	time.Sleep(time.Second)

	linkChecker.RequireNoErrors()
	link2 = linkChecker.RequireOneActiveLink()
	ctx.Req.Equal(link1.Id, link2.Id)

	// restart router 1
	ctx.Req.NoError(router1.Shutdown())
	ctx.Req.NoError(ctx.waitForPortClose("localhost:6004", 2*time.Second))
	router1 = ctx.startRouter(1)
	defer func() {
		ctx.Req.NoError(router1.Shutdown())
	}()

	router1cc = testutil.StartLinkTest(linkChecker, "router-1", ctrlListener, ctx.Req)
	ctx.Req.NoError(protobufs.MarshalTyped(peerUpdates2).WithTimeout(time.Second).SendAndWaitForWire(router1cc))

	linkChecker.RequireNoErrors()

	//time.Sleep(time.Minute)
	//
	//linkCheck1.RequireNoErrors()
	//link1 = linkCheck1.RequireOneActiveLink()
	//
	//linkCheck2.RequireNoErrors()
	//link2 = linkCheck1.RequireOneActiveLink()

	ctx.Req.Equal(link1.Id, link2.Id)

	ctx.Teardown()
	_ = router1cc.Close()
	_ = router2cc.Close()
	_ = ctrlListener.Close()
}
