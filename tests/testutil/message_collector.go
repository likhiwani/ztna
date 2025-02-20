package testutil

import (
	"fmt"
	"time"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func NewMessageCollector(id string) *MessageCollector {
	logtrace.LogWithFunctionName()
	return &MessageCollector{
		id:       id,
		msgs:     make(chan *channel.Message, 16),
		decoders: []channel.TraceMessageDecoder{channel.Decoder{}, ctrl_pb.Decoder{}},
	}
}

type MessageCollector struct {
	id       string
	msgs     chan *channel.Message
	decoders []channel.TraceMessageDecoder
}

func (self *MessageCollector) HandleReceive(m *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	if m.ContentType == -33 || m.ContentType == 5 {
		logrus.Debug("ignoring heartbeats and reconnect ping")
		return
	}
	select {
	case self.msgs <- m:
		decoded := fmt.Sprintf("ContentType: %v", m.ContentType)
		for _, decoder := range self.decoders {
			if val, ok := decoder.Decode(m); ok {
				decoded += ", decoded=" + string(val)
				break
			}
		}
		logrus.Infof("%v: received %v", self.id, decoded)
	case <-time.After(time.Second * 5):
		logrus.Error("timed out trying to queue message, closing channel")
		_ = ch.Close()
	}
}

func (self *MessageCollector) Next(timeout time.Duration) (*channel.Message, error) {
	logtrace.LogWithFunctionName()
	select {
	case msg := <-self.msgs:
		return msg, nil
	case <-time.After(timeout):
		return nil, errors.New("timed out")
	}
}

func (self *MessageCollector) NoMessages(timeout time.Duration, req require.Assertions) {
	logtrace.LogWithFunctionName()
	select {
	case msg := <-self.msgs:
		req.Nil(msg)
	case <-time.After(timeout):
	}
}
