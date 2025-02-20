package xlink_transport

import (
	"crypto/x509"
	"fmt"
	"net"
	"testing"
	"time"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
	"github.com/openziti/metrics"
	"github.com/stretchr/testify/assert"
)

type testUnderlay struct {
}

func (t testUnderlay) Rx() (*channel.Message, error) {
	logtrace.LogWithFunctionName()
	time.Sleep(time.Hour)
	return nil, nil
}

func (t testUnderlay) Tx(*channel.Message) error {
	logtrace.LogWithFunctionName()
	time.Sleep(10 * time.Microsecond)
	return nil
}

func (t testUnderlay) Id() string {
	logtrace.LogWithFunctionName()
	return "test"
}

func (t testUnderlay) LogicalName() string {
	logtrace.LogWithFunctionName()
	return "test"
}

func (t testUnderlay) ConnectionId() string {
	logtrace.LogWithFunctionName()
	return "test"
}

func (t testUnderlay) Certificates() []*x509.Certificate {
	logtrace.LogWithFunctionName()
	return nil
}

func (t testUnderlay) Label() string {
	logtrace.LogWithFunctionName()
	return "test"
}

func (t testUnderlay) Close() error {
	logtrace.LogWithFunctionName()
	return nil
}

func (t testUnderlay) IsClosed() bool {
	logtrace.LogWithFunctionName()
	return false
}

func (t testUnderlay) Headers() map[int32][]byte {
	logtrace.LogWithFunctionName()
	return nil
}

func (t testUnderlay) SetWriteTimeout(time.Duration) error {
	logtrace.LogWithFunctionName()
	return nil
}

func (t testUnderlay) SetWriteDeadline(time.Time) error {
	logtrace.LogWithFunctionName()
	return nil
}

func (t testUnderlay) GetLocalAddr() net.Addr {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

func (t testUnderlay) GetRemoteAddr() net.Addr {
	logtrace.LogWithFunctionName()
	panic("implement me")
}

type testUnderlayFactory struct {
	underlay testUnderlay
}

func (t testUnderlayFactory) Create(time.Duration) (channel.Underlay, error) {
	logtrace.LogWithFunctionName()
	return t.underlay, nil
}

func Test_Throughput(t *testing.T) {
	logtrace.LogWithFunctionName()
	t.SkipNow()

	underlayFactory := testUnderlayFactory{
		underlay: testUnderlay{},
	}

	options := channel.DefaultOptions()
	options.OutQueueSize = 64
	ch, err := channel.NewChannel("test", underlayFactory, nil, options)
	assert.NoError(t, err)

	registry := metrics.NewRegistry("test", nil)
	drops := registry.Meter("drops")
	msgs := registry.Meter("msgs")

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			v := registry.Poll().Meters["drops"]
			fmt.Printf("drops - m1: %v, count: %v\n", v.M1Rate, v.Count)
			v = registry.Poll().Meters["msgs"]
			fmt.Printf("msgs  - m1: %v, count: %v\n", v.M1Rate, v.Count)
		}
	}()

	go func() {
		for {
			m := channel.NewMessage(1, nil)
			sent, err := ch.TrySend(m)
			assert.NoError(t, err)
			if !sent {
				drops.Mark(1)
			}
			msgs.Mark(1)
			time.Sleep(10 * time.Microsecond)
		}
	}()

	time.Sleep(time.Minute)
}
