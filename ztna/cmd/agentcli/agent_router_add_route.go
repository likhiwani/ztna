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

	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/common/pb/mgmt_pb"
	"ztna-core/ztna/router"
	"ztna-core/ztna/ztna/cmd/common"
	"github.com/openziti/channel/v3"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

type AgentRouteAction struct {
	AgentOptions
}

func NewRouteCmd(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	action := &AgentRouteAction{
		AgentOptions: AgentOptions{
			CommonOptions: p(),
		},
	}

	cmd := &cobra.Command{
		Args: cobra.ExactArgs(4),
		Use:  "route <controller id> <circuit id> <source-address> <destination-address>",
		RunE: func(cmd *cobra.Command, args []string) error {
			action.Cmd = cmd
			action.Args = args
			return action.MakeChannelRequest(router.AgentAppId, action.makeRequest)
		},
	}

	action.AddAgentOptions(cmd)

	return cmd
}

func (self *AgentRouteAction) makeRequest(ch channel.Channel) error {
	logtrace.LogWithFunctionName()
	route := &ctrl_pb.Route{
		CircuitId: self.Args[1],
		Forwards: []*ctrl_pb.Route_Forward{
			{
				SrcAddress: self.Args[2],
				DstAddress: self.Args[3],
			},
		},
	}

	buf, err := proto.Marshal(route)
	if err != nil {
		return err
	}

	msg := channel.NewMessage(int32(mgmt_pb.ContentType_RouterDebugUpdateRouteRequestType), buf)
	msg.PutStringHeader(int32(mgmt_pb.Header_ControllerId), self.Args[0])
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
