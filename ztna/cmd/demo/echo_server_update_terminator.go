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

package demo

import (
	"fmt"
	"time"

	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/ztna/cmd/agentcli"
	"ztna-core/ztna/ztna/cmd/common"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"

	"github.com/openziti/agent"
	"github.com/openziti/channel/v3"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type AgentEchoServerUpdateTerminatorAction struct {
	agentcli.AgentOptions
	precedence string
	cost       uint16
}

func NewAgentEchoServerUpdateTerminatorCmd(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	action := &AgentEchoServerUpdateTerminatorAction{
		AgentOptions: agentcli.AgentOptions{
			CommonOptions: p(),
		},
	}

	updateTerminatorCmd := &cobra.Command{
		Args: cobra.RangeArgs(0, 1),
		Use:  "update-terminator <optional-target>",
		Run: func(cmd *cobra.Command, args []string) {
			action.Cmd = cmd
			action.Args = args
			err := action.Run()
			cmdhelper.CheckErr(err)
		},
	}

	updateTerminatorCmd.Flags().StringVarP(&action.precedence, "precedence", "p", "", "Target precedence [default, required, failed]")
	updateTerminatorCmd.Flags().Uint16VarP(&action.cost, "cost", "c", 0, "Target cost")

	return updateTerminatorCmd
}

// Run implements the command
func (self *AgentEchoServerUpdateTerminatorAction) Run() error {
	logtrace.LogWithFunctionName()
	var addr string
	var err error

	if len(self.Args) == 1 {
		addr, err = agent.ParseGopsAddress(self.Args)
		if err != nil {
			return err
		}
	}

	return agentcli.MakeAgentChannelRequest(addr, EchoServerAppId, self.makeRequest)
}

func (self *AgentEchoServerUpdateTerminatorAction) makeRequest(ch channel.Channel) error {
	logtrace.LogWithFunctionName()
	msg := channel.NewMessage(EchoServerUpdateTerminator, nil)

	if self.Cmd.Flag("precedence").Changed {
		msg.PutStringHeader(EchoServerPrecedenceHeader, self.precedence)
	}
	if self.Cmd.Flag("cost").Changed {
		msg.PutUint16Header(EchoServerCostHeader, self.cost)
	}

	reply, err := msg.WithTimeout(5 * time.Second).SendForReply(ch)
	if err != nil {
		return err
	}
	if reply.ContentType == channel.ContentTypeResultType {
		result := channel.UnmarshalResult(reply)
		if result.Success {
			fmt.Println("success")
		} else {
			fmt.Printf("error: %v\n", result.Message)
		}
	} else {
		return errors.Errorf("unexpected response type: %v", reply.ContentType)
	}
	return nil
}
