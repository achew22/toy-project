// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: api/v1/helloworld.proto

package api

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

type GreetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *GreetRequest) Reset() {
	*x = GreetRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_helloworld_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GreetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GreetRequest) ProtoMessage() {}

func (x *GreetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_helloworld_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GreetRequest.ProtoReflect.Descriptor instead.
func (*GreetRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_helloworld_proto_rawDescGZIP(), []int{0}
}

func (x *GreetRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type GreetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *GreetResponse) Reset() {
	*x = GreetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_helloworld_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GreetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GreetResponse) ProtoMessage() {}

func (x *GreetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_helloworld_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GreetResponse.ProtoReflect.Descriptor instead.
func (*GreetResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_helloworld_proto_rawDescGZIP(), []int{1}
}

func (x *GreetResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_api_v1_helloworld_proto protoreflect.FileDescriptor

var file_api_v1_helloworld_proto_rawDesc = []byte{
	0x0a, 0x17, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x77, 0x6f,
	0x72, 0x6c, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1b, 0x63, 0x6d, 0x64, 0x2e, 0x61,
	0x63, 0x68, 0x65, 0x77, 0x2e, 0x74, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x22, 0x22, 0x0a, 0x0c, 0x47, 0x72, 0x65, 0x65, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x29, 0x0a, 0x0d, 0x47, 0x72,
	0x65, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x32, 0x6c, 0x0a, 0x0a, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x57, 0x6f,
	0x72, 0x6c, 0x64, 0x12, 0x5e, 0x0a, 0x05, 0x47, 0x72, 0x65, 0x65, 0x74, 0x12, 0x29, 0x2e, 0x63,
	0x6d, 0x64, 0x2e, 0x61, 0x63, 0x68, 0x65, 0x77, 0x2e, 0x74, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x72, 0x65, 0x65, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x63, 0x6d, 0x64, 0x2e, 0x61, 0x63,
	0x68, 0x65, 0x77, 0x2e, 0x74, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2e, 0x61,
	0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x72, 0x65, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x42, 0x2b, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x61, 0x63, 0x68, 0x65, 0x77, 0x32, 0x32, 0x2f, 0x74, 0x6f, 0x79, 0x2d, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x3b, 0x61, 0x70, 0x69,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_v1_helloworld_proto_rawDescOnce sync.Once
	file_api_v1_helloworld_proto_rawDescData = file_api_v1_helloworld_proto_rawDesc
)

func file_api_v1_helloworld_proto_rawDescGZIP() []byte {
	file_api_v1_helloworld_proto_rawDescOnce.Do(func() {
		file_api_v1_helloworld_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_v1_helloworld_proto_rawDescData)
	})
	return file_api_v1_helloworld_proto_rawDescData
}

var file_api_v1_helloworld_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_v1_helloworld_proto_goTypes = []any{
	(*GreetRequest)(nil),  // 0: cmd.achew.toyproject.api.v1.GreetRequest
	(*GreetResponse)(nil), // 1: cmd.achew.toyproject.api.v1.GreetResponse
}
var file_api_v1_helloworld_proto_depIdxs = []int32{
	0, // 0: cmd.achew.toyproject.api.v1.HelloWorld.Greet:input_type -> cmd.achew.toyproject.api.v1.GreetRequest
	1, // 1: cmd.achew.toyproject.api.v1.HelloWorld.Greet:output_type -> cmd.achew.toyproject.api.v1.GreetResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_v1_helloworld_proto_init() }
func file_api_v1_helloworld_proto_init() {
	if File_api_v1_helloworld_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_v1_helloworld_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*GreetRequest); i {
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
		file_api_v1_helloworld_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*GreetResponse); i {
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
			RawDescriptor: file_api_v1_helloworld_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_v1_helloworld_proto_goTypes,
		DependencyIndexes: file_api_v1_helloworld_proto_depIdxs,
		MessageInfos:      file_api_v1_helloworld_proto_msgTypes,
	}.Build()
	File_api_v1_helloworld_proto = out.File
	file_api_v1_helloworld_proto_rawDesc = nil
	file_api_v1_helloworld_proto_goTypes = nil
	file_api_v1_helloworld_proto_depIdxs = nil
}
