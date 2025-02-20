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

package handler_edge_ctrl

import (
	"ztna-core/sdk-golang/ziti/edge"
	"ztna-core/ztna/logtrace"
)

type controllerError interface {
	error
	ErrorCode() uint32
}

func internalError(err error) controllerError {
	logtrace.LogWithFunctionName()
	if err == nil {
		return nil
	}
	return internalErrorWrapper{error: err}
}

type internalErrorWrapper struct {
	error
}

func (internalErrorWrapper) ErrorCode() uint32 {
	logtrace.LogWithFunctionName()
	return edge.ErrorCodeInternal
}

type InvalidApiSessionError struct{}

func (InvalidApiSessionError) Error() string {
	logtrace.LogWithFunctionName()
	return "invalid api session"
}

func (self InvalidApiSessionError) ErrorCode() uint32 {
	logtrace.LogWithFunctionName()
	return edge.ErrorCodeInvalidApiSession
}

type InvalidSessionError struct{}

func (InvalidSessionError) Error() string {
	logtrace.LogWithFunctionName()
	return "invalid session"
}

func (self InvalidSessionError) ErrorCode() uint32 {
	logtrace.LogWithFunctionName()
	return edge.ErrorCodeInvalidSession
}

type WrongSessionTypeError struct{}

func (WrongSessionTypeError) Error() string {
	logtrace.LogWithFunctionName()
	return "incorrect session type"
}

func (self WrongSessionTypeError) ErrorCode() uint32 {
	logtrace.LogWithFunctionName()
	return edge.ErrorCodeWrongSessionType
}

type InvalidEdgeRouterForSessionError struct{}

func (InvalidEdgeRouterForSessionError) Error() string {
	logtrace.LogWithFunctionName()
	return "invalid edge router for session"
}

func (self InvalidEdgeRouterForSessionError) ErrorCode() uint32 {
	logtrace.LogWithFunctionName()
	return edge.ErrorCodeInvalidEdgeRouterForSession
}

type InvalidServiceError struct{}

func (InvalidServiceError) Error() string {
	logtrace.LogWithFunctionName()
	return "invalid service"
}

func (self InvalidServiceError) ErrorCode() uint32 {
	logtrace.LogWithFunctionName()
	return edge.ErrorCodeInvalidService
}

type TunnelingNotEnabledError struct{}

func (TunnelingNotEnabledError) Error() string {
	logtrace.LogWithFunctionName()
	return "tunneling not enabled"
}

func (self TunnelingNotEnabledError) ErrorCode() uint32 {
	logtrace.LogWithFunctionName()
	return edge.ErrorCodeTunnelingNotEnabled
}

func invalidTerminator(msg string) controllerError {
	logtrace.LogWithFunctionName()
	return &genericControllerError{
		msg:       msg,
		errorCode: edge.ErrorCodeInvalidTerminator,
	}
}

func invalidCost(msg string) controllerError {
	logtrace.LogWithFunctionName()
	return &genericControllerError{
		msg:       msg,
		errorCode: edge.ErrorCodeInvalidCost,
	}
}

func invalidPrecedence(msg string) controllerError {
	logtrace.LogWithFunctionName()
	return &genericControllerError{
		msg:       msg,
		errorCode: edge.ErrorCodeInvalidPrecedence,
	}
}

func encryptionDataMissing(msg string) controllerError {
	logtrace.LogWithFunctionName()
	return &genericControllerError{
		msg:       msg,
		errorCode: edge.ErrorCodeEncryptionDataMissing,
	}
}

type genericControllerError struct {
	msg       string
	errorCode uint32
}

func (self *genericControllerError) Error() string {
	logtrace.LogWithFunctionName()
	return self.msg
}

func (self *genericControllerError) ErrorCode() uint32 {
	logtrace.LogWithFunctionName()
	return self.errorCode
}
