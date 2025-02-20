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
	"crypto/x509"
	"time"
	"ztna-core/ztna/common/pb/edge_cmd_pb"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/command"
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/event"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	nfpem "github.com/openziti/foundation/v2/pem"
	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

func NewControllerManager(env Env) *ControllerManager {
	logtrace.LogWithFunctionName()
	manager := &ControllerManager{
		baseEntityManager: newBaseEntityManager[*Controller, *db.Controller](env, env.GetStores().Controller),
	}
	manager.impl = manager

	RegisterManagerDecoder[*Controller](env, manager)

	return manager
}

type ControllerManager struct {
	baseEntityManager[*Controller, *db.Controller]
}

func (self *ControllerManager) newModelEntity() *Controller {
	logtrace.LogWithFunctionName()
	return &Controller{}
}

func (self *ControllerManager) Create(entity *Controller, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchCreate[*Controller](self, entity, ctx)
}

func (self *ControllerManager) ApplyCreate(cmd *command.CreateEntityCommand[*Controller], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	_, err := self.createEntity(cmd.Entity, ctx)
	return err
}

func (self *ControllerManager) Update(entity *Controller, checker fields.UpdatedFields, ctx *change.Context) error {
	logtrace.LogWithFunctionName()
	return DispatchUpdate[*Controller](self, entity, checker, ctx)
}

func (self *ControllerManager) ApplyUpdate(cmd *command.UpdateEntityCommand[*Controller], ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	return self.updateEntity(cmd.Entity, cmd.UpdatedFields, ctx)
}

func (self *ControllerManager) Read(id string) (*Controller, error) {
	logtrace.LogWithFunctionName()
	modelEntity := &Controller{}
	if err := self.readEntity(id, modelEntity); err != nil {
		return nil, err
	}
	return modelEntity, nil
}

func (self *ControllerManager) readInTx(tx *bbolt.Tx, id string) (*Controller, error) {
	logtrace.LogWithFunctionName()
	modelEntity := &Controller{}
	if err := self.readEntityInTx(tx, id, modelEntity); err != nil {
		return nil, err
	}
	return modelEntity, nil
}

func (self *ControllerManager) ReadByName(name string) (*Controller, error) {
	logtrace.LogWithFunctionName()
	modelEntity := &Controller{}
	nameIndex := self.env.GetStores().Controller.GetNameIndex()
	if err := self.readEntityWithIndex("name", []byte(name), nameIndex, modelEntity); err != nil {
		return nil, err
	}
	return modelEntity, nil
}

func (self *ControllerManager) Marshall(entity *Controller) ([]byte, error) {
	logtrace.LogWithFunctionName()
	msg := &edge_cmd_pb.Controller{
		Id:           entity.Id,
		Name:         entity.Name,
		Address:      entity.CtrlAddress,
		CertPem:      entity.CertPem,
		Fingerprint:  entity.Fingerprint,
		IsOnline:     entity.IsOnline,
		LastJoinedAt: timePtrToPb(&entity.LastJoinedAt),
		ApiAddresses: map[string]*edge_cmd_pb.ApiAddressList{},
	}

	for apiKey, instances := range entity.ApiAddresses {
		msg.ApiAddresses[apiKey] = &edge_cmd_pb.ApiAddressList{}
		for _, instance := range instances {
			msg.ApiAddresses[apiKey].Addresses = append(msg.ApiAddresses[apiKey].Addresses, &edge_cmd_pb.ApiAddress{
				Url:     instance.Url,
				Version: instance.Version,
			})
		}
	}

	return proto.Marshal(msg)
}

func (self *ControllerManager) Unmarshall(bytes []byte) (*Controller, error) {
	logtrace.LogWithFunctionName()
	msg := &edge_cmd_pb.Controller{}

	if err := proto.Unmarshal(bytes, msg); err != nil {
		return nil, err
	}

	lastJoinedAt := time.Time{}
	if msg.LastJoinedAt != nil {
		lastJoinedAt = *pbTimeToTimePtr(msg.LastJoinedAt)
	}

	controller := &Controller{
		BaseEntity: models.BaseEntity{
			Id:   msg.Id,
			Tags: edge_cmd_pb.DecodeTags(msg.Tags),
		},
		Name:         msg.Name,
		CtrlAddress:  msg.Address,
		CertPem:      msg.CertPem,
		Fingerprint:  msg.Fingerprint,
		IsOnline:     msg.IsOnline,
		LastJoinedAt: lastJoinedAt,
		ApiAddresses: map[string][]ApiAddress{},
	}

	for apiKey, instanceList := range msg.ApiAddresses {
		controller.ApiAddresses[apiKey] = nil
		if instanceList != nil {
			for _, instance := range instanceList.Addresses {
				controller.ApiAddresses[apiKey] = append(controller.ApiAddresses[apiKey], ApiAddress{
					Url:     instance.Url,
					Version: instance.Version,
				})
			}
		}
	}

	return controller, nil
}

func (self *ControllerManager) getCurrentAsClusterPeer() *event.ClusterPeer {
	logtrace.LogWithFunctionName()
	addr, id, version := self.env.GetRaftInfo()
	tlsConfig, _, _ := self.env.GetServerCert()
	var leaderCerts []*x509.Certificate

	for _, certBytes := range tlsConfig.Certificate {
		if cert, err := x509.ParseCertificate(certBytes); err == nil {
			leaderCerts = append(leaderCerts, cert)
		}
	}

	apiAddresses, _ := self.env.GetApiAddresses()

	return &event.ClusterPeer{
		Id:           id,
		Addr:         addr,
		Version:      version,
		ServerCert:   leaderCerts,
		ApiAddresses: apiAddresses,
	}
}

func (self *ControllerManager) PeersConnected(peers []*event.ClusterPeer, peerConnectedEvent bool) {
	logtrace.LogWithFunctionName()
	controllers := map[string]*Controller{}

	result, err := self.BaseList("true limit none")
	if err != nil {
		pfxlog.Logger().WithError(err).Error("failed to list controllers")
		return
	} else {
		for _, ctrl := range result.Entities {
			controllers[ctrl.Id] = ctrl
		}
	}

	changeCtx := change.New()
	if peerConnectedEvent {
		changeCtx.SetSourceType("raft.peers.connected").
			SetChangeAuthorType(change.AuthorTypeController)
	} else {
		changeCtx.SetSourceType("raft.leadership.gained").
			SetChangeAuthorType(change.AuthorTypeController)
	}

	selfAsPeer := self.getCurrentAsClusterPeer()
	peerFingerprints := ""
	for _, peer := range peers {
		if len(peer.ServerCert) > 0 {
			fingerprint := nfpem.FingerprintFromCertificate(peer.ServerCert[0])

			if peerFingerprints == "" {
				peerFingerprints = fingerprint
			} else {
				peerFingerprints = peerFingerprints + ", " + fingerprint
			}
		}
	}

	pfxlog.Logger().Infof("acting as leader, updating controllers with peers, self: %s, peers: %s", nfpem.FingerprintFromCertificate(selfAsPeer.ServerCert[0]), peerFingerprints)

	if !peerConnectedEvent {
		// add this controller as a "peer" when leadership is gained
		peers = append(peers, selfAsPeer)
	}

	for _, peer := range peers {
		if len(peer.ServerCert) < 1 {
			pfxlog.Logger().Errorf("peer %s has no certificate", peer.Id)
			continue
		}

		newController := &Controller{
			BaseEntity: models.BaseEntity{
				Id: peer.Id,
			},
			Name:         peer.ServerCert[0].Subject.CommonName,
			CertPem:      nfpem.EncodeToString(peer.ServerCert[0]),
			Fingerprint:  nfpem.FingerprintFromCertificate(peer.ServerCert[0]),
			CtrlAddress:  peer.Addr,
			IsOnline:     true,
			LastJoinedAt: time.Now(),
			ApiAddresses: map[string][]ApiAddress{},
		}

		for apiKey, instances := range peer.ApiAddresses {
			newController.ApiAddresses[apiKey] = nil

			for _, instance := range instances {
				newController.ApiAddresses[apiKey] = append(newController.ApiAddresses[apiKey], ApiAddress{
					Url:     instance.Url,
					Version: instance.Version,
				})
			}
		}

		existing := controllers[peer.Id]
		if existing == nil {
			if err = self.Create(newController, changeCtx); err != nil {
				pfxlog.Logger().WithError(err).WithField("ctrlId", peer.Id).
					Error("could not create controller during peer(s) connection")
			}
		} else if peerConnectedEvent || existing.IsChanged(newController) {
			if err = self.Update(newController, nil, changeCtx); err != nil {
				pfxlog.Logger().WithError(err).WithField("ctrlId", peer.Id).
					Error("could not update controller during peer(s) connection")
			}
		}
	}
}

func (self *ControllerManager) PeersDisconnected(peers []*event.ClusterPeer) {
	logtrace.LogWithFunctionName()
	changeCtx := change.New()
	changeCtx.SetSourceType("raft.peers.disconnected").
		SetChangeAuthorType(change.AuthorTypeController)

	disconnectFields := fields.UpdatedFieldsMap{
		db.FieldControllerIsOnline: struct{}{},
	}
	for _, peer := range peers {
		controller := &Controller{
			BaseEntity: models.BaseEntity{
				Id: peer.Id,
			},
			IsOnline: false,
		}

		if err := self.Update(controller, disconnectFields, changeCtx); err != nil {
			pfxlog.Logger().WithError(err).Error("could not update controller during peer(s) disconnection")
		}
	}
}
