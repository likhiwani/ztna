package handler_ctrl

import (
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/common/trace"
	logtrace "ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	trace_pb "github.com/openziti/channel/v3/trace/pb"
	"google.golang.org/protobuf/proto"
)

type traceHandler struct {
	dispatcher trace.EventHandler
}

func newTraceHandler(dispatcher trace.EventHandler) *traceHandler {
	logtrace.LogWithFunctionName()
	return &traceHandler{dispatcher: dispatcher}
}

func (*traceHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(ctrl_pb.ContentType_TraceEventType)
}

func (handler *traceHandler) HandleReceive(msg *channel.Message, _ channel.Channel) {
	logtrace.LogWithFunctionName()
	event := &trace_pb.ChannelMessage{}
	if err := proto.Unmarshal(msg.Body, event); err == nil {
		go handler.dispatcher.Accept(event)
	} else {
		pfxlog.Logger().Errorf("unexpected error decoding trace message (%s)", err)
	}
}
