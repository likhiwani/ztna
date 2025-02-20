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
	"fmt"
	"io"
	"os"

	"ztna-core/edge-api/rest_management_api_client/certificate_authority"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/ztna/cmd/api"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"
	"ztna-core/ztna/ztna/util"

	"github.com/spf13/cobra"
)

type createCaOptions struct {
	api.EntityOptions
	Ca                     rest_model.CaCreate
	IdentityRolesFromFlags []string
}

// newCreateCaCmd creates the 'edge controller create ca local' command for the given entity type
func newCreateCaCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	logtrace.LogWithFunctionName()
	options := &createCaOptions{
		EntityOptions: api.NewEntityOptions(out, errOut),
		Ca: rest_model.CaCreate{
			CertPem: Ptr(""),
			ExternalIDClaim: &rest_model.ExternalIDClaim{
				Index:           Ptr(int64(0)),
				Location:        Ptr(""),
				Matcher:         Ptr(""),
				MatcherCriteria: Ptr(""),
				Parser:          Ptr(""),
				ParserCriteria:  Ptr(""),
			},
			IdentityNameFormat:        "",
			IdentityRoles:             []string{},
			IsAuthEnabled:             Ptr(false),
			IsAutoCaEnrollmentEnabled: Ptr(false),
			IsOttCaEnrollmentEnabled:  Ptr(false),
			Name:                      Ptr(""),
			Tags: &rest_model.Tags{
				SubTags: map[string]interface{}{},
			},
		},
	}

	cmd := &cobra.Command{
		Use:   "ca <name> <pemCertFile> [--autoca, --ottca, --auth]",
		Short: "creates a ca managed by the Ziti Edge Controller",
		Long:  "creates a ca managed by the Ziti Edge Controller",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("requires at least %d arg(s), only received %d", 2, len(args))
			}

			options.Ca.Name = &args[0]

			caFile := args[1]

			pemBytes, err := os.ReadFile(caFile)

			options.Ca.CertPem = Ptr(string(pemBytes))

			if err != nil {
				return fmt.Errorf("could not read CA certificate file: %s", err)
			}

			return nil

		},
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := runCreateCa(options)
			cmdhelper.CheckErr(err)
		},
		SuggestFor: []string{},
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	cmd.Flags().BoolVarP(options.Ca.IsAuthEnabled, "auth", "e", false, "Whether the CA can be used for authentication or not")
	cmd.Flags().BoolVarP(options.Ca.IsOttCaEnrollmentEnabled, "ottca", "o", false, "Whether the CA can be used for one-time-token CA enrollment")
	cmd.Flags().BoolVarP(options.Ca.IsAutoCaEnrollmentEnabled, "autoca", "u", false, "Whether the CA can be used for auto CA enrollment")
	cmd.Flags().StringSliceVarP(&options.IdentityRolesFromFlags, "role-attributes", "a", []string{}, "A csv string of role attributes enrolling identities receive")
	cmd.Flags().StringVarP(&options.Ca.IdentityNameFormat, "identity-name-format", "f", "", "The naming format to use for identities enrolling via the CA")

	//ExternalIdClaim
	cmd.Flags().Int64VarP(options.Ca.ExternalIDClaim.Index, "index", "d", 0, "the index to use if multiple external ids are found, default 0")
	cmd.Flags().StringVarP(options.Ca.ExternalIDClaim.Location, "location", "l", "", "the location to search for external ids")
	cmd.Flags().StringVarP(options.Ca.ExternalIDClaim.Matcher, "matcher", "m", "", "the matcher to use at the given location")
	cmd.Flags().StringVarP(options.Ca.ExternalIDClaim.MatcherCriteria, "matcher-criteria", "x", "", "criteria used with the given matcher")
	cmd.Flags().StringVarP(options.Ca.ExternalIDClaim.Parser, "parser", "p", "", "the parser to use on found external ids")
	cmd.Flags().StringVarP(options.Ca.ExternalIDClaim.ParserCriteria, "parser-criteria", "z", "", "criteria used with the given parser")

	options.AddCommonFlags(cmd)

	return cmd
}

func runCreateCa(options *createCaOptions) (err error) {
	logtrace.LogWithFunctionName()
	managementClient, err := util.NewEdgeManagementClient(options)

	if err != nil {
		return err
	}

	params := certificate_authority.NewCreateCaParams()
	params.Ca = &options.Ca

	for _, attr := range options.IdentityRolesFromFlags {
		params.Ca.IdentityRoles = append(params.Ca.IdentityRoles, attr)
	}

	for k, v := range options.GetTags() {
		params.Ca.Tags.SubTags[k] = v
	}

	//clear external id claims if location is not set
	if params.Ca.ExternalIDClaim.Location == nil || *params.Ca.ExternalIDClaim.Location == "" {
		params.Ca.ExternalIDClaim = nil
	}

	resp, err := managementClient.CertificateAuthority.CreateCa(params, nil)

	if err != nil {
		return util.WrapIfApiError(err)
	}

	checkId := resp.GetPayload().Data.ID

	if _, err = fmt.Fprintf(options.Out, "%v\n", checkId); err != nil {
		panic(err)
	}

	return err
}
