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

package trace

import (
	logtrace "ztna-core/ztna/logtrace"

	trace_pb "github.com/openziti/channel/v3/trace/pb"
)

type Filter interface {
	Accept(event *trace_pb.ChannelMessage) bool
}

func NewAllowAllFilter() Filter {
	logtrace.LogWithFunctionName()
	return &allowAllFilter{}
}

func NewIncludeFilter(includedContentTypes []int32) Filter {
	logtrace.LogWithFunctionName()
	return &includeFilter{contentTypes: includedContentTypes}
}

func NewExcludeFilter(excludedContentTypes []int32) Filter {
	logtrace.LogWithFunctionName()
	return &excludeFilter{contentTypes: excludedContentTypes}
}

type allowAllFilter struct{}

func (*allowAllFilter) Accept(event *trace_pb.ChannelMessage) bool {
	logtrace.LogWithFunctionName()
	return true
}

type includeFilter struct {
	contentTypes []int32
}

func (filter *includeFilter) Accept(event *trace_pb.ChannelMessage) bool {
	logtrace.LogWithFunctionName()
	for _, contentType := range filter.contentTypes {
		if event.ContentType == contentType {
			return true
		}
	}
	return false
}

type excludeFilter struct {
	contentTypes []int32
}

func (filter *excludeFilter) Accept(event *trace_pb.ChannelMessage) bool {
	logtrace.LogWithFunctionName()
	for _, contentType := range filter.contentTypes {
		if event.ContentType == contentType {
			return false
		}
	}
	return true
}
