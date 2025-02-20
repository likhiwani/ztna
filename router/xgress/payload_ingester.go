package xgress

import (
	"time"
	"ztna-core/ztna/logtrace"
)

var payloadIngester *PayloadIngester

func InitPayloadIngester(closeNotify <-chan struct{}) {
	logtrace.LogWithFunctionName()
	payloadIngester = NewPayloadIngester(closeNotify)
}

type payloadEntry struct {
	payload *Payload
	x       *Xgress
}

type PayloadIngester struct {
	payloadIngest         chan *payloadEntry
	payloadSendReq        chan *Xgress
	receiveBufferInspects chan *receiveBufferInspectEvent
	closeNotify           <-chan struct{}
}

func NewPayloadIngester(closeNotify <-chan struct{}) *PayloadIngester {
	logtrace.LogWithFunctionName()
	pi := &PayloadIngester{
		payloadIngest:         make(chan *payloadEntry, 16),
		payloadSendReq:        make(chan *Xgress, 16),
		receiveBufferInspects: make(chan *receiveBufferInspectEvent, 4),
		closeNotify:           closeNotify,
	}

	go pi.run()

	return pi
}

func (self *PayloadIngester) inspect(evt *receiveBufferInspectEvent, timeout <-chan time.Time) bool {
	logtrace.LogWithFunctionName()
	select {
	case self.receiveBufferInspects <- evt:
		return true
	case <-self.closeNotify:
	case <-timeout:
	}
	return false
}

func (self *PayloadIngester) ingest(payload *Payload, x *Xgress) {
	logtrace.LogWithFunctionName()
	self.payloadIngest <- &payloadEntry{
		payload: payload,
		x:       x,
	}
}

func (self *PayloadIngester) run() {
	logtrace.LogWithFunctionName()
	for {
		select {
		case payloadEntry := <-self.payloadIngest:
			payloadEntry.x.payloadIngester(payloadEntry.payload)
		case x := <-self.payloadSendReq:
			x.queueSends()
		case evt := <-self.receiveBufferInspects:
			evt.handle()
		case <-self.closeNotify:
			return
		}
	}
}
