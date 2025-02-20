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

	"github.com/Jeffail/gabs"
	"github.com/spf13/cobra"
)

type createServiceEdgeRouterPolicyOptions struct {
	api.EntityOptions
	edgeRouterRoles []string
	serviceRoles    []string
	semantic        string
}

// NewCreateServiceEdgeRouterPolicyCmd creates the 'edge controller create service-edge-router-policy' command
func NewCreateServiceEdgeRouterPolicyCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	logtrace.LogWithFunctionName()
	options := &createServiceEdgeRouterPolicyOptions{
		EntityOptions: api.NewEntityOptions(out, errOut),
	}

	cmd := &cobra.Command{
		Use:     "service-edge-router-policy <name>",
		Aliases: []string{"serp"},
		Short:   "creates a service-edge-router-policy managed by the Ziti Edge Controller",
		Long:    "creates a service-edge-router-policy managed by the Ziti Edge Controller",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := runCreateServiceEdgeRouterPolicy(options)
			cmdhelper.CheckErr(err)
		},
		SuggestFor: []string{},
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	cmd.Flags().StringSliceVar(&options.edgeRouterRoles, "edge-router-roles", nil, "Edge router roles of the new service edge router policy")
	cmd.Flags().StringSliceVar(&options.serviceRoles, "service-roles", nil, "Identity roles of the new service edge router policy")
	cmd.Flags().StringVar(&options.semantic, "semantic", "AnyOf", "Semantic dictating how multiple attributes should be interpreted. Valid values: AnyOf, AllOf")
	options.AddCommonFlags(cmd)

	return cmd
}

// runCreateServiceEdgeRouterPolicy create a new edgeRouterPolicy on the Ziti Edge Controller
func runCreateServiceEdgeRouterPolicy(o *createServiceEdgeRouterPolicyOptions) error {
	logtrace.LogWithFunctionName()
	edgeRouterRoles, err := convertNamesToIds(o.edgeRouterRoles, "edge-routers", o.Options)
	if err != nil {
		return err
	}

	serviceRoles, err := convertNamesToIds(o.serviceRoles, "services", o.Options)
	if err != nil {
		return err
	}
	entityData := gabs.New()
	api.SetJSONValue(entityData, o.Args[0], "name")
	api.SetJSONValue(entityData, edgeRouterRoles, "edgeRouterRoles")
	api.SetJSONValue(entityData, serviceRoles, "serviceRoles")
	if o.semantic != "" {
		api.SetJSONValue(entityData, o.semantic, "semantic")
	}
	o.SetTags(entityData)

	result, err := CreateEntityOfType("service-edge-router-policies", entityData.String(), &o.Options)
	return o.LogCreateResult("service edge router policy", result, err)
}
