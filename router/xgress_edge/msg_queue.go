package xgress_edge

import (
	"context"
	"sync/atomic"
	"ztna-core/sdk-golang/ziti/edge"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
	"github.com/openziti/foundation/v2/sequencer"
	"golang.org/x/sync/semaphore"
)

type MsgQueue interface {
	Push(msg *channel.Message) error
	Pop() *channel.Message
	Close()
}

type BoundedMsgQueue interface {
	MsgQueue
	TryPopMax(size int) *channel.Message
}

func NewMsgQueue(queueDepth int) *baseMsgQ {
	logtrace.LogWithFunctionName()
	return &baseMsgQ{
		ch:          make(chan *channel.Message, queueDepth),
		closeNotify: make(chan struct{}),
	}
}

type baseMsgQ struct {
	ch          chan *channel.Message
	closeNotify chan struct{}
	closed      atomic.Bool
}

func (mq *baseMsgQ) Push(msg *channel.Message) error {
	logtrace.LogWithFunctionName()
	select {
	case mq.ch <- msg:
		return nil
	case <-mq.closeNotify:
		return sequencer.ErrClosed
	}
}

func (mq *baseMsgQ) Pop() *channel.Message {
	logtrace.LogWithFunctionName()
	select {
	case msg := <-mq.ch:
		return msg
	case <-mq.closeNotify:
		// If we're closed, return any buffered values, otherwise return nil
		select {
		case msg := <-mq.ch:
			return msg
		default:
			return nil
		}
	}
}

func (mq *baseMsgQ) Close() {
	logtrace.LogWithFunctionName()
	if mq.closed.CompareAndSwap(false, true) {
		close(mq.closeNotify)
	}
}

func NewSizeLimitedMsgQueue(size int32) *sizeLimitedMsgQ {
	logtrace.LogWithFunctionName()
	return &sizeLimitedMsgQ{
		ch:          make(chan *channel.Message, 16),
		closeNotify: make(chan struct{}),
		sizeSem:     *semaphore.NewWeighted(int64(size)),
	}
}

type sizeLimitedMsgQ struct {
	ch          chan *channel.Message
	closeNotify chan struct{}
	closed      atomic.Bool
	sizeSem     semaphore.Weighted
	next        *channel.Message
}

func (mq *sizeLimitedMsgQ) Push(msg *channel.Message) error {
	logtrace.LogWithFunctionName()
	if size := len(msg.Body); size > 0 {
		// TODO: Handle if size > mtu
		if err := mq.sizeSem.Acquire(context.Background(), int64(size)); err != nil {
			return err
		}
	}
	select {
	case mq.ch <- msg:
		return nil
	case <-mq.closeNotify:
		return sequencer.ErrClosed
	}
}

func (mq *sizeLimitedMsgQ) Pop() *channel.Message {
	logtrace.LogWithFunctionName()
	var msg *channel.Message

	if mq.next != nil {
		msg = mq.next
		mq.next = nil
	} else {
		select {
		case msg = <-mq.ch:
		case <-mq.closeNotify:
			// If we're closed, return any buffered values, otherwise return nil
			select {
			case msg = <-mq.ch:
			default:
				return nil
			}
		}
	}

	if msg != nil {
		if size := len(msg.Body); size > 0 {
			mq.sizeSem.Release(int64(size))
		}
	}

	return msg
}

func (mq *sizeLimitedMsgQ) TryPopMax(size int) *channel.Message {
	logtrace.LogWithFunctionName()
	var msg *channel.Message

	if mq.next == nil {
		select {
		case mq.next = <-mq.ch:
		default:
		}
	}

	if mq.next == nil || mq.next.ContentType != edge.ContentTypeData {
		return nil
	}

	nextSize := len(mq.next.Body)
	if nextSize > size {
		return nil
	}

	msg = mq.next
	mq.next = nil

	if nextSize > 0 {
		mq.sizeSem.Release(int64(nextSize))
	}

	return msg
}

func (mq *sizeLimitedMsgQ) Close() {
	logtrace.LogWithFunctionName()
	if mq.closed.CompareAndSwap(false, true) {
		close(mq.closeNotify)
	}
}
