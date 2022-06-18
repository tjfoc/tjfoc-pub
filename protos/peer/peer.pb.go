// Code generated by protoc-gen-go. DO NOT EDIT.
// source: peer.proto

/*
Package peer is a generated protocol buffer package.

It is generated from these files:
	peer.proto

It has these top-level messages:
	BlockchainBool
	BlockchainHash
	PeerInfo
	PeerUpdateInfo
	MemberListInfo
	BlockchainNumber
	SearchMes
	SearchRes
*/
package peer

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import block "github.com/tjfoc/tjfoc/protos/block"
import transaction "github.com/tjfoc/tjfoc/protos/transaction"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type BlockchainBool struct {
	Ok  bool   `protobuf:"varint,1,opt,name=ok" json:"ok,omitempty"`
	Err string `protobuf:"bytes,2,opt,name=err" json:"err,omitempty"`
}

func (m *BlockchainBool) Reset()                    { *m = BlockchainBool{} }
func (m *BlockchainBool) String() string            { return proto.CompactTextString(m) }
func (*BlockchainBool) ProtoMessage()               {}
func (*BlockchainBool) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *BlockchainBool) GetOk() bool {
	if m != nil {
		return m.Ok
	}
	return false
}

func (m *BlockchainBool) GetErr() string {
	if m != nil {
		return m.Err
	}
	return ""
}

type BlockchainHash struct {
	HashData []byte `protobuf:"bytes,1,opt,name=hashData,proto3" json:"hashData,omitempty"`
}

func (m *BlockchainHash) Reset()                    { *m = BlockchainHash{} }
func (m *BlockchainHash) String() string            { return proto.CompactTextString(m) }
func (*BlockchainHash) ProtoMessage()               {}
func (*BlockchainHash) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *BlockchainHash) GetHashData() []byte {
	if m != nil {
		return m.HashData
	}
	return nil
}

type PeerInfo struct {
	Id    string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Addr  string `protobuf:"bytes,2,opt,name=addr" json:"addr,omitempty"`
	State int32  `protobuf:"varint,3,opt,name=state" json:"state,omitempty"`
}

func (m *PeerInfo) Reset()                    { *m = PeerInfo{} }
func (m *PeerInfo) String() string            { return proto.CompactTextString(m) }
func (*PeerInfo) ProtoMessage()               {}
func (*PeerInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *PeerInfo) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *PeerInfo) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

func (m *PeerInfo) GetState() int32 {
	if m != nil {
		return m.State
	}
	return 0
}

type PeerUpdateInfo struct {
	Typ   int32  `protobuf:"varint,1,opt,name=typ" json:"typ,omitempty"`
	Id    string `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
	Addr  string `protobuf:"bytes,3,opt,name=addr" json:"addr,omitempty"`
	Data  string `protobuf:"bytes,4,opt,name=data" json:"data,omitempty"`
	Sign  string `protobuf:"bytes,5,opt,name=sign" json:"sign,omitempty"`
	Admin string `protobuf:"bytes,6,opt,name=admin" json:"admin,omitempty"`
}

func (m *PeerUpdateInfo) Reset()                    { *m = PeerUpdateInfo{} }
func (m *PeerUpdateInfo) String() string            { return proto.CompactTextString(m) }
func (*PeerUpdateInfo) ProtoMessage()               {}
func (*PeerUpdateInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *PeerUpdateInfo) GetTyp() int32 {
	if m != nil {
		return m.Typ
	}
	return 0
}

func (m *PeerUpdateInfo) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *PeerUpdateInfo) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

func (m *PeerUpdateInfo) GetData() string {
	if m != nil {
		return m.Data
	}
	return ""
}

func (m *PeerUpdateInfo) GetSign() string {
	if m != nil {
		return m.Sign
	}
	return ""
}

func (m *PeerUpdateInfo) GetAdmin() string {
	if m != nil {
		return m.Admin
	}
	return ""
}

type MemberListInfo struct {
	MemberList []*PeerInfo `protobuf:"bytes,1,rep,name=memberList" json:"memberList,omitempty"`
}

func (m *MemberListInfo) Reset()                    { *m = MemberListInfo{} }
func (m *MemberListInfo) String() string            { return proto.CompactTextString(m) }
func (*MemberListInfo) ProtoMessage()               {}
func (*MemberListInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *MemberListInfo) GetMemberList() []*PeerInfo {
	if m != nil {
		return m.MemberList
	}
	return nil
}

type BlockchainNumber struct {
	Number uint64 `protobuf:"varint,1,opt,name=number" json:"number,omitempty"`
}

func (m *BlockchainNumber) Reset()                    { *m = BlockchainNumber{} }
func (m *BlockchainNumber) String() string            { return proto.CompactTextString(m) }
func (*BlockchainNumber) ProtoMessage()               {}
func (*BlockchainNumber) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *BlockchainNumber) GetNumber() uint64 {
	if m != nil {
		return m.Number
	}
	return 0
}

type SearchMes struct {
	Key  [][]byte `protobuf:"bytes,1,rep,name=key,proto3" json:"key,omitempty"`
	Type uint32   `protobuf:"varint,2,opt,name=type" json:"type,omitempty"`
}

func (m *SearchMes) Reset()                    { *m = SearchMes{} }
func (m *SearchMes) String() string            { return proto.CompactTextString(m) }
func (*SearchMes) ProtoMessage()               {}
func (*SearchMes) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *SearchMes) GetKey() [][]byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *SearchMes) GetType() uint32 {
	if m != nil {
		return m.Type
	}
	return 0
}

type SearchRes struct {
	Res map[string]string `protobuf:"bytes,1,rep,name=res" json:"res,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *SearchRes) Reset()                    { *m = SearchRes{} }
func (m *SearchRes) String() string            { return proto.CompactTextString(m) }
func (*SearchRes) ProtoMessage()               {}
func (*SearchRes) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *SearchRes) GetRes() map[string]string {
	if m != nil {
		return m.Res
	}
	return nil
}

func init() {
	proto.RegisterType((*BlockchainBool)(nil), "peer.BlockchainBool")
	proto.RegisterType((*BlockchainHash)(nil), "peer.BlockchainHash")
	proto.RegisterType((*PeerInfo)(nil), "peer.PeerInfo")
	proto.RegisterType((*PeerUpdateInfo)(nil), "peer.PeerUpdateInfo")
	proto.RegisterType((*MemberListInfo)(nil), "peer.MemberListInfo")
	proto.RegisterType((*BlockchainNumber)(nil), "peer.BlockchainNumber")
	proto.RegisterType((*SearchMes)(nil), "peer.SearchMes")
	proto.RegisterType((*SearchRes)(nil), "peer.SearchRes")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Peer service

type PeerClient interface {
	Search(ctx context.Context, in *SearchMes, opts ...grpc.CallOption) (*SearchRes, error)
	NewTransaction(ctx context.Context, in *transaction.Transaction, opts ...grpc.CallOption) (*BlockchainBool, error)
	BlockchainGetHeight(ctx context.Context, in *BlockchainBool, opts ...grpc.CallOption) (*BlockchainNumber, error)
	BlockchainGetBlockByHash(ctx context.Context, in *BlockchainHash, opts ...grpc.CallOption) (*block.Block, error)
	BlockchainGetBlockByHeight(ctx context.Context, in *BlockchainNumber, opts ...grpc.CallOption) (*block.Block, error)
	BlockchainGetTransaction(ctx context.Context, in *BlockchainHash, opts ...grpc.CallOption) (*transaction.Transaction, error)
	BlockchainGetTransactionIndex(ctx context.Context, in *BlockchainHash, opts ...grpc.CallOption) (*BlockchainNumber, error)
	BlockchainGetTransactionBlock(ctx context.Context, in *BlockchainHash, opts ...grpc.CallOption) (*BlockchainNumber, error)
	GetMemberList(ctx context.Context, in *BlockchainBool, opts ...grpc.CallOption) (*MemberListInfo, error)
	UpdatePeer(ctx context.Context, in *PeerUpdateInfo, opts ...grpc.CallOption) (*BlockchainBool, error)
}

type peerClient struct {
	cc *grpc.ClientConn
}

func NewPeerClient(cc *grpc.ClientConn) PeerClient {
	return &peerClient{cc}
}

func (c *peerClient) Search(ctx context.Context, in *SearchMes, opts ...grpc.CallOption) (*SearchRes, error) {
	out := new(SearchRes)
	err := grpc.Invoke(ctx, "/peer.Peer/Search", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *peerClient) NewTransaction(ctx context.Context, in *transaction.Transaction, opts ...grpc.CallOption) (*BlockchainBool, error) {
	out := new(BlockchainBool)
	err := grpc.Invoke(ctx, "/peer.Peer/NewTransaction", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *peerClient) BlockchainGetHeight(ctx context.Context, in *BlockchainBool, opts ...grpc.CallOption) (*BlockchainNumber, error) {
	out := new(BlockchainNumber)
	err := grpc.Invoke(ctx, "/peer.Peer/BlockchainGetHeight", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *peerClient) BlockchainGetBlockByHash(ctx context.Context, in *BlockchainHash, opts ...grpc.CallOption) (*block.Block, error) {
	out := new(block.Block)
	err := grpc.Invoke(ctx, "/peer.Peer/BlockchainGetBlockByHash", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *peerClient) BlockchainGetBlockByHeight(ctx context.Context, in *BlockchainNumber, opts ...grpc.CallOption) (*block.Block, error) {
	out := new(block.Block)
	err := grpc.Invoke(ctx, "/peer.Peer/BlockchainGetBlockByHeight", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *peerClient) BlockchainGetTransaction(ctx context.Context, in *BlockchainHash, opts ...grpc.CallOption) (*transaction.Transaction, error) {
	out := new(transaction.Transaction)
	err := grpc.Invoke(ctx, "/peer.Peer/BlockchainGetTransaction", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *peerClient) BlockchainGetTransactionIndex(ctx context.Context, in *BlockchainHash, opts ...grpc.CallOption) (*BlockchainNumber, error) {
	out := new(BlockchainNumber)
	err := grpc.Invoke(ctx, "/peer.Peer/BlockchainGetTransactionIndex", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *peerClient) BlockchainGetTransactionBlock(ctx context.Context, in *BlockchainHash, opts ...grpc.CallOption) (*BlockchainNumber, error) {
	out := new(BlockchainNumber)
	err := grpc.Invoke(ctx, "/peer.Peer/BlockchainGetTransactionBlock", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *peerClient) GetMemberList(ctx context.Context, in *BlockchainBool, opts ...grpc.CallOption) (*MemberListInfo, error) {
	out := new(MemberListInfo)
	err := grpc.Invoke(ctx, "/peer.Peer/GetMemberList", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *peerClient) UpdatePeer(ctx context.Context, in *PeerUpdateInfo, opts ...grpc.CallOption) (*BlockchainBool, error) {
	out := new(BlockchainBool)
	err := grpc.Invoke(ctx, "/peer.Peer/UpdatePeer", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Peer service

type PeerServer interface {
	Search(context.Context, *SearchMes) (*SearchRes, error)
	NewTransaction(context.Context, *transaction.Transaction) (*BlockchainBool, error)
	BlockchainGetHeight(context.Context, *BlockchainBool) (*BlockchainNumber, error)
	BlockchainGetBlockByHash(context.Context, *BlockchainHash) (*block.Block, error)
	BlockchainGetBlockByHeight(context.Context, *BlockchainNumber) (*block.Block, error)
	BlockchainGetTransaction(context.Context, *BlockchainHash) (*transaction.Transaction, error)
	BlockchainGetTransactionIndex(context.Context, *BlockchainHash) (*BlockchainNumber, error)
	BlockchainGetTransactionBlock(context.Context, *BlockchainHash) (*BlockchainNumber, error)
	GetMemberList(context.Context, *BlockchainBool) (*MemberListInfo, error)
	UpdatePeer(context.Context, *PeerUpdateInfo) (*BlockchainBool, error)
}

func RegisterPeerServer(s *grpc.Server, srv PeerServer) {
	s.RegisterService(&_Peer_serviceDesc, srv)
}

func _Peer_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchMes)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PeerServer).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/peer.Peer/Search",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PeerServer).Search(ctx, req.(*SearchMes))
	}
	return interceptor(ctx, in, info, handler)
}

func _Peer_NewTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(transaction.Transaction)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PeerServer).NewTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/peer.Peer/NewTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PeerServer).NewTransaction(ctx, req.(*transaction.Transaction))
	}
	return interceptor(ctx, in, info, handler)
}

func _Peer_BlockchainGetHeight_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockchainBool)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PeerServer).BlockchainGetHeight(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/peer.Peer/BlockchainGetHeight",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PeerServer).BlockchainGetHeight(ctx, req.(*BlockchainBool))
	}
	return interceptor(ctx, in, info, handler)
}

func _Peer_BlockchainGetBlockByHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockchainHash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PeerServer).BlockchainGetBlockByHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/peer.Peer/BlockchainGetBlockByHash",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PeerServer).BlockchainGetBlockByHash(ctx, req.(*BlockchainHash))
	}
	return interceptor(ctx, in, info, handler)
}

func _Peer_BlockchainGetBlockByHeight_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockchainNumber)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PeerServer).BlockchainGetBlockByHeight(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/peer.Peer/BlockchainGetBlockByHeight",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PeerServer).BlockchainGetBlockByHeight(ctx, req.(*BlockchainNumber))
	}
	return interceptor(ctx, in, info, handler)
}

func _Peer_BlockchainGetTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockchainHash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PeerServer).BlockchainGetTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/peer.Peer/BlockchainGetTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PeerServer).BlockchainGetTransaction(ctx, req.(*BlockchainHash))
	}
	return interceptor(ctx, in, info, handler)
}

func _Peer_BlockchainGetTransactionIndex_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockchainHash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PeerServer).BlockchainGetTransactionIndex(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/peer.Peer/BlockchainGetTransactionIndex",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PeerServer).BlockchainGetTransactionIndex(ctx, req.(*BlockchainHash))
	}
	return interceptor(ctx, in, info, handler)
}

func _Peer_BlockchainGetTransactionBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockchainHash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PeerServer).BlockchainGetTransactionBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/peer.Peer/BlockchainGetTransactionBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PeerServer).BlockchainGetTransactionBlock(ctx, req.(*BlockchainHash))
	}
	return interceptor(ctx, in, info, handler)
}

func _Peer_GetMemberList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockchainBool)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PeerServer).GetMemberList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/peer.Peer/GetMemberList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PeerServer).GetMemberList(ctx, req.(*BlockchainBool))
	}
	return interceptor(ctx, in, info, handler)
}

func _Peer_UpdatePeer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PeerUpdateInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PeerServer).UpdatePeer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/peer.Peer/UpdatePeer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PeerServer).UpdatePeer(ctx, req.(*PeerUpdateInfo))
	}
	return interceptor(ctx, in, info, handler)
}

var _Peer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "peer.Peer",
	HandlerType: (*PeerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Search",
			Handler:    _Peer_Search_Handler,
		},
		{
			MethodName: "NewTransaction",
			Handler:    _Peer_NewTransaction_Handler,
		},
		{
			MethodName: "BlockchainGetHeight",
			Handler:    _Peer_BlockchainGetHeight_Handler,
		},
		{
			MethodName: "BlockchainGetBlockByHash",
			Handler:    _Peer_BlockchainGetBlockByHash_Handler,
		},
		{
			MethodName: "BlockchainGetBlockByHeight",
			Handler:    _Peer_BlockchainGetBlockByHeight_Handler,
		},
		{
			MethodName: "BlockchainGetTransaction",
			Handler:    _Peer_BlockchainGetTransaction_Handler,
		},
		{
			MethodName: "BlockchainGetTransactionIndex",
			Handler:    _Peer_BlockchainGetTransactionIndex_Handler,
		},
		{
			MethodName: "BlockchainGetTransactionBlock",
			Handler:    _Peer_BlockchainGetTransactionBlock_Handler,
		},
		{
			MethodName: "GetMemberList",
			Handler:    _Peer_GetMemberList_Handler,
		},
		{
			MethodName: "UpdatePeer",
			Handler:    _Peer_UpdatePeer_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "peer.proto",
}

func init() { proto.RegisterFile("peer.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 558 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0x6d, 0x6b, 0xdb, 0x30,
	0x10, 0xf6, 0x6b, 0x48, 0xaf, 0xad, 0xdb, 0x69, 0xa1, 0x18, 0xc3, 0x20, 0xe8, 0x53, 0x28, 0xc5,
	0xb0, 0x0c, 0xc6, 0x28, 0x0c, 0x4a, 0x48, 0x69, 0x0b, 0x4d, 0x37, 0xb4, 0xed, 0x07, 0x28, 0xf1,
	0xad, 0x36, 0x49, 0xec, 0x60, 0xa9, 0xdb, 0xf2, 0x7d, 0xbf, 0x68, 0xbf, 0x70, 0x48, 0x4a, 0x62,
	0xe7, 0xc5, 0x83, 0x7d, 0x7b, 0xee, 0xf1, 0xe9, 0xb9, 0xe7, 0xce, 0x27, 0x01, 0x2c, 0x10, 0xcb,
	0x78, 0x51, 0x16, 0xb2, 0x20, 0x9e, 0xc2, 0xd1, 0xf1, 0x78, 0x56, 0x4c, 0xa6, 0x86, 0x8a, 0x5e,
	0xc9, 0x92, 0xe7, 0x82, 0x4f, 0x64, 0x56, 0xe4, 0x86, 0xa2, 0x7d, 0x08, 0x06, 0x2a, 0x63, 0x92,
	0xf2, 0x2c, 0x1f, 0x14, 0xc5, 0x8c, 0x04, 0xe0, 0x14, 0xd3, 0xd0, 0xee, 0xda, 0xbd, 0x36, 0x73,
	0x8a, 0x29, 0x39, 0x07, 0x17, 0xcb, 0x32, 0x74, 0xba, 0x76, 0xef, 0x88, 0x29, 0x48, 0xaf, 0xea,
	0x67, 0xee, 0xb9, 0x48, 0x49, 0x04, 0xed, 0x94, 0x8b, 0x74, 0xc8, 0x25, 0xd7, 0x27, 0x4f, 0xd8,
	0x26, 0xa6, 0x43, 0x68, 0x7f, 0x46, 0x2c, 0x1f, 0xf2, 0xef, 0x85, 0xd2, 0xce, 0x12, 0x9d, 0x71,
	0xc4, 0x9c, 0x2c, 0x21, 0x04, 0x3c, 0x9e, 0x24, 0x6b, 0x71, 0x8d, 0x49, 0x07, 0x7c, 0x21, 0xb9,
	0xc4, 0xd0, 0xed, 0xda, 0x3d, 0x9f, 0x99, 0x80, 0xfe, 0xb6, 0x21, 0x50, 0x32, 0xdf, 0x16, 0x09,
	0x97, 0xa8, 0xc5, 0xce, 0xc1, 0x95, 0xcb, 0x85, 0x56, 0xf3, 0x99, 0x82, 0x2b, 0x79, 0x67, 0x4f,
	0xde, 0xad, 0xc9, 0x13, 0xf0, 0x12, 0x65, 0xd3, 0x33, 0x9c, 0xc2, 0x8a, 0x13, 0xd9, 0x73, 0x1e,
	0xfa, 0x86, 0x53, 0x58, 0xd9, 0xe0, 0xc9, 0x3c, 0xcb, 0xc3, 0x96, 0x26, 0x4d, 0x40, 0x6f, 0x20,
	0x18, 0xe1, 0x7c, 0x8c, 0xe5, 0x63, 0x26, 0xa4, 0x76, 0x11, 0x03, 0xcc, 0x37, 0x4c, 0x68, 0x77,
	0xdd, 0xde, 0x71, 0x3f, 0x88, 0xf5, 0x7f, 0x58, 0xb7, 0xcd, 0x6a, 0x19, 0xf4, 0x12, 0xce, 0xab,
	0xe1, 0x3d, 0xbd, 0x28, 0x9e, 0x5c, 0x40, 0x2b, 0xd7, 0x48, 0x37, 0xe3, 0xb1, 0x55, 0x44, 0xdf,
	0xc2, 0xd1, 0x17, 0xe4, 0xe5, 0x24, 0x1d, 0xa1, 0x50, 0xed, 0x4e, 0x71, 0xa9, 0x2b, 0x9c, 0x30,
	0x05, 0x95, 0x6d, 0xb9, 0x5c, 0xa0, 0x6e, 0xf8, 0x94, 0x69, 0x4c, 0x8b, 0xf5, 0x11, 0x86, 0x82,
	0x5c, 0x82, 0x5b, 0xa2, 0x58, 0x99, 0x0a, 0x8d, 0xa9, 0xcd, 0xd7, 0x98, 0xa1, 0xb8, 0xcd, 0x65,
	0xb9, 0x64, 0x2a, 0x29, 0x7a, 0x0f, 0xed, 0x35, 0x51, 0x95, 0xd2, 0xbf, 0x5c, 0x95, 0xea, 0x80,
	0xff, 0x83, 0xcf, 0x5e, 0x70, 0x35, 0x5c, 0x13, 0x5c, 0x3b, 0x1f, 0xec, 0xfe, 0x1f, 0x1f, 0x3c,
	0xd5, 0x28, 0xb9, 0x82, 0x96, 0xd1, 0x26, 0x67, 0xf5, 0x4a, 0x23, 0x14, 0xd1, 0xd9, 0x4e, 0x69,
	0x6a, 0x91, 0x01, 0x04, 0x4f, 0xf8, 0xf3, 0x6b, 0xb5, 0x8f, 0x24, 0x8c, 0xeb, 0xdb, 0x59, 0xfb,
	0x12, 0x75, 0xcc, 0xf1, 0xed, 0x3d, 0xa5, 0x16, 0xb9, 0x85, 0xd7, 0x15, 0x77, 0x87, 0xf2, 0x1e,
	0xb3, 0xe7, 0x54, 0x92, 0x83, 0xe9, 0xd1, 0xc5, 0x2e, 0x6b, 0x66, 0x4f, 0x2d, 0x72, 0x03, 0xe1,
	0x96, 0x8c, 0x0e, 0x06, 0x4b, 0xbd, 0xd8, 0x7b, 0x5a, 0x8a, 0x8d, 0x4e, 0x62, 0x73, 0xab, 0x34,
	0x4d, 0x2d, 0x32, 0x84, 0xe8, 0xa0, 0x82, 0xf1, 0xd3, 0x50, 0x79, 0x4f, 0xe5, 0x71, 0xc7, 0x47,
	0x7d, 0x38, 0x87, 0x7d, 0x34, 0x8e, 0x8c, 0x5a, 0xe4, 0x13, 0xbc, 0x69, 0x52, 0x7b, 0xc8, 0x13,
	0xfc, 0xd5, 0x20, 0xd9, 0x3c, 0xa6, 0x7f, 0x08, 0x6a, 0xfe, 0xbf, 0x05, 0x3f, 0xc2, 0xe9, 0x1d,
	0xca, 0xea, 0x3a, 0x35, 0xfc, 0xb8, 0x15, 0xbb, 0x7d, 0xed, 0xa8, 0x45, 0xae, 0x01, 0xcc, 0x63,
	0xa0, 0xb7, 0xaf, 0x53, 0x5d, 0xb9, 0xea, 0x89, 0x68, 0xda, 0x9c, 0x71, 0x4b, 0x3f, 0x7e, 0xef,
	0xfe, 0x06, 0x00, 0x00, 0xff, 0xff, 0x98, 0x9f, 0x8d, 0x3f, 0x30, 0x05, 0x00, 0x00,
}