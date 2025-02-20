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
	"fmt"
	"regexp"
	"strings"
	logtrace "ztna-core/ztna/logtrace"

	trace_pb "github.com/openziti/channel/v3/trace/pb"
)

type SourceType int

const (
	SourceTypePipe SourceType = iota
)

type ToggleVerbosity int

const (
	ToggleVerbosityNone ToggleVerbosity = iota
	ToggleVerbosityMatches
	ToggleVerbosityMisses
	ToggleVerbosityAll
)

func (val ToggleVerbosity) Show(matched bool) bool {
	logtrace.LogWithFunctionName()
	if matched {
		return val == ToggleVerbosityMatches || val == ToggleVerbosityAll
	}
	return val == ToggleVerbosityMisses || val == ToggleVerbosityAll
}

func GetVerbosity(verbosity trace_pb.TraceToggleVerbosity) ToggleVerbosity {
	logtrace.LogWithFunctionName()
	switch verbosity {
	case trace_pb.TraceToggleVerbosity_ReportNone:
		return ToggleVerbosityNone
	case trace_pb.TraceToggleVerbosity_ReportMatches:
		return ToggleVerbosityMatches
	case trace_pb.TraceToggleVerbosity_ReportMisses:
		return ToggleVerbosityMisses
	case trace_pb.TraceToggleVerbosity_ReportAll:
		return ToggleVerbosityAll
	}
	return ToggleVerbosityNone
}

type SourceMatcher interface {
	Matches(source string) bool
}

type PipeToggleMatchers struct {
	AppMatcher  SourceMatcher
	PipeMatcher SourceMatcher
}

func NewPipeToggleMatchers(request *trace_pb.TogglePipeTracesRequest) (*PipeToggleMatchers, *ToggleResult) {
	logtrace.LogWithFunctionName()
	result := &ToggleResult{Success: true, Message: &strings.Builder{}}

	appRegex, err := regexp.Compile(request.AppRegex)
	if err != nil {
		result.Success = false
		errMsg := fmt.Sprintf("Failed to parse app id regex '%v' with error: %v\n", request.AppRegex, err)
		result.Message.WriteString(errMsg)
	}

	pipeRegex, err := regexp.Compile(request.PipeRegex)
	if err != nil {
		result.Success = false
		errMsg := fmt.Sprintf("Failed to parse pipe id regex '%v' with error: %v\n", request.PipeRegex, err)
		result.Message.WriteString(errMsg)
	}

	appMatcher := NewSourceMatcher(appRegex)
	pipeMatcher := NewSourceMatcher(pipeRegex)

	return &PipeToggleMatchers{appMatcher, pipeMatcher}, result
}

type ToggleResult struct {
	Success bool
	Message *strings.Builder
}

func (result *ToggleResult) Append(msg string) {
	logtrace.LogWithFunctionName()
	result.Message.WriteString(msg)
	result.Message.WriteString("\n")
}

type ToggleApplyResult interface {
	IsMatched() bool
	GetMessage() string
	Append(result *ToggleResult, verbosity ToggleVerbosity)
}

type ToggleApplyResultImpl struct {
	Matched bool
	Message string
}

func (applyResult *ToggleApplyResultImpl) IsMatched() bool {
	logtrace.LogWithFunctionName()
	return applyResult.Matched
}

func (applyResult *ToggleApplyResultImpl) GetMessage() string {
	logtrace.LogWithFunctionName()
	return applyResult.Message
}

func (applyResult *ToggleApplyResultImpl) Append(result *ToggleResult, verbosity ToggleVerbosity) {
	logtrace.LogWithFunctionName()
	if verbosity.Show(applyResult.Matched) {
		result.Append(applyResult.Message)
	}
}

type sourceRegexpMatcher struct {
	regex *regexp.Regexp
}

func (matcher *sourceRegexpMatcher) Matches(source string) bool {
	logtrace.LogWithFunctionName()
	return matcher.regex.Match([]byte(source))
}

func NewSourceMatcher(regex *regexp.Regexp) SourceMatcher {
	logtrace.LogWithFunctionName()
	return &sourceRegexpMatcher{regex}
}

type Source interface {
	EnableTracing(sourceType SourceType, matcher SourceMatcher, handler EventHandler, resultChan chan<- ToggleApplyResult)
	DisableTracing(sourceType SourceType, matcher SourceMatcher, handler EventHandler, resultChan chan<- ToggleApplyResult)
}

type Controller interface {
	Source
	AddSource(source Source)
	RemoveSource(source Source)
	EventHandler
}

func NewController(closeNotify <-chan struct{}) Controller {
	logtrace.LogWithFunctionName()
	controller := &controllerImpl{
		events:      make(chan controllerEvent, 25),
		sources:     make(map[Source]Source),
		closeNotify: closeNotify,
	}
	go controller.run()
	return controller
}

type enableSourcesEvent struct {
	sourceType SourceType
	matcher    SourceMatcher
	handler    EventHandler
	resultChan chan<- ToggleApplyResult
}

func (event *enableSourcesEvent) handle(controller *controllerImpl) {
	logtrace.LogWithFunctionName()
	for source := range controller.sources {
		source.EnableTracing(event.sourceType, event.matcher, event.handler, event.resultChan)
	}
	close(event.resultChan)
}

type disableSourcesEvent struct {
	sourceType SourceType
	matcher    SourceMatcher
	handler    EventHandler
	resultChan chan<- ToggleApplyResult
}

func (event *disableSourcesEvent) handle(controller *controllerImpl) {
	logtrace.LogWithFunctionName()
	for handler := range controller.sources {
		handler.DisableTracing(event.sourceType, event.matcher, event.handler, event.resultChan)
	}
	close(event.resultChan)
}

type sourceAddedEvent struct {
	source Source
}

func (event *sourceAddedEvent) handle(controller *controllerImpl) {
	logtrace.LogWithFunctionName()
	controller.sources[event.source] = event.source
}

type sourceRemovedEvent struct {
	source Source
}

func (event *sourceRemovedEvent) handle(controller *controllerImpl) {
	logtrace.LogWithFunctionName()
	delete(controller.sources, event.source)
}

type controllerEvent interface {
	handle(impl *controllerImpl)
}

type controllerImpl struct {
	events      chan controllerEvent
	sources     map[Source]Source
	closeNotify <-chan struct{}
}

func (controller *controllerImpl) Accept(event *trace_pb.ChannelMessage) {
	logtrace.LogWithFunctionName()
	select {
	case controller.events <- &eventWrapper{wrapped: event}:
	case <-controller.closeNotify:
	}
}

func (controller *controllerImpl) EnableTracing(sourceType SourceType, matcher SourceMatcher, handler EventHandler, resultChan chan<- ToggleApplyResult) {
	logtrace.LogWithFunctionName()
	select {
	case controller.events <- &enableSourcesEvent{sourceType: sourceType, matcher: matcher, handler: handler, resultChan: resultChan}:
	case <-controller.closeNotify:
	}
}

func (controller *controllerImpl) DisableTracing(sourceType SourceType, matcher SourceMatcher, handler EventHandler, resultChan chan<- ToggleApplyResult) {
	logtrace.LogWithFunctionName()
	select {
	case controller.events <- &disableSourcesEvent{sourceType: sourceType, matcher: matcher, handler: handler, resultChan: resultChan}:
	case <-controller.closeNotify:
	}
}

func (controller *controllerImpl) AddSource(source Source) {
	logtrace.LogWithFunctionName()
	select {
	case controller.events <- &sourceAddedEvent{source}:
	case <-controller.closeNotify:
	}
}

func (controller *controllerImpl) RemoveSource(source Source) {
	logtrace.LogWithFunctionName()
	select {
	case controller.events <- &sourceRemovedEvent{source}:
	case <-controller.closeNotify:
	}
}

func (controller *controllerImpl) run() {
	logtrace.LogWithFunctionName()
	for {
		select {
		case <-controller.closeNotify:
			return
		case evt := <-controller.events:
			evt.handle(controller)
		}
	}
}
