package handler_common

import (
	"time"
	"ztna-core/ztna/common/ctrl_msg"
	logtrace "ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
)

func SendSuccess(request *channel.Message, ch channel.Channel, message string) {
	logtrace.LogWithFunctionName()
	SendResult(request, ch, message, true)
}

func SendFailure(request *channel.Message, ch channel.Channel, message string) {
	logtrace.LogWithFunctionName()
	SendResult(request, ch, message, false)
}

func SendResult(request *channel.Message, ch channel.Channel, message string, success bool) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label())
	if !success {
		log.Errorf("%v error (%s)", ch.LogicalName(), message)
	}

	response := channel.NewResult(success, message)
	response.ReplyTo(request)
	if err := response.WithTimeout(5 * time.Second).SendAndWaitForWire(ch); err != nil {
		log.WithError(err).Error("failed to send result")
	}
}

func SendOpResult(request *channel.Message, ch channel.Channel, op string, message string, success bool) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label()).WithField("operation", op)
	if !success {
		log.Errorf("%v error performing %v: (%s)", ch.LogicalName(), op, message)
	}

	response := channel.NewResult(success, message)
	response.ReplyTo(request)
	if err := response.WithTimeout(5 * time.Second).SendAndWaitForWire(ch); err != nil {
		log.WithError(err).Error("failed to send result")
	}
}

func SendServerBusy(request *channel.Message, ch channel.Channel, op string) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label()).WithField("operation", op)
	log.Errorf("%v error performing %v: (%s)", ch.LogicalName(), op, "server too busy")

	response := channel.NewResult(false, "server too busy")
	response.ReplyTo(request)
	response.Headers.PutUint32Header(ctrl_msg.HeaderResultErrorCode, ctrl_msg.ResultErrorRateLimited)
	if err := response.WithTimeout(5 * time.Second).SendAndWaitForWire(ch); err != nil {
		log.WithError(err).Error("failed to send result")
	}
}

func WasRateLimited(msg *channel.Message) bool {
	logtrace.LogWithFunctionName()
	val, found := msg.GetUint32Header(ctrl_msg.HeaderResultErrorCode)
	return found && val == ctrl_msg.ResultErrorRateLimited
}
