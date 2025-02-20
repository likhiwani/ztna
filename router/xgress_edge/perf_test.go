package xgress_edge

import (
	"crypto/x509"
	"testing"
	"time"
	"ztna-core/sdk-golang/ziti/edge"
	"ztna-core/ztna/common/inspect"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/forwarder"
	"ztna-core/ztna/router/handler_xgress"
	metrics2 "ztna-core/ztna/router/metrics"
	"ztna-core/ztna/router/xgress"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMirrorLink(fwd *forwarder.Forwarder) *mirrorLink {
	logtrace.LogWithFunctionName()
	result := &mirrorLink{
		fwd:  fwd,
		acks: make(chan *xgress.Acknowledgement, 4),
	}
	go result.run()
	return result
}

type mirrorLink struct {
	fwd  *forwarder.Forwarder
	acks chan *xgress.Acknowledgement
}

func (link *mirrorLink) DialAddress() string {
	logtrace.LogWithFunctionName()
	return "tcp:localhost:1234"
}

func (link *mirrorLink) GetAddresses() []*ctrl_pb.LinkConn {
	logtrace.LogWithFunctionName()
	return nil
}

func (link *mirrorLink) IsClosed() bool {
	logtrace.LogWithFunctionName()
	return false
}

func (link *mirrorLink) InspectCircuit(circuitDetail *inspect.CircuitInspectDetail) {
	logtrace.LogWithFunctionName()
}

func (link *mirrorLink) InspectLink() *inspect.LinkInspectDetail {
	logtrace.LogWithFunctionName()
	return nil
}

func (link *mirrorLink) CloseNotified() error {
	logtrace.LogWithFunctionName()
	return nil
}

func (link *mirrorLink) DestVersion() string {
	logtrace.LogWithFunctionName()
	return "0.0.0"
}

func (link *mirrorLink) LinkProtocol() string {
	logtrace.LogWithFunctionName()
	return "tls"
}

func (link *mirrorLink) HandleCloseNotification(f func()) {
	logtrace.LogWithFunctionName()
	f()
}

func (link *mirrorLink) DestinationId() string {
	logtrace.LogWithFunctionName()
	return "test"
}

func (link *mirrorLink) Id() string {
	logtrace.LogWithFunctionName()
	return "router1"
}

func (link *mirrorLink) SendPayload(payload *xgress.Payload, _ time.Duration, _ xgress.PayloadType) error {
	logtrace.LogWithFunctionName()
	ack := &xgress.Acknowledgement{
		CircuitId:      "test",
		Flags:          0,
		RecvBufferSize: 0,
		RTT:            payload.RTT,
	}
	ack.Sequence = append(ack.Sequence, payload.Sequence)
	link.acks <- ack
	return nil
}

func (link *mirrorLink) run() {
	logtrace.LogWithFunctionName()
	for ack := range link.acks {
		err := link.fwd.ForwardAcknowledgement("router1", ack)
		if err != nil {
			pfxlog.Logger().WithError(err).Infof("unable to forward ack")
		}
	}
}

func (link *mirrorLink) SendAcknowledgement(*xgress.Acknowledgement) error {
	logtrace.LogWithFunctionName()
	return nil
}

func (link *mirrorLink) SendControl(*xgress.Control) error {
	logtrace.LogWithFunctionName()
	return nil
}

func (link *mirrorLink) Close() error {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func Benchmark_CowMapWritePerf(b *testing.B) {
	logtrace.LogWithFunctionName()
	mux := edge.NewCowMapMsgMux()
	writePerf(b, mux)
}

func writePerf(b *testing.B, mux edge.MsgMux) {
	logtrace.LogWithFunctionName()
	testChannel := &NoopTestChannel{}

	listener := &listener{}

	proxy := &edgeClientConn{
		msgMux:       mux,
		listener:     listener,
		fingerprints: nil,
		ch:           testChannel,
	}

	conn := &edgeXgressConn{
		MsgChannel: *edge.NewEdgeMsgChannel(proxy.ch, 1),
		seq:        NewMsgQueue(4),
	}

	req := require.New(b)
	req.NoError(mux.AddMsgSink(conn))

	metricsRegistry := metrics.NewUsageRegistry("test", map[string]string{}, nil)
	xgress.InitMetrics(metricsRegistry)

	fwdOptions := forwarder.DefaultOptions()
	fwd := forwarder.NewForwarder(metricsRegistry, nil, fwdOptions, nil)

	link := newMirrorLink(fwd)

	err := fwd.RegisterLink(link)
	assert.NoError(b, err)

	err = fwd.Route("test", &ctrl_pb.Route{
		CircuitId: "test",
		Egress:    nil,
		Forwards: []*ctrl_pb.Route_Forward{
			{SrcAddress: "test", DstAddress: "router1"},
			{SrcAddress: "router1", DstAddress: "test"},
		},
	})
	assert.NoError(b, err)

	x := xgress.NewXgress("test", "test", "test", conn, xgress.Initiator, xgress.DefaultOptions(), nil)
	x.SetReceiveHandler(handler_xgress.NewReceiveHandler(fwd))
	x.AddPeekHandler(metrics2.NewXgressPeekHandler(fwd.MetricsRegistry()))

	//x.SetCloseHandler(bindHandler.closeHandler)
	fwd.RegisterDestination(x.CircuitId(), x.Address(), x)

	x.Start()

	b.ReportAllocs()
	b.ResetTimer()

	data := make([]byte, 1024)

	for i := 0; i < b.N; i++ {
		msg := edge.NewDataMsg(1, uint32(i+1), data)
		mux.HandleReceive(msg, testChannel)
		b.SetBytes(1024)
	}
}

type simpleTestXgConn struct {
	ch chan []byte
}

func (conn *simpleTestXgConn) write(data []byte) {
	logtrace.LogWithFunctionName()
	conn.ch <- data
}

func (conn *simpleTestXgConn) Close() error {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (conn *simpleTestXgConn) LogContext() string {
	logtrace.LogWithFunctionName()
	return "test"
}

func (conn *simpleTestXgConn) ReadPayload() ([]byte, map[uint8][]byte, error) {
	logtrace.LogWithFunctionName()
	result := <-conn.ch
	return result, nil, nil
}

func (conn *simpleTestXgConn) WritePayload([]byte, map[uint8][]byte) (int, error) {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (conn *simpleTestXgConn) HandleControlMsg(xgress.ControlType, channel.Headers, xgress.ControlReceiver) error {
	logtrace.LogWithFunctionName()
	return nil
}

func Benchmark_BaselinePerf(b *testing.B) {
	logtrace.LogWithFunctionName()
	conn := &simpleTestXgConn{
		ch: make(chan []byte),
	}
	xgOptions := xgress.DefaultOptions()

	metricsRegistry := metrics.NewUsageRegistry("test", map[string]string{}, nil)
	xgress.InitMetrics(metricsRegistry)

	fwdOptions := forwarder.DefaultOptions()
	fwd := forwarder.NewForwarder(metricsRegistry, nil, fwdOptions, nil)

	link := newMirrorLink(fwd)

	err := fwd.RegisterLink(link)
	assert.NoError(b, err)

	err = fwd.Route("test", &ctrl_pb.Route{
		CircuitId: "test",
		Egress:    nil,
		Forwards: []*ctrl_pb.Route_Forward{
			{SrcAddress: "test", DstAddress: "router1"},
			{SrcAddress: "router1", DstAddress: "test"},
		},
	})
	assert.NoError(b, err)

	x := xgress.NewXgress("test", "test", "test", conn, xgress.Initiator, xgOptions, nil)
	x.SetReceiveHandler(handler_xgress.NewReceiveHandler(fwd))
	x.AddPeekHandler(metrics2.NewXgressPeekHandler(fwd.MetricsRegistry()))

	//x.SetCloseHandler(bindHandler.closeHandler)
	fwd.RegisterDestination(x.CircuitId(), x.Address(), x)

	x.Start()

	b.ReportAllocs()
	b.ResetTimer()

	data := make([]byte, 1024)

	for i := 0; i < b.N; i++ {
		conn.write(data)
		b.SetBytes(1024)
	}
}

type NoopTestChannel struct {
}

func (ch *NoopTestChannel) Underlay() channel.Underlay {
	logtrace.LogWithFunctionName()
	//TODO implement me
	panic("implement me")
}

func (ch *NoopTestChannel) StartRx() {
	logtrace.LogWithFunctionName()
}

func (ch *NoopTestChannel) Id() string {
	logtrace.LogWithFunctionName()
	panic("implement Id()")
}

func (ch *NoopTestChannel) LogicalName() string {
	logtrace.LogWithFunctionName()
	panic("implement LogicalName()")
}

func (ch *NoopTestChannel) ConnectionId() string {
	logtrace.LogWithFunctionName()
	panic("implement ConnectionId()")
}

func (ch *NoopTestChannel) Certificates() []*x509.Certificate {
	logtrace.LogWithFunctionName()
	panic("implement Certificates()")
}

func (ch *NoopTestChannel) Label() string {
	logtrace.LogWithFunctionName()
	return "testchannel"
}

func (ch *NoopTestChannel) SetLogicalName(string) {
	logtrace.LogWithFunctionName()
	panic("implement SetLogicalName")
}

func (ch *NoopTestChannel) TrySend(channel.Sendable) (bool, error) {
	logtrace.LogWithFunctionName()
	return true, nil
}

func (ch *NoopTestChannel) Send(channel.Sendable) error {
	logtrace.LogWithFunctionName()
	return nil
}

func (ch *NoopTestChannel) Close() error {
	logtrace.LogWithFunctionName()
	panic("implement Close")
}

func (ch *NoopTestChannel) IsClosed() bool {
	logtrace.LogWithFunctionName()
	panic("implement IsClosed")
}

func (ch *NoopTestChannel) GetTimeSinceLastRead() time.Duration {
	logtrace.LogWithFunctionName()
	return 0
}
