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

package main

import (
	"os"
	"ztna-core/ztna/common/build"
	"ztna-core/ztna/common/version"
	logtrace "ztna-core/ztna/logtrace"
	"ztna-core/ztna/ztna/cmd"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/transport/v2"
	"github.com/openziti/transport/v2/dtls"
	"github.com/openziti/transport/v2/tcp"
	"github.com/openziti/transport/v2/tls"
	"github.com/openziti/transport/v2/transwarp"
	"github.com/openziti/transport/v2/transwarptls"
	"github.com/openziti/transport/v2/udp"
	"github.com/openziti/transport/v2/ws"
	"github.com/openziti/transport/v2/wss"
	"github.com/sirupsen/logrus"
)

func init() {
	file, err := os.OpenFile("/Users/lbaswani/make-it-happen/ztna/ztna.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatalf("Failed to open log file: %v", err)
	}
	logrus.SetOutput(file)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	logtrace.LogWithFunctionName()
	options := pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti/").NoColor()
	pfxlog.GlobalInit(logrus.InfoLevel, options)

	transport.AddAddressParser(tls.AddressParser{})
	transport.AddAddressParser(dtls.AddressParser{})
	transport.AddAddressParser(tcp.AddressParser{})
	transport.AddAddressParser(transwarp.AddressParser{})
	transport.AddAddressParser(transwarptls.AddressParser{})
	transport.AddAddressParser(ws.AddressParser{})
	transport.AddAddressParser(wss.AddressParser{})
	transport.AddAddressParser(udp.AddressParser{})

	build.InitBuildInfo(version.GetCmdBuildInfo())
}

func main() {
	logtrace.LogWithFunctionName()
	cmd.Execute()
}
