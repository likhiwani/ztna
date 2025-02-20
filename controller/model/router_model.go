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
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
	"github.com/openziti/foundation/v2/genext"
	"github.com/openziti/foundation/v2/versions"
	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
)

type Listener interface {
	AdvertiseAddress() string
	Protocol() string
	Groups() []string
}

type Router struct {
	models.BaseEntity
	Name        string
	Fingerprint *string
	Listeners   []*ctrl_pb.Listener
	Control     channel.Channel
	Connected   atomic.Bool
	ConnectTime time.Time
	VersionInfo *versions.VersionInfo
	routerLinks RouterLinks
	Cost        uint16
	NoTraversal bool
	Disabled    bool
	Metadata    *ctrl_pb.RouterMetadata
}

func (entity *Router) GetLinks() []*Link {
	logtrace.LogWithFunctionName()
	return entity.routerLinks.GetLinks()
}

func (entity *Router) toBoltEntityForUpdate(tx *bbolt.Tx, env Env, _ boltz.FieldChecker) (*db.Router, error) {
	logtrace.LogWithFunctionName()
	return entity.toBoltEntityForCreate(tx, env)
}

func (entity *Router) toBoltEntityForCreate(*bbolt.Tx, Env) (*db.Router, error) {
	logtrace.LogWithFunctionName()
	return &db.Router{
		BaseExtEntity: *boltz.NewExtEntity(entity.Id, entity.Tags),
		Name:          entity.Name,
		Fingerprint:   entity.Fingerprint,
		Cost:          entity.Cost,
		NoTraversal:   entity.NoTraversal,
		Disabled:      entity.Disabled,
	}, nil
}

func (entity *Router) fillFrom(_ Env, _ *bbolt.Tx, boltRouter *db.Router) error {
	logtrace.LogWithFunctionName()
	entity.Name = boltRouter.Name
	entity.Fingerprint = boltRouter.Fingerprint
	entity.Cost = boltRouter.Cost
	entity.NoTraversal = boltRouter.NoTraversal
	entity.Disabled = boltRouter.Disabled
	entity.FillCommon(boltRouter)
	return nil
}

func (entity *Router) AddLinkListener(addr, linkProtocol string, linkCostTags []string, groups []string) {
	logtrace.LogWithFunctionName()
	entity.Listeners = append(entity.Listeners, &ctrl_pb.Listener{
		Address:  addr,
		Protocol: linkProtocol,
		CostTags: linkCostTags,
		Groups:   groups,
	})
}

func (entity *Router) SetLinkListeners(listeners []*ctrl_pb.Listener) {
	logtrace.LogWithFunctionName()
	entity.Listeners = listeners
}

func (entity *Router) SetMetadata(metadata *ctrl_pb.RouterMetadata) {
	logtrace.LogWithFunctionName()
	entity.Metadata = metadata
}

func (entity *Router) HasCapability(capability ctrl_pb.RouterCapability) bool {
	logtrace.LogWithFunctionName()
	return entity.Metadata != nil && genext.Contains(entity.Metadata.Capabilities, capability)
}

func (entity *Router) SupportsRouterLinkMgmt() bool {
	logtrace.LogWithFunctionName()
	if entity.VersionInfo == nil {
		return true
	}
	supportsLinkMgmt, err := entity.VersionInfo.HasMinimumVersion("0.32.1")
	return err != nil || supportsLinkMgmt
}
