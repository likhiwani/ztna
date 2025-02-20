package xgress

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"
	"ztna-core/ztna/controller/idgen"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/metrics"
	cmap "github.com/orcaman/concurrent-map/v2"
	metrics2 "github.com/rcrowley/go-metrics"
	"github.com/sirupsen/logrus"
)

func newTestXgConn(bufferSize int, targetSends uint32, targetReceives uint32) *testXgConn {
	logtrace.LogWithFunctionName()
	return &testXgConn{
		bufferSize:     bufferSize,
		targetSends:    targetSends,
		targetReceives: targetReceives,
		done:           make(chan struct{}),
		errs:           make(chan error, 1),
	}
}

type testXgConn struct {
	sndMsgCounter  uint32
	rcvMsgCounter  uint32
	bufferSize     int
	targetSends    uint32
	targetReceives uint32
	sendCounter    uint32
	recvCounter    uint32
	done           chan struct{}
	errs           chan error
	bufCounter     uint32
}

func (self *testXgConn) Close() error {
	logtrace.LogWithFunctionName()
	return nil
}

func (self *testXgConn) LogContext() string {
	logtrace.LogWithFunctionName()
	return "test"
}

func (self *testXgConn) ReadPayload() ([]byte, map[uint8][]byte, error) {
	logtrace.LogWithFunctionName()
	self.sndMsgCounter++
	if self.targetSends == 0 {
		time.Sleep(time.Minute)
	}
	var m map[uint8][]byte
	buf := make([]byte, self.bufferSize)
	sl := buf
	for len(sl) > 0 && self.sendCounter < self.targetSends {
		binary.BigEndian.PutUint32(sl, self.sendCounter)
		self.sendCounter++
		sl = sl[4:]
	}

	if len(sl) > 0 {
		buf = buf[:len(buf)-len(sl)]
	}

	if self.sndMsgCounter%10 == 0 {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, self.sndMsgCounter)
		m = map[uint8][]byte{
			5: b,
		}
		if self.sndMsgCounter%20 == 0 {
			m[10] = []byte("hello")
		}
	}

	if self.sendCounter >= self.targetSends {
		//fmt.Printf("sending final %d bytes\n", len(buf))
		return buf, nil, io.EOF
	}

	//fmt.Printf("sending %d bytes\n", len(buf))

	return buf, m, nil
}

func (self *testXgConn) WritePayload(buf []byte, m map[uint8][]byte) (int, error) {
	logtrace.LogWithFunctionName()
	self.rcvMsgCounter++
	sl := buf
	for len(sl) > 0 {
		next := binary.BigEndian.Uint32(sl)
		sl = sl[4:]
		if next != self.recvCounter {
			select {
			case self.errs <- fmt.Errorf("expected counter %d, got %d, buf: %d", self.recvCounter, next, self.bufCounter):
			default:
			}
		}
		self.recvCounter++
		if self.recvCounter == self.targetReceives {
			close(self.done)
		} else if self.recvCounter > self.targetReceives {
			select {
			case self.errs <- fmt.Errorf("exceeded expected counter %d, got %d, buf: %d", self.targetReceives, self.recvCounter, self.bufCounter):
			default:
			}
		}
	}

	if self.rcvMsgCounter%10 == 0 {
		b, ok := m[5]
		if !ok {
			select {
			case self.errs <- fmt.Errorf("expected header 5, got %+v headers, rcv count: %d", m, self.rcvMsgCounter):
			default:
			}
		} else if len(b) != 4 {
			select {
			case self.errs <- fmt.Errorf("expected header 5, len 4, got %+v, rcv count: %d", b, self.rcvMsgCounter):
			default:
			}
		} else {
			v := binary.BigEndian.Uint32(b)
			if v != self.rcvMsgCounter {
				select {
				case self.errs <- fmt.Errorf("expected header counter %d, got %d", self.rcvMsgCounter, v):
				default:
				}
			}
		}
		if self.rcvMsgCounter%20 == 0 {
			if string(m[10]) != "hello" {
				select {
				case self.errs <- fmt.Errorf("missing 10:hello in map, counter %d", self.recvCounter):
				default:
				}
			}
		}
	}

	//fmt.Printf("received %d bytes\n", len(buf))
	self.bufCounter++

	return len(buf), nil
}

func (self *testXgConn) HandleControlMsg(controlType ControlType, headers channel.Headers, responder ControlReceiver) error {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

type testIntermediary struct {
	circuitId          string
	dest               *Xgress
	msgs               channel.MessageStrategy
	payloadTransformer PayloadTransformer
	counter            uint64
	bytesCallback      func([]byte)
}

func (self *testIntermediary) HandleXgressReceive(payload *Payload, x *Xgress) {
	logtrace.LogWithFunctionName()
	m := payload.Marshall()
	self.payloadTransformer.Tx(m, nil)
	b, err := self.msgs.GetMarshaller()(m)
	if err != nil {
		panic(err)
	}

	if self.bytesCallback != nil {
		self.bytesCallback(b)
	}

	m, err = self.msgs.GetPacketProducer()(b)
	if err != nil {
		logrus.WithError(err).Error("error get next msg")
		panic(err)
	}

	if err = self.validateMessage(m); err != nil {
		panic(err)
	}

	payload, err = UnmarshallPayload(m)
	if err != nil {
		panic(err)
	}

	if err = self.dest.SendPayload(payload, 0, PayloadTypeXg); err != nil {
		panic(err)
	}
	//fmt.Printf("transmitted payload %d from %s -> %s\n", payload.Sequence, x.address, self.dest.address)
}

func (self *testIntermediary) validateMessage(m *channel.Message) error {
	logtrace.LogWithFunctionName()
	circuitId, found := m.GetStringHeader(HeaderKeyCircuitId)
	if !found {
		return errors.New("no circuit id found")
	}

	if circuitId != self.circuitId {
		return fmt.Errorf("expected circuit id %s, got %s", self.circuitId, circuitId)
	}

	seq, found := m.GetUint64Header(HeaderKeySequence)
	if !found {
		return errors.New("no sequence found")
	}
	if seq != self.counter {
		return fmt.Errorf("expected sequence %d, got %d", self.counter, seq)
	}
	self.counter++

	return nil
}

func (self *testIntermediary) HandleControlReceive(control *Control, x *Xgress) {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

type testAcker struct {
	destinations cmap.ConcurrentMap[string, *Xgress]
}

func (self *testAcker) ack(ack *Acknowledgement, address Address) {
	logtrace.LogWithFunctionName()
	dest, _ := self.destinations.Get(string(address))
	if dest != nil {
		if err := dest.SendAcknowledgement(ack); err != nil {
			panic(err)
		}
	} else {
		panic(fmt.Errorf("no xgress found with id %s", string(address)))
	}
}

type mockForwarder struct{}

func (m mockForwarder) RetransmitPayload(srcAddr Address, payload *Payload) error {
	logtrace.LogWithFunctionName()
	return nil
}

func (m mockForwarder) ForwardAcknowledgement(srcAddr Address, acknowledgement *Acknowledgement) error {
	logtrace.LogWithFunctionName()
	return nil
}

type mockFaulter struct{}

func (m mockFaulter) ReportForwardingFault(circuitId string, ctrlId string) {
	logtrace.LogWithFunctionName()
}

func Test_MinimalPayloadMarshalling(t *testing.T) {
	logtrace.LogWithFunctionName()
	logOptions := pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti/").NoColor()
	pfxlog.GlobalInit(logrus.InfoLevel, logOptions)
	pfxlog.SetFormatter(pfxlog.NewFormatter(pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti/").StartingToday()))

	metricsRegistry := metrics.NewRegistry("test", nil)
	InitMetrics(metricsRegistry)

	closeNotify := make(chan struct{})
	defer func() {
		close(closeNotify)
	}()

	InitPayloadIngester(closeNotify)
	InitRetransmitter(mockForwarder{}, mockFaulter{}, metricsRegistry, closeNotify)

	ackHandler := &testAcker{
		destinations: cmap.New[*Xgress](),
	}
	acker = ackHandler
	options := DefaultOptions()
	options.Mtu = 1400

	circuitId := idgen.New()
	srcTestConn := newTestXgConn(10_000, 100_000, 0)
	dstTestConn := newTestXgConn(10_000, 0, 100_000)

	srcXg := NewXgress(circuitId, "ctrl", "src", srcTestConn, Initiator, options, nil)
	dstXg := NewXgress(circuitId, "ctrl", "dst", dstTestConn, Terminator, options, nil)

	ackHandler.destinations.Set("src", dstXg)
	ackHandler.destinations.Set("dst", srcXg)

	msgStrategy := channel.DatagramMessageStrategy(UnmarshallPacketPayload)
	srcXg.receiveHandler = &testIntermediary{
		circuitId: circuitId,
		dest:      dstXg,
		msgs:      msgStrategy,
	}

	dstXg.receiveHandler = &testIntermediary{
		circuitId: circuitId,
		dest:      srcXg,
		msgs:      msgStrategy,
	}

	srcXg.Start()
	dstXg.Start()

	select {
	case <-dstTestConn.done:
	case err := <-dstTestConn.errs:
		t.Fatal(err)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func Test_PayloadSize(t *testing.T) {
	logtrace.LogWithFunctionName()
	logOptions := pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti/").NoColor()
	pfxlog.GlobalInit(logrus.InfoLevel, logOptions)
	pfxlog.SetFormatter(pfxlog.NewFormatter(pfxlog.DefaultOptions().SetTrimPrefix("github.com/openziti/").StartingToday()))

	metricsRegistry := metrics.NewRegistry("test", nil)
	InitMetrics(metricsRegistry)

	closeNotify := make(chan struct{})
	defer func() {
		close(closeNotify)
	}()

	InitPayloadIngester(closeNotify)
	InitRetransmitter(mockForwarder{}, mockFaulter{}, metricsRegistry, closeNotify)

	ackHandler := &testAcker{
		destinations: cmap.New[*Xgress](),
	}
	acker = ackHandler
	options := DefaultOptions()
	//options.Mtu = 1435

	h := metricsRegistry.Histogram("msg_size")

	circuitId := idgen.New()
	srcTestConn := newTestXgConn(200, 100_000, 0)
	dstTestConn := newTestXgConn(200, 0, 100_000)

	srcXg := NewXgress(circuitId, "ctrl", "src", srcTestConn, Initiator, options, nil)
	dstXg := NewXgress(circuitId, "ctrl", "dst", dstTestConn, Terminator, options, nil)

	ackHandler.destinations.Set("src", dstXg)
	ackHandler.destinations.Set("dst", srcXg)

	msgStrategy := channel.DatagramMessageStrategy(UnmarshallPacketPayload)
	srcXg.receiveHandler = &testIntermediary{
		circuitId: circuitId,
		dest:      dstXg,
		msgs:      msgStrategy,
		bytesCallback: func(bytes []byte) {
			h.Update(int64(len(bytes)))
		},
	}

	dstXg.receiveHandler = &testIntermediary{
		circuitId: circuitId,
		dest:      srcXg,
		msgs:      msgStrategy,
	}

	srcXg.Start()
	dstXg.Start()

	select {
	case <-dstTestConn.done:
	case err := <-dstTestConn.errs:
		t.Fatal(err)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	fmt.Printf("max msg size: %d\n", h.(metrics2.Histogram).Max())
}
