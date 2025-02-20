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

package handler_xgress

import (
	"time"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/forwarder"
	"ztna-core/ztna/router/xgress"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
)

type receiveHandler struct {
	forwarder *forwarder.Forwarder
}

func NewReceiveHandler(forwarder *forwarder.Forwarder) *receiveHandler {
	logtrace.LogWithFunctionName()
	return &receiveHandler{
		forwarder: forwarder,
	}
}

func (xrh *receiveHandler) HandleXgressReceive(payload *xgress.Payload, x *xgress.Xgress) {
	logtrace.LogWithFunctionName()
	for {
		if err := xrh.forwarder.ForwardPayload(x.Address(), payload, time.Second); err != nil {
			if !channel.IsTimeout(err) {
				pfxlog.ContextLogger(x.Label()).WithFields(payload.GetLoggerFields()).WithError(err).Error("unable to forward payload")
				xrh.forwarder.ReportForwardingFault(payload.CircuitId, x.CtrlId())
				return
			}
		} else {
			return
		}
	}
}

func (xrh *receiveHandler) HandleControlReceive(control *xgress.Control, x *xgress.Xgress) {
	logtrace.LogWithFunctionName()
	if err := xrh.forwarder.ForwardControl(x.Address(), control); err != nil {
		pfxlog.ContextLogger(x.Label()).WithFields(control.GetLoggerFields()).WithError(err).Error("unable to forward control")
	}
}
