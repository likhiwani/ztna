package oidc_auth

import (
	"context"
	"fmt"
	"net/http"
	"ztna-core/ztna/common"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/logtrace"

	"github.com/zitadel/oidc/v2/pkg/oidc"
)

// contextKey is a private type used to restrict context value access
type contextKey string

// contextKeyHttpRequest is the key value to retrieve the current http.Request from a context
const contextKeyHttpRequest contextKey = "oidc_request"
const contextKeyTokenState contextKey = "oidc_token_state"

// NewChangeCtx creates a change.Context scoped to oidc_auth package
func NewChangeCtx() *change.Context {
	logtrace.LogWithFunctionName()
	ctx := change.New()

	ctx.SetSourceType(SourceTypeOidc).
		SetChangeAuthorType(change.AuthorTypeController)

	return ctx
}

// NewHttpChangeCtx creates a change.Context scoped to oidc_auth package and supplied http.Request
func NewHttpChangeCtx(r *http.Request) *change.Context {
	logtrace.LogWithFunctionName()
	ctx := NewChangeCtx()

	ctx.SetSourceLocal(r.Host).
		SetSourceRemote(r.RemoteAddr).
		SetSourceMethod(r.Method)

	return ctx
}

type TokenState struct {
	AccessClaims  *common.AccessClaims
	RefreshClaims *common.RefreshClaims
}

func TokenStateFromContext(ctx context.Context) (*TokenState, error) {
	logtrace.LogWithFunctionName()
	val := ctx.Value(contextKeyTokenState)

	if val == nil {
		srvErr := oidc.ErrServerError()
		srvErr.Description = "token state context was nil"
		return nil, srvErr
	}

	tokenState := val.(*TokenState)

	if tokenState == nil {
		srvErr := oidc.ErrServerError()
		srvErr.Description = fmt.Sprintf("could not cast token state context value from %T to %T", val, tokenState)
		return nil, srvErr
	}

	return tokenState, nil
}

// HttpRequestFromContext returns the initiating http.Request for the current OIDC context
func HttpRequestFromContext(ctx context.Context) (*http.Request, error) {
	logtrace.LogWithFunctionName()
	httpVal := ctx.Value(contextKeyHttpRequest)

	if httpVal == nil {
		return nil, oidc.ErrServerError()
	}

	httpRequest := httpVal.(*http.Request)

	if httpRequest == nil {
		return nil, oidc.ErrServerError()
	}

	return httpRequest, nil
}
