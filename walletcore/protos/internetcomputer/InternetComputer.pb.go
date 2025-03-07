// SPDX-License-Identifier: Apache-2.0
//
// Copyright © 2017 Trust Wallet.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v5.26.1
// source: InternetComputer.proto

package internetcomputer

import (
	common "github.com/vultisig/vultisigner/walletcore/protos/common"
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

// Internet Computer Transactions
type Transaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Payload transfer
	//
	// Types that are assignable to TransactionOneof:
	//
	//	*Transaction_Transfer_
	TransactionOneof isTransaction_TransactionOneof `protobuf_oneof:"transaction_oneof"`
}

func (x *Transaction) Reset() {
	*x = Transaction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_InternetComputer_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Transaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction) ProtoMessage() {}

func (x *Transaction) ProtoReflect() protoreflect.Message {
	mi := &file_InternetComputer_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transaction.ProtoReflect.Descriptor instead.
func (*Transaction) Descriptor() ([]byte, []int) {
	return file_InternetComputer_proto_rawDescGZIP(), []int{0}
}

func (m *Transaction) GetTransactionOneof() isTransaction_TransactionOneof {
	if m != nil {
		return m.TransactionOneof
	}
	return nil
}

func (x *Transaction) GetTransfer() *Transaction_Transfer {
	if x, ok := x.GetTransactionOneof().(*Transaction_Transfer_); ok {
		return x.Transfer
	}
	return nil
}

type isTransaction_TransactionOneof interface {
	isTransaction_TransactionOneof()
}

type Transaction_Transfer_ struct {
	Transfer *Transaction_Transfer `protobuf:"bytes,1,opt,name=transfer,proto3,oneof"`
}

func (*Transaction_Transfer_) isTransaction_TransactionOneof() {}

// Input data necessary to create a signed transaction.
type SigningInput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PrivateKey  []byte       `protobuf:"bytes,1,opt,name=private_key,json=privateKey,proto3" json:"private_key,omitempty"`
	Transaction *Transaction `protobuf:"bytes,2,opt,name=transaction,proto3" json:"transaction,omitempty"`
}

func (x *SigningInput) Reset() {
	*x = SigningInput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_InternetComputer_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SigningInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SigningInput) ProtoMessage() {}

func (x *SigningInput) ProtoReflect() protoreflect.Message {
	mi := &file_InternetComputer_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SigningInput.ProtoReflect.Descriptor instead.
func (*SigningInput) Descriptor() ([]byte, []int) {
	return file_InternetComputer_proto_rawDescGZIP(), []int{1}
}

func (x *SigningInput) GetPrivateKey() []byte {
	if x != nil {
		return x.PrivateKey
	}
	return nil
}

func (x *SigningInput) GetTransaction() *Transaction {
	if x != nil {
		return x.Transaction
	}
	return nil
}

// Transaction signing output.
type SigningOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Signed and encoded transaction bytes.
	// NOTE: Before sending to the Rosetta node, this value should be hex-encoded before using with the JSON structure.
	SignedTransaction []byte              `protobuf:"bytes,1,opt,name=signed_transaction,json=signedTransaction,proto3" json:"signed_transaction,omitempty"`
	Error             common.SigningError `protobuf:"varint,2,opt,name=error,proto3,enum=TW.Common.Proto.SigningError" json:"error,omitempty"`
	ErrorMessage      string              `protobuf:"bytes,3,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

func (x *SigningOutput) Reset() {
	*x = SigningOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_InternetComputer_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SigningOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SigningOutput) ProtoMessage() {}

func (x *SigningOutput) ProtoReflect() protoreflect.Message {
	mi := &file_InternetComputer_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SigningOutput.ProtoReflect.Descriptor instead.
func (*SigningOutput) Descriptor() ([]byte, []int) {
	return file_InternetComputer_proto_rawDescGZIP(), []int{2}
}

func (x *SigningOutput) GetSignedTransaction() []byte {
	if x != nil {
		return x.SignedTransaction
	}
	return nil
}

func (x *SigningOutput) GetError() common.SigningError {
	if x != nil {
		return x.Error
	}
	return common.SigningError(0)
}

func (x *SigningOutput) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

// ICP ledger transfer arguments
type Transaction_Transfer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ToAccountIdentifier   string `protobuf:"bytes,1,opt,name=to_account_identifier,json=toAccountIdentifier,proto3" json:"to_account_identifier,omitempty"`
	Amount                uint64 `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
	Memo                  uint64 `protobuf:"varint,3,opt,name=memo,proto3" json:"memo,omitempty"`
	CurrentTimestampNanos uint64 `protobuf:"varint,4,opt,name=current_timestamp_nanos,json=currentTimestampNanos,proto3" json:"current_timestamp_nanos,omitempty"`
	PermittedDrift        uint64 `protobuf:"varint,5,opt,name=permitted_drift,json=permittedDrift,proto3" json:"permitted_drift,omitempty"`
}

func (x *Transaction_Transfer) Reset() {
	*x = Transaction_Transfer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_InternetComputer_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Transaction_Transfer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction_Transfer) ProtoMessage() {}

func (x *Transaction_Transfer) ProtoReflect() protoreflect.Message {
	mi := &file_InternetComputer_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transaction_Transfer.ProtoReflect.Descriptor instead.
func (*Transaction_Transfer) Descriptor() ([]byte, []int) {
	return file_InternetComputer_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Transaction_Transfer) GetToAccountIdentifier() string {
	if x != nil {
		return x.ToAccountIdentifier
	}
	return ""
}

func (x *Transaction_Transfer) GetAmount() uint64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *Transaction_Transfer) GetMemo() uint64 {
	if x != nil {
		return x.Memo
	}
	return 0
}

func (x *Transaction_Transfer) GetCurrentTimestampNanos() uint64 {
	if x != nil {
		return x.CurrentTimestampNanos
	}
	return 0
}

func (x *Transaction_Transfer) GetPermittedDrift() uint64 {
	if x != nil {
		return x.PermittedDrift
	}
	return 0
}

var File_InternetComputer_proto protoreflect.FileDescriptor

var file_InternetComputer_proto_rawDesc = []byte{
	0x0a, 0x16, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6d, 0x70, 0x75, 0x74,
	0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19, 0x54, 0x57, 0x2e, 0x49, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6d, 0x70, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xbf, 0x02, 0x0a, 0x0b, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x4d, 0x0a, 0x08, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x54, 0x57, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65,
	0x74, 0x43, 0x6f, 0x6d, 0x70, 0x75, 0x74, 0x65, 0x72, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x54, 0x72, 0x61, 0x6e,
	0x73, 0x66, 0x65, 0x72, 0x48, 0x00, 0x52, 0x08, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72,
	0x1a, 0xcb, 0x01, 0x0a, 0x08, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x12, 0x32, 0x0a,
	0x15, 0x74, 0x6f, 0x5f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x13, 0x74, 0x6f,
	0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65,
	0x72, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x65, 0x6d,
	0x6f, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x04, 0x6d, 0x65, 0x6d, 0x6f, 0x12, 0x36, 0x0a,
	0x17, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x5f, 0x6e, 0x61, 0x6e, 0x6f, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x15,
	0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x4e, 0x61, 0x6e, 0x6f, 0x73, 0x12, 0x27, 0x0a, 0x0f, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x74, 0x74,
	0x65, 0x64, 0x5f, 0x64, 0x72, 0x69, 0x66, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0e,
	0x70, 0x65, 0x72, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x44, 0x72, 0x69, 0x66, 0x74, 0x42, 0x13,
	0x0a, 0x11, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6f, 0x6e,
	0x65, 0x6f, 0x66, 0x22, 0x79, 0x0a, 0x0c, 0x53, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x49, 0x6e,
	0x70, 0x75, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x5f, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74,
	0x65, 0x4b, 0x65, 0x79, 0x12, 0x48, 0x0a, 0x0b, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x54, 0x57, 0x2e, 0x49,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x43, 0x6f, 0x6d, 0x70, 0x75, 0x74, 0x65, 0x72, 0x2e,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x0b, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x98,
	0x01, 0x0a, 0x0d, 0x53, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74,
	0x12, 0x2d, 0x0a, 0x12, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x11, 0x73, 0x69,
	0x67, 0x6e, 0x65, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x33, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1d,
	0x2e, 0x54, 0x57, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x53, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0x17, 0x0a, 0x15, 0x77, 0x61, 0x6c,
	0x6c, 0x65, 0x74, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x6a, 0x6e, 0x69, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_InternetComputer_proto_rawDescOnce sync.Once
	file_InternetComputer_proto_rawDescData = file_InternetComputer_proto_rawDesc
)

func file_InternetComputer_proto_rawDescGZIP() []byte {
	file_InternetComputer_proto_rawDescOnce.Do(func() {
		file_InternetComputer_proto_rawDescData = protoimpl.X.CompressGZIP(file_InternetComputer_proto_rawDescData)
	})
	return file_InternetComputer_proto_rawDescData
}

var file_InternetComputer_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_InternetComputer_proto_goTypes = []interface{}{
	(*Transaction)(nil),          // 0: TW.InternetComputer.Proto.Transaction
	(*SigningInput)(nil),         // 1: TW.InternetComputer.Proto.SigningInput
	(*SigningOutput)(nil),        // 2: TW.InternetComputer.Proto.SigningOutput
	(*Transaction_Transfer)(nil), // 3: TW.InternetComputer.Proto.Transaction.Transfer
	(common.SigningError)(0),     // 4: TW.Common.Proto.SigningError
}
var file_InternetComputer_proto_depIdxs = []int32{
	3, // 0: TW.InternetComputer.Proto.Transaction.transfer:type_name -> TW.InternetComputer.Proto.Transaction.Transfer
	0, // 1: TW.InternetComputer.Proto.SigningInput.transaction:type_name -> TW.InternetComputer.Proto.Transaction
	4, // 2: TW.InternetComputer.Proto.SigningOutput.error:type_name -> TW.Common.Proto.SigningError
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_InternetComputer_proto_init() }
func file_InternetComputer_proto_init() {
	if File_InternetComputer_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_InternetComputer_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Transaction); i {
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
		file_InternetComputer_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SigningInput); i {
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
		file_InternetComputer_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SigningOutput); i {
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
		file_InternetComputer_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Transaction_Transfer); i {
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
	file_InternetComputer_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Transaction_Transfer_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_InternetComputer_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_InternetComputer_proto_goTypes,
		DependencyIndexes: file_InternetComputer_proto_depIdxs,
		MessageInfos:      file_InternetComputer_proto_msgTypes,
	}.Build()
	File_InternetComputer_proto = out.File
	file_InternetComputer_proto_rawDesc = nil
	file_InternetComputer_proto_goTypes = nil
	file_InternetComputer_proto_depIdxs = nil
}
