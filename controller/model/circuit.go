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

package model

import (
	"sync/atomic"
	"time"
	"ztna-core/ztna/common/datastructures"
	"ztna-core/ztna/common/logcontext"
	"ztna-core/ztna/controller/xt"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/identity"
	"github.com/openziti/storage/objectz"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type Circuit struct {
	Id         string
	ClientId   string
	ServiceId  string
	Terminator xt.CostedTerminator
	Path       *Path
	Tags       map[string]string
	Rerouting  atomic.Bool
	PeerData   xt.PeerData
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (self *Circuit) GetId() string {
	logtrace.LogWithFunctionName()
	return self.Id
}

func (self *Circuit) SetId(string) {
	logtrace.LogWithFunctionName()
	// id cannot be updated
}

func (self *Circuit) GetCreatedAt() time.Time {
	logtrace.LogWithFunctionName()
	return self.CreatedAt
}

func (self *Circuit) GetUpdatedAt() time.Time {
	logtrace.LogWithFunctionName()
	return self.UpdatedAt
}

func (self *Circuit) GetTags() map[string]interface{} {
	logtrace.LogWithFunctionName()
	result := map[string]interface{}{}
	for k, v := range self.Tags {
		result[k] = v
	}
	return result
}

func (self *Circuit) IsSystemEntity() bool {
	logtrace.LogWithFunctionName()
	return false
}

func (self *Circuit) HasRouter(routerId string) bool {
	logtrace.LogWithFunctionName()
	if self == nil || self.Path == nil {
		return false
	}
	for _, node := range self.Path.Nodes {
		if node.Id == routerId {
			return true
		}
	}
	return false
}

func (self *Circuit) IsEndpointRouter(routerId string) bool {
	logtrace.LogWithFunctionName()
	if self == nil || self.Path == nil || len(self.Path.Nodes) == 0 {
		return false
	}
	return self.Path.Nodes[0].Id == routerId || self.Path.Nodes[len(self.Path.Nodes)-1].Id == routerId
}

type CircuitManager struct {
	circuits cmap.ConcurrentMap[string, *Circuit]
	store    *objectz.ObjectStore[*Circuit]
}

func NewCircuitController() *CircuitManager {
	logtrace.LogWithFunctionName()
	result := &CircuitManager{
		circuits: cmap.New[*Circuit](),
	}
	result.store = objectz.NewObjectStore[*Circuit](func() objectz.ObjectIterator[*Circuit] {
		return datastructures.IterateCMap(result.circuits)
	})
	result.store.AddStringSymbol("id", func(entity *Circuit) *string {
		return &entity.Id
	})
	result.store.AddStringSymbol("clientId", func(entity *Circuit) *string {
		return &entity.ClientId
	})
	result.store.AddStringSymbol("service", func(entity *Circuit) *string {
		return &entity.ServiceId
	})
	result.store.AddStringSymbol("terminator", func(entity *Circuit) *string {
		val := entity.Terminator.GetId()
		return &val
	})
	result.store.AddDatetimeSymbol("createdAt", func(entity *Circuit) *time.Time {
		return &entity.CreatedAt
	})
	result.store.AddDatetimeSymbol("updatedAt", func(entity *Circuit) *time.Time {
		return &entity.CreatedAt
	})
	return result
}

func (self *CircuitManager) GetStore() *objectz.ObjectStore[*Circuit] {
	logtrace.LogWithFunctionName()
	return self.store
}

func (self *CircuitManager) Add(circuit *Circuit) {
	logtrace.LogWithFunctionName()
	self.circuits.Set(circuit.Id, circuit)
}

func (self *CircuitManager) Get(id string) (*Circuit, bool) {
	logtrace.LogWithFunctionName()
	if circuit, found := self.circuits.Get(id); found {
		return circuit, true
	}
	return nil, false
}

func (self *CircuitManager) All() []*Circuit {
	logtrace.LogWithFunctionName()
	var circuits []*Circuit
	self.circuits.IterCb(func(_ string, circuit *Circuit) {
		circuits = append(circuits, circuit)
	})
	return circuits
}

func (self *CircuitManager) Remove(circuit *Circuit) {
	logtrace.LogWithFunctionName()
	self.circuits.Remove(circuit.Id)
}

type CreateCircuitParams interface {
	GetServiceId() string
	GetSourceRouter() *Router
	GetClientId() *identity.TokenId
	GetCircuitTags(terminator xt.CostedTerminator) map[string]string
	GetLogContext() logcontext.Context
	GetDeadline() time.Time
}
