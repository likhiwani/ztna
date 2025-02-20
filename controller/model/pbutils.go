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

package model

import (
	"time"
	"ztna-core/ztna/common/pb/edge_cmd_pb"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/logtrace"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func timePtrToPb(t *time.Time) *timestamppb.Timestamp {
	logtrace.LogWithFunctionName()
	if t == nil {
		return nil
	}
	result := timestamppb.New(*t)
	return result
}

func pbTimeToTimePtr(pb *timestamppb.Timestamp) *time.Time {
	logtrace.LogWithFunctionName()
	if pb == nil {
		return nil
	}
	result := pb.AsTime()
	return &result
}

func ContextToProtobuf(context *change.Context) *edge_cmd_pb.ChangeContext {
	logtrace.LogWithFunctionName()
	if context == nil {
		return nil
	}
	return &edge_cmd_pb.ChangeContext{
		Attributes: context.Attributes,
		RaftIndex:  context.RaftIndex,
	}
}

func ProtobufToContext(context *edge_cmd_pb.ChangeContext) *change.Context {
	logtrace.LogWithFunctionName()
	if context == nil {
		return nil
	}
	return &change.Context{
		Attributes: context.Attributes,
		RaftIndex:  context.RaftIndex,
	}
}
