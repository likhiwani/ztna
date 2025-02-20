package xgress

import (
	"sync/atomic"
	"ztna-core/ztna/logtrace"

	"github.com/ef-ds/deque"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/metrics"
)

var acker ackSender

type ackSender interface {
	ack(ack *Acknowledgement, address Address)
}

func InitAcker(forwarder PayloadBufferForwarder, metrics metrics.Registry, closeNotify <-chan struct{}) {
	logtrace.LogWithFunctionName()
	acker = NewAcker(forwarder, metrics, closeNotify)
}

type ackEntry struct {
	Address
	*Acknowledgement
}

// Note: if altering this struct, be sure to account for 64 bit alignment on 32 bit arm arch
// https://pkg.go.dev/sync/atomic#pkg-note-BUG
// https://github.com/golang/go/issues/36606
type Acker struct {
	acksQueueSize int64
	forwarder     PayloadBufferForwarder
	acks          *deque.Deque
	ackIngest     chan *ackEntry
	ackSend       chan *ackEntry
	closeNotify   <-chan struct{}
}

func NewAcker(forwarder PayloadBufferForwarder, metrics metrics.Registry, closeNotify <-chan struct{}) *Acker {
	logtrace.LogWithFunctionName()
	result := &Acker{
		forwarder:   forwarder,
		acks:        deque.New(),
		ackIngest:   make(chan *ackEntry, 16),
		ackSend:     make(chan *ackEntry, 1),
		closeNotify: closeNotify,
	}

	go result.ackIngester()
	go result.ackSender()

	metrics.FuncGauge("xgress.acks.queue_size", func() int64 {
		return atomic.LoadInt64(&result.acksQueueSize)
	})

	return result
}

func (acker *Acker) ack(ack *Acknowledgement, address Address) {
	logtrace.LogWithFunctionName()
	acker.ackIngest <- &ackEntry{
		Acknowledgement: ack,
		Address:         address,
	}
}

func (acker *Acker) ackIngester() {
	logtrace.LogWithFunctionName()
	var next *ackEntry
	for {
		if next == nil {
			if val, _ := acker.acks.PopFront(); val != nil {
				next = val.(*ackEntry)
			}
		}

		if next == nil {
			select {
			case ack := <-acker.ackIngest:
				acker.acks.PushBack(ack)
			case <-acker.closeNotify:
				return
			}
		} else {
			select {
			case ack := <-acker.ackIngest:
				acker.acks.PushBack(ack)
			case acker.ackSend <- next:
				next = nil
			case <-acker.closeNotify:
				return
			}
		}
		atomic.StoreInt64(&acker.acksQueueSize, int64(acker.acks.Len()))
	}
}

func (acker *Acker) ackSender() {
	logtrace.LogWithFunctionName()
	logger := pfxlog.Logger()
	for {
		select {
		case nextAck := <-acker.ackSend:
			if err := acker.forwarder.ForwardAcknowledgement(nextAck.Address, nextAck.Acknowledgement); err != nil {
				logger.WithError(err).Debugf("unexpected error while sending ack from %v", nextAck.Address)
				ackFailures.Mark(1)
			} else {
				ackTxMeter.Mark(1)
			}
		case <-acker.closeNotify:
			return
		}
	}
}
