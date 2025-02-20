package handler_ctrl

import (
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var updateClusterLeaderHandlerInstance *updateClusterLeaderHandler

type updateClusterLeaderHandler struct {
	callback       CtrlAddressUpdater
	currentVersion uint64
}

func (handler *updateClusterLeaderHandler) ContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(ctrl_pb.ContentType_UpdateClusterLeaderRequestType)
}

func (handler *updateClusterLeaderHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	log := pfxlog.ContextLogger(ch.Label()).Entry
	upd := &ctrl_pb.UpdateClusterLeader{}
	if err := proto.Unmarshal(msg.Body, upd); err != nil {
		log.WithError(err).Error("error unmarshalling")
		return
	}

	log = log.WithFields(logrus.Fields{
		"localVersion":  handler.currentVersion,
		"remoteVersion": upd.Index,
		"ctrlId":        ch.Id(),
	})

	if handler.currentVersion == 0 || handler.currentVersion < upd.Index {
		log.Info("handling update of cluster leader")
		handler.callback.UpdateLeader(ch.Id())
	} else {
		log.Info("ignoring outdated update cluster leader message")
	}
}

func newUpdateClusterLeaderHandler(callback CtrlAddressUpdater) channel.TypedReceiveHandler {
	logtrace.LogWithFunctionName()
	if updateClusterLeaderHandlerInstance == nil {
		updateClusterLeaderHandlerInstance = &updateClusterLeaderHandler{
			callback: callback,
		}
	}
	return updateClusterLeaderHandlerInstance
}
