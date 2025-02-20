package edge

import (
	"path/filepath"
	"strings"
	"ztna-core/ztna/logtrace"
	zitilib_actions "ztna-core/ztna/zititest/zitilab/actions"
	"ztna-core/ztna/zititest/zitilab/cli"

	"github.com/openziti/fablab/kernel/libssh"
	"github.com/openziti/fablab/kernel/model"
)

func InitIdentities(componentSpec string, concurrency int) model.Action {
	logtrace.LogWithFunctionName()
	return &initIdentitiesAction{
		componentSpec: componentSpec,
		concurrency:   concurrency,
	}
}

func (action *initIdentitiesAction) Execute(run model.Run) error {
	logtrace.LogWithFunctionName()
	return run.GetModel().ForEachComponent(action.componentSpec, action.concurrency, func(c *model.Component) error {
		if err := zitilib_actions.EdgeExec(run.GetModel(), "delete", "identity", c.Id); err != nil {
			return err
		}

		return action.createAndEnrollIdentity(run, c)
	})
}

func (action *initIdentitiesAction) createAndEnrollIdentity(run model.Run, c *model.Component) error {
	logtrace.LogWithFunctionName()
	ssh := c.GetHost().NewSshConfigFactory()

	jwtFileName := filepath.Join(run.GetTmpDir(), c.Id+".jwt")

	err := zitilib_actions.EdgeExec(c.GetModel(), "create", "identity", c.Id,
		"--jwt-output-file", jwtFileName,
		"-a", strings.Join(c.Tags, ","))

	if err != nil {
		return err
	}

	configFileName := filepath.Join(run.GetTmpDir(), c.Id+".json")

	_, err = cli.Exec(c.GetModel(), "edge", "enroll", "--jwt", jwtFileName, "--out", configFileName)

	if err != nil {
		return err
	}

	remoteConfigFile := "/home/ubuntu/fablab/cfg/" + c.Id + ".json"
	return libssh.SendFile(ssh, configFileName, remoteConfigFile)
}

type initIdentitiesAction struct {
	componentSpec string
	concurrency   int
}
