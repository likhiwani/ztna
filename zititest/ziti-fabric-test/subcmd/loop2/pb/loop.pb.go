// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: loop.proto

package loop2_pb

import (
	"ztna-core/ztna/logtrace"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Test struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name            string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	TxRequests      int32  `protobuf:"varint,2,opt,name=txRequests,proto3" json:"txRequests,omitempty"`
	TxPacing        int32  `protobuf:"varint,3,opt,name=txPacing,proto3" json:"txPacing,omitempty"`
	TxMaxJitter     int32  `protobuf:"varint,4,opt,name=txMaxJitter,proto3" json:"txMaxJitter,omitempty"`
	RxRequests      int32  `protobuf:"varint,5,opt,name=rxRequests,proto3" json:"rxRequests,omitempty"`
	RxTimeout       int32  `protobuf:"varint,6,opt,name=rxTimeout,proto3" json:"rxTimeout,omitempty"`
	PayloadMinBytes int32  `protobuf:"varint,7,opt,name=payloadMinBytes,proto3" json:"payloadMinBytes,omitempty"`
	PayloadMaxBytes int32  `protobuf:"varint,8,opt,name=payloadMaxBytes,proto3" json:"payloadMaxBytes,omitempty"`
}

func (x *Test) Reset() {
    logtrace.LogWithFunctionName()
	*x = Test{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loop_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Test) String() string {
    logtrace.LogWithFunctionName()
	return protoimpl.X.MessageStringOf(x)
}

func (*Test) ProtoMessage() {}
    logtrace.LogWithFunctionName()

func (x *Test) ProtoReflect() protoreflect.Message {
    logtrace.LogWithFunctionName()
	mi := &file_loop_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Test.ProtoReflect.Descriptor instead.
func (*Test) Descriptor() ([]byte, []int) {
    logtrace.LogWithFunctionName()
	return file_loop_proto_rawDescGZIP(), []int{0}
}

func (x *Test) GetName() string {
    logtrace.LogWithFunctionName()
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Test) GetTxRequests() int32 {
    logtrace.LogWithFunctionName()
	if x != nil {
		return x.TxRequests
	}
	return 0
}

func (x *Test) GetTxPacing() int32 {
    logtrace.LogWithFunctionName()
	if x != nil {
		return x.TxPacing
	}
	return 0
}

func (x *Test) GetTxMaxJitter() int32 {
    logtrace.LogWithFunctionName()
	if x != nil {
		return x.TxMaxJitter
	}
	return 0
}

func (x *Test) GetRxRequests() int32 {
    logtrace.LogWithFunctionName()
	if x != nil {
		return x.RxRequests
	}
	return 0
}

func (x *Test) GetRxTimeout() int32 {
    logtrace.LogWithFunctionName()
	if x != nil {
		return x.RxTimeout
	}
	return 0
}

func (x *Test) GetPayloadMinBytes() int32 {
    logtrace.LogWithFunctionName()
	if x != nil {
		return x.PayloadMinBytes
	}
	return 0
}

func (x *Test) GetPayloadMaxBytes() int32 {
    logtrace.LogWithFunctionName()
	if x != nil {
		return x.PayloadMaxBytes
	}
	return 0
}

type Block struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sequence int32  `protobuf:"varint,1,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Data     []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	Hash     []byte `protobuf:"bytes,3,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *Block) Reset() {
    logtrace.LogWithFunctionName()
	*x = Block{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loop_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Block) String() string {
    logtrace.LogWithFunctionName()
	return protoimpl.X.MessageStringOf(x)
}

func (*Block) ProtoMessage() {}
    logtrace.LogWithFunctionName()

func (x *Block) ProtoReflect() protoreflect.Message {
    logtrace.LogWithFunctionName()
	mi := &file_loop_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Block.ProtoReflect.Descriptor instead.
func (*Block) Descriptor() ([]byte, []int) {
    logtrace.LogWithFunctionName()
	return file_loop_proto_rawDescGZIP(), []int{1}
}

func (x *Block) GetSequence() int32 {
    logtrace.LogWithFunctionName()
	if x != nil {
		return x.Sequence
	}
	return 0
}

func (x *Block) GetData() []byte {
    logtrace.LogWithFunctionName()
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Block) GetHash() []byte {
    logtrace.LogWithFunctionName()
	if x != nil {
		return x.Hash
	}
	return nil
}

type Result struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *Result) Reset() {
    logtrace.LogWithFunctionName()
	*x = Result{}
	if protoimpl.UnsafeEnabled {
		mi := &file_loop_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Result) String() string {
    logtrace.LogWithFunctionName()
	return protoimpl.X.MessageStringOf(x)
}

func (*Result) ProtoMessage() {}
    logtrace.LogWithFunctionName()

func (x *Result) ProtoReflect() protoreflect.Message {
    logtrace.LogWithFunctionName()
	mi := &file_loop_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Result.ProtoReflect.Descriptor instead.
func (*Result) Descriptor() ([]byte, []int) {
    logtrace.LogWithFunctionName()
	return file_loop_proto_rawDescGZIP(), []int{2}
}

func (x *Result) GetSuccess() bool {
    logtrace.LogWithFunctionName()
	if x != nil {
		return x.Success
	}
	return false
}

func (x *Result) GetMessage() string {
    logtrace.LogWithFunctionName()
	if x != nil {
		return x.Message
	}
	return ""
}

var File_loop_proto protoreflect.FileDescriptor

var file_loop_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x6c, 0x6f, 0x6f, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x7a, 0x69,
	0x74, 0x69, 0x5f, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x6c, 0x6f, 0x6f, 0x70, 0x32, 0x2e, 0x70, 0x62,
	0x22, 0x8a, 0x02, 0x0a, 0x04, 0x54, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1e, 0x0a,
	0x0a, 0x74, 0x78, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x0a, 0x74, 0x78, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x12, 0x1a, 0x0a,
	0x08, 0x74, 0x78, 0x50, 0x61, 0x63, 0x69, 0x6e, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x08, 0x74, 0x78, 0x50, 0x61, 0x63, 0x69, 0x6e, 0x67, 0x12, 0x20, 0x0a, 0x0b, 0x74, 0x78, 0x4d,
	0x61, 0x78, 0x4a, 0x69, 0x74, 0x74, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b,
	0x74, 0x78, 0x4d, 0x61, 0x78, 0x4a, 0x69, 0x74, 0x74, 0x65, 0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x72,
	0x78, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x0a, 0x72, 0x78, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x72,
	0x78, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09,
	0x72, 0x78, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x28, 0x0a, 0x0f, 0x70, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x69, 0x6e, 0x42, 0x79, 0x74, 0x65, 0x73, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x69, 0x6e, 0x42, 0x79,
	0x74, 0x65, 0x73, 0x12, 0x28, 0x0a, 0x0f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x61,
	0x78, 0x42, 0x79, 0x74, 0x65, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0f, 0x70, 0x61,
	0x79, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x61, 0x78, 0x42, 0x79, 0x74, 0x65, 0x73, 0x22, 0x4b, 0x0a,
	0x05, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e,
	0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e,
	0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x22, 0x3c, 0x0a, 0x06, 0x52, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x18,
	0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0x42, 0x5a, 0x40, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6f, 0x70, 0x65, 0x6e, 0x7a, 0x69, 0x74, 0x69, 0x2f,
	0x2f, 0x7a, 0x69, 0x74, 0x69, 0x2f, 0x7a, 0x69, 0x74, 0x69, 0x2d, 0x66, 0x61, 0x62, 0x72, 0x69,
	0x63, 0x2d, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x73, 0x75, 0x62, 0x63, 0x6d, 0x64, 0x2f, 0x6c, 0x6f,
	0x6f, 0x70, 0x32, 0x2f, 0x6c, 0x6f, 0x6f, 0x70, 0x32, 0x5f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_loop_proto_rawDescOnce sync.Once
	file_loop_proto_rawDescData = file_loop_proto_rawDesc
)

func file_loop_proto_rawDescGZIP() []byte {
    logtrace.LogWithFunctionName()
	file_loop_proto_rawDescOnce.Do(func() {
		file_loop_proto_rawDescData = protoimpl.X.CompressGZIP(file_loop_proto_rawDescData)
	})
	return file_loop_proto_rawDescData
}

var file_loop_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_loop_proto_goTypes = []interface{}{
	(*Test)(nil),   // 0: ziti_test.loop2.pb.Test
	(*Block)(nil),  // 1: ziti_test.loop2.pb.Block
	(*Result)(nil), // 2: ziti_test.loop2.pb.Result
}
var file_loop_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_loop_proto_init() }
    logtrace.LogWithFunctionName()
func file_loop_proto_init() {
    logtrace.LogWithFunctionName()
	if File_loop_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_loop_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Test); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_loop_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Block); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_loop_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Result); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_loop_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_loop_proto_goTypes,
		DependencyIndexes: file_loop_proto_depIdxs,
		MessageInfos:      file_loop_proto_msgTypes,
	}.Build()
	File_loop_proto = out.File
	file_loop_proto_rawDesc = nil
	file_loop_proto_goTypes = nil
	file_loop_proto_depIdxs = nil
}
