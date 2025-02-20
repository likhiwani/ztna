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

package agentcli

import (
	logtrace "ztna-core/ztna/logtrace"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"time"

	"ztna-core/ztna/ztna/cmd/common"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"
	"github.com/openziti/agent"
	"github.com/spf13/cobra"
)

type AgentSetGcAction struct {
	AgentOptions
}

func NewSetGcCmd(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	action := &AgentSetGcAction{
		AgentOptions: AgentOptions{
			CommonOptions: p(),
		},
	}

	cmd := &cobra.Command{
		Use:   "setgc gc-percentage",
		Short: "Sets the GC percentage in the target application",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			action.Cmd = cmd
			action.Args = args
			err := action.Run()
			cmdhelper.CheckErr(err)
		},
	}

	action.AddAgentOptions(cmd)

	return cmd
}

// Run implements the command
func (self *AgentSetGcAction) Run() error {
	logtrace.LogWithFunctionName()
	if self.Cmd.Flags().Changed("timeout") {
		time.AfterFunc(self.timeout, func() {
			fmt.Println("operation timed out")
			os.Exit(-1)
		})
	}

	pctArg := self.Args[0]

	perc, err := strconv.ParseInt(pctArg, 10, strconv.IntSize)
	if err != nil {
		return err
	}
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buf, perc)

	return self.MakeRequest(agent.SetGCPercent, buf, self.CopyToWriter(os.Stdout))
}
