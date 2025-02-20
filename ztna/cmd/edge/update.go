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

package edge

import (
	"io"

	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/ztna/cmd/api"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"
	"ztna-core/ztna/ztna/util"

	"github.com/Jeffail/gabs"
	"github.com/spf13/cobra"
	"gopkg.in/resty.v1"
)

// newUpdateCmd creates a command object for the "controller update" command
func newUpdateCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	logtrace.LogWithFunctionName()
	cmd := &cobra.Command{
		Use:   "update",
		Short: "updates various entities managed by the Ziti Edge Controller",
		Long:  "updates various entities managed by the Ziti Edge Controller",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			cmdhelper.CheckErr(err)
		},
	}

	cmd.AddCommand(newUpdateAuthenticatorCmd(out, errOut))
	cmd.AddCommand(newUpdateConfigCmd(out, errOut))
	cmd.AddCommand(newUpdateConfigTypeCmd(out, errOut))
	cmd.AddCommand(newUpdateCaCmd(out, errOut))
	cmd.AddCommand(newUpdateEdgeRouterCmd(out, errOut))
	cmd.AddCommand(newUpdateEdgeRouterPolicyCmd(out, errOut))
	cmd.AddCommand(newUpdateIdentityCmd(out, errOut))
	cmd.AddCommand(newUpdateIdentityConfigsCmd(out, errOut))
	cmd.AddCommand(newUpdateServiceCmd(out, errOut))
	cmd.AddCommand(newUpdateServicePolicyCmd(out, errOut))
	cmd.AddCommand(newUpdateServiceEdgeRouterPolicyCmd(out, errOut))
	cmd.AddCommand(newUpdateTerminatorCmd(out, errOut))
	cmd.AddCommand(newUpdatePostureCheckCmd(out, errOut))
	cmd.AddCommand(newUpdateExtJwtSignerCmd(out, errOut))
	cmd.AddCommand(newUpdateAuthPolicySignerCmd(out, errOut))

	return cmd
}

func putEntityOfType(entityType string, body string, options *api.Options) (*gabs.Container, error) {
	logtrace.LogWithFunctionName()
	return updateEntityOfType(entityType, body, options, resty.MethodPut)
}

func patchEntityOfType(entityType string, body string, options *api.Options) (*gabs.Container, error) {
	logtrace.LogWithFunctionName()
	return updateEntityOfType(entityType, body, options, resty.MethodPatch)
}

func postEntityOfType(entityType string, body string, options *api.Options) (*gabs.Container, error) {
	logtrace.LogWithFunctionName()
	return updateEntityOfType(entityType, body, options, resty.MethodPost)
}

func deleteEntityOfTypeWithBody(entityType string, body string, options *api.Options) (*gabs.Container, error) {
	logtrace.LogWithFunctionName()
	return updateEntityOfType(entityType, body, options, resty.MethodDelete)
}

// updateEntityOfType updates an entity of the given type on the Ziti Edge Controller
func updateEntityOfType(entityType string, body string, options *api.Options, method string) (*gabs.Container, error) {
	logtrace.LogWithFunctionName()
	return util.ControllerUpdate(util.EdgeAPI, entityType, body, options.Out, method, options.OutputJSONRequest, options.OutputJSONResponse, options.Timeout, options.Verbose)
}

func doRequest(entityType string, options *api.Options, doRequest func(request *resty.Request, url string) (*resty.Response, error)) (*gabs.Container, error) {
	logtrace.LogWithFunctionName()
	return util.EdgeControllerRequest(entityType, options.Out, options.OutputJSONResponse, options.Timeout, options.Verbose, doRequest)
}
