package oidc_auth

import (
	"fmt"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/foundation/v2/errorz"
)

func newNotAcceptableError(acceptHeader string) *errorz.ApiError {
	logtrace.LogWithFunctionName()
	return &errorz.ApiError{
		Code:    "NOT_ACCEPTABLE",
		Message: fmt.Sprintf("the request is not acceptable, the accept header did not have any supported options: %s (supported: %s, %s)", acceptHeader, JsonContentType, HtmlContentType),
	}
}
