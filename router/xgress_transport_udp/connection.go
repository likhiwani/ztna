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

package xgress_transport_udp

import (
	"net"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/xgress"

	"github.com/openziti/channel/v3"
	"github.com/openziti/foundation/v2/info"
	"github.com/pkg/errors"
)

func (p *packetConn) LogContext() string {
	logtrace.LogWithFunctionName()
	return p.RemoteAddr().String()
}

func (p *packetConn) ReadPayload() ([]byte, map[uint8][]byte, error) {
	logtrace.LogWithFunctionName()
	buffer := make([]byte, info.MaxUdpPacketSize)
	n, err := p.Read(buffer)
	if err != nil {
		return nil, nil, err
	}
	return buffer[:n], nil, nil
}

func (p *packetConn) WritePayload(data []byte, headers map[uint8][]byte) (n int, err error) {
	logtrace.LogWithFunctionName()
	return p.Write(data)
}

func (self *packetConn) HandleControlMsg(controlType xgress.ControlType, headers channel.Headers, responder xgress.ControlReceiver) error {
	logtrace.LogWithFunctionName()
	if controlType == xgress.ControlTypeTraceRoute {
		xgress.RespondToTraceRequest(headers, "xgress/transport_udp", "", responder)
		return nil
	}
	return errors.Errorf("unhandled control type: %v", controlType)
}

func newPacketConn(conn net.Conn) xgress.Connection {
	logtrace.LogWithFunctionName()
	return &packetConn{conn}
}

type packetConn struct {
	net.Conn
}
