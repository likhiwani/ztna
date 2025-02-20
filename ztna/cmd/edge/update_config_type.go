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
	"encoding/json"
	"fmt"
	"io"
	"os"

	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/ztna/cmd/api"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"

	"github.com/pkg/errors"

	"github.com/Jeffail/gabs"
	"github.com/spf13/cobra"
)

type updateConfigTypeAction struct {
	api.EntityOptions
	name     string
	data     string
	jsonFile string
}

func newUpdateConfigTypeCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	logtrace.LogWithFunctionName()
	action := &updateConfigTypeAction{
		EntityOptions: api.NewEntityOptions(out, errOut),
	}

	cmd := &cobra.Command{
		Use:   "config-type <idOrName>",
		Short: "updates a config type managed by the Ziti Edge Controller",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			action.Cmd = cmd
			action.Args = args
			err := action.run()
			cmdhelper.CheckErr(err)
		},
		SuggestFor: []string{},
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	cmd.Flags().StringVarP(&action.name, "name", "n", "", "Set the name of the config")
	cmd.Flags().StringVarP(&action.data, "data", "d", "", "Set the data of the config")
	cmd.Flags().StringVarP(&action.jsonFile, "json-file", "f", "", "Read config JSON from a file instead of the command line")

	action.AddCommonFlags(cmd)

	return cmd
}

// runUpdateConfigType update a new config on the Ziti Edge Controller
func (self *updateConfigTypeAction) run() error {
	logtrace.LogWithFunctionName()
	id, err := mapNameToID("config-types", self.Args[0], self.Options)
	if err != nil {
		return err
	}
	entityData := gabs.New()
	change := false

	if self.Cmd.Flags().Changed("name") {
		api.SetJSONValue(entityData, self.name, "name")
		change = true
	}

	var jsonBytes []byte

	if self.Cmd.Flags().Changed("data") {
		jsonBytes = []byte(self.data)
	}

	if self.Cmd.Flags().Changed("json-file") {
		if self.Cmd.Flags().Changed("data") {
			return errors.New("only one of --data and --json-file is allowed")
		}
		var err error
		if jsonBytes, err = os.ReadFile(self.jsonFile); err != nil {
			return fmt.Errorf("failed to read config json file %v: %w", self.jsonFile, err)
		}
	}

	if self.TagsProvided() {
		self.SetTags(entityData)
		change = true
	}

	if len(jsonBytes) > 0 {
		dataMap := map[string]interface{}{}
		if err := json.Unmarshal(jsonBytes, &dataMap); err != nil {
			fmt.Printf("Attempted to parse: %v\n", string(jsonBytes))
			fmt.Printf("Failing parsing JSON: %+v\n", err)
			return errors.Errorf("unable to parse data as json: %v", err)
		}
		api.SetJSONValue(entityData, dataMap, "schema")
		change = true
	}

	if !change {
		return errors.New("no change specified. must specify at least one attribute to change")
	}

	_, err = patchEntityOfType(fmt.Sprintf("config-types/%v", id), entityData.String(), &self.Options)

	return err
}
