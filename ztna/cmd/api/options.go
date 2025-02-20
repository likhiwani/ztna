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

package api

import (
	logtrace "ztna-core/ztna/logtrace"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Jeffail/gabs"
	"ztna-core/ztna/ztna/cmd/common"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Options are common options for edge controller commands
type Options struct {
	common.CommonOptions
	OutputJSONRequest  bool
	OutputJSONResponse bool
	OutputCSV          bool
}

func (options *Options) OutputResponseJson() bool {
	logtrace.LogWithFunctionName()
	return options.OutputJSONResponse
}

func (options *Options) OutputRequestJson() bool {
	logtrace.LogWithFunctionName()
	return options.OutputJSONRequest
}

func (options *Options) OutputWriter() io.Writer {
	logtrace.LogWithFunctionName()
	return options.CommonOptions.Out
}

func (options *Options) ErrOutputWriter() io.Writer {
	logtrace.LogWithFunctionName()
	return options.CommonOptions.Err
}

func (options *Options) AddCommonFlags(cmd *cobra.Command) {
	logtrace.LogWithFunctionName()
	cmd.Flags().StringVarP(&common.CliIdentity, "cli-identity", "i", "", "Specify the saved identity you want the CLI to use when connect to the controller with")
	cmd.Flags().BoolVarP(&options.OutputJSONResponse, "output-json", "j", false, "Output the full JSON response from the Ziti Edge Controller")
	cmd.Flags().BoolVar(&options.OutputJSONRequest, "output-request-json", false, "Output the full JSON request to the Ziti Edge Controller")
	cmd.Flags().IntVarP(&options.Timeout, "timeout", "", 5, "Timeout for REST operations (specified in seconds)")
	cmd.Flags().BoolVarP(&options.Verbose, "verbose", "", false, "Enable verbose logging")
}

func (options *Options) LogCreateResult(entityType string, result *gabs.Container, err error) error {
	logtrace.LogWithFunctionName()
	if err != nil {
		return err
	}

	if !options.OutputJSONResponse {
		id := result.S("data", "id").Data()
		_, err = fmt.Fprintf(options.Out, "New %v %v created with id: %v\n", entityType, options.Args[0], id)
		return err
	}
	return nil
}

type EntityOptions struct {
	Options
	Tags     map[string]string
	TagsJson string
}

func (self *EntityOptions) AddCommonFlags(cmd *cobra.Command) {
	logtrace.LogWithFunctionName()
	self.Options.AddCommonFlags(cmd)
	if cmd.Flags().ShorthandLookup("t") == nil {
		cmd.Flags().StringToStringVarP(&self.Tags, "tags", "t", nil, "Add tags to entity definition")
	} else {
		cmd.Flags().StringToStringVar(&self.Tags, "tags", nil, "Add tags to entity definition")
	}
	cmd.Flags().StringVar(&self.TagsJson, "tags-json", "", "Add tags defined in JSON to entity definition")
}

func (self *EntityOptions) GetTags() map[string]interface{} {
	logtrace.LogWithFunctionName()
	result := map[string]interface{}{}
	if len(self.TagsJson) > 0 {
		if err := json.Unmarshal([]byte(self.TagsJson), &result); err != nil {
			logrus.Fatalf("invalid tags JSON: '%s'", self.TagsJson)
		}
	}
	for k, v := range self.Tags {
		result[k] = v
	}
	return result
}

func (self *EntityOptions) TagsProvided() bool {
	logtrace.LogWithFunctionName()
	return self.Cmd.Flags().Changed("tags") || self.Cmd.Flags().Changed("tags-json")
}

func (self *EntityOptions) SetTags(container *gabs.Container) {
	logtrace.LogWithFunctionName()
	tags := self.GetTags()
	SetJSONValue(container, tags, "tags")
}

func NewEntityOptions(out, errOut io.Writer) EntityOptions {
	logtrace.LogWithFunctionName()
	return EntityOptions{
		Options: Options{
			CommonOptions: common.CommonOptions{
				Out: out,
				Err: errOut,
			},
		},
	}
}
