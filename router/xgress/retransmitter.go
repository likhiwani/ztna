package xgress

import (
	"sync/atomic"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/metrics"
)

var retransmitter *Retransmitter

func InitRetransmitter(forwarder PayloadBufferForwarder, faultReporter RetransmitterFaultReporter, metrics metrics.Registry, closeNotify <-chan struct{}) {
	logtrace.LogWithFunctionName()
	retransmitter = NewRetransmitter(forwarder, faultReporter, metrics, closeNotify)
}

type RetransmitterFaultReporter interface {
	ReportForwardingFault(circuitId string, ctrlId string)
}

type Retransmitter struct {
	forwarder            PayloadBufferForwarder
	faultReporter        RetransmitterFaultReporter
	retxTail             *txPayload
	retxHead             *txPayload
	retransmitIngest     chan *txPayload
	retransmitSend       chan *txPayload
	retransmitsQueueSize int64
	closeNotify          <-chan struct{}
}

func NewRetransmitter(forwarder PayloadBufferForwarder, faultReporter RetransmitterFaultReporter, metrics metrics.Registry, closeNotify <-chan struct{}) *Retransmitter {
	logtrace.LogWithFunctionName()
	ctrl := &Retransmitter{
		forwarder:        forwarder,
		retransmitIngest: make(chan *txPayload, 16),
		retransmitSend:   make(chan *txPayload, 1),
		closeNotify:      closeNotify,
		faultReporter:    faultReporter,
	}

	go ctrl.retransmitIngester()
	go ctrl.retransmitSender()

	metrics.FuncGauge("xgress.retransmits.queue_size", func() int64 {
		return atomic.LoadInt64(&ctrl.retransmitsQueueSize)
	})

	return ctrl
}

func (self *Retransmitter) queue(p *txPayload) {
	logtrace.LogWithFunctionName()
	self.retransmitIngest <- p
}

func (self *Retransmitter) popHead() *txPayload {
	logtrace.LogWithFunctionName()
	if self.retxHead == nil {
		return nil
	}

	result := self.retxHead
	if result.prev == nil {
		self.retxHead = nil
		self.retxTail = nil
	} else {
		self.retxHead = result.prev
		result.prev.next = nil
	}

	result.prev = nil
	result.next = nil

	atomic.AddInt64(&self.retransmitsQueueSize, -1)

	return result
}

func (self *Retransmitter) pushTail(txp *txPayload) {
	logtrace.LogWithFunctionName()
	if txp.prev != nil || txp.next != nil || txp == self.retxHead {
		return
	}
	if self.retxHead == nil {
		self.retxTail = txp
		self.retxHead = txp
	} else {
		txp.next = self.retxTail
		self.retxTail.prev = txp
		self.retxTail = txp
	}
	atomic.AddInt64(&self.retransmitsQueueSize, 1)
}

func (self *Retransmitter) delete(txp *txPayload) {
	logtrace.LogWithFunctionName()
	if self.retxHead == txp {
		self.popHead()
	} else if txp == self.retxTail {
		self.retxTail = txp.next
		self.retxTail.prev = nil
		atomic.AddInt64(&self.retransmitsQueueSize, -1)
	} else if txp.prev != nil {
		txp.prev.next = txp.next
		txp.next.prev = txp.prev
		atomic.AddInt64(&self.retransmitsQueueSize, -1)
	}

	txp.prev = nil
	txp.next = nil
}

func (self *Retransmitter) retransmitIngester() {
	logtrace.LogWithFunctionName()
	var next *txPayload
	for {
		if next == nil {
			next = self.popHead()
		}

		if next == nil {
			select {
			case retransmit := <-self.retransmitIngest:
				self.acceptRetransmit(retransmit)
			case <-self.closeNotify:
				return
			}
		} else {
			select {
			case retransmit := <-self.retransmitIngest:
				self.acceptRetransmit(retransmit)
			case self.retransmitSend <- next:
				next = nil
			case <-self.closeNotify:
				return
			}
		}
	}
}

func (self *Retransmitter) acceptRetransmit(txp *txPayload) {
	logtrace.LogWithFunctionName()
	if txp.isAcked() {
		self.delete(txp)
	} else {
		self.pushTail(txp)
	}
}

func (self *Retransmitter) retransmitSender() {
	logtrace.LogWithFunctionName()
	logger := pfxlog.Logger()
	for {
		select {
		case retransmit := <-self.retransmitSend:
			if !retransmit.isAcked() {
				if err := self.forwarder.RetransmitPayload(retransmit.x.address, retransmit.payload); err != nil {
					// if xgress is closed, don't log the error. We still want to try retransmitting in case we're re-sending end of circuit
					if !retransmit.x.Closed() {
						logger.WithError(err).Errorf("unexpected error while retransmitting payload from [@/%v]", retransmit.x.address)
						retransmissionFailures.Mark(1)
						self.faultReporter.ReportForwardingFault(retransmit.payload.CircuitId, retransmit.x.ctrlId)
					} else {
						logger.WithError(err).Tracef("unexpected error while retransmitting payload from [@/%v] (already closed)", retransmit.x.address)
					}
				} else {
					retransmit.markSent()
					retransmissions.Mark(1)
				}
				retransmit.dequeued()
			}
		case <-self.closeNotify:
			return
		}
	}
}
