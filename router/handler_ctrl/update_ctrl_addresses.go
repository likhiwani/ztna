package handler_ctrl

import (
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var updateCtrlAddressesHandlerInstance *updateCtrlAddressesHandler

type CtrlAddressUpdater interface {
	UpdateCtrlEndpoints(endpoints []string)
	UpdateLeader(leaderId string)
}

type updateCtrlAddressesHandler struct {
	callback       CtrlAddressUpdater
	currentVersion uint64
}

func (handler *updateCtrlAddressesHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(ctrl_pb.ContentType_UpdateCtrlAddressesType)
}

func (handler *updateCtrlAddressesHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label()).Entry
	upd := &ctrl_pb.UpdateCtrlAddresses{}
	if err := proto.Unmarshal(msg.Body, upd); err != nil {
		log.WithError(err).Error("error unmarshalling")
		return
	}

	log = log.WithFields(logrus.Fields{
		"endpoints":     upd.Addresses,
		"localVersion":  handler.currentVersion,
		"remoteVersion": upd.Index,
		"isLeader":      upd.IsLeader,
	})

	log.Info("update ctrl endpoints message received")

	if handler.currentVersion == 0 || handler.currentVersion < upd.Index {
		log.Info("updating to newer controller endpoints")
		handler.callback.UpdateCtrlEndpoints(upd.Addresses)
		handler.currentVersion = upd.Index

		if upd.IsLeader {
			handler.callback.UpdateLeader(ch.Id())
		}
	} else {
		log.Info("ignoring outdated controller endpoints")
	}
}

func newUpdateCtrlAddressesHandler(callback CtrlAddressUpdater) channel.TypedReceiveHandler {
	logtrace.LogWithFunctionName()
	if updateCtrlAddressesHandlerInstance == nil {
		updateCtrlAddressesHandlerInstance = &updateCtrlAddressesHandler{
			callback: callback,
		}
	}
	return updateCtrlAddressesHandlerInstance
}
