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
	"errors"
	"fmt"
	"math"

	"github.com/Jeffail/gabs"
	"ztna-core/ztna/ztna/cmd/api"
	"ztna-core/ztna/ztna/cmd/common"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"
	"ztna-core/ztna/ztna/util"
	"github.com/openziti/foundation/v2/stringz"
	errors2 "github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type updateTerminatorOptions struct {
	api.Options
	router     string
	address    string
	binding    string
	cost       int32
	precedence string
	tags       map[string]string
}

func newUpdateTerminatorCmd(p common.OptionsProvider) *cobra.Command {
	options := &updateTerminatorOptions{
		Options: api.Options{CommonOptions: p()},
	}

	cmd := &cobra.Command{
		Use:   "terminator <id>",
		Short: "updates a service terminator",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := runUpdateTerminator(options)
			cmdhelper.CheckErr(err)
		},
		SuggestFor: []string{},
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	cmd.Flags().StringVar(&options.router, "router", "", "Set the terminator router")
	cmd.Flags().StringVar(&options.address, "address", "", "Set the terminator address")
	cmd.Flags().StringVar(&options.binding, "binding", "", "Set the terminator binding")
	cmd.Flags().Int32VarP(&options.cost, "cost", "c", 0, "Set the terminator cost")
	cmd.Flags().StringVarP(&options.precedence, "precedence", "p", "", "Set the terminator precedence ('default', 'required' or 'failed')")
	cmd.Flags().StringToStringVar(&options.tags, "tags", nil, "Custom management tags")
	options.AddCommonFlags(cmd)

	return cmd
}

// runUpdateTerminator implements the command to update a Terminator
func runUpdateTerminator(o *updateTerminatorOptions) (err error) {
	entityData := gabs.New()

	change := false
	if o.Cmd.Flags().Changed("router") {
		router, err := api.MapNameToID(util.FabricAPI, "routers", &o.Options, o.router)
		if err != nil {
			return err
		}

		api.SetJSONValue(entityData, router, "router")
		change = true
	}

	if o.Cmd.Flags().Changed("binding") {
		api.SetJSONValue(entityData, o.binding, "binding")
		change = true
	}

	if o.Cmd.Flags().Changed("address") {
		api.SetJSONValue(entityData, o.address, "address")
		change = true
	}

	if o.Cmd.Flags().Changed("cost") {
		if o.cost > math.MaxUint16 {
			return errors2.Errorf("Invalid cost %v. Must be positive number less than or equal to %v", o.cost, math.MaxUint16)
		}
		api.SetJSONValue(entityData, o.cost, "cost")
		change = true
	}

	if o.Cmd.Flags().Changed("precedence") {
		validValues := []string{"default", "required", "failed"}
		if !stringz.Contains(validValues, o.precedence) {
			return errors2.Errorf("Invalid precedence %v. Must be one of %+v", o.precedence, validValues)
		}
		api.SetJSONValue(entityData, o.precedence, "precedence")
		change = true
	}

	if o.Cmd.Flags().Changed("tags") {
		api.SetJSONValue(entityData, o.tags, "tags")
		change = true
	}

	if !change {
		return errors.New("no change specified. must specify at least one attribute to change")
	}

	_, err = patchEntityOfType(fmt.Sprintf("terminators/%v", o.Args[0]), entityData.String(), &o.Options)
	return err
}
