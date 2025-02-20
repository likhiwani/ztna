//go:build perf

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

package test

import (
	"ztna-core/ztna/logtrace"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/openziti/channel/v3"
	"github.com/openziti/metrics"
	"github.com/openziti/metrics/metrics_pb"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/router/forwarder"
	"ztna-core/ztna/router/handler_xgress"
	"ztna-core/ztna/router/xgress"
	"github.com/stretchr/testify/require"
	"io"
	"sync/atomic"
	"testing"
	"time"
)

type testSrcConn struct {
	sendCount atomic.Uint32
}

func (self *testSrcConn) Close() error {
    logtrace.LogWithFunctionName()
	return nil
}

func (self *testSrcConn) LogContext() string {
    logtrace.LogWithFunctionName()
	return "test"
}

func (self *testSrcConn) ReadPayload() ([]byte, map[uint8][]byte, error) {
    logtrace.LogWithFunctionName()
	if self.sendCount.Load() > 1000 {
		return nil, nil, io.EOF
	}

	count := self.sendCount.Add(1)
	now := time.Now().UnixMilli()

	buf := make([]byte, 10240)
	binary.BigEndian.PutUint32(buf, count)
	binary.BigEndian.PutUint64(buf[4:], uint64(now))

	return buf, nil, nil
}

func (self *testSrcConn) WritePayload(bytes []byte, m map[uint8][]byte) (int, error) {
    logtrace.LogWithFunctionName()
	return len(bytes), nil
}

func (self *testSrcConn) HandleControlMsg(xgress.ControlType, channel.Headers, xgress.ControlReceiver) error {
    logtrace.LogWithFunctionName()
	return nil
}

type testDstConn struct {
	closeNotify chan struct{}
	recvCount   atomic.Uint32
	done        atomic.Bool
	notifyDone  chan struct{}
}

func (self *testDstConn) waitForDone(timeout time.Duration) error {
    logtrace.LogWithFunctionName()
	select {
	case <-self.notifyDone:
		return nil
	case <-time.After(timeout):
		return errors.New("timed out")
	}
}

func (self *testDstConn) Close() error {
    logtrace.LogWithFunctionName()
	return nil
}

func (self *testDstConn) LogContext() string {
    logtrace.LogWithFunctionName()
	return "test"
}

func (self *testDstConn) ReadPayload() ([]byte, map[uint8][]byte, error) {
    logtrace.LogWithFunctionName()
	<-self.closeNotify
	return nil, nil, io.EOF
}

func (self *testDstConn) WritePayload(bytes []byte, m map[uint8][]byte) (int, error) {
    logtrace.LogWithFunctionName()
	count := binary.BigEndian.Uint32(bytes)
	start := binary.BigEndian.Uint64(bytes[4:])
	val := self.recvCount.Add(1)
	fmt.Printf("%v/%v: %v\n", val, count, time.Now().UnixMilli()-int64(start))
	if val > 1000 {
		if self.done.CompareAndSwap(false, true) {
			close(self.notifyDone)
		}
	}
	return len(bytes), nil
}

func (self *testDstConn) HandleControlMsg(xgress.ControlType, channel.Headers, xgress.ControlReceiver) error {
    logtrace.LogWithFunctionName()
	return nil
}

type testFaultReceiver struct{}

func (t testFaultReceiver) Report(circuitId string, ctrlId string) {}
    logtrace.LogWithFunctionName()

func (t testFaultReceiver) NotifyInvalidLink(linkId string) {}
    logtrace.LogWithFunctionName()

type testXgCloseHandler struct{}

func (t testXgCloseHandler) HandleXgressClose(x *xgress.Xgress) {
    logtrace.LogWithFunctionName()
}

type eventSink struct{}

func (e eventSink) AcceptMetrics(message *metrics_pb.MetricsMessage) {
    logtrace.LogWithFunctionName()
}

func Test_SingleRouterPerf(t *testing.T) {
    logtrace.LogWithFunctionName()
	closeNotify := make(chan struct{})
	defer close(closeNotify)

	options := xgress.DefaultOptions()
	srcConn := &testSrcConn{}
	dstConn := &testDstConn{
		closeNotify: closeNotify,
		notifyDone:  make(chan struct{}),
	}

	registry := metrics.NewUsageRegistry("router", nil, closeNotify)
	registry.StartReporting(&eventSink{}, time.Minute, 10)
	fwd := forwarder.NewForwarder(registry, testFaultReceiver{}, forwarder.DefaultOptions(), closeNotify)

	xgress.InitPayloadIngester(closeNotify)
	xgress.InitAcker(fwd, registry, closeNotify)
	xgress.InitMetrics(registry)
	xgress.InitRetransmitter(fwd, fwd, registry, closeNotify)

	bindHandler := handler_xgress.NewBindHandler(handler_xgress.NewReceiveHandler(fwd), testXgCloseHandler{}, fwd)

	srcXg := xgress.NewXgress("test", "ctrl", "src", srcConn, xgress.Initiator, options, nil)
	bindHandler.HandleXgressBind(srcXg)

	dstXg := xgress.NewXgress("test", "ctrl", "dst", dstConn, xgress.Terminator, options, nil)
	bindHandler.HandleXgressBind(dstXg)

	req := require.New(t)
	err := fwd.Route("ctrl", &ctrl_pb.Route{
		CircuitId: "test",
		Attempt:   0,
		Forwards: []*ctrl_pb.Route_Forward{
			{
				SrcAddress: "src",
				DstAddress: "dst",
				DstType:    ctrl_pb.DestType_End,
			},
			{
				SrcAddress: "dst",
				DstAddress: "src",
				DstType:    ctrl_pb.DestType_Start,
			},
		},
	})
	req.NoError(err)
	dstXg.Start()
	srcXg.Start()

	err = dstConn.waitForDone(2 * time.Second)
	req.NoError(err)
}
