//go:build apitests
// +build apitests

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

package tests

import (
	"os"
	"testing"
	"ztna-core/ztna/logtrace"

	"github.com/spf13/cobra"
)

// for usages with main
func TestMain(m *testing.M) {
	logtrace.LogWithFunctionName()

	root := &cobra.Command{
		Use:   "go test",
		Short: "Test for the Ziti Edge API",
		Run: func(cmd *cobra.Command, args []string) {
			os.Exit(m.Run())
		},
	}
	testContext := GetTestContext()
	root.Flags().StringVarP(&testContext.AdminAuthenticator.Username, "username", "u", "admin", "The default admin username")
	root.Flags().StringVarP(&testContext.AdminAuthenticator.Password, "password", "p", "admin", "The default admin password")
	root.Flags().StringVarP(&testContext.ApiHost, "api", "a", "127.0.0.1:1281", "The Edge API host:port to connect to")

	if err := root.Execute(); err != nil {
		panic(err)
	}
}
