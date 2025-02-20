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

package network

import (
	"fmt"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/controller/xt_smartrouting"
	"ztna-core/ztna/logtrace"

	"ztna-core/ztna/controller/db"

	"github.com/openziti/transport/v2"
	"github.com/openziti/transport/v2/tcp"
)

func newTestEntityHelper(ctx *model.TestContext, network *Network) *testEntityHelper {
	logtrace.LogWithFunctionName()
	addr := "tcp:0.0.0.0:0"
	transportAddr, err := tcp.AddressParser{}.Parse(addr)
	ctx.NoError(err)

	return &testEntityHelper{
		ctx:           ctx.TestContext,
		network:       network,
		transportAddr: transportAddr,
	}
}

type testEntityHelper struct {
	ctx           *db.TestContext
	network       *Network
	routerIdx     int
	serviceIdx    int
	terminatorIdx int
	transportAddr transport.Address
}

func (self *testEntityHelper) addTestRouter() *model.Router {
	logtrace.LogWithFunctionName()
	router := model.NewRouterForTest(fmt.Sprintf("router-%03d", self.routerIdx), "", self.transportAddr, nil, 0, false)
	self.network.Router.MarkConnected(router)
	self.ctx.NoError(self.network.Router.Create(router, change.New()))
	self.routerIdx++
	return router
}

func (self *testEntityHelper) addTestTerminator(serviceName string, routerName string, instanceId string, isSystem bool) *model.Terminator {
	logtrace.LogWithFunctionName()
	id := fmt.Sprintf("terminator-#%d", self.terminatorIdx)
	term := &model.Terminator{
		BaseEntity: models.BaseEntity{
			Id:       id,
			IsSystem: isSystem,
		},
		Service:    serviceName,
		Router:     routerName,
		InstanceId: instanceId,
		Address:    "ToDo",
	}
	self.ctx.NoError(self.network.Terminator.Create(term, change.New()))
	self.terminatorIdx++
	return term
}

func (self *testEntityHelper) addTestService(serviceName string) *model.Service {
	logtrace.LogWithFunctionName()
	id := fmt.Sprintf("service-#%d", self.serviceIdx)
	svc := &model.Service{
		BaseEntity:         models.BaseEntity{Id: id},
		Name:               serviceName,
		TerminatorStrategy: xt_smartrouting.Name,
	}
	self.serviceIdx++
	self.ctx.NoError(self.network.Service.Create(svc, change.New()))
	return svc
}

func (self *testEntityHelper) discardControllerEvents() {
	logtrace.LogWithFunctionName()
	for {
		select {
		case <-self.network.assembleAndCleanC:
		default:
			return
		}
	}
}

// these debug methods can be used to dump routing evaluation steps to a file for easier analysis

//func initDebug(path string) {
//	f, err := os.Create(path)
//	if err != nil {
//		panic(err)
//	}
//	dbg = &debugger{f: f}
//}
//
//func stopDebug() {
//	if err := dbg.f.Close(); err != nil {
//		panic(err)
//	}
//}
//
//var dbg *debugger
//
//type debugger struct {
//	f   *os.File
//	err error
//}
//
//func debugf(v string, args ...interface{}) {
//	if dbg.err == nil {
//		_, dbg.err = fmt.Fprintf(dbg.f, v, args...)
//	}
//}
//
//func debugDumpDistance(dist map[*Router]int64) {
//	var keys []*Router
//	for k := range dist {
//		keys = append(keys, k)
//	}
//	sort.Slice(keys, func(i, j int) bool {
//		return keys[i].Id < keys[j].Id
//	})
//	for _, k := range keys {
//		debugf("   ->%v = %v\n", k.Id, dist[k])
//	}
//}
//
//func debugPath(p *Path) {
//	nodes := p.Nodes
//	debugf("[r/%v]", nodes[0].Id)
//	if len(p.Links) > 0 {
//		nodes = nodes[1:]
//		for _, link := range p.Links {
//			debugf(" -> [l/%v cost=%v] -> [r/%v]", link.Id, link.Cost, nodes[0].Id)
//		}
//	}
//	debugf("\n")
//}
//
//func debugNetwork(n *Network) {
//	routers := n.AllConnectedRouters()
//	sort.Slice(routers, func(i, j int) bool {
//		return routers[i].Id < routers[j].Id
//	})
//
//	for rIdx, router := range routers {
//		debugf("%v router: %v\n", rIdx, router.Id)
//		links := router.routerLinks.GetLinks()
//		sort.Slice(links, func(i, j int) bool {
//			return links[i].Id < links[j].Id
//		})
//		for lIdx, link := range links {
//			debugf("    %v link %v for (%v -> %v) c: %v sc: %v sl:%v dl: %v\n",
//				lIdx, link.Id, link.Src.Id, link.Dst.Id, link.GetCost(), link.StaticCost, link.SrcLatency, link.DstLatency)
//		}
//	}
//}
