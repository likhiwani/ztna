package cmd_pb

import (
	"encoding/binary"
	logtrace "ztna-core/ztna/logtrace"

	"google.golang.org/protobuf/proto"
)

func (request *AddPeerRequest) GetContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(ContentType_AddPeerRequestType)
}

func (request *RemovePeerRequest) GetContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(ContentType_RemovePeerRequestType)
}

func (request *TransferLeadershipRequest) GetContentType() int32 {
	logtrace.LogWithFunctionName()
	return int32(ContentType_TransferLeadershipRequestType)
}

// TypedMessage instances are protobuf messages which know their command type
type TypedMessage interface {
	proto.Message
	GetCommandType() int32
}

// EncodeProtobuf returns the encoded message, prefixed with the command type
func EncodeProtobuf(v TypedMessage) ([]byte, error) {
	logtrace.LogWithFunctionName()
	b, err := proto.Marshal(v)
	if err != nil {
		return nil, err
	}
	result := make([]byte, len(b)+4)
	binary.BigEndian.PutUint32(result, uint32(v.GetCommandType()))
	copy(result[4:], b)
	return result, nil
}

func (x *CreateEntityCommand) GetCommandType() int32 {
	logtrace.LogWithFunctionName()
	return int32(CommandType_CreateEntityType)
}

func (x *UpdateEntityCommand) GetCommandType() int32 {
	logtrace.LogWithFunctionName()
	return int32(CommandType_UpdateEntityType)
}

func (x *DeleteEntityCommand) GetCommandType() int32 {
	logtrace.LogWithFunctionName()
	return int32(CommandType_DeleteEntityType)
}

func (x *DeleteTerminatorsBatchCommand) GetCommandType() int32 {
	logtrace.LogWithFunctionName()
	return int32(CommandType_DeleteTerminatorsBatchType)
}

func (x *SyncSnapshotCommand) GetCommandType() int32 {
	logtrace.LogWithFunctionName()
	return int32(CommandType_SyncSnapshot)
}

func (x *InitClusterIdCommand) GetCommandType() int32 {
	logtrace.LogWithFunctionName()
	return int32(CommandType_InitClusterId)
}

func EncodeTags(tags map[string]interface{}) (map[string]*TagValue, error) {
	logtrace.LogWithFunctionName()
	if len(tags) == 0 {
		return nil, nil
	}
	result := map[string]*TagValue{}

	for k, v := range tags {
		if v == nil {
			result[k] = &TagValue{
				Value: &TagValue_NilValue{
					NilValue: true,
				},
			}
		} else {
			switch val := v.(type) {
			case string:
				result[k] = &TagValue{
					Value: &TagValue_StringValue{
						StringValue: val,
					},
				}
			case bool:
				result[k] = &TagValue{
					Value: &TagValue_BoolValue{
						BoolValue: val,
					},
				}
			case float64:
				result[k] = &TagValue{
					Value: &TagValue_FpValue{
						FpValue: val,
					},
				}
			}
		}
	}
	return result, nil
}

func DecodeTags(tags map[string]*TagValue) map[string]interface{} {
	logtrace.LogWithFunctionName()
	if len(tags) == 0 {
		return nil
	}
	result := map[string]interface{}{}

	for k, v := range tags {
		switch v.Value.(type) {
		case *TagValue_NilValue:
			result[k] = nil
		case *TagValue_BoolValue:
			result[k] = v.GetBoolValue()
		case *TagValue_StringValue:
			result[k] = v.GetStringValue()
		case *TagValue_FpValue:
			result[k] = v.GetFpValue()
		}
	}

	return result
}
