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

package version

import (
	"runtime"
	logtrace "ztna-core/ztna/logtrace"

	"github.com/openziti/foundation/v2/versions"
)

type cmdBuildInfo struct{}

func (c cmdBuildInfo) EncoderDecoder() versions.VersionEncDec {
	logtrace.LogWithFunctionName()
	return &versions.StdVersionEncDec
}

func (c cmdBuildInfo) Version() string {
	logtrace.LogWithFunctionName()
	return Version
}

func (c cmdBuildInfo) Revision() string {
	logtrace.LogWithFunctionName()
	return Revision
}

func (c cmdBuildInfo) BuildDate() string {
	logtrace.LogWithFunctionName()
	return BuildDate
}

func (c cmdBuildInfo) AsVersionInfo() *versions.VersionInfo {
	logtrace.LogWithFunctionName()
	return &versions.VersionInfo{
		Version:   c.Version(),
		Revision:  c.Revision(),
		BuildDate: c.BuildDate(),
		OS:        GetOS(),
		Arch:      GetArchitecture(),
	}
}

func GetCmdBuildInfo() versions.VersionProvider {
	logtrace.LogWithFunctionName()
	return cmdBuildInfo{}
}

func GetVersion() string {
	logtrace.LogWithFunctionName()
	return Version
}

func GetRevision() string {
	logtrace.LogWithFunctionName()
	return Revision
}

func GetBuildDate() string {
	logtrace.LogWithFunctionName()
	return BuildDate
}

func GetGoVersion() string {
	logtrace.LogWithFunctionName()
	return runtime.Version()
}

func GetOS() string {
	logtrace.LogWithFunctionName()
	return runtime.GOOS
}

func GetArchitecture() string {
	logtrace.LogWithFunctionName()
	return runtime.GOARCH
}
