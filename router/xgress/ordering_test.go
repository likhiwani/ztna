package xgress

import (
	"encoding/binary"
	"io"
	"sync/atomic"
	"testing"
	"time"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
	"github.com/openziti/metrics"
	"github.com/stretchr/testify/require"
)

type testConn struct {
	ch          chan uint64
	closeNotify chan struct{}
	closed      atomic.Bool
}

func (conn *testConn) Close() error {
	logtrace.LogWithFunctionName()
	if conn.closed.CompareAndSwap(false, true) {
		close(conn.closeNotify)
	}
	return nil
}

func (conn *testConn) LogContext() string {
	logtrace.LogWithFunctionName()
	return "test"
}

func (conn *testConn) ReadPayload() ([]byte, map[uint8][]byte, error) {
	logtrace.LogWithFunctionName()
	<-conn.closeNotify
	return nil, nil, io.EOF
}

func (conn *testConn) WritePayload(bytes []byte, _ map[uint8][]byte) (int, error) {
	logtrace.LogWithFunctionName()
	val := binary.LittleEndian.Uint64(bytes)
	conn.ch <- val
	return len(bytes), nil
}

func (conn *testConn) HandleControlMsg(ControlType, channel.Headers, ControlReceiver) error {
	logtrace.LogWithFunctionName()
	return nil
}

type noopForwarder struct{}

func (n noopForwarder) ForwardPayload(Address, *Payload) error {
	logtrace.LogWithFunctionName()
	return nil
}

func (n noopForwarder) ForwardAcknowledgement(Address, *Acknowledgement) error {
	logtrace.LogWithFunctionName()
	return nil
}

func (n noopForwarder) RetransmitPayload(Address, *Payload) error {
	logtrace.LogWithFunctionName()
	return nil
}

type noopReceiveHandler struct{}

func (n noopReceiveHandler) HandleXgressReceive(*Payload, *Xgress) {
	logtrace.LogWithFunctionName()
}

func (n noopReceiveHandler) HandleControlReceive(*Control, *Xgress) {
	logtrace.LogWithFunctionName()
}

func Test_Ordering(t *testing.T) {
	logtrace.LogWithFunctionName()
	closeNotify := make(chan struct{})
	metricsRegistry := metrics.NewUsageRegistry("test", map[string]string{}, closeNotify)
	InitPayloadIngester(closeNotify)
	InitMetrics(metricsRegistry)
	InitAcker(&noopForwarder{}, metricsRegistry, closeNotify)

	conn := &testConn{
		ch:          make(chan uint64, 1),
		closeNotify: make(chan struct{}),
	}

	x := NewXgress("test", "ctrl", "test", conn, Initiator, DefaultOptions(), nil)
	x.receiveHandler = noopReceiveHandler{}
	go x.tx()

	defer x.Close()

	msgCount := 100000

	errorCh := make(chan error, 1)

	go func() {
		for i := 0; i < msgCount; i++ {
			data := make([]byte, 8)
			binary.LittleEndian.PutUint64(data, uint64(i))
			payload := &Payload{
				CircuitId: "test",
				Flags:     SetOriginatorFlag(0, Terminator),
				RTT:       0,
				Sequence:  int32(i),
				Headers:   nil,
				Data:      data,
			}
			if err := x.SendPayload(payload, 0, PayloadTypeXg); err != nil {
				errorCh <- err
				x.Close()
				return
			}
		}
	}()

	timeout := time.After(20 * time.Second)

	req := require.New(t)
	for i := 0; i < msgCount; i++ {
		select {
		case next := <-conn.ch:
			req.Equal(uint64(i), next)
		case <-conn.closeNotify:
			req.Fail("test failed with count at %v", i)
		case err := <-errorCh:
			req.NoError(err)
		case <-timeout:
			req.Failf("timed out", "count at %v", i)
		}
	}
}
