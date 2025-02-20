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
	"fmt"

	"ztna-core/ztna/common/pb/mgmt_pb"
	"ztna-core/ztna/router"
	"ztna-core/ztna/ztna/cmd/common"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"
	"github.com/openziti/channel/v3"
	"github.com/spf13/cobra"
)

type ToggleCtrlChannelAgentAction struct {
	AgentOptions
	enable bool
}

func NewToggleCtrlChannelAgentCmd(p common.OptionsProvider, name string, enable bool) *cobra.Command {
	logtrace.LogWithFunctionName()
	options := &ToggleCtrlChannelAgentAction{
		AgentOptions: AgentOptions{
			CommonOptions: p(),
		},
		enable: enable,
	}

	cmd := &cobra.Command{
		Args: cobra.RangeArgs(0, 1),
		Use:  name + " <ctrl-id>",
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			cmdhelper.CheckErr(err)
		},
	}

	options.AddAgentOptions(cmd)

	return cmd
}

// Run implements the command
func (self *ToggleCtrlChannelAgentAction) Run() error {
	logtrace.LogWithFunctionName()
	return self.MakeChannelRequest(router.AgentAppId, self.makeRequest)
}

func (self *ToggleCtrlChannelAgentAction) makeRequest(ch channel.Channel) error {
	logtrace.LogWithFunctionName()
	var ctrlId string
	if len(self.Args) > 0 {
		ctrlId = self.Args[0]
	}

	msg := channel.NewMessage(int32(mgmt_pb.ContentType_RouterDebugToggleCtrlChannelRequestType), []byte(ctrlId))
	msg.PutBoolHeader(int32(mgmt_pb.Header_CtrlChanToggle), self.enable)
	reply, err := msg.WithTimeout(self.timeout).SendForReply(ch)
	if err != nil {
		return err
	}
	if reply.ContentType == channel.ContentTypeResultType {
		result := channel.UnmarshalResult(reply)
		if result.Success {
			if len(result.Message) > 0 {
				fmt.Printf("success: %v\n", result.Message)
			} else {
				fmt.Println("success")
			}
		} else {
			fmt.Printf("error: %v\n", result.Message)
		}
	} else {
		fmt.Printf("unexpected response type %v\n", reply.ContentType)
	}
	return nil
}
