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
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"ztna-core/ztna/common/inspect"
	"ztna-core/ztna/common/pb/cmd_pb"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/common/pb/mgmt_pb"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/command"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/xt"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3/protobufs"
	"google.golang.org/protobuf/proto"

	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/models"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/storage/boltz"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

const (
	RouterQuiesceFlag   uint32 = 1
	RouterDequiesceFlag uint32 = 2
)

type RouterPresenceHandler interface {
	RouterConnected(r *Router)
	RouterDisconnected(r *Router)
}

type SyncRouterPresenceHandler interface {
	InvokeRouterConnectedSynchronously() bool
	RouterPresenceHandler
}

func NewRouter(id, name, fingerprint string, cost uint16, noTraversal bool) *Router {
	logtrace.LogWithFunctionName()
	if name == "" {
		name = id
	}
	result := &Router{
		BaseEntity:  models.BaseEntity{Id: id},
		Name:        name,
		Fingerprint: &fingerprint,
		Cost:        cost,
		NoTraversal: noTraversal,
	}
	result.routerLinks.allLinks.Store([]*Link{})
	result.routerLinks.linkByRouter.Store(map[string][]*Link{})
	return result
}

type RouterManager struct {
	baseEntityManager[*Router, *db.Router]
	cache     cmap.ConcurrentMap[string, *Router]
	connected cmap.ConcurrentMap[string, *Router]
}

func newRouterManager(env Env) *RouterManager {
	logtrace.LogWithFunctionName()
	routerStore := env.GetStores().Router
	result := &RouterManager{
		baseEntityManager: newBaseEntityManager[*Router, *db.Router](env, routerStore),
		cache:             cmap.New[*Router](),
		connected:         cmap.New[*Router](),
	}
	result.impl = result

	routerStore.AddEntityIdListener(result.UpdateCachedRouter, boltz.EntityUpdated)
	routerStore.AddEntityIdListener(result.HandleRouterDelete, boltz.EntityDeleted)

	RegisterManagerDecoder[*Router](env, result)

	isConnectedSymbol := boltz.NewBoolFuncSymbol(routerStore, "connected", func(id string) bool {
		return result.connected.Has(id)
	})

	if store, ok := routerStore.(boltz.ConfigurableStore); ok {
		store.AddEntitySymbol(isConnectedSymbol)
		store.MakeSymbolPublic(isConnectedSymbol.GetName())
	} else {
		panic("router store is not boltz.ConfigurableStore")
	}

	return result
}

func (self *RouterManager) newModelEntity() *Router {
	logtrace.LogWithFunctionName()
	return &Router{}
}

func (self *RouterManager) MarkConnected(r *Router) {
	logtrace.LogWithFunctionName()
	if router, _ := self.connected.Get(r.Id); router != nil {
		if ch := router.Control; ch != nil {
			if err := ch.Close(); err != nil {
				pfxlog.Logger().WithError(err).Error("error closing control channel")
			}
		}
	}

	r.Connected.Store(true)
	self.connected.Set(r.Id, r)
}

func (self *RouterManager) MarkDisconnected(r *Router) {
	logtrace.LogWithFunctionName()
	r.Connected.Store(false)
	self.connected.RemoveCb(r.Id, func(key string, v *Router, exists bool) bool {
		if exists && v != r {
			pfxlog.Logger().WithField("routerId", r.Id).Info("router not current connect, not clearing from connected map")
			return false
		}
		return exists
	})
	r.routerLinks.Clear()
}

func (self *RouterManager) IsConnected(id string) bool {
	logtrace.LogWithFunctionName()
	return self.connected.Has(id)
}

func (self *RouterManager) GetConnected(id string) *Router {
	logtrace.LogWithFunctionName()
	if router, found := self.connected.Get(id); found {
		return router
	}
	return nil
}

func (self *RouterManager) AllConnected() []*Router {
	logtrace.LogWithFunctionName()
	var routers []*Router
	self.connected.IterCb(func(_ string, router *Router) {
		routers = append(routers, router)
	})
	return routers
}

func (self *RouterManager) ConnectedCount() int {
	logtrace.LogWithFunctionName()
	return self.connected.Count()
}

func (self *RouterManager) Create(entity *Router, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchCreate[*Router](self, entity, ctx)
}

func (self *RouterManager) ApplyCreate(cmd *command.CreateEntityCommand[*Router], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	router := cmd.Entity
	routerId, err := self.createEntity(router, ctx)
	if err == nil {
		self.cache.Set(routerId, router)
	}
	return err
}

func (self *RouterManager) Read(id string) (entity *Router, err error) {
	logtrace.LogWithFunctionName()
	err = self.GetDb().View(func(tx *bbolt.Tx) error {
		entity, err = self.readInTx(tx, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return entity, err
}

func (self *RouterManager) Exists(id string) (bool, error) {
	logtrace.LogWithFunctionName()
	exists := false
	err := self.GetDb().View(func(tx *bbolt.Tx) error {
		exists = self.Store.IsEntityPresent(tx, id)
		return nil
	})
	return exists, err
}

func (self *RouterManager) readUncached(id string) (*Router, error) {
	logtrace.LogWithFunctionName()
	entity := &Router{}
	err := self.GetDb().View(func(tx *bbolt.Tx) error {
		return self.readEntityInTx(tx, id, entity)
	})
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (self *RouterManager) readInTx(tx *bbolt.Tx, id string) (*Router, error) {
	logtrace.LogWithFunctionName()
	if router, _ := self.cache.Get(id); router != nil {
		return router, nil
	}

	entity := &Router{}
	if err := self.readEntityInTx(tx, id, entity); err != nil {
		return nil, err
	}

	self.cache.Set(id, entity)
	return entity, nil
}

func (self *RouterManager) Update(entity *Router, updatedFields fields.UpdatedFields, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchUpdate[*Router](self, entity, updatedFields, ctx)
}

func (self *RouterManager) ApplyUpdate(cmd *command.UpdateEntityCommand[*Router], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	if cmd.Flags == RouterQuiesceFlag {
		return self.ApplyQuiesce(cmd, ctx)
	} else if cmd.Flags == RouterDequiesceFlag {
		return self.ApplyDequiesce(cmd, ctx)
	}

	return self.updateEntity(cmd.Entity, cmd.UpdatedFields, ctx)
}

// QuiesceRouter marks all terminators on the router as failed, so that new traffic will avoid this router, if there's
// any alternative path
func (self *RouterManager) QuiesceRouter(entity *Router, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	cmd := &command.UpdateEntityCommand[*Router]{
		Context:       ctx,
		Updater:       self,
		Entity:        entity,
		UpdatedFields: nil,
		Flags:         RouterQuiesceFlag,
	}

	return self.Dispatch(cmd)
}

// DequiesceRouter returns all routers with a saved precedence that are in a failed state back to their saved state
func (self *RouterManager) DequiesceRouter(entity *Router, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	cmd := &command.UpdateEntityCommand[*Router]{
		Context:       ctx,
		Updater:       self,
		Entity:        entity,
		UpdatedFields: nil,
		Flags:         RouterDequiesceFlag,
	}

	return self.Dispatch(cmd)
}

func (self *RouterManager) ApplyQuiesce(cmd *command.UpdateEntityCommand[*Router], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	return self.UpdateTerminators(cmd.Entity, ctx, func(terminator *db.Terminator) error {
		if terminator.Precedence == xt.Precedences.Failed.String() {
			return nil
		}

		currentPrecedence := terminator.Precedence
		terminator.SavedPrecedence = &currentPrecedence
		terminator.Precedence = xt.Precedences.Failed.String()

		return self.env.GetStores().Terminator.Update(ctx.GetSystemContext(), terminator, boltz.MapFieldChecker{
			db.FieldTerminatorPrecedence:      struct{}{},
			db.FieldTerminatorSavedPrecedence: struct{}{},
		})
	})
}

func (self *RouterManager) ApplyDequiesce(cmd *command.UpdateEntityCommand[*Router], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	return self.UpdateTerminators(cmd.Entity, ctx, func(terminator *db.Terminator) error {
		if terminator.SavedPrecedence == nil || terminator.Precedence != xt.Precedences.Failed.String() {
			return nil
		}

		terminator.Precedence = *terminator.SavedPrecedence
		terminator.SavedPrecedence = nil

		return self.env.GetStores().Terminator.Update(ctx.GetSystemContext(), terminator, boltz.MapFieldChecker{
			db.FieldTerminatorPrecedence:      struct{}{},
			db.FieldTerminatorSavedPrecedence: struct{}{},
		})
	})
}

func (self *RouterManager) UpdateTerminators(router *Router, ctx boltz.MutateContext, f func(terminator *db.Terminator) error) error {
	logtrace.LogWithFunctionName()
	return self.GetDb().Update(ctx, func(ctx boltz.MutateContext) error {
		terminatorIds := self.Store.GetRelatedEntitiesIdList(ctx.Tx(), router.Id, db.EntityTypeTerminators)
		for _, terminatorId := range terminatorIds {
			terminator, _, err := self.env.GetStores().Terminator.FindById(ctx.Tx(), terminatorId)
			if err != nil {
				return err
			}
			if err = f(terminator); err != nil {
				return err
			}
		}
		return nil
	})
}

func (self *RouterManager) HandleRouterDelete(id string) {
	logtrace.LogWithFunctionName()
	log := pfxlog.Logger().WithField("routerId", id)
	log.Info("processing router delete")
	self.cache.Remove(id)

	// if we close the control channel, the router will get removed from the connected cache. We don't do it
	// here because it results in deadlock
	if router, found := self.connected.Get(id); found {
		if ctrl := router.Control; ctrl != nil {
			_ = ctrl.Close()
			log.Warn("connected router deleted, disconnecting router")
		} else {
			log.Warn("deleted router in connected cache doesn't have a connected control channel")
		}
	} else {
		log.Debug("deleted router not connected, no further action required")
	}
}

func (self *RouterManager) UpdateCachedRouter(id string) {
	logtrace.LogWithFunctionName()
	log := pfxlog.Logger().WithField("routerId", id)
	if router, err := self.readUncached(id); err != nil {
		log.WithError(err).Error("failed to read router for cache update")
	} else {
		updateCb := func(key string, v *Router, exist bool) bool {
			if !exist {
				return false
			}

			v.Name = router.Name
			v.Fingerprint = router.Fingerprint
			v.Cost = router.Cost
			v.NoTraversal = router.NoTraversal
			v.Disabled = router.Disabled

			if v.Disabled {
				if ctrl := v.Control; ctrl != nil {
					_ = ctrl.Close()
					log.Warn("connected router disabled, disconnecting router")
				}
			}

			return false
		}

		self.cache.RemoveCb(id, updateCb)
		self.connected.RemoveCb(id, updateCb)
	}
}

func (self *RouterManager) RemoveFromCache(id string) {
	logtrace.LogWithFunctionName()
	self.cache.Remove(id)
}

func (self *RouterManager) Marshall(entity *Router) ([]byte, error) {
	logtrace.LogWithFunctionName()
	tags, err := cmd_pb.EncodeTags(entity.Tags)
	if err != nil {
		return nil, err
	}

	var fingerprint []byte
	if entity.Fingerprint != nil {
		fingerprint = []byte(*entity.Fingerprint)
	}

	msg := &cmd_pb.Router{
		Id:          entity.Id,
		Name:        entity.Name,
		Fingerprint: fingerprint,
		Cost:        uint32(entity.Cost),
		NoTraversal: entity.NoTraversal,
		Disabled:    entity.Disabled,
		Tags:        tags,
	}

	return proto.Marshal(msg)
}

func (self *RouterManager) Unmarshall(bytes []byte) (*Router, error) {
	logtrace.LogWithFunctionName()
	msg := &cmd_pb.Router{}
	if err := proto.Unmarshal(bytes, msg); err != nil {
		return nil, err
	}

	var fingerprint *string
	if msg.Fingerprint != nil {
		tmp := string(msg.Fingerprint)
		fingerprint = &tmp
	}

	return &Router{
		BaseEntity: models.BaseEntity{
			Id:   msg.Id,
			Tags: cmd_pb.DecodeTags(msg.Tags),
		},
		Name:        msg.Name,
		Fingerprint: fingerprint,
		Cost:        uint16(msg.Cost),
		NoTraversal: msg.NoTraversal,
		Disabled:    msg.Disabled,
	}, nil
}

func (self *RouterManager) ValidateRouterSdkTerminators(router *Router, cb func(detail *mgmt_pb.RouterSdkTerminatorsDetails)) {
	logtrace.LogWithFunctionName()
	request := &ctrl_pb.InspectRequest{RequestedValues: []string{"sdk-terminators"}}
	resp := &ctrl_pb.InspectResponse{}
	respMsg, err := protobufs.MarshalTyped(request).WithTimeout(time.Minute).SendForReply(router.Control)
	if err = protobufs.TypedResponse(resp).Unmarshall(respMsg, err); err != nil {
		self.ReportRouterSdkTerminatorsError(router, err, cb)
		return
	}

	var inspectResult *inspect.SdkTerminatorInspectResult
	for _, val := range resp.Values {
		if val.Name == "sdk-terminators" {
			if err = json.Unmarshal([]byte(val.Value), &inspectResult); err != nil {
				self.ReportRouterSdkTerminatorsError(router, err, cb)
				return
			}
		}
	}

	if inspectResult == nil {
		if len(resp.Errors) > 0 {
			err = errors.New(strings.Join(resp.Errors, ","))
			self.ReportRouterSdkTerminatorsError(router, err, cb)
			return
		}
		self.ReportRouterSdkTerminatorsError(router, errors.New("no terminator details returned from router"), cb)
		return
	}

	listResult, err := self.env.GetManagers().Terminator.BaseList(fmt.Sprintf(`router="%s" and binding="edge" limit none`, router.Id))
	if err != nil {
		self.ReportRouterSdkTerminatorsError(router, err, cb)
		return
	}

	result := &mgmt_pb.RouterSdkTerminatorsDetails{
		RouterId:        router.Id,
		RouterName:      router.Name,
		ValidateSuccess: true,
	}

	terminators := map[string]*Terminator{}

	for _, terminator := range listResult.Entities {
		terminators[terminator.Id] = terminator
	}

	for _, entry := range inspectResult.Entries {
		detail := &mgmt_pb.RouterSdkTerminatorDetail{
			TerminatorId:    entry.Id,
			RouterState:     entry.State,
			IsValid:         true,
			OperationActive: entry.OperationActive,
			CreateTime:      entry.CreateTime,
			LastAttempt:     entry.LastAttempt,
		}
		result.Details = append(result.Details, detail)

		if entry.State != "established" {
			detail.IsValid = false
		}

		if _, found := terminators[entry.Id]; found {
			detail.CtrlState = mgmt_pb.TerminatorState_Valid
			delete(terminators, entry.Id)
		} else {
			detail.CtrlState = mgmt_pb.TerminatorState_Unknown
			detail.IsValid = false
		}
	}

	for _, terminator := range terminators {
		detail := &mgmt_pb.RouterSdkTerminatorDetail{
			TerminatorId: terminator.Id,
			CtrlState:    mgmt_pb.TerminatorState_Valid,
			RouterState:  "unknown",
			IsValid:      false,
		}
		result.Details = append(result.Details, detail)
	}

	cb(result)
}

func (self *RouterManager) ReportRouterSdkTerminatorsError(router *Router, err error, cb func(detail *mgmt_pb.RouterSdkTerminatorsDetails)) {
	logtrace.LogWithFunctionName()
	result := &mgmt_pb.RouterSdkTerminatorsDetails{
		RouterId:        router.Id,
		RouterName:      router.Name,
		ValidateSuccess: false,
		Message:         err.Error(),
	}
	cb(result)
}

type RouterLinks struct {
	sync.Mutex
	allLinks     atomic.Value
	linkByRouter atomic.Value
}

func (self *RouterLinks) GetLinks() []*Link {
	logtrace.LogWithFunctionName()
	result := self.allLinks.Load()
	if result == nil {
		return nil
	}
	return result.([]*Link)
}

func (self *RouterLinks) GetLinksByRouter() map[string][]*Link {
	logtrace.LogWithFunctionName()
	result := self.linkByRouter.Load()
	if result == nil {
		return nil
	}
	return result.(map[string][]*Link)
}

func (self *RouterLinks) Add(link *Link, otherRouterId string) {
	logtrace.LogWithFunctionName()
	self.Lock()
	defer self.Unlock()
	links := self.GetLinks()
	newLinks := make([]*Link, 0, len(links)+1)
	newLinks = append(newLinks, links...)
	newLinks = append(newLinks, link)
	self.allLinks.Store(newLinks)

	byRouter := self.GetLinksByRouter()
	newLinksByRouter := map[string][]*Link{}
	maps.Copy(newLinksByRouter, byRouter)
	forRouterList := newLinksByRouter[otherRouterId]
	newForRouterList := append([]*Link{link}, forRouterList...)
	newLinksByRouter[otherRouterId] = newForRouterList
	self.linkByRouter.Store(newLinksByRouter)
}

func (self *RouterLinks) Remove(link *Link, otherRouterId string) {
	logtrace.LogWithFunctionName()
	self.Lock()
	defer self.Unlock()
	links := self.GetLinks()
	newLinks := make([]*Link, 0, len(links))
	for _, l := range links {
		if link != l {
			newLinks = append(newLinks, l)
		}
	}
	self.allLinks.Store(newLinks)

	byRouter := self.GetLinksByRouter()
	newLinksByRouter := map[string][]*Link{}
	maps.Copy(newLinksByRouter, byRouter)
	forRouterList := newLinksByRouter[otherRouterId]
	var newForRouterList []*Link
	for _, l := range forRouterList {
		if l != link {
			newForRouterList = append(newForRouterList, l)
		}
	}
	if len(newForRouterList) == 0 {
		delete(newLinksByRouter, otherRouterId)
	} else {
		newLinksByRouter[otherRouterId] = newForRouterList
	}

	self.linkByRouter.Store(newLinksByRouter)
}

func (self *RouterLinks) Clear() {
	logtrace.LogWithFunctionName()
	self.allLinks.Store([]*Link{})
	self.linkByRouter.Store(map[string][]*Link{})
}
