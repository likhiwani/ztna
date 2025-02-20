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

package log

import (
	"fmt"
	logtrace "ztna-core/ztna/logtrace"

	"github.com/fatih/color"
)

func Infof(msg string, args ...interface{}) {
	logtrace.LogWithFunctionName()
	Info(fmt.Sprintf(msg, args...))
}

func Info(msg string) {
	logtrace.LogWithFunctionName()
	fmt.Print(msg)
}

func Infoln(msg string) {
	logtrace.LogWithFunctionName()
	fmt.Println(msg)
}

func Warnf(msg string, args ...interface{}) {
	logtrace.LogWithFunctionName()
	Warn(fmt.Sprintf(msg, args...))
}

func Warn(msg string) {
	logtrace.LogWithFunctionName()
	color.Yellow(msg)
}

func Fatalf(msg string, args ...interface{}) {
	logtrace.LogWithFunctionName()
	Fatal(fmt.Sprintf(msg, args...))
}

func Fatal(msg string) {
	logtrace.LogWithFunctionName()
	color.Red(msg)
}
