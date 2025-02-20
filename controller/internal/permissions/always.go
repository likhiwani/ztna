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

package permissions

import "ztna-core/ztna/logtrace"

type AlwaysAllow struct{}

var always = &AlwaysAllow{}

func (a *AlwaysAllow) IsAllowed(_ ...string) bool {
	logtrace.LogWithFunctionName()
	return true
}

func Always() *AlwaysAllow {
	logtrace.LogWithFunctionName()
	return always
}
