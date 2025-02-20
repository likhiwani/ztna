package model

import (
	"ztna-core/ztna/controller/db"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/controller/xt"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
)

type Terminator struct {
	models.BaseEntity
	Service         string
	Router          string
	Binding         string
	Address         string
	InstanceId      string
	InstanceSecret  []byte
	Cost            uint16
	Precedence      xt.Precedence
	PeerData        map[uint32][]byte
	HostId          string
	SavedPrecedence xt.Precedence
	SourceCtrl      string
}

func (entity *Terminator) GetServiceId() string {
	logtrace.LogWithFunctionName()
	return entity.Service
}

func (entity *Terminator) GetRouterId() string {
	logtrace.LogWithFunctionName()
	return entity.Router
}

func (entity *Terminator) GetBinding() string {
	logtrace.LogWithFunctionName()
	return entity.Binding
}

func (entity *Terminator) GetAddress() string {
	logtrace.LogWithFunctionName()
	return entity.Address
}

func (entity *Terminator) GetInstanceId() string {
	logtrace.LogWithFunctionName()
	return entity.InstanceId
}

func (entity *Terminator) GetInstanceSecret() []byte {
	logtrace.LogWithFunctionName()
	return entity.InstanceSecret
}

func (entity *Terminator) GetCost() uint16 {
	logtrace.LogWithFunctionName()
	return entity.Cost
}

func (entity *Terminator) GetPrecedence() xt.Precedence {
	logtrace.LogWithFunctionName()
	return entity.Precedence
}

func (entity *Terminator) GetPeerData() xt.PeerData {
	logtrace.LogWithFunctionName()
	return entity.PeerData
}

func (entity *Terminator) GetHostId() string {
	logtrace.LogWithFunctionName()
	return entity.HostId
}

func (entity *Terminator) GetSourceCtrl() string {
	logtrace.LogWithFunctionName()
	return entity.SourceCtrl
}

func (entity *Terminator) toBoltEntityForUpdate(tx *bbolt.Tx, env Env, _ boltz.FieldChecker) (*db.Terminator, error) {
	logtrace.LogWithFunctionName()
	return entity.toBoltEntityForCreate(tx, env)
}

func (entity *Terminator) toBoltEntityForCreate(*bbolt.Tx, Env) (*db.Terminator, error) {
	logtrace.LogWithFunctionName()
	precedence := xt.Precedences.Default.String()
	if entity.Precedence != nil {
		precedence = entity.Precedence.String()
	}

	var savedPrecedence *string
	if entity.SavedPrecedence != nil {
		precedenceStr := entity.SavedPrecedence.String()
		savedPrecedence = &precedenceStr
	}

	return &db.Terminator{
		BaseExtEntity:   *entity.ToBoltBaseExtEntity(),
		Service:         entity.Service,
		Router:          entity.Router,
		Binding:         entity.Binding,
		Address:         entity.Address,
		InstanceId:      entity.InstanceId,
		InstanceSecret:  entity.InstanceSecret,
		Cost:            entity.Cost,
		Precedence:      precedence,
		PeerData:        entity.PeerData,
		HostId:          entity.HostId,
		SavedPrecedence: savedPrecedence,
		SourceCtrl:      entity.SourceCtrl,
	}, nil
}

func (entity *Terminator) fillFrom(_ Env, _ *bbolt.Tx, boltTerminator *db.Terminator) error {
	logtrace.LogWithFunctionName()
	entity.Service = boltTerminator.Service
	entity.Router = boltTerminator.Router
	entity.Binding = boltTerminator.Binding
	entity.Address = boltTerminator.Address
	entity.InstanceId = boltTerminator.InstanceId
	entity.InstanceSecret = boltTerminator.InstanceSecret
	entity.PeerData = boltTerminator.PeerData
	entity.Cost = boltTerminator.Cost
	entity.Precedence = xt.GetPrecedenceForName(boltTerminator.Precedence)
	entity.HostId = boltTerminator.HostId
	entity.SourceCtrl = boltTerminator.SourceCtrl
	entity.FillCommon(boltTerminator)

	if boltTerminator.SavedPrecedence != nil {
		entity.SavedPrecedence = xt.GetPrecedenceForName(*boltTerminator.SavedPrecedence)
	}

	return nil
}
