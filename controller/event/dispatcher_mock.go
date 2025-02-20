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

package event

import (
	"regexp"
	"time"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/metrics/metrics_pb"
	"github.com/openziti/storage/boltz"
)

var _ Dispatcher = DispatcherMock{}

type DispatcherMock struct{}

func (d DispatcherMock) AcceptConnectEvent(event *ConnectEvent) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AcceptSdkEvent(event *SdkEvent) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AcceptApiSessionEvent(event *ApiSessionEvent) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddApiSessionEventHandler(handler ApiSessionEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RemoveApiSessionEventHandler(handler ApiSessionEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddSessionEventHandler(handler SessionEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RemoveSessionEventHandler(handler SessionEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddEntityCountEventHandler(handler EntityCountEventHandler, interval time.Duration, onlyLeaderEvents bool) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RemoveEntityCountEventHandler(handler EntityCountEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddEntityChangeSource(store boltz.Store) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddGlobalEntityChangeMetadata(k string, v any) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddEntityChangeEventHandler(handler EntityChangeEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RemoveEntityChangeEventHandler(handler EntityChangeEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AcceptEntityChangeEvent(event *EntityChangeEvent) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) GetFormatterFactory(formatterType string) FormatterFactory {
	logtrace.LogWithFunctionName()
	return nil
}

func (d DispatcherMock) RegisterFormatterFactory(string, FormatterFactory) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RegisterEventTypeFunctions(string, RegistrationHandler, UnregistrationHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) ProcessSubscriptions(interface{}, []*Subscription) error {
	logtrace.LogWithFunctionName()
	return nil
}

func (d DispatcherMock) RemoveAllSubscriptions(interface{}) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RegisterEventType(string, TypeRegistrar) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RegisterEventHandlerFactory(string, HandlerFactory) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddCircuitEventHandler(CircuitEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RemoveCircuitEventHandler(CircuitEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddLinkEventHandler(LinkEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RemoveLinkEventHandler(LinkEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddMetricsMapper(MetricsMapper) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddMetricsEventHandler(MetricsEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RemoveMetricsEventHandler(MetricsEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddMetricsMessageHandler(MetricsMessageHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RemoveMetricsMessageHandler(MetricsMessageHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) NewFilteredMetricsAdapter(*regexp.Regexp, *regexp.Regexp, MetricsEventHandler) MetricsMessageHandler {
	logtrace.LogWithFunctionName()
	return nil
}

func (d DispatcherMock) AddRouterEventHandler(RouterEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RemoveRouterEventHandler(RouterEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddServiceEventHandler(ServiceEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RemoveServiceEventHandler(ServiceEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddTerminatorEventHandler(TerminatorEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RemoveTerminatorEventHandler(TerminatorEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddUsageEventHandler(UsageEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RemoveUsageEventHandler(UsageEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AcceptCircuitEvent(*CircuitEvent) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AcceptLinkEvent(*LinkEvent) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AcceptMetricsEvent(*MetricsEvent) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AcceptMetricsMsg(*metrics_pb.MetricsMessage) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AcceptRouterEvent(*RouterEvent) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AcceptServiceEvent(*ServiceEvent) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AcceptTerminatorEvent(*TerminatorEvent) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AcceptUsageEvent(*UsageEvent) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AddClusterEventHandler(ClusterEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) RemoveClusterEventHandler(ClusterEventHandler) {
	logtrace.LogWithFunctionName()
}

func (d DispatcherMock) AcceptClusterEvent(*ClusterEvent) {
	logtrace.LogWithFunctionName()
}
