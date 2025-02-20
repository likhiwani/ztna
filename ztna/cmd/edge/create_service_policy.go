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
	"strings"

	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/ztna/cmd/api"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"

	"github.com/pkg/errors"

	"github.com/Jeffail/gabs"
	"github.com/spf13/cobra"
)

type createServicePolicyOptions struct {
	api.EntityOptions
	serviceRoles      []string
	identityRoles     []string
	postureCheckRoles []string
	semantic          string
}

// newCreateServicePolicyCmd creates the 'edge controller create service-policy' command
func newCreateServicePolicyCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	logtrace.LogWithFunctionName()
	options := &createServicePolicyOptions{
		EntityOptions: api.NewEntityOptions(out, errOut),
	}

	cmd := &cobra.Command{
		Use:     "service-policy <name> <type>",
		Aliases: []string{"sp"},
		Short:   "creates a service-policy managed by the Ziti Edge Controller",
		Long:    "creates a service-policy managed by the Ziti Edge Controller",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := runCreateServicePolicy(options)
			cmdhelper.CheckErr(err)
		},
		SuggestFor: []string{},
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	cmd.Flags().StringSliceVar(&options.serviceRoles, "service-roles", nil, "Service roles of the new service policy")
	cmd.Flags().StringSliceVar(&options.identityRoles, "identity-roles", nil, "Identity roles of the new service policy")
	cmd.Flags().StringVar(&options.semantic, "semantic", "AnyOf", "Semantic dictating how multiple attributes should be interpreted. Valid values: AnyOf, AllOf")
	cmd.Flags().StringSliceVarP(&options.postureCheckRoles, "posture-check-roles", "p", nil, "Posture check roles of the new service policy")
	options.AddCommonFlags(cmd)

	return cmd
}

// runCreateServicePolicy create a new servicePolicy on the Ziti Edge Controller
func runCreateServicePolicy(o *createServicePolicyOptions) error {
	logtrace.LogWithFunctionName()
	policyType := o.Args[1]
	if policyType != "Bind" && policyType != "Dial" {
		return errors.Errorf("Invalid policy type '%v'. Valid values: [Bind, Dial]", policyType)
	}

	serviceRoles, err := convertNamesToIds(o.serviceRoles, "services", o.Options)
	if err != nil {
		return err
	}

	identityRoles, err := convertNamesToIds(o.identityRoles, "identities", o.Options)
	if err != nil {
		return err
	}

	postureCheckRoles, err := convertNamesToIds(o.postureCheckRoles, "posture-checks", o.Options)
	if err != nil {
		return err
	}

	entityData := gabs.New()
	api.SetJSONValue(entityData, o.Args[0], "name")
	api.SetJSONValue(entityData, o.Args[1], "type")
	api.SetJSONValue(entityData, serviceRoles, "serviceRoles")
	api.SetJSONValue(entityData, identityRoles, "identityRoles")
	api.SetJSONValue(entityData, postureCheckRoles, "postureCheckRoles")
	if o.semantic != "" {
		api.SetJSONValue(entityData, o.semantic, "semantic")
	}
	o.SetTags(entityData)

	result, err := CreateEntityOfType("service-policies", entityData.String(), &o.Options)
	return o.LogCreateResult("service policy", result, err)
}

func convertNamesToIds(roles []string, entityType string, o api.Options) ([]string, error) {
	logtrace.LogWithFunctionName()
	var result []string
	for _, val := range roles {
		if strings.HasPrefix(val, "@") {
			idOrName := strings.TrimPrefix(val, "@")
			id, err := mapNameToID(entityType, idOrName, o)
			if err != nil {
				return nil, err
			}
			result = append(result, "@"+id)
		} else {
			result = append(result, val)
		}
	}
	// The REST endpoints treat an empty slice differently from a nil,
	// and in this case it's important to pass in an empty slice
	if result == nil {
		return []string{}, nil
	}
	return result, nil
}
