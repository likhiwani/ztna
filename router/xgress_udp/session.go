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

package xgress_udp

import (
	"io"
	"net"
	"time"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/xgress"

	"github.com/openziti/channel/v3"
	"github.com/pkg/errors"
)

func NewPacketSesssion(l Listener, addr net.Addr, timeout int64) Session {
	logtrace.LogWithFunctionName()
	return &PacketSession{
		listener:             l,
		readC:                make(chan []byte, 10),
		addr:                 addr,
		state:                SessionStateNew,
		timeoutIntervalNanos: timeout,
	}
}

func (s *PacketSession) State() SessionState {
	logtrace.LogWithFunctionName()
	return s.state
}

func (s *PacketSession) SetState(state SessionState) {
	logtrace.LogWithFunctionName()
	s.state = state
}

func (s *PacketSession) Address() net.Addr {
	logtrace.LogWithFunctionName()
	return s.addr
}

func (s *PacketSession) ReadPayload() ([]byte, map[uint8][]byte, error) {
	logtrace.LogWithFunctionName()
	buffer, chanOpen := <-s.readC
	if !chanOpen {
		return buffer, nil, io.EOF
	}
	return buffer, nil, nil
}

func (s *PacketSession) Write(p []byte) (n int, err error) {
	logtrace.LogWithFunctionName()
	s.listener.QueueEvent((*SessionUpdateEvent)(s))
	return s.listener.WriteTo(p, s.addr)
}

func (s *PacketSession) WritePayload(p []byte, _ map[uint8][]byte) (n int, err error) {
	logtrace.LogWithFunctionName()
	return s.Write(p)
}

func (s *PacketSession) HandleControlMsg(controlType xgress.ControlType, headers channel.Headers, responder xgress.ControlReceiver) error {
	logtrace.LogWithFunctionName()
	if controlType == xgress.ControlTypeTraceRoute {
		xgress.RespondToTraceRequest(headers, "xgress/udp", "", responder)
		return nil
	}
	return errors.Errorf("unhandled control type: %v", controlType)
}

func (s *PacketSession) QueueRead(data []byte) {
	logtrace.LogWithFunctionName()
	s.readC <- data
}

func (s *PacketSession) Close() error {
	logtrace.LogWithFunctionName()
	s.listener.QueueEvent((*sessionCloseEvent)(s))
	return nil
}

func (s *PacketSession) LogContext() string {
	logtrace.LogWithFunctionName()
	return s.addr.String()
}

func (s *PacketSession) TimeoutNanos() int64 {
	logtrace.LogWithFunctionName()
	return s.timeoutNanos
}

func (s *PacketSession) MarkActivity() {
	logtrace.LogWithFunctionName()
	s.timeoutNanos = time.Now().UnixNano() + s.timeoutIntervalNanos
}

func (s *PacketSession) SessionId() string {
	logtrace.LogWithFunctionName()
	return s.addr.String()
}

type PacketSession struct {
	listener             Listener
	readC                chan []byte
	addr                 net.Addr
	state                SessionState
	timeoutIntervalNanos int64
	timeoutNanos         int64
	closed               bool
}

func (s *SessionUpdateEvent) Handle(_ Listener) {
	logtrace.LogWithFunctionName()
	(*PacketSession)(s).MarkActivity()
}

type SessionUpdateEvent PacketSession

func (e *sessionCloseEvent) Handle(l Listener) {
	logtrace.LogWithFunctionName()
	session := (*PacketSession)(e)
	if !session.closed {
		close(session.readC)
		l.DeleteSession(session.SessionId())
		session.closed = true
	}
}

type sessionCloseEvent PacketSession
