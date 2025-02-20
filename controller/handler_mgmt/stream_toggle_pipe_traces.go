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

package handler_mgmt

import (
	"fmt"
	"sync"
	"time"
	"ztna-core/ztna/controller/model"
	"ztna-core/ztna/logtrace"

	"ztna-core/ztna/common/handler_common"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/common/pb/mgmt_pb"
	"ztna-core/ztna/common/trace"
	"ztna-core/ztna/controller/network"

	"github.com/openziti/channel/v3"
	trace_pb "github.com/openziti/channel/v3/trace/pb"
	"google.golang.org/protobuf/proto"
)

type traceTogglePipeHandler struct {
	eventHandler trace.EventHandler
	network      *network.Network
}

func newTogglePipeTracesHandler(network *network.Network) *traceTogglePipeHandler {
	logtrace.LogWithFunctionName()
	return &traceTogglePipeHandler{
		eventHandler: network.GetTraceController(),
		network:      network,
	}
}

func (*traceTogglePipeHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(mgmt_pb.ContentType_TogglePipeTracesRequestType)
}

func (handler *traceTogglePipeHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	request := &trace_pb.TogglePipeTracesRequest{}

	if err := proto.Unmarshal(msg.Body, request); err != nil {
		handler_common.SendFailure(msg, ch, err.Error())
		return
	}

	matchers, result := trace.NewPipeToggleMatchers(request)

	if !result.Success {
		handler.complete(msg, ch, result)
		return
	}

	resultChan := make(chan trace.ToggleApplyResult)

	verbosity := trace.GetVerbosity(request.Verbosity)

	if checkMatch(handler.network.GetAppId(), matchers, verbosity, result) {
		if request.Enable {
			handler.network.GetTraceController().EnableTracing(trace.SourceTypePipe, matchers.PipeMatcher, handler.eventHandler, resultChan)
		} else {
			handler.network.GetTraceController().DisableTracing(trace.SourceTypePipe, matchers.PipeMatcher, handler.eventHandler, resultChan)
		}
		getApplyResults(resultChan, verbosity, result)
	}

	if !result.Success {
		handler.complete(msg, ch, result)
		return
	}

	remoteResultChan := make(chan *remoteToggleResult)
	waitGroup := &sync.WaitGroup{}

	for _, router := range handler.network.AllConnectedRouters() {
		if checkMatch(router.Id, matchers, verbosity, result) {
			waitGroup.Add(1)
			go handleResponse(router, msg, remoteResultChan, waitGroup)
		}
	}

	// Close chan once all results have been queued
	go func() {
		waitGroup.Wait()
		close(remoteResultChan)
	}()

	for remoteToggleResult := range remoteResultChan {
		if !remoteToggleResult.success {
			result.Success = false
		}
		result.Message.WriteString(remoteToggleResult.message)
	}

	handler.complete(msg, ch, result)
}

func (handler *traceTogglePipeHandler) complete(msg *channel.Message, ch channel.Channel, result *trace.ToggleResult) {
	logtrace.LogWithFunctionName()
	if result.Success {
		handler_common.SendSuccess(msg, ch, result.Message.String())
	} else {
		handler_common.SendFailure(msg, ch, result.Message.String())
	}
}

func checkMatch(appId string, matchers *trace.PipeToggleMatchers, verbosity trace.ToggleVerbosity, result *trace.ToggleResult) bool {
	logtrace.LogWithFunctionName()
	appMatches := matchers.AppMatcher.Matches(appId)
	applyResult := &trace.ToggleApplyResultImpl{
		Matched: appMatches,
		Message: fmt.Sprintf("App %v matched? %v", appId, appMatches),
	}
	applyResult.Append(result, verbosity)
	return appMatches
}

func getApplyResults(resultChan chan trace.ToggleApplyResult, verbosity trace.ToggleVerbosity, result *trace.ToggleResult) {
	logtrace.LogWithFunctionName()
	timeout := time.After(time.Second * 5)
	for {
		select {
		case applyResult, chanOpen := <-resultChan:
			if !chanOpen {
				return
			}
			applyResult.Append(result, verbosity)
		case <-timeout:
			result.Success = false
			result.Append("Timed out waiting for toggle to be applied to controller")
			return
		}
	}
}

func handleResponse(router *model.Router, mgmtReq *channel.Message, msgsCh chan<- *remoteToggleResult, waitGroup *sync.WaitGroup) {
	logtrace.LogWithFunctionName()
	defer waitGroup.Done()

	msg := channel.NewMessage(int32(ctrl_pb.ContentType_TogglePipeTracesRequestType), mgmtReq.Body)
	response, err := msg.WithTimeout(5 * time.Second).SendForReply(router.Control)

	if err != nil {
		msgsCh <- &remoteToggleResult{false, err.Error()}
	} else if response.ContentType == channel.ContentTypeResultType {
		result := channel.UnmarshalResult(response)
		msgsCh <- &remoteToggleResult{result.Success, result.Message}
	} else {
		msg := fmt.Sprintf("Unexpected response type from router %v: %v\n", router.Id, response.ContentType)
		msgsCh <- &remoteToggleResult{false, msg}
	}
}

type remoteToggleResult struct {
	success bool
	message string
}
