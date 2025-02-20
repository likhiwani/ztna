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

package zitilib_actions

import (
	"fmt"
	"os"
	"path/filepath"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/fablab/kernel/libssh"

	"github.com/openziti/fablab/kernel/model"
	"github.com/openziti/foundation/v2/info"
	"github.com/sirupsen/logrus"
)

func Logs() model.Action {
	logtrace.LogWithFunctionName()
	return &logs{}
}

func (self *logs) Execute(run model.Run) error {
	logtrace.LogWithFunctionName()
	if !run.GetModel().IsBound() {
		return fmt.Errorf("model not bound")
	}

	snapshot := fmt.Sprintf("%d", info.NowInMilliseconds())
	for rn, r := range run.GetModel().Regions {
		for hn, h := range r.Hosts {
			ssh := h.NewSshConfigFactory()
			if err := self.forHost(snapshot, rn, hn, ssh); err != nil {
				return fmt.Errorf("error retrieving logs for [%s/%s] (%w)", rn, hn, err)
			}
		}
	}
	return nil
}

func (self *logs) forHost(snapshot, rn, hn string, ssh libssh.SshConfigFactory) error {
	logtrace.LogWithFunctionName()
	path := filepath.Join(model.AllocateForensicScenario(snapshot, "logs"), rn, hn)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("error creating logs path [%s] for host [%s/%s] (%w)", path, rn, hn, err)
	}
	logrus.Infof("=> [%s]", path)

	fis, err := libssh.RemoteFileList(ssh, ".")
	if err != nil {
		return fmt.Errorf("error retrieving home directory for host [%s/%s] (%w)", rn, hn, err)
	}
	hasLogs := false
	for _, fi := range fis {
		if fi.Name() == "logs" && fi.IsDir() {
			hasLogs = true
			break
		}
	}

	if hasLogs {
		if err := self.forHostDir(path, "logs", ssh); err != nil {
			return fmt.Errorf("error retrieving files from host [%s/%s] (%w)", rn, hn, err)
		}
	}

	return nil
}

func (self *logs) forHostDir(localPath, remotePath string, ssh libssh.SshConfigFactory) error {
	logtrace.LogWithFunctionName()
	fis, err := libssh.RemoteFileList(ssh, remotePath)
	if err != nil {
		return err
	}
	var paths []string
	for _, fi := range fis {
		if fi.IsDir() {
			nextLocalPath := filepath.Join(localPath, fi.Name())
			if err := os.MkdirAll(nextLocalPath, os.ModePerm); err != nil {
				return fmt.Errorf("error creating local path [%s] (%w)", nextLocalPath, err)
			}
			nextRemotePath := filepath.Join(remotePath, fi.Name())
			if err := self.forHostDir(nextLocalPath, nextRemotePath, ssh); err != nil {
				return fmt.Errorf("error recursing into tree (%w)", err)
			}

		} else {
			paths = append(paths, filepath.Join(remotePath, fi.Name()))
		}
	}
	if err := libssh.RetrieveRemoteFiles(ssh, localPath, paths...); err != nil {
		return fmt.Errorf("error retrieving from [%s] (%w)", localPath, err)
	}
	return nil
}

type logs struct{}
