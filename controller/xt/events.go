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

package xt

import "ztna-core/ztna/logtrace"

func NewStrategyChangeEvent(serviceId string, current, added, changed, removed []Terminator) StrategyChangeEvent {
	logtrace.LogWithFunctionName()
	return &strategyChangeEvent{
		serviceId: serviceId,
		current:   current,
		added:     added,
		changed:   changed,
		removed:   removed,
	}
}

func TList(terminators ...Terminator) []Terminator {
	logtrace.LogWithFunctionName()
	return terminators
}

type strategyChangeEvent struct {
	serviceId string
	current   []Terminator
	added     []Terminator
	changed   []Terminator
	removed   []Terminator
}

func (event *strategyChangeEvent) GetServiceId() string {
	logtrace.LogWithFunctionName()
	return event.serviceId
}

func (event *strategyChangeEvent) GetCurrent() []Terminator {
	logtrace.LogWithFunctionName()
	return event.current
}

func (event *strategyChangeEvent) GetAdded() []Terminator {
	logtrace.LogWithFunctionName()
	return event.added
}

func (event *strategyChangeEvent) GetChanged() []Terminator {
	logtrace.LogWithFunctionName()
	return event.changed
}

func (event *strategyChangeEvent) GetRemoved() []Terminator {
	logtrace.LogWithFunctionName()
	return event.removed
}

func NewDialFailedEvent(terminator Terminator) TerminatorEvent {
	logtrace.LogWithFunctionName()
	return &defaultEvent{
		terminator: terminator,
		eventType:  eventTypeFailed,
	}
}

func NewDialSucceeded(terminator Terminator) TerminatorEvent {
	logtrace.LogWithFunctionName()
	return &defaultEvent{
		terminator: terminator,
		eventType:  eventTypeSucceeded,
	}
}

func NewCircuitRemoved(terminator Terminator) TerminatorEvent {
	logtrace.LogWithFunctionName()
	return &defaultEvent{
		terminator: terminator,
		eventType:  eventTypeCircuitRemoved,
	}
}

type eventType int

const (
	eventTypeFailed eventType = iota
	eventTypeSucceeded
	eventTypeCircuitRemoved
)

type defaultEvent struct {
	terminator Terminator
	eventType  eventType
}

func (event *defaultEvent) GetTerminator() Terminator {
	logtrace.LogWithFunctionName()
	return event.terminator
}

func (event *defaultEvent) Accept(visitor EventVisitor) {
	logtrace.LogWithFunctionName()
	if event.eventType == eventTypeFailed {
		visitor.VisitDialFailed(event)
	} else if event.eventType == eventTypeSucceeded {
		visitor.VisitDialSucceeded(event)
	} else if event.eventType == eventTypeCircuitRemoved {
		visitor.VisitCircuitRemoved(event)
	}
}

var _ EventVisitor = DefaultEventVisitor{}

type DefaultEventVisitor struct{}

func (visitor DefaultEventVisitor) VisitDialFailed(TerminatorEvent) {
	logtrace.LogWithFunctionName()
}
func (visitor DefaultEventVisitor) VisitDialSucceeded(TerminatorEvent) {
	logtrace.LogWithFunctionName()
}
func (visitor DefaultEventVisitor) VisitCircuitRemoved(TerminatorEvent) {
	logtrace.LogWithFunctionName()
}
