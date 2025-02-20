package env

import (
	"time"
	"ztna-core/ztna/logtrace"
)

func init() {
	logtrace.LogWithFunctionName()
	IntervalSize = time.Minute
}

var IntervalSize time.Duration
