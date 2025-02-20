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

package util

import (
	"fmt"
	"os"
	"strings"

	"ztna-core/ztna/common/getziti"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/ztna/constants"

	"github.com/fatih/color"

	"ztna-core/ztna/common/version"

	"github.com/blang/semver"
	"github.com/michaelquigley/pfxlog"
)

func LogReleaseVersionCheck(zitiComponent string) {
	logtrace.LogWithFunctionName()
	logger := pfxlog.Logger()
	if strings.ToLower(os.Getenv("ZITI_CHECK_VERSION")) == "true" {
		logger.Debug("ZITI_CHECK_VERSION is true. starting version check")
		developmentSemver, _ := semver.Parse("0.0.0")
		latestGithubRelease, err := getziti.GetHighestVersionGitHubReleaseInfo(constants.OpenZitiOrg, constants.ZITI, false)
		if err != nil {
			logger.Debugf("failed to find latest GitHub version with error: %s", err)
			return // soft-fail version check if GitHub API is unavailable
		}
		// compose current build's semver as version string and semver object
		currentBuildVersion := version.GetVersion()
		currentBuildSemver, err := semver.ParseTolerant(currentBuildVersion) // ParseTolerant trims leading "v"
		if err != nil {
			logger.Warnf("failed to parse current build version as semver: '%s' with error: %s", version.GetVersion(), err)
			return
		}
		// ignore non-release builds and current release build
		if currentBuildSemver.EQ(developmentSemver) {
			logger.Debugf(
				"this build of %s is unreleased v%s",
				zitiComponent,
				developmentSemver,
			)
		} else if latestGithubRelease.SemVer.GT(currentBuildSemver) {
			yellow := color.New(color.FgYellow).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Fprintf(os.Stderr,
				`
*********************************************************************************

An update with %s is available for %s %s from 
https://github.com/openziti/%s/releases/latest/

*********************************************************************************
`,
				green("v"+latestGithubRelease.SemVer.String()),
				zitiComponent,
				yellow("v"+currentBuildSemver.String()),
				constants.ZITI,
			)
			logger.Debugf(
				"this v%s build of %s is superseded by v%s",
				currentBuildSemver,
				zitiComponent,
				latestGithubRelease,
			)
		} else if latestGithubRelease.SemVer.EQ(currentBuildSemver) {
			logger.Debugf(
				"this build of %s is the latest release v%s",
				zitiComponent,
				currentBuildSemver,
			)
		}
	} else {
		logger.Debug("ZITI_CHECK_VERSION is not 'true'. skipping version check")
	}
}
