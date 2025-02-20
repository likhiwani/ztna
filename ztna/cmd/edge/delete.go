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
	"net/url"
	"strings"

	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/ztna/cmd/api"
	"ztna-core/ztna/ztna/cmd/common"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"
	"ztna-core/ztna/ztna/util"

	"github.com/fatih/color"
	"github.com/openziti/storage/boltz"

	"github.com/spf13/cobra"
)

type deleteOptions struct {
	*api.Options
	ignoreMissing bool
}

// newDeleteCmd creates a command object for the "edge controller delete" command
func newDeleteCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	logtrace.LogWithFunctionName()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "deletes various entities managed by the Ziti Edge Controller",
		Long:  "deletes various entities managed by the Ziti Edge Controller",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			cmdhelper.CheckErr(err)
		},
	}

	newOptions := func() *deleteOptions {
		return &deleteOptions{
			Options: &api.Options{
				CommonOptions: common.CommonOptions{
					Out: out,
					Err: errOut,
				},
			},
		}
	}

	cmd.AddCommand(newDeleteCmdForEntityType("api-session", newOptions()))
	cmd.AddCommand(newDeleteCmdForEntityType("authenticator", newOptions()))
	cmd.AddCommand(newDeleteCmdForEntityType("enrollment", newOptions()))
	cmd.AddCommand(newDeleteCmdForEntityType("ca", newOptions()))
	cmd.AddCommand(newDeleteCmdForEntityType("config", newOptions()))
	cmd.AddCommand(newDeleteCmdForEntityType("config-type", newOptions()))
	cmd.AddCommand(newDeleteCmdForEntityType("edge-router", newOptions(), "er", "ers"))
	cmd.AddCommand(newDeleteCmdForEntityType("edge-router-policy", newOptions(), "erp", "erps"))
	cmd.AddCommand(newDeleteCmdForEntityType("identity", newOptions()))
	cmd.AddCommand(newDeleteCmdForEntityType("posture-check", newOptions()))
	cmd.AddCommand(newDeleteCmdForEntityType("service", newOptions()))
	cmd.AddCommand(newDeleteCmdForEntityType("service-edge-router-policy", newOptions(), "serp", "serps"))
	cmd.AddCommand(newDeleteCmdForEntityType("service-policy", newOptions(), "sp", "sps"))
	cmd.AddCommand(newDeleteCmdForEntityType("session", newOptions()))
	cmd.AddCommand(newDeleteCmdForEntityType("terminator", newOptions()))
	cmd.AddCommand(newDeleteCmdForEntityType("transit-router", newOptions()))
	cmd.AddCommand(newDeleteCmdForEntityType("auth-policy", newOptions()))
	cmd.AddCommand(newDeleteCmdForEntityType("external-jwt-signer", newOptions(), "ext-jwt-signer", "ext-jwt-signers", "external-jwt-signers"))

	return cmd
}

// newDeleteCmdForEntityType creates the delete command for the given entity type
func newDeleteCmdForEntityType(entityType string, options *deleteOptions, aliases ...string) *cobra.Command {
	logtrace.LogWithFunctionName()
	cmd := &cobra.Command{
		Use:     entityType + " <id>",
		Short:   "deletes " + getPlural(entityType) + " managed by the Ziti Edge Controller",
		Args:    cobra.MinimumNArgs(1),
		Aliases: append(aliases, getPlural(entityType)),
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := runDeleteEntityOfType(options, getPlural(entityType))
			cmdhelper.CheckErr(err)
		},
		SuggestFor: []string{},
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	options.AddCommonFlags(cmd)
	cmd.Flags().BoolVar(&options.ignoreMissing, "ignore-missing", false, "don't error if the entity can't be found to be deleted")

	cmd.AddCommand(newDeleteWhereCmdForEntityType(entityType, options))

	return cmd
}

func newDeleteWhereCmdForEntityType(entityType string, options *deleteOptions) *cobra.Command {
	logtrace.LogWithFunctionName()
	cmd := &cobra.Command{
		Use:   "where <filter>",
		Short: "deletes " + getPlural(entityType) + " matching the filter managed by the Ziti Edge Controller",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := runDeleteEntityOfTypeWhere(options, getPlural(entityType))
			cmdhelper.CheckErr(err)
		},
		SuggestFor: []string{},
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	options.AddCommonFlags(cmd)

	return cmd
}

// runDeleteEntityOfType implements the commands to delete various entity types
func runDeleteEntityOfType(o *deleteOptions, entityType string) error {
	logtrace.LogWithFunctionName()
	var err error
	ids := o.Args
	if entityType != "terminators" && entityType != "api-sessions" && entityType != "sessions" && entityType != "authenticators" && entityType != "enrollments" {
		if ids, err = mapNamesToIDs(entityType, *o.Options, true, ids...); err != nil {
			return err
		}
	}
	return deleteEntitiesOfType(o, entityType, ids)
}

func deleteEntitiesOfType(o *deleteOptions, entityType string, ids []string) error {
	logtrace.LogWithFunctionName()
	for _, id := range ids {
		err, statusCode := util.ControllerDelete("edge", entityType, id, "", o.Out, o.OutputJSONRequest, o.OutputJSONResponse, o.Timeout, o.Verbose)
		if err != nil {
			if statusCode != nil && o.ignoreMissing {
				o.Printf("delete of %v with id %v: %v\n", boltz.GetSingularEntityType(entityType), id, color.New(color.FgYellow, color.Bold).Sprint("NOT FOUND"))
				return nil
			}
			o.Printf("delete of %v with id %v: %v\n", boltz.GetSingularEntityType(entityType), id, color.New(color.FgRed, color.Bold).Sprint("FAIL"))
			return err
		}
		o.Printf("delete of %v with id %v: %v\n", boltz.GetSingularEntityType(entityType), id, color.New(color.FgGreen, color.Bold).Sprint("OK"))
	}
	return nil
}

// runDeleteEntityOfType implements the commands to delete various entity types
func runDeleteEntityOfTypeWhere(options *deleteOptions, entityType string) error {
	logtrace.LogWithFunctionName()
	filter := strings.Join(options.Args, " ")

	params := url.Values{}
	params.Add("filter", filter)

	children, pageInfo, err := ListEntitiesOfType(entityType, params, options.OutputJSONResponse, options.Out, options.Timeout, options.Verbose)
	if err != nil {
		return err
	}

	options.Printf("filter returned ")
	pageInfo.Output(options.Options)

	var ids []string
	for _, entity := range children {
		id, _ := entity.Path("id").Data().(string)
		ids = append(ids, id)
	}

	return deleteEntitiesOfType(options, entityType, ids)
}

func getPlural(entityType string) string {
	logtrace.LogWithFunctionName()
	if strings.HasSuffix(entityType, "y") {
		return strings.TrimSuffix(entityType, "y") + "ies"
	}
	return entityType + "s"
}
