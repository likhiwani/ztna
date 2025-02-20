package db

import (
	"ztna-core/ztna/logtrace"

	"github.com/openziti/storage/boltz"
	log "github.com/sirupsen/logrus"
)

func (m *Migrations) createInterceptV1ConfigType(step *boltz.MigrationStep) {
	logtrace.LogWithFunctionName()
	cfg, _ := m.stores.ConfigType.LoadOneByName(step.Ctx.Tx(), interceptV1ConfigType.Name)
	if cfg == nil {
		step.SetError(m.stores.ConfigType.Create(step.Ctx, interceptV1ConfigType))
	} else {
		log.Debugf("'%s' config type already exists. not creating.", interceptV1ConfigType.Name)
	}
}

func (m *Migrations) createHostV1ConfigType(step *boltz.MigrationStep) {
	logtrace.LogWithFunctionName()
	cfg, _ := m.stores.ConfigType.LoadOneByName(step.Ctx.Tx(), hostV1ConfigType.Name)
	if cfg == nil {
		step.SetError(m.stores.ConfigType.Create(step.Ctx, hostV1ConfigType))
	} else {
		log.Debugf("'%s' config type already exists. not creating.", hostV1ConfigType.Name)
	}
}
