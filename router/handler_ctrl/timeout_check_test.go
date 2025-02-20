package handler_ctrl

import (
	"fmt"
	"testing"
	"ztna-core/ztna/logtrace"

	"github.com/stretchr/testify/assert"
)

type testTimeoutErr struct{}

func (t testTimeoutErr) Error() string {
	logtrace.LogWithFunctionName()
	return "test"
}

func (t testTimeoutErr) Timeout() bool {
	logtrace.LogWithFunctionName()
	return true
}

func (t testTimeoutErr) Temporary() bool {
	logtrace.LogWithFunctionName()
	return true
}

func Test_TimeoutCheck(t *testing.T) {
	logtrace.LogWithFunctionName()
	err := testTimeoutErr{}
	req := assert.New(t)
	req.True(isNetworkTimeout(err))

	wrapped := fmt.Errorf("there was an error (%w)", err)
	req.True(isNetworkTimeout(wrapped))
}
