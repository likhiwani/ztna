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

package client

import (
	"fmt"
	"io"
	"os"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/xgress_transport"
	"ztna-core/ztna/zititest/ziti-fabric-test/subcmd"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/info"
	"github.com/openziti/identity"
	"github.com/openziti/identity/dotziti"
	"github.com/openziti/transport/v2"
	"github.com/spf13/cobra"
)

func init() {
	logtrace.LogWithFunctionName()
	ncCmd.Flags().StringVarP(&ncCmdIdentity, "identityName", "i", "default", "dotzeet identity name")
	ncCmd.Flags().StringVarP(&ncCmdIngress, "ingressEndpoint", "e", "tls:127.0.0.1:7002", "ingress endpoint address")
	subcmd.Root.AddCommand(ncCmd)
}

var ncCmd = &cobra.Command{
	Use:   "nc <service>",
	Short: "Simple NetCat",
	Args:  cobra.ExactArgs(1),
	Run:   doNC,
}
var ncCmdIdentity string
var ncCmdIngress string

func doNC(cmd *cobra.Command, args []string) {
	logtrace.LogWithFunctionName()
	_, id, err := dotziti.LoadIdentity(ncCmdIdentity)
	if err != nil {
		panic(err)
	}

	ingressAddr, err := transport.ParseAddress(ncCmdIngress)

	if err != nil {
		panic(err)
	}

	serviceId := &identity.TokenId{Token: args[0]}
	fmt.Fprintf(os.Stderr, "Dialing fabric ingress %v\n", ncCmdIngress)
	conn, err := xgress_transport.ClientDial(ingressAddr, id, serviceId, nil)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "Successfully authenticated to ingress %v. Beginning nc.\n", ncCmdIngress)

	pfxlog.Logger().Debug("connected")
	go Copy(conn, os.Stdin)
	Copy(os.Stdout, conn)
}

func Copy(writer io.Writer, reader io.Reader) {
	logtrace.LogWithFunctionName()
	buf := make([]byte, info.MaxUdpPacketSize)
	bytesCopied, err := io.CopyBuffer(writer, reader, buf)
	pfxlog.Logger().Infof("Copied %v bytes", bytesCopied)
	if err != nil {
		pfxlog.Logger().Errorf("error while copying bytes (%v)", err)
	}
}

// CopyAndLog does what io.Copy does but with additional logging
func CopyAndLog(context string, writer io.Writer, reader io.Reader) {
	logtrace.LogWithFunctionName()
	buf := make([]byte, info.MaxUdpPacketSize)

	var bytesRead, totalBytesRead, bytesWritten, totalBytesWritten int
	var readErr, writeErr error

	for {
		bytesRead, readErr = reader.Read(buf)
		totalBytesRead += bytesRead
		if bytesRead > 0 {
			bytesWritten, writeErr = writer.Write(buf[:bytesRead])
			totalBytesWritten += bytesWritten
			if writeErr != nil {
				pfxlog.Logger().WithError(writeErr).Error("Write failure on copy")
			}
		}

		if readErr != nil && readErr != io.EOF {
			pfxlog.Logger().WithError(readErr).Error("Read failure on copy")
		}

		if readErr != nil || writeErr != nil {
			return
		}

		_, _ = fmt.Fprintf(os.Stderr, "%v: Read %v (%v total), Wrote %v (%v total)\n",
			context, bytesRead, totalBytesRead, bytesWritten, totalBytesWritten)
	}
}
