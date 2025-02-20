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

package fabric

import (
	logtrace "ztna-core/ztna/logtrace"
	"github.com/Jeffail/gabs"
	fabric_rest_client "ztna-core/ztna/controller/rest_client"
	"ztna-core/ztna/ztna/cmd/api"
	"ztna-core/ztna/ztna/cmd/common"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"
	"ztna-core/ztna/ztna/util"
	"github.com/spf13/cobra"
	"gopkg.in/resty.v1"
)

// NewFabricCmd creates a command object for the fabric command
func NewFabricCmd(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	fabricCmd := util.NewEmptyParentCmd("fabric", "Manage the Fabric components of a Ziti network using the Ziti Fabric REST and WebSocket APIs")

	fabricCmd.AddCommand(newCreateCommand(p), newListCmd(p), newUpdateCommand(p), newDeleteCmd(p))
	fabricCmd.AddCommand(newInspectCmd(p))
	fabricCmd.AddCommand(newDbCmd(p))
	fabricCmd.AddCommand(newStreamCommand(p))
	fabricCmd.AddCommand(newRaftCmd(p))
	fabricCmd.AddCommand(newValidateCommand(p))
	return fabricCmd
}

func newCreateCommand(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "creates various entities managed by the Ziti Controller",
		Run: func(cmd *cobra.Command, args []string) {
			cmdhelper.CheckErr(cmd.Help())
		},
	}

	createCmd.AddCommand(newCreateRouterCmd(p))
	createCmd.AddCommand(newCreateServiceCmd(p))
	createCmd.AddCommand(newCreateTerminatorCmd(p))

	return createCmd
}

func newUpdateCommand(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "update various entities managed by the Ziti Controller",
		Run: func(cmd *cobra.Command, args []string) {
			cmdhelper.CheckErr(cmd.Help())
		},
	}

	updateCmd.AddCommand(newUpdateLinkCmd(p))
	updateCmd.AddCommand(newUpdateRouterCmd(p))
	updateCmd.AddCommand(newUpdateServiceCmd(p))
	updateCmd.AddCommand(newUpdateTerminatorCmd(p))

	return updateCmd
}

func newStreamCommand(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	streamCmd := &cobra.Command{
		Use:   "stream",
		Short: "stream fabric operational data",
		Run: func(cmd *cobra.Command, args []string) {
			cmdhelper.CheckErr(cmd.Help())
		},
	}

	streamCmd.AddCommand(NewStreamEventsCmd(p))
	streamTracesCmd := NewStreamTracesCmd(p)
	streamCmd.AddCommand(streamTracesCmd)

	var toggleTracesCmd = &cobra.Command{
		Use:   "toggle",
		Short: "Toggle traces on or off",
		Run: func(cmd *cobra.Command, args []string) {
			cmdhelper.CheckErr(cmd.Help())
		},
	}
	streamTracesCmd.AddCommand(toggleTracesCmd)
	toggleTracesCmd.AddCommand(NewStreamTogglePipeTracesCmd(p))

	return streamCmd
}

func newValidateCommand(p common.OptionsProvider) *cobra.Command {
	logtrace.LogWithFunctionName()
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "validate model data",
		Run: func(cmd *cobra.Command, args []string) {
			cmdhelper.CheckErr(cmd.Help())
		},
	}

	validateCmd.AddCommand(NewValidateTerminatorsCmd(p))
	validateCmd.AddCommand(NewValidateRouterLinksCmd(p))
	validateCmd.AddCommand(NewValidateRouterSdkTerminatorsCmd(p))
	validateCmd.AddCommand(NewValidateRouterDataModelCmd(p))
	validateCmd.AddCommand(NewValidateIdentityConnectionStatusesCmd(p))
	return validateCmd
}

// createEntityOfType create an entity of the given type on the Ziti Controller
func createEntityOfType(entityType string, body string, options *api.Options) (*gabs.Container, error) {
	logtrace.LogWithFunctionName()
	return util.ControllerCreate("fabric", entityType, body, options.Out, options.OutputJSONRequest, options.OutputJSONResponse, options.Timeout, options.Verbose)
}

func patchEntityOfType(entityType string, body string, options *api.Options) (*gabs.Container, error) {
	logtrace.LogWithFunctionName()
	return updateEntityOfType(entityType, body, options, resty.MethodPatch)
}

// updateEntityOfType updates an entity of the given type on the Ziti Edge Controller
func updateEntityOfType(entityType string, body string, options *api.Options, method string) (*gabs.Container, error) {
	logtrace.LogWithFunctionName()
	return util.ControllerUpdate(util.FabricAPI, entityType, body, options.Out, method, options.OutputJSONRequest, options.OutputJSONResponse, options.Timeout, options.Verbose)
}

func WithFabricClient(clientOpts util.ClientOpts, f func(client *fabric_rest_client.ZitiFabric) error) error {
	logtrace.LogWithFunctionName()
	client, err := util.NewFabricManagementClient(clientOpts)
	if err != nil {
		return err
	}
	return f(client)
}
