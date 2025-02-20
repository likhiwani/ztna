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

package helpers

import (
	logtrace "ztna-core/ztna/logtrace"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	edge "ztna-core/ztna/controller/config"
	"ztna-core/ztna/router/xgress_edge_tunnel"
	"ztna-core/ztna/ztna/constants"
	"github.com/pkg/errors"
)

func HomeDir() string {
	logtrace.LogWithFunctionName()
	if h := os.Getenv("HOME"); h != "" {
		return NormalizePath(h)
	}
	h := os.Getenv("USERPROFILE") // windows
	if h == "" {
		h = "."
	}
	return NormalizePath(h)
}

func WorkingDir() (string, error) {
	logtrace.LogWithFunctionName()
	wd, err := os.Getwd()
	if wd == "" || err != nil {
		return "", err
	}

	return NormalizePath(wd), nil
}

func GetZitiHome() string {
	logtrace.LogWithFunctionName()
	// Get path from env variable
	retVal := os.Getenv(constants.ZitiHomeVarName)

	if retVal == "" {
		// If not set, create a default path of the current working directory
		workingDir, err := WorkingDir()
		if err != nil {
			// If there is an error just use .
			workingDir = "."
		}

		_ = os.Setenv(constants.ZitiHomeVarName, workingDir)
		retVal = os.Getenv(constants.ZitiHomeVarName)
	}

	return NormalizePath(retVal)
}

func HostnameOrNetworkName() string {
	logtrace.LogWithFunctionName()
	val := os.Getenv("ZITI_NETWORK_NAME")
	if val == "" {
		h, err := os.Hostname()
		if err != nil {
			return "localhost"
		}
		return h
	}
	return val
}

func defaultValue(val string) func() string {
	logtrace.LogWithFunctionName()
	return func() string {
		return val
	}
}
func defaultIntValue(val int64) func() string {
	logtrace.LogWithFunctionName()
	return func() string {
		return strconv.FormatInt(val, 10)
	}
}

func GetCtrlBindAddress() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.CtrlBindAddressVarName, defaultValue(constants.DefaultCtrlBindAddress))
}

func GetCtrlAdvertisedAddress() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.CtrlAdvertisedAddressVarName, HostnameOrNetworkName)
}

func GetEdgeRouterIpOvderride() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.ZitiEdgeRouterIPOverrideVarName, defaultValue(""))
}

func GetCtrlAdvertisedPort() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.CtrlAdvertisedPortVarName, defaultValue(constants.DefaultCtrlAdvertisedPort))
}

func GetCtrlDatabaseFile() string {
	logtrace.LogWithFunctionName()
	path := getFromEnv(constants.CtrlDatabaseFileVarName, defaultValue(constants.DefaultCtrlDatabaseFile))
	if !filepath.IsAbs(path) {
		path = filepath.Join(GetZitiHome(), path)
	}
	return NormalizePath(path)
}

func GetCtrlEdgeBindAddress() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.CtrlEdgeBindAddressVarName, defaultValue(constants.DefaultCtrlEdgeBindAddress))
}

func GetCtrlEdgeAdvertisedAddress() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.CtrlEdgeAdvertisedAddressVarName, HostnameOrNetworkName)
}

func GetCtrlEdgeAltAdvertisedAddress() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.CtrlEdgeAltAdvertisedAddressVarName, GetCtrlEdgeAdvertisedAddress)
}

func GetCtrlEdgeAdvertisedPort() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.CtrlEdgeAdvertisedPortVarName, defaultValue(constants.DefaultCtrlEdgeAdvertisedPort))
}

func GetZitiEdgeRouterPort() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.ZitiEdgeRouterPortVarName, defaultValue(constants.DefaultZitiEdgeRouterPort))
}

func GetZitiEdgeRouterListenerBindPort() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.ZitiEdgeRouterListenerBindPortVarName, defaultValue(constants.DefaultZitiEdgeRouterListenerBindPort))
}

func GetCtrlEdgeIdentityEnrollmentDuration() time.Duration {
	logtrace.LogWithFunctionName()
	retVal := getFromEnv(constants.CtrlEdgeIdentityEnrollmentDurationVarName, defaultIntValue(int64(edge.DefaultEdgeEnrollmentDuration.Minutes())))
	retValInt, err := strconv.Atoi(retVal)
	if err != nil {
		err := errors.Wrap(err, "Unable to get "+constants.CtrlEdgeIdentityEnrollmentDurationVarDescription)
		if err != nil {
			return edge.DefaultEdgeEnrollmentDuration
		}
	}

	return time.Duration(retValInt) * time.Minute
}

func GetCtrlEdgeRouterEnrollmentDuration() time.Duration {
	logtrace.LogWithFunctionName()
	retVal := getFromEnv(constants.CtrlEdgeRouterEnrollmentDurationVarName, defaultIntValue(int64(edge.DefaultEdgeEnrollmentDuration.Minutes())))
	retValInt, err := strconv.Atoi(retVal)
	if err != nil {
		err := errors.Wrap(err, "Unable to get "+constants.CtrlEdgeRouterEnrollmentDurationVarDescription)
		if err != nil {
			return edge.DefaultEdgeEnrollmentDuration
		}
	}

	return time.Duration(retValInt) * time.Minute
}

func GetZitiEdgeRouterC() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.ZitiEdgeRouterCsrCVarName, defaultValue(constants.DefaultEdgeRouterCsrC))
}

func GetZitiEdgeRouterST() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.ZitiEdgeRouterCsrSTVarName, defaultValue(constants.DefaultEdgeRouterCsrST))
}

func GetZitiEdgeRouterL() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.ZitiEdgeRouterCsrLVarName, defaultValue(constants.DefaultEdgeRouterCsrL))
}

func GetZitiEdgeRouterO() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.ZitiEdgeRouterCsrOVarName, defaultValue(constants.DefaultEdgeRouterCsrO))
}

func GetZitiEdgeRouterOU() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.ZitiEdgeRouterCsrOUVarName, defaultValue(constants.DefaultEdgeRouterCsrOU))
}

type envVarNotFound func() string

func getFromEnv(envVarName string, onNotFound envVarNotFound) string {
	logtrace.LogWithFunctionName()
	// Get path from env variable
	retVal := os.Getenv(envVarName)
	if retVal != "" {
		return retVal
	}
	return onNotFound()
}

// NormalizePath replaces windows \ with / which windows allows for
func NormalizePath(input string) string {
	logtrace.LogWithFunctionName()
	return strings.ReplaceAll(input, "\\", "/")
}

func GetRouterAdvertisedAddress() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.ZitiEdgeRouterAdvertisedAddressVarName, HostnameOrNetworkName)
}
func GetZitiEdgeRouterResolver() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.ZitiEdgeRouterResolverVarName, defaultValue(xgress_edge_tunnel.DefaultDnsResolver))
}
func GetZitiEdgeRouterDnsSvcIpRange() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.ZitiEdgeRouterDnsSvcIpRangeVarName, defaultValue(xgress_edge_tunnel.DefaultDnsServiceIpRange))
}
func GetRouterSans() string {
	logtrace.LogWithFunctionName()
	return getFromEnv(constants.ZitiRouterCsrSansDnsVarName, GetRouterAdvertisedAddress)
}
