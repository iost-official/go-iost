// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.2
// source: consensus/synchro/pb/message.proto

package msgpb

import (
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

type RequireType int32

const (
	RequireType_GETBLOCKHASHES         RequireType = 0
	RequireType_GETBLOCKHASHESBYNUMBER RequireType = 1
)

// Enum value maps for RequireType.
var (
	RequireType_name = map[int32]string{
		0: "GETBLOCKHASHES",
		1: "GETBLOCKHASHESBYNUMBER",
	}
	RequireType_value = map[string]int32{
		"GETBLOCKHASHES":         0,
		"GETBLOCKHASHESBYNUMBER": 1,
	}
)

func (x RequireType) Enum() *RequireType {
	p := new(RequireType)
	*p = x
	return p
}

func (x RequireType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RequireType) Descriptor() protoreflect.EnumDescriptor {
	return file_consensus_synchro_pb_message_proto_enumTypes[0].Descriptor()
}

func (RequireType) Type() protoreflect.EnumType {
	return &file_consensus_synchro_pb_message_proto_enumTypes[0]
}

func (x RequireType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RequireType.Descriptor instead.
func (RequireType) EnumDescriptor() ([]byte, []int) {
	return file_consensus_synchro_pb_message_proto_rawDescGZIP(), []int{0}
}

type BlockInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Number int64  `protobuf:"varint,1,opt,name=number,proto3" json:"number,omitempty"`
	Hash   []byte `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *BlockInfo) Reset() {
	*x = BlockInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_consensus_synchro_pb_message_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockInfo) ProtoMessage() {}

func (x *BlockInfo) ProtoReflect() protoreflect.Message {
	mi := &file_consensus_synchro_pb_message_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockInfo.ProtoReflect.Descriptor instead.
func (*BlockInfo) Descriptor() ([]byte, []int) {
	return file_consensus_synchro_pb_message_proto_rawDescGZIP(), []int{0}
}

func (x *BlockInfo) GetNumber() int64 {
	if x != nil {
		return x.Number
	}
	return 0
}

func (x *BlockInfo) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

type BlockHashQuery struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReqType RequireType `protobuf:"varint,1,opt,name=reqType,proto3,enum=msgpb.RequireType" json:"reqType,omitempty"`
	Start   int64       `protobuf:"varint,2,opt,name=start,proto3" json:"start,omitempty"`
	End     int64       `protobuf:"varint,3,opt,name=end,proto3" json:"end,omitempty"`
	Nums    []int64     `protobuf:"varint,4,rep,packed,name=nums,proto3" json:"nums,omitempty"`
}

func (x *BlockHashQuery) Reset() {
	*x = BlockHashQuery{}
	if protoimpl.UnsafeEnabled {
		mi := &file_consensus_synchro_pb_message_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockHashQuery) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockHashQuery) ProtoMessage() {}

func (x *BlockHashQuery) ProtoReflect() protoreflect.Message {
	mi := &file_consensus_synchro_pb_message_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockHashQuery.ProtoReflect.Descriptor instead.
func (*BlockHashQuery) Descriptor() ([]byte, []int) {
	return file_consensus_synchro_pb_message_proto_rawDescGZIP(), []int{1}
}

func (x *BlockHashQuery) GetReqType() RequireType {
	if x != nil {
		return x.ReqType
	}
	return RequireType_GETBLOCKHASHES
}

func (x *BlockHashQuery) GetStart() int64 {
	if x != nil {
		return x.Start
	}
	return 0
}

func (x *BlockHashQuery) GetEnd() int64 {
	if x != nil {
		return x.End
	}
	return 0
}

func (x *BlockHashQuery) GetNums() []int64 {
	if x != nil {
		return x.Nums
	}
	return nil
}

type BlockHashResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BlockInfos []*BlockInfo `protobuf:"bytes,1,rep,name=blockInfos,proto3" json:"blockInfos,omitempty"`
}

func (x *BlockHashResponse) Reset() {
	*x = BlockHashResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_consensus_synchro_pb_message_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockHashResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockHashResponse) ProtoMessage() {}

func (x *BlockHashResponse) ProtoReflect() protoreflect.Message {
	mi := &file_consensus_synchro_pb_message_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockHashResponse.ProtoReflect.Descriptor instead.
func (*BlockHashResponse) Descriptor() ([]byte, []int) {
	return file_consensus_synchro_pb_message_proto_rawDescGZIP(), []int{2}
}

func (x *BlockHashResponse) GetBlockInfos() []*BlockInfo {
	if x != nil {
		return x.BlockInfos
	}
	return nil
}

type SyncHeight struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Height int64 `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Time   int64 `protobuf:"varint,2,opt,name=time,proto3" json:"time,omitempty"`
}

func (x *SyncHeight) Reset() {
	*x = SyncHeight{}
	if protoimpl.UnsafeEnabled {
		mi := &file_consensus_synchro_pb_message_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncHeight) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncHeight) ProtoMessage() {}

func (x *SyncHeight) ProtoReflect() protoreflect.Message {
	mi := &file_consensus_synchro_pb_message_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncHeight.ProtoReflect.Descriptor instead.
func (*SyncHeight) Descriptor() ([]byte, []int) {
	return file_consensus_synchro_pb_message_proto_rawDescGZIP(), []int{3}
}

func (x *SyncHeight) GetHeight() int64 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *SyncHeight) GetTime() int64 {
	if x != nil {
		return x.Time
	}
	return 0
}

var File_consensus_synchro_pb_message_proto protoreflect.FileDescriptor

var file_consensus_synchro_pb_message_proto_rawDesc = []byte{
	0x0a, 0x22, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75, 0x73, 0x2f, 0x73, 0x79, 0x6e, 0x63,
	0x68, 0x72, 0x6f, 0x2f, 0x70, 0x62, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x6d, 0x73, 0x67, 0x70, 0x62, 0x22, 0x37, 0x0a, 0x09, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x16, 0x0a, 0x06, 0x6e, 0x75, 0x6d, 0x62,
	0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72,
	0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04,
	0x68, 0x61, 0x73, 0x68, 0x22, 0x7a, 0x0a, 0x0e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73,
	0x68, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x2c, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x54, 0x79, 0x70,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x12, 0x2e, 0x6d, 0x73, 0x67, 0x70, 0x62, 0x2e,
	0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x07, 0x72, 0x65, 0x71,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x05, 0x73, 0x74, 0x61, 0x72, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x65, 0x6e,
	0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x65, 0x6e, 0x64, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x75, 0x6d, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x03, 0x52, 0x04, 0x6e, 0x75, 0x6d, 0x73,
	0x22, 0x45, 0x0a, 0x11, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x30, 0x0a, 0x0a, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x49, 0x6e,
	0x66, 0x6f, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x6d, 0x73, 0x67, 0x70,
	0x62, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0a, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x73, 0x22, 0x38, 0x0a, 0x0a, 0x53, 0x79, 0x6e, 0x63, 0x48,
	0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x12, 0x0a,
	0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x74, 0x69, 0x6d,
	0x65, 0x2a, 0x3d, 0x0a, 0x0b, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x12, 0x0a, 0x0e, 0x47, 0x45, 0x54, 0x42, 0x4c, 0x4f, 0x43, 0x4b, 0x48, 0x41, 0x53, 0x48,
	0x45, 0x53, 0x10, 0x00, 0x12, 0x1a, 0x0a, 0x16, 0x47, 0x45, 0x54, 0x42, 0x4c, 0x4f, 0x43, 0x4b,
	0x48, 0x41, 0x53, 0x48, 0x45, 0x53, 0x42, 0x59, 0x4e, 0x55, 0x4d, 0x42, 0x45, 0x52, 0x10, 0x01,
	0x42, 0x3d, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x69,
	0x6f, 0x73, 0x74, 0x2d, 0x6f, 0x66, 0x66, 0x69, 0x63, 0x69, 0x61, 0x6c, 0x2f, 0x67, 0x6f, 0x2d,
	0x69, 0x6f, 0x73, 0x74, 0x2f, 0x76, 0x33, 0x2f, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75,
	0x73, 0x2f, 0x73, 0x79, 0x6e, 0x63, 0x68, 0x72, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_consensus_synchro_pb_message_proto_rawDescOnce sync.Once
	file_consensus_synchro_pb_message_proto_rawDescData = file_consensus_synchro_pb_message_proto_rawDesc
)

func file_consensus_synchro_pb_message_proto_rawDescGZIP() []byte {
	file_consensus_synchro_pb_message_proto_rawDescOnce.Do(func() {
		file_consensus_synchro_pb_message_proto_rawDescData = protoimpl.X.CompressGZIP(file_consensus_synchro_pb_message_proto_rawDescData)
	})
	return file_consensus_synchro_pb_message_proto_rawDescData
}

var file_consensus_synchro_pb_message_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_consensus_synchro_pb_message_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_consensus_synchro_pb_message_proto_goTypes = []interface{}{
	(RequireType)(0),          // 0: msgpb.RequireType
	(*BlockInfo)(nil),         // 1: msgpb.BlockInfo
	(*BlockHashQuery)(nil),    // 2: msgpb.BlockHashQuery
	(*BlockHashResponse)(nil), // 3: msgpb.BlockHashResponse
	(*SyncHeight)(nil),        // 4: msgpb.SyncHeight
}
var file_consensus_synchro_pb_message_proto_depIdxs = []int32{
	0, // 0: msgpb.BlockHashQuery.reqType:type_name -> msgpb.RequireType
	1, // 1: msgpb.BlockHashResponse.blockInfos:type_name -> msgpb.BlockInfo
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_consensus_synchro_pb_message_proto_init() }
func file_consensus_synchro_pb_message_proto_init() {
	if File_consensus_synchro_pb_message_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_consensus_synchro_pb_message_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockInfo); i {
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
		file_consensus_synchro_pb_message_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockHashQuery); i {
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
		file_consensus_synchro_pb_message_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockHashResponse); i {
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
		file_consensus_synchro_pb_message_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncHeight); i {
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
			RawDescriptor: file_consensus_synchro_pb_message_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_consensus_synchro_pb_message_proto_goTypes,
		DependencyIndexes: file_consensus_synchro_pb_message_proto_depIdxs,
		EnumInfos:         file_consensus_synchro_pb_message_proto_enumTypes,
		MessageInfos:      file_consensus_synchro_pb_message_proto_msgTypes,
	}.Build()
	File_consensus_synchro_pb_message_proto = out.File
	file_consensus_synchro_pb_message_proto_rawDesc = nil
	file_consensus_synchro_pb_message_proto_goTypes = nil
	file_consensus_synchro_pb_message_proto_depIdxs = nil
}
