package build

import (
	logtrace "ztna-core/ztna/logtrace"
)

var defaultBuildInfo = defaultInfo{}
var info Info = defaultBuildInfo

func GetBuildInfo() Info {
	logtrace.LogWithFunctionName()
	return info
}

func InitBuildInfo(buildInfo Info) {
	logtrace.LogWithFunctionName()
	if info == defaultBuildInfo {
		info = buildInfo
	}
}

type Info interface {
	Version() string
	Revision() string
	BuildDate() string
}

type defaultInfo struct{}

func (d defaultInfo) Version() string {
	logtrace.LogWithFunctionName()
	return "unknown"
}

func (d defaultInfo) Revision() string {
	logtrace.LogWithFunctionName()
	return "unknown"
}

func (d defaultInfo) BuildDate() string {
	logtrace.LogWithFunctionName()
	return "unknown"
}
