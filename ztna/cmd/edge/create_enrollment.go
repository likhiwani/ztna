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
	"context"
	"fmt"
	"io"
	"time"

	"ztna-core/edge-api/rest_management_api_client/enrollment"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/ztna/cmd/api"
	"ztna-core/ztna/ztna/cmd/common"
	cmdhelper "ztna-core/ztna/ztna/cmd/helpers"
	"ztna-core/ztna/ztna/util"

	"github.com/go-openapi/strfmt"
	"github.com/spf13/cobra"
)

func newCreateEnrollmentCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	logtrace.LogWithFunctionName()
	cmd := &cobra.Command{
		Use: "enrollment",
	}

	cmd.AddCommand(newCreateEnrollmentOtt(out, errOut))
	cmd.AddCommand(newCreateEnrollmentOttCa(out, errOut))
	cmd.AddCommand(newCreateEnrollmentUpdb(out, errOut))

	return cmd
}

type createEnrollmentOptions struct {
	api.Options
	jwtOutputFile string
	duration      int64
}

func newCreateEnrollmentOtt(out io.Writer, errOut io.Writer) *cobra.Command {
	logtrace.LogWithFunctionName()
	options := &createEnrollmentOptions{
		Options: api.Options{
			CommonOptions: common.CommonOptions{
				Out: out,
				Err: errOut,
			},
		},
	}

	cmd := &cobra.Command{
		Use:   "ott <identityIdOrName> [-duration <minutes>]",
		Short: "creates a one-time-token (ott) enrollment for the given identity",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args

			err := runCreateEnrollmentOtt(options)
			cmdhelper.CheckErr(err)
		},
		SuggestFor: []string{},
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	options.AddCommonFlags(cmd)
	cmd.Flags().Int64VarP(&options.duration, "duration", "d", 30, "the duration of time the enrollment should valid for")
	cmd.Flags().StringVarP(&options.jwtOutputFile, "jwt-output-file", "o", "", "File to which to output the enrollment JWT ")

	return cmd
}

func newCreateEnrollmentOttCa(out io.Writer, errOut io.Writer) *cobra.Command {
	logtrace.LogWithFunctionName()
	options := &createEnrollmentOptions{
		Options: api.Options{
			CommonOptions: common.CommonOptions{
				Out: out,
				Err: errOut,
			},
		},
	}

	cmd := &cobra.Command{
		Use:   "ottca <identityIdOrName> <caIdOrName> [-duration <minutes>]",
		Short: "creates a one-time-token ca (ottca) enrollment for the given identity and ca",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args

			err := runCreateEnrollmentOttCa(options)
			cmdhelper.CheckErr(err)
		},
		SuggestFor: []string{},
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	options.AddCommonFlags(cmd)
	cmd.Flags().Int64VarP(&options.duration, "duration", "d", 30, "the duration of time the enrollment should valid for")

	return cmd
}

func runCreateEnrollmentOtt(options *createEnrollmentOptions) error {
	logtrace.LogWithFunctionName()
	managementClient, err := util.NewEdgeManagementClient(options)

	if err != nil {
		return err
	}

	identityId, err := mapNameToID("identities", options.Args[0], options.Options)

	if err != nil {
		return err
	}

	method := rest_model.EnrollmentCreateMethodOtt
	expiresAt := strfmt.DateTime(time.Now().Add(time.Duration(options.duration) * time.Minute))

	params := &enrollment.CreateEnrollmentParams{
		Enrollment: &rest_model.EnrollmentCreate{
			ExpiresAt:  &expiresAt,
			IdentityID: &identityId,
			Method:     &method,
		},
		Context: context.Background(),
	}

	resp, err := managementClient.Enrollment.CreateEnrollment(params, nil)

	if err != nil {
		return util.WrapIfApiError(err)
	}

	enrollmentID := resp.GetPayload().Data.ID

	if _, err = fmt.Fprintf(options.Out, "%v\n", enrollmentID); err != nil {
		panic(err)
	}

	if options.jwtOutputFile != "" {
		if err = getIdentityJwt(&options.Options, identityId, options.jwtOutputFile, "ott", options.Options.Timeout, options.Options.Verbose); err != nil {
			return err
		}
	}
	return err
}

func runCreateEnrollmentOttCa(options *createEnrollmentOptions) error {
	logtrace.LogWithFunctionName()
	managementClient, err := util.NewEdgeManagementClient(options)

	if err != nil {
		return err
	}

	identityId, err := mapNameToID("identities", options.Args[0], options.Options)

	if err != nil {
		return err
	}

	caId, err := mapNameToID("cas", options.Args[1], options.Options)

	if err != nil {
		return err
	}

	method := rest_model.EnrollmentCreateMethodOttca
	expiresAt := strfmt.DateTime(time.Now().Add(time.Duration(options.duration) * time.Minute))

	params := &enrollment.CreateEnrollmentParams{
		Enrollment: &rest_model.EnrollmentCreate{
			ExpiresAt:  &expiresAt,
			IdentityID: &identityId,
			Method:     &method,
			CaID:       &caId,
		},
		Context: context.Background(),
	}

	resp, err := managementClient.Enrollment.CreateEnrollment(params, nil)

	if err != nil {
		return util.WrapIfApiError(err)
	}

	enrollmentID := resp.GetPayload().Data.ID

	if _, err = fmt.Fprintf(options.Out, "%v\n", enrollmentID); err != nil {
		panic(err)
	}

	return err
}

func newCreateEnrollmentUpdb(out io.Writer, errOut io.Writer) *cobra.Command {
	logtrace.LogWithFunctionName()
	options := &createEnrollmentOptions{
		Options: api.Options{
			CommonOptions: common.CommonOptions{
				Out: out,
				Err: errOut,
			},
		},
	}

	cmd := &cobra.Command{
		Use:   "updb <identityIdOrName> <username> [-duration <minutes>]",
		Short: "creates a username password (updb) enrollment for the given identity",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args

			err := runCreateEnrollmentUpdb(options)
			cmdhelper.CheckErr(err)
		},
		SuggestFor: []string{},
	}

	// allow interspersing positional args and flags
	cmd.Flags().SetInterspersed(true)
	options.AddCommonFlags(cmd)
	cmd.Flags().Int64VarP(&options.duration, "duration", "d", 30, "the duration of time the enrollment should valid for")

	return cmd
}

func runCreateEnrollmentUpdb(options *createEnrollmentOptions) error {
	logtrace.LogWithFunctionName()
	managementClient, err := util.NewEdgeManagementClient(options)

	if err != nil {
		return err
	}

	identityId, err := mapNameToID("identities", options.Args[0], options.Options)

	if err != nil {
		return err
	}

	username := options.Args[1]

	method := rest_model.EnrollmentCreateMethodUpdb
	expiresAt := strfmt.DateTime(time.Now().Add(time.Duration(options.duration) * time.Minute))

	params := &enrollment.CreateEnrollmentParams{
		Enrollment: &rest_model.EnrollmentCreate{
			ExpiresAt:  &expiresAt,
			IdentityID: &identityId,
			Method:     &method,
			Username:   &username,
		},
		Context: context.Background(),
	}

	resp, err := managementClient.Enrollment.CreateEnrollment(params, nil)

	if err != nil {
		return util.WrapIfApiError(err)
	}

	enrollmentID := resp.GetPayload().Data.ID

	if _, err = fmt.Fprintf(options.Out, "%v\n", enrollmentID); err != nil {
		panic(err)
	}

	return err
}
