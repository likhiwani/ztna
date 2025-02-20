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

package controller

import (
	"fmt"

	"ztna-core/ztna/common/version"
	edgeSubCmd "ztna-core/ztna/controller/subcmd"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/ztna/cmd/common"
	"ztna-core/ztna/ztna/constants"
	"ztna-core/ztna/ztna/util"

	"github.com/michaelquigley/pfxlog"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewControllerCmd() *cobra.Command {
	logtrace.LogWithFunctionName()
	cmd := &cobra.Command{
		Use:   "controller",
		Short: "Ziti Controller",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				logrus.SetLevel(logrus.DebugLevel)
			}

			switch logFormatter {
			case "pfxlog":
				pfxlog.SetFormatter(pfxlog.NewFormatter(pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti/").StartingToday()))
			case "json":
				pfxlog.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05.000Z"})
			case "text":
				pfxlog.SetFormatter(&logrus.TextFormatter{})
			default:
				// let logrus do its own thing
			}

			util.LogReleaseVersionCheck(constants.ZITI_CONTROLLER)
		},
	}

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	cmd.PersistentFlags().BoolVarP(&cliAgentEnabled, "cliagent", "a", true, "Enable/disabled CLI Agent (enabled by default)")
	cmd.PersistentFlags().StringVar(&cliAgentAddr, "cli-agent-addr", "", "Specify where CLI Agent should list (ex: unix:/tmp/myfile.sock or tcp:127.0.0.1:10001)")
	cmd.PersistentFlags().StringVar(&cliAgentAlias, "cli-agent-alias", "", "Alias which can be used by ziti agent commands to find this instance")
	cmd.PersistentFlags().StringVar(&logFormatter, "log-formatter", "", "Specify log formatter [json|pfxlog|text]")

	cmd.AddCommand(NewRunCmd())
	cmd.AddCommand(NewDeleteSessionsFromConfigCmd())
	cmd.AddCommand(NewDeleteSessionsFromDbCmd())

	versionCmd := common.NewVersionCmd()
	versionCmd.Hidden = true
	versionCmd.Deprecated = "use 'ziti version' instead of 'ziti controller version'"
	cmd.AddCommand(versionCmd)

	edgeSubCmd.AddCommands(cmd, version.GetCmdBuildInfo())

	return cmd
}

var verbose bool
var cliAgentEnabled bool
var cliAgentAddr string
var cliAgentAlias string
var logFormatter string

func Execute() {
	logtrace.LogWithFunctionName()
	if err := NewControllerCmd().Execute(); err != nil {
		fmt.Printf("error: %s\n", err)
	}
}
