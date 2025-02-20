/*
	(c) Copyright NetFoundry Inc.

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

package link

import (
	"errors"
	"testing"
	"time"
	"ztna-core/ztna/common/inspect"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/controller/idgen"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/env"
	"ztna-core/ztna/router/xgress"
	"ztna-core/ztna/router/xlink"

	"github.com/openziti/channel/v3"
	"github.com/openziti/foundation/v2/goroutines"
	"github.com/openziti/identity"
	"github.com/openziti/metrics"
	"github.com/openziti/transport/v2"
	"github.com/stretchr/testify/require"
)

type testEnv struct {
	metricsRegistry metrics.UsageRegistry
	closeNotify     chan struct{}
	ctrls           env.NetworkControllers
}

func (self *testEnv) GetRouterId() *identity.TokenId {
	logtrace.LogWithFunctionName()
	return &identity.TokenId{
		Token: "test",
	}
}

func (self *testEnv) GetNetworkControllers() env.NetworkControllers {
	logtrace.LogWithFunctionName()
	return self.ctrls
}

func (self *testEnv) GetXlinkDialers() []xlink.Dialer {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testEnv) GetCloseNotify() <-chan struct{} {
	logtrace.LogWithFunctionName()
	return self.closeNotify
}

func (self *testEnv) GetLinkDialerPool() goroutines.Pool {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testEnv) GetRateLimiterPool() goroutines.Pool {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testEnv) GetMetricsRegistry() metrics.UsageRegistry {
	logtrace.LogWithFunctionName()
	return self.metricsRegistry
}

type testLink struct {
	id     string
	key    string
	destId string
}

func (self *testLink) Id() string {
	logtrace.LogWithFunctionName()
	return self.id
}

func (self *testLink) SendPayload(payload *xgress.Payload, timeout time.Duration, payloadType xgress.PayloadType) error {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) SendAcknowledgement(acknowledgement *xgress.Acknowledgement) error {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) SendControl(control *xgress.Control) error {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) InspectCircuit(circuitDetail *inspect.CircuitInspectDetail) {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) Key() string {
	logtrace.LogWithFunctionName()
	return self.key
}

func (self *testLink) Init(metricsRegistry metrics.Registry) error {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) Close() error {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) CloseNotified() error {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) DestinationId() string {
	logtrace.LogWithFunctionName()
	return self.destId
}

func (self *testLink) DestVersion() string {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) LinkProtocol() string {
	logtrace.LogWithFunctionName()
	return "tls"
}

func (self *testLink) DialAddress() string {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) CloseOnce(f func()) {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) IsClosed() bool {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) InspectLink() *inspect.LinkInspectDetail {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) GetAddresses() []*ctrl_pb.LinkConn {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) IsDialed() bool {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) Iteration() uint32 {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) AreFaultsSent() bool {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (self *testLink) DuplicatesRejected() uint32 {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func newTestLink(reg *linkRegistryImpl) *testLink {
	logtrace.LogWithFunctionName()
	linkId := idgen.NewUUIDString()
	destId := idgen.NewUUIDString()
	linkKey := reg.GetLinkKey("", "tls", destId, "")
	return &testLink{
		id:     linkId,
		key:    linkKey,
		destId: destId,
	}
}

func newTestEnv() *testEnv {
	logtrace.LogWithFunctionName()
	closeNotify := make(chan struct{})
	metricsRegistry := metrics.NewUsageRegistry("test", nil, closeNotify)

	ctrls := env.NewNetworkControllers(time.Second, func(address transport.Address, bindHandler channel.BindHandler) error {
		return errors.New("implement me")
	}, env.NewDefaultHeartbeatOptions())
	return &testEnv{
		metricsRegistry: metricsRegistry,
		closeNotify:     closeNotify,
		ctrls:           ctrls,
	}

}

func Test_gcLinkMetrics(t *testing.T) {
	logtrace.LogWithFunctionName()
	routerEnv := newTestEnv()
	defer close(routerEnv.closeNotify)

	reg := NewLinkRegistry(routerEnv).(*linkRegistryImpl)
	orphaned := reg.getOrphanedLinkMetrics()

	req := require.New(t)
	req.Equal(0, len(orphaned))

	linkId := idgen.NewUUIDString()
	registry := routerEnv.metricsRegistry

	linkMetric := func(linkId, name string) string {
		return "link." + linkId + "." + name
	}

	addLinkMetrics := func(linkId string) map[string]metrics.Metric {
		result := map[string]metrics.Metric{}
		addHist := func(name string) {
			result[linkMetric(linkId, name)] = registry.Histogram(linkMetric(linkId, name))
		}

		addMeter := func(name string) {
			result[linkMetric(linkId, name)] = registry.Meter(linkMetric(linkId, name))
		}

		addHist("latency")
		addHist("queue_time")
		addMeter("tx.bytesrate")
		addMeter("tx.msgrate")
		addHist("tx.msgsize")
		addMeter("rx.bytesrate")
		addMeter("rx.msgrate")
		addHist("rx.msgsize")

		metricId := "link.dropped_msgs:" + linkId
		result[metricId] = registry.Meter(metricId)

		return result
	}

	checkLinkMetrics := func(linkId string, m map[string]metrics.Metric, contains bool) {
		checkMetric := func(name string) {
			metricId := linkMetric(linkId, name)
			if contains {
				req.True(m[metricId] != nil, "missing metric %s", metricId)
			} else {
				req.False(m[metricId] != nil, "should not have metric %s", metricId)
			}
		}
		checkPostFixMetric := func(name string) {
			metricId := "link." + name + ":" + linkId
			if contains {
				req.True(m[metricId] != nil, "missing metric %s", metricId)
			} else {
				req.False(m[metricId] != nil, "should not have metric %s", metricId)
			}
		}
		checkMetric("latency")
		checkMetric("queue_time")
		checkMetric("tx.bytesrate")
		checkMetric("tx.msgrate")
		checkMetric("tx.msgsize")
		checkMetric("rx.bytesrate")
		checkMetric("rx.msgrate")
		checkMetric("rx.msgsize")
		checkPostFixMetric("dropped_msgs")
	}

	checkLinkMetricsContains := func(linkId string, m map[string]metrics.Metric) {
		checkLinkMetrics(linkId, m, true)
	}

	checkLinkMetricsDoesntHave := func(linkId string, m map[string]metrics.Metric) {
		checkLinkMetrics(linkId, m, false)
	}

	getRegistryMetrics := func() map[string]metrics.Metric {
		result := map[string]metrics.Metric{}
		registry.EachMetric(func(name string, metric metrics.Metric) {
			result[name] = metric
		})
		return result
	}

	l := addLinkMetrics(linkId)
	registry.Histogram("unrelated.to.links.hist")
	registry.Meter("unrelated.to.links.meter")

	orphaned = reg.getOrphanedLinkMetrics()
	req.Equal(len(l), len(orphaned))
	checkLinkMetricsContains(linkId, orphaned)

	orphaned = reg.gcLinkMetrics(nil)
	req.Equal(len(l), len(orphaned))
	checkLinkMetricsContains(linkId, orphaned)

	orphaned = reg.gcLinkMetrics(orphaned)
	req.Equal(0, len(orphaned))
	checkLinkMetricsDoesntHave(linkId, getRegistryMetrics())

	req.Equal(2, len(getRegistryMetrics()))

	linkId2 := idgen.NewUUIDString()
	link3 := newTestLink(reg)
	link4 := newTestLink(reg)
	linkId5 := idgen.NewUUIDString()

	reg.linkByIdMap[link3.id] = link3
	reg.linkMap[link3.Key()] = link4

	dest := newLinkDest(link4.DestinationId())
	reg.destinations[link4.DestinationId()] = dest
	dest.linkMap[link4.key] = &linkState{
		linkKey: link4.key,
		linkId:  link4.id,
		status:  StatusPending,
		dest:    dest,
	}

	addLinkMetrics(linkId2)
	addLinkMetrics(link3.id)
	addLinkMetrics(link4.id)
	addLinkMetrics(linkId5)

	req.Equal(9*4+2, len(getRegistryMetrics()))

	orphaned = reg.gcLinkMetrics(nil)
	req.Equal(18, len(orphaned))
	checkLinkMetricsContains(linkId2, orphaned)
	checkLinkMetricsContains(linkId5, orphaned)
	checkLinkMetricsDoesntHave(link3.id, orphaned)
	checkLinkMetricsDoesntHave(link4.id, orphaned)

	req.Equal(9*4+2, len(getRegistryMetrics()))

	orphaned = reg.gcLinkMetrics(orphaned)
	req.Equal(0, len(orphaned))
	req.Equal(9*2+2, len(getRegistryMetrics()))

	checkLinkMetricsContains(link3.id, getRegistryMetrics())
	checkLinkMetricsContains(link4.id, getRegistryMetrics())
	checkLinkMetricsDoesntHave(linkId2, getRegistryMetrics())
	checkLinkMetricsDoesntHave(linkId5, getRegistryMetrics())
}
