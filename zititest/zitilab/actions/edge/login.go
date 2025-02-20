package edge

import (
	"errors"
	"path/filepath"

	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/zititest/zitilab/cli"
	"ztna-core/ztna/ztna/cmd/common"
	"ztna-core/ztna/ztna/util"

	"github.com/openziti/fablab/kernel/model"
)

func Login(componentSelector string) model.Action {
	logtrace.LogWithFunctionName()
	return &login{
		componentSelector: componentSelector,
	}
}

func (l *login) Execute(run model.Run) error {
	logtrace.LogWithFunctionName()
	m := run.GetModel()
	ctrl, err := m.SelectComponent(l.componentSelector)
	if err != nil {
		return err
	}

	username := m.MustStringVariable("credentials.edge.username")
	password := m.MustStringVariable("credentials.edge.password")
	edgeApiBaseUrl := ctrl.Host.PublicIp + ":1280"

	caChain := filepath.Join(model.KitBuild(), model.BuildPkiDir, ctrl.Id, "certs", ctrl.Id+".cert")

	if username == "" {
		return errors.New("variable credentials/edge/username must be a string")
	}

	if password == "" {
		return errors.New("variable credentials/edge/password must be a string")
	}

	if _, err = cli.Exec(m, "edge", "login", edgeApiBaseUrl, "-i", model.ActiveInstanceId(), "--ca", caChain, "-u", username, "-p", password); err != nil {
		return err
	}

	if _, err = cli.Exec(m, "edge", "use", model.ActiveInstanceId()); err != nil {
		return err
	}

	common.CliIdentity = model.ActiveInstanceId()
	util.ReloadConfig()

	return nil
}

type login struct {
	componentSelector string
}
