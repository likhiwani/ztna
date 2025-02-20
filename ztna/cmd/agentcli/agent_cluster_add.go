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
	"ztna-core/ztna/controller"
	"ztna-core/ztna/ztna/cmd/common"
	"github.com/openziti/channel/v3"
	"github.com/spf13/cobra"
)

type AgentClusterAddAction struct {
	AgentOptions
	Voter    bool
	MemberId string
}

func NewAgentClusterAdd(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	action := &AgentClusterAddAction{
		AgentOptions: AgentOptions{
			CommonOptions: p(),
		},
	}

	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(1),
		Use:   "add <addr>",
		Short: "adds a node to the controller cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			action.Cmd = cmd
			action.Args = args
			return action.MakeChannelRequest(byte(AgentAppController), action.makeRequest)
		},
	}

	action.AddAgentOptions(cmd)
	cmd.Flags().BoolVar(&action.Voter, "voter", true, "Is this member a voting member")
	cmd.Flags().StringVar(&action.MemberId, "id", "", "The member id. If not provided, it will be looked up")

	return cmd
}

func (self *AgentClusterAddAction) makeRequest(ch channel.Channel) error {
	logtrace.LogWithFunctionName()
	msg := channel.NewMessage(int32(mgmt_pb.ContentType_RaftAddPeerRequestType), nil)
	msg.PutStringHeader(controller.AgentAddrHeader, self.Args[0])
	msg.PutBoolHeader(controller.AgentIsVoterHeader, self.Voter)

	if self.MemberId != "" {
		msg.PutStringHeader(controller.AgentIdHeader, self.MemberId)
	}

	reply, err := msg.WithTimeout(self.timeout).SendForReply(ch)
	if err != nil {
		return err
	}
	result := channel.UnmarshalResult(reply)
	if !result.Success {
		return fmt.Errorf("cluster add failed: %s", result.Message)
	}
	fmt.Println(result.Message)
	return nil
}
