/*
	Copyright 2020 NetFoundry Inc.

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

package zitilib_runlevel_5_operation

import (
	"fmt"
	"strings"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/fablab/kernel/libssh"
	"github.com/openziti/fablab/kernel/model"
	"github.com/sirupsen/logrus"
)

func Loop3Dialer(host *model.Host, scenario, endpoint string, joiner chan struct{}, extraArgs ...string) model.Stage {
	logtrace.LogWithFunctionName()
	return &loopDialer{
		host:      host,
		scenario:  scenario,
		endpoint:  endpoint,
		joiner:    joiner,
		subcmd:    "loop3",
		extraArgs: extraArgs,
	}
}

func LoopDialer(host *model.Host, scenario, endpoint string, joiner chan struct{}, extraArgs ...string) model.Stage {
	logtrace.LogWithFunctionName()
	return &loopDialer{
		host:      host,
		scenario:  scenario,
		endpoint:  endpoint,
		joiner:    joiner,
		subcmd:    "loop2",
		extraArgs: extraArgs,
	}
}

func (self *loopDialer) Execute(run model.Run) error {
	logtrace.LogWithFunctionName()
	ssh := self.host.NewSshConfigFactory()
	if err := libssh.RemoteKill(ssh, fmt.Sprintf("ziti-fabric-test %v dialer", self.subcmd)); err != nil {
		return fmt.Errorf("error killing %v listeners (%w)", self.subcmd, err)
	}

	go self.run(run)
	return nil
}

func (self *loopDialer) run(ctx model.Run) {
	logtrace.LogWithFunctionName()
	defer func() {
		if self.joiner != nil {
			close(self.joiner)
			logrus.Debugf("closed joiner")
		}
	}()

	ssh := self.host.NewSshConfigFactory()
	logFile := fmt.Sprintf("/home/%s/logs/%v-dialer-%s.log", ssh.User(), self.subcmd, ctx.GetId())
	dialerCmd := fmt.Sprintf("/home/%s/fablab/bin/ziti-fabric-test %v dialer /home/%s/fablab/cfg/%s -e %s -s %s %s >> %s 2>&1",
		ssh.User(), self.subcmd, ssh.User(), self.scenario, self.endpoint, self.host.GetId(), strings.Join(self.extraArgs, " "), logFile)
	if output, err := libssh.RemoteExec(ssh, dialerCmd); err != nil {
		logrus.Errorf("error starting loop dialer [%s] (%v)", output, err)
	}
}

type loopDialer struct {
	host      *model.Host
	endpoint  string
	scenario  string
	joiner    chan struct{}
	subcmd    string
	extraArgs []string
}
