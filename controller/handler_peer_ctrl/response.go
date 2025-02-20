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

package handler_peer_ctrl

import (
	"encoding/json"
	"fmt"
	"ztna-core/ztna/common/pb/cmd_pb"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/controller/peermsg"
	"ztna-core/ztna/logtrace"

	"github.com/hashicorp/raft"
	"github.com/openziti/channel/v3"
	"github.com/openziti/foundation/v2/errorz"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func sendErrorResponseCalculateType(m *channel.Message, ch channel.Channel, err error) {
	logtrace.LogWithFunctionName()
	if errors.Is(err, raft.ErrNotLeader) {
		sendErrorResponse(m, ch, err, peermsg.ErrorCodeNotLeader)
	} else {
		sendApiErrorResponse(m, ch, models.ToApiError(err))
	}
}

func sendErrorResponse(m *channel.Message, ch channel.Channel, err error, errorCode uint32) {
	logtrace.LogWithFunctionName()
	resp := channel.NewMessage(int32(cmd_pb.ContentType_ErrorResponseType), []byte(err.Error()))
	resp.ReplyTo(m)
	resp.PutUint32Header(peermsg.HeaderErrorCode, errorCode)

	if sendErr := ch.Send(resp); sendErr != nil {
		logrus.WithError(sendErr).Error("error while sending error response")
	}
}

func sendApiErrorResponse(m *channel.Message, ch channel.Channel, err *errorz.ApiError) {
	logtrace.LogWithFunctionName()
	encodingMap := map[string]interface{}{}
	encodingMap["code"] = err.Code
	encodingMap["message"] = err.Message
	encodingMap["status"] = err.Status
	encodingMap["cause"] = err.Cause
	if err.Cause != nil {
		encodingMap["causeType"] = fmt.Sprintf("%T", err.Cause)
	}

	buf, encodeErr := json.Marshal(encodingMap)
	if encodeErr != nil {
		logrus.WithError(encodeErr).WithField("apiErr", err).Error("unable to encode api error")
		sendErrorResponse(m, ch, err, peermsg.ErrorCodeGeneric)
		return
	}
	resp := channel.NewMessage(int32(cmd_pb.ContentType_ErrorResponseType), buf)
	resp.ReplyTo(m)
	resp.PutUint32Header(peermsg.HeaderErrorCode, peermsg.ErrorCodeApiError)

	if sendErr := ch.Send(resp); sendErr != nil {
		logrus.WithError(sendErr).Error("error while sending error response")
	}
}

func sendSuccessResponse(m *channel.Message, ch channel.Channel, index uint64) {
	logtrace.LogWithFunctionName()
	resp := channel.NewMessage(int32(cmd_pb.ContentType_SuccessResponseType), nil)
	resp.ReplyTo(m)
	resp.PutUint64Header(peermsg.HeaderIndex, index)
	if sendErr := ch.Send(resp); sendErr != nil {
		logrus.WithError(sendErr).Error("error while sending success response")
	}
}
