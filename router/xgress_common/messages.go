package xgress_common

import (
	"ztna-core/ztna/common/pb/edge_ctrl_pb"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
	"github.com/openziti/channel/v3/protobufs"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

const (
	PayloadFlagsHeader uint8 = 0x10
)

func GetResultOrFailure(msg *channel.Message, err error, result protobufs.TypedMessage) error {
	logtrace.LogWithFunctionName()
	if err != nil {
		return err
	}

	if msg.ContentType == int32(edge_ctrl_pb.ContentType_ErrorType) {
		msg := string(msg.Body)
		if msg == "" {
			msg = "error state returned from controller with no message"
		}
		return errors.New(msg)
	}

	if msg.ContentType != result.GetContentType() {
		return errors.Errorf("unexpected response type %v to request. expected %v", msg.ContentType, result.GetContentType())
	}

	return proto.Unmarshal(msg.Body, result)
}

func CheckForFailureResult(msg *channel.Message, err error, successType edge_ctrl_pb.ContentType) error {
	logtrace.LogWithFunctionName()
	if err != nil {
		return err
	}

	if msg.ContentType == int32(edge_ctrl_pb.ContentType_ErrorType) {
		msg := string(msg.Body)
		if msg == "" {
			msg = "error state returned from controller with no message"
		}
		return errors.New(msg)
	}

	if msg.ContentType != int32(successType) {
		return errors.Errorf("unexpected response type %v to request. expected %v", msg.ContentType, successType)
	}

	return nil
}

func GetFinHeaders() map[uint8][]byte {
	logtrace.LogWithFunctionName()
	return map[uint8][]byte{
		PayloadFlagsHeader: {0x1, 0, 0, 0},
	}
}
