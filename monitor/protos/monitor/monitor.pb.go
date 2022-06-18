// Code generated by protoc-gen-go. DO NOT EDIT.
// source: monitor.proto

/*
Package monitor is a generated protocol buffer package.

It is generated from these files:
	monitor.proto

It has these top-level messages:
	Result
	TransactionKey
	TransactionResults
	DealResults
	CryptNumber
	CompareNumber
*/
package monitor

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import privateTransaction "github.com/tjfoc/tjfoc/monitor/protos/privateTransaction"

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

type Result struct {
	Req bool `protobuf:"varint,1,opt,name=req" json:"req,omitempty"`
}

func (m *Result) Reset()                    { *m = Result{} }
func (m *Result) String() string            { return proto.CompactTextString(m) }
func (*Result) ProtoMessage()               {}
func (*Result) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Result) GetReq() bool {
	if m != nil {
		return m.Req
	}
	return false
}

// 该结构用于查询一个加密的kv
type TransactionKey struct {
	Key         []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value       []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Status      []byte `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	TxId        []byte `protobuf:"bytes,4,opt,name=txId,proto3" json:"txId,omitempty"`
	PeerID      []byte `protobuf:"bytes,5,opt,name=peerID,proto3" json:"peerID,omitempty"`
	RandNum     []byte `protobuf:"bytes,6,opt,name=randNum,proto3" json:"randNum,omitempty"`
	Sign        []byte `protobuf:"bytes,7,opt,name=sign,proto3" json:"sign,omitempty"`
	Secretvalue []byte `protobuf:"bytes,8,opt,name=secretvalue,proto3" json:"secretvalue,omitempty"`
}

func (m *TransactionKey) Reset()                    { *m = TransactionKey{} }
func (m *TransactionKey) String() string            { return proto.CompactTextString(m) }
func (*TransactionKey) ProtoMessage()               {}
func (*TransactionKey) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *TransactionKey) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *TransactionKey) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *TransactionKey) GetStatus() []byte {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *TransactionKey) GetTxId() []byte {
	if m != nil {
		return m.TxId
	}
	return nil
}

func (m *TransactionKey) GetPeerID() []byte {
	if m != nil {
		return m.PeerID
	}
	return nil
}

func (m *TransactionKey) GetRandNum() []byte {
	if m != nil {
		return m.RandNum
	}
	return nil
}

func (m *TransactionKey) GetSign() []byte {
	if m != nil {
		return m.Sign
	}
	return nil
}

func (m *TransactionKey) GetSecretvalue() []byte {
	if m != nil {
		return m.Secretvalue
	}
	return nil
}

// 该结构用于隐私交易双方执行完成之后扩散隐私交易结果
type TransactionResults struct {
	Kv     map[string][]byte `protobuf:"bytes,1,rep,name=kv" json:"kv,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Action map[string]int32  `protobuf:"bytes,2,rep,name=action" json:"action,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
	TxId   []byte            `protobuf:"bytes,3,opt,name=txId,proto3" json:"txId,omitempty"`
}

func (m *TransactionResults) Reset()                    { *m = TransactionResults{} }
func (m *TransactionResults) String() string            { return proto.CompactTextString(m) }
func (*TransactionResults) ProtoMessage()               {}
func (*TransactionResults) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *TransactionResults) GetKv() map[string][]byte {
	if m != nil {
		return m.Kv
	}
	return nil
}

func (m *TransactionResults) GetAction() map[string]int32 {
	if m != nil {
		return m.Action
	}
	return nil
}

func (m *TransactionResults) GetTxId() []byte {
	if m != nil {
		return m.TxId
	}
	return nil
}

// 该结构用于整个block交易执行结束后，对需要进行加密的key进行加密处理
type DealResults struct {
	Kv      map[string][]byte `protobuf:"bytes,1,rep,name=kv" json:"kv,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Tongtai map[string]bool   `protobuf:"bytes,2,rep,name=tongtai" json:"tongtai,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
	PeerID  []byte            `protobuf:"bytes,3,opt,name=peerID,proto3" json:"peerID,omitempty"`
}

func (m *DealResults) Reset()                    { *m = DealResults{} }
func (m *DealResults) String() string            { return proto.CompactTextString(m) }
func (*DealResults) ProtoMessage()               {}
func (*DealResults) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *DealResults) GetKv() map[string][]byte {
	if m != nil {
		return m.Kv
	}
	return nil
}

func (m *DealResults) GetTongtai() map[string]bool {
	if m != nil {
		return m.Tongtai
	}
	return nil
}

func (m *DealResults) GetPeerID() []byte {
	if m != nil {
		return m.PeerID
	}
	return nil
}

// 改结构用于在执行同态加减法时，将需要加上或者减去的明文数值转换为同态加密数值
type CryptNumber struct {
	TxId   []byte `protobuf:"bytes,1,opt,name=txId,proto3" json:"txId,omitempty"`
	Value  int64  `protobuf:"varint,2,opt,name=value" json:"value,omitempty"`
	Svalue []byte `protobuf:"bytes,3,opt,name=svalue,proto3" json:"svalue,omitempty"`
}

func (m *CryptNumber) Reset()                    { *m = CryptNumber{} }
func (m *CryptNumber) String() string            { return proto.CompactTextString(m) }
func (*CryptNumber) ProtoMessage()               {}
func (*CryptNumber) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *CryptNumber) GetTxId() []byte {
	if m != nil {
		return m.TxId
	}
	return nil
}

func (m *CryptNumber) GetValue() int64 {
	if m != nil {
		return m.Value
	}
	return 0
}

func (m *CryptNumber) GetSvalue() []byte {
	if m != nil {
		return m.Svalue
	}
	return nil
}

// 加密过的(100) - 没有加密过的1000
type CompareNumber struct {
	TxId   []byte `protobuf:"bytes,1,opt,name=txId,proto3" json:"txId,omitempty"`
	Svalue []byte `protobuf:"bytes,2,opt,name=svalue,proto3" json:"svalue,omitempty"`
	Value  int64  `protobuf:"varint,3,opt,name=value" json:"value,omitempty"`
	Result int64  `protobuf:"varint,4,opt,name=result" json:"result,omitempty"`
}

func (m *CompareNumber) Reset()                    { *m = CompareNumber{} }
func (m *CompareNumber) String() string            { return proto.CompactTextString(m) }
func (*CompareNumber) ProtoMessage()               {}
func (*CompareNumber) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *CompareNumber) GetTxId() []byte {
	if m != nil {
		return m.TxId
	}
	return nil
}

func (m *CompareNumber) GetSvalue() []byte {
	if m != nil {
		return m.Svalue
	}
	return nil
}

func (m *CompareNumber) GetValue() int64 {
	if m != nil {
		return m.Value
	}
	return 0
}

func (m *CompareNumber) GetResult() int64 {
	if m != nil {
		return m.Result
	}
	return 0
}

func init() {
	proto.RegisterType((*Result)(nil), "monitor.Result")
	proto.RegisterType((*TransactionKey)(nil), "monitor.TransactionKey")
	proto.RegisterType((*TransactionResults)(nil), "monitor.TransactionResults")
	proto.RegisterType((*DealResults)(nil), "monitor.DealResults")
	proto.RegisterType((*CryptNumber)(nil), "monitor.CryptNumber")
	proto.RegisterType((*CompareNumber)(nil), "monitor.CompareNumber")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Monitor service

type MonitorClient interface {
	NewTransaction(ctx context.Context, in *privateTransaction.Transaction, opts ...grpc.CallOption) (*privateTransaction.Transaction, error)
	// 用于查询key对应的value
	QueryKey(ctx context.Context, in *TransactionKey, opts ...grpc.CallOption) (*TransactionKey, error)
	Authentication(ctx context.Context, in *TransactionKey, opts ...grpc.CallOption) (*TransactionKey, error)
	// 用于扩散隐私交易结果
	SendTxResult(ctx context.Context, in *TransactionResults, opts ...grpc.CallOption) (*Result, error)
	GetTxResult(ctx context.Context, in *TransactionResults, opts ...grpc.CallOption) (*TransactionResults, error)
	// 用于对整个block中需要进行加密的key进行加密
	SendTotalTxResult(ctx context.Context, in *DealResults, opts ...grpc.CallOption) (*DealResults, error)
	// 用于对某个明文数值进行同态加密
	RequireCrypt(ctx context.Context, in *CryptNumber, opts ...grpc.CallOption) (*CryptNumber, error)
	// 用于比较
	RequireCompare(ctx context.Context, in *CompareNumber, opts ...grpc.CallOption) (*CompareNumber, error)
}

type monitorClient struct {
	cc *grpc.ClientConn
}

func NewMonitorClient(cc *grpc.ClientConn) MonitorClient {
	return &monitorClient{cc}
}

func (c *monitorClient) NewTransaction(ctx context.Context, in *privateTransaction.Transaction, opts ...grpc.CallOption) (*privateTransaction.Transaction, error) {
	out := new(privateTransaction.Transaction)
	err := grpc.Invoke(ctx, "/monitor.Monitor/NewTransaction", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *monitorClient) QueryKey(ctx context.Context, in *TransactionKey, opts ...grpc.CallOption) (*TransactionKey, error) {
	out := new(TransactionKey)
	err := grpc.Invoke(ctx, "/monitor.Monitor/QueryKey", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *monitorClient) Authentication(ctx context.Context, in *TransactionKey, opts ...grpc.CallOption) (*TransactionKey, error) {
	out := new(TransactionKey)
	err := grpc.Invoke(ctx, "/monitor.Monitor/Authentication", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *monitorClient) SendTxResult(ctx context.Context, in *TransactionResults, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := grpc.Invoke(ctx, "/monitor.Monitor/SendTxResult", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *monitorClient) GetTxResult(ctx context.Context, in *TransactionResults, opts ...grpc.CallOption) (*TransactionResults, error) {
	out := new(TransactionResults)
	err := grpc.Invoke(ctx, "/monitor.Monitor/GetTxResult", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *monitorClient) SendTotalTxResult(ctx context.Context, in *DealResults, opts ...grpc.CallOption) (*DealResults, error) {
	out := new(DealResults)
	err := grpc.Invoke(ctx, "/monitor.Monitor/SendTotalTxResult", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *monitorClient) RequireCrypt(ctx context.Context, in *CryptNumber, opts ...grpc.CallOption) (*CryptNumber, error) {
	out := new(CryptNumber)
	err := grpc.Invoke(ctx, "/monitor.Monitor/RequireCrypt", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *monitorClient) RequireCompare(ctx context.Context, in *CompareNumber, opts ...grpc.CallOption) (*CompareNumber, error) {
	out := new(CompareNumber)
	err := grpc.Invoke(ctx, "/monitor.Monitor/RequireCompare", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Monitor service

type MonitorServer interface {
	NewTransaction(context.Context, *privateTransaction.Transaction) (*privateTransaction.Transaction, error)
	// 用于查询key对应的value
	QueryKey(context.Context, *TransactionKey) (*TransactionKey, error)
	Authentication(context.Context, *TransactionKey) (*TransactionKey, error)
	// 用于扩散隐私交易结果
	SendTxResult(context.Context, *TransactionResults) (*Result, error)
	GetTxResult(context.Context, *TransactionResults) (*TransactionResults, error)
	// 用于对整个block中需要进行加密的key进行加密
	SendTotalTxResult(context.Context, *DealResults) (*DealResults, error)
	// 用于对某个明文数值进行同态加密
	RequireCrypt(context.Context, *CryptNumber) (*CryptNumber, error)
	// 用于比较
	RequireCompare(context.Context, *CompareNumber) (*CompareNumber, error)
}

func RegisterMonitorServer(s *grpc.Server, srv MonitorServer) {
	s.RegisterService(&_Monitor_serviceDesc, srv)
}

func _Monitor_NewTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(privateTransaction.Transaction)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitorServer).NewTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monitor.Monitor/NewTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitorServer).NewTransaction(ctx, req.(*privateTransaction.Transaction))
	}
	return interceptor(ctx, in, info, handler)
}

func _Monitor_QueryKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransactionKey)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitorServer).QueryKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monitor.Monitor/QueryKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitorServer).QueryKey(ctx, req.(*TransactionKey))
	}
	return interceptor(ctx, in, info, handler)
}

func _Monitor_Authentication_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransactionKey)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitorServer).Authentication(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monitor.Monitor/Authentication",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitorServer).Authentication(ctx, req.(*TransactionKey))
	}
	return interceptor(ctx, in, info, handler)
}

func _Monitor_SendTxResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransactionResults)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitorServer).SendTxResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monitor.Monitor/SendTxResult",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitorServer).SendTxResult(ctx, req.(*TransactionResults))
	}
	return interceptor(ctx, in, info, handler)
}

func _Monitor_GetTxResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransactionResults)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitorServer).GetTxResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monitor.Monitor/GetTxResult",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitorServer).GetTxResult(ctx, req.(*TransactionResults))
	}
	return interceptor(ctx, in, info, handler)
}

func _Monitor_SendTotalTxResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DealResults)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitorServer).SendTotalTxResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monitor.Monitor/SendTotalTxResult",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitorServer).SendTotalTxResult(ctx, req.(*DealResults))
	}
	return interceptor(ctx, in, info, handler)
}

func _Monitor_RequireCrypt_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CryptNumber)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitorServer).RequireCrypt(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monitor.Monitor/RequireCrypt",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitorServer).RequireCrypt(ctx, req.(*CryptNumber))
	}
	return interceptor(ctx, in, info, handler)
}

func _Monitor_RequireCompare_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CompareNumber)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitorServer).RequireCompare(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/monitor.Monitor/RequireCompare",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitorServer).RequireCompare(ctx, req.(*CompareNumber))
	}
	return interceptor(ctx, in, info, handler)
}

var _Monitor_serviceDesc = grpc.ServiceDesc{
	ServiceName: "monitor.Monitor",
	HandlerType: (*MonitorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NewTransaction",
			Handler:    _Monitor_NewTransaction_Handler,
		},
		{
			MethodName: "QueryKey",
			Handler:    _Monitor_QueryKey_Handler,
		},
		{
			MethodName: "Authentication",
			Handler:    _Monitor_Authentication_Handler,
		},
		{
			MethodName: "SendTxResult",
			Handler:    _Monitor_SendTxResult_Handler,
		},
		{
			MethodName: "GetTxResult",
			Handler:    _Monitor_GetTxResult_Handler,
		},
		{
			MethodName: "SendTotalTxResult",
			Handler:    _Monitor_SendTotalTxResult_Handler,
		},
		{
			MethodName: "RequireCrypt",
			Handler:    _Monitor_RequireCrypt_Handler,
		},
		{
			MethodName: "RequireCompare",
			Handler:    _Monitor_RequireCompare_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "monitor.proto",
}

func init() { proto.RegisterFile("monitor.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 574 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x54, 0xdd, 0x6e, 0xd3, 0x30,
	0x14, 0xce, 0x4f, 0xdb, 0x94, 0x93, 0xae, 0x80, 0x35, 0x8d, 0x28, 0x20, 0x51, 0xcc, 0x05, 0xbb,
	0x40, 0xbd, 0xd8, 0x84, 0x04, 0x03, 0x0d, 0x95, 0x15, 0xa1, 0xaa, 0xa2, 0x88, 0xd0, 0x17, 0xf0,
	0xda, 0xa3, 0x11, 0xb5, 0x4d, 0x3a, 0xc7, 0x29, 0xcb, 0x35, 0x2f, 0xc0, 0xcb, 0xf0, 0x16, 0x3c,
	0x14, 0x8a, 0x9d, 0xb4, 0x1e, 0x64, 0xed, 0xb4, 0xbb, 0xf3, 0xf3, 0x7d, 0xe7, 0xc4, 0x9f, 0x3f,
	0x07, 0xf6, 0x16, 0x71, 0x14, 0x8a, 0x98, 0x77, 0x97, 0x3c, 0x16, 0x31, 0x71, 0x8a, 0xd4, 0xf7,
	0x96, 0x3c, 0x5c, 0x31, 0x81, 0x63, 0xce, 0xa2, 0x84, 0x4d, 0x44, 0x18, 0x47, 0x0a, 0x42, 0x7d,
	0x68, 0x04, 0x98, 0xa4, 0x73, 0x41, 0x1e, 0x80, 0xcd, 0xf1, 0xd2, 0x33, 0x3b, 0xe6, 0x61, 0x33,
	0xc8, 0x43, 0xfa, 0xc7, 0x84, 0xb6, 0xc6, 0x18, 0x62, 0x96, 0x83, 0x66, 0x98, 0x49, 0x50, 0x2b,
	0xc8, 0x43, 0xb2, 0x0f, 0xf5, 0x15, 0x9b, 0xa7, 0xe8, 0x59, 0xb2, 0xa6, 0x12, 0x72, 0x00, 0x8d,
	0x44, 0x30, 0x91, 0x26, 0x9e, 0x2d, 0xcb, 0x45, 0x46, 0x08, 0xd4, 0xc4, 0xd5, 0x60, 0xea, 0xd5,
	0x64, 0x55, 0xc6, 0x39, 0x76, 0x89, 0xc8, 0x07, 0x7d, 0xaf, 0xae, 0xb0, 0x2a, 0x23, 0x1e, 0x38,
	0x9c, 0x45, 0xd3, 0x51, 0xba, 0xf0, 0x1a, 0xb2, 0x51, 0xa6, 0xf9, 0x94, 0x24, 0xbc, 0x88, 0x3c,
	0x47, 0x4d, 0xc9, 0x63, 0xd2, 0x01, 0x37, 0xc1, 0x09, 0x47, 0xa1, 0xbe, 0xa6, 0x29, 0x5b, 0x7a,
	0x89, 0xfe, 0xb2, 0x80, 0x68, 0xc7, 0x51, 0xc7, 0x4e, 0xc8, 0x31, 0x58, 0xb3, 0x95, 0x67, 0x76,
	0xec, 0x43, 0xf7, 0xe8, 0x79, 0xb7, 0x14, 0xf0, 0x7f, 0x60, 0x77, 0xb8, 0xfa, 0x18, 0x09, 0x9e,
	0x05, 0xd6, 0x6c, 0x45, 0xde, 0x43, 0x43, 0x35, 0x3d, 0x4b, 0x12, 0x5f, 0x6c, 0x23, 0xf6, 0x64,
	0xa6, 0xc8, 0x05, 0x6d, 0x2d, 0x84, 0xbd, 0x11, 0xc2, 0x7f, 0x05, 0x4e, 0xb1, 0x43, 0xd7, 0xf9,
	0xde, 0x16, 0x9d, 0x4f, 0xac, 0xd7, 0xa6, 0xff, 0x06, 0x5c, 0x6d, 0xc3, 0x2e, 0x6a, 0x5d, 0xa3,
	0xd2, 0x9f, 0x16, 0xb8, 0x7d, 0x64, 0xf3, 0x52, 0x8b, 0x97, 0x9a, 0x16, 0x4f, 0xd6, 0x47, 0xd2,
	0x10, 0xd7, 0x44, 0x78, 0x0b, 0x8e, 0x88, 0xa3, 0x0b, 0xc1, 0xc2, 0x42, 0x85, 0x67, 0x95, 0x94,
	0xb1, 0xc2, 0x28, 0x5e, 0xc9, 0xd0, 0x6e, 0xdd, 0xd6, 0x6f, 0xfd, 0xae, 0x22, 0x9c, 0x40, 0x4b,
	0xdf, 0xb3, 0x8b, 0xdb, 0xd4, 0x55, 0xf8, 0x02, 0xee, 0x19, 0xcf, 0x96, 0x62, 0x94, 0x2e, 0xce,
	0x91, 0xaf, 0xaf, 0xc6, 0xd4, 0x3c, 0x7a, 0x8d, 0x6c, 0xeb, 0x2e, 0x57, 0xe5, 0xd2, 0xe5, 0xca,
	0x69, 0x21, 0xec, 0x9d, 0xc5, 0x8b, 0x25, 0xe3, 0xb8, 0x65, 0xe4, 0x86, 0x6c, 0xe9, 0xe4, 0xcd,
	0x2a, 0xfb, 0x9f, 0x55, 0x5c, 0xea, 0x29, 0x9f, 0x8e, 0x1d, 0x14, 0xd9, 0xd1, 0xef, 0x1a, 0x38,
	0x9f, 0x95, 0xe8, 0x64, 0x0c, 0xed, 0x11, 0xfe, 0xd0, 0x0c, 0x48, 0x9e, 0x76, 0x2b, 0x1e, 0xbe,
	0x16, 0xfb, 0xbb, 0x00, 0xd4, 0x20, 0xa7, 0xd0, 0xfc, 0x9a, 0x22, 0xcf, 0xf2, 0xe7, 0xff, 0xa8,
	0xca, 0xe6, 0x43, 0xcc, 0xfc, 0x9b, 0x1a, 0xd4, 0x20, 0x7d, 0x68, 0xf7, 0x52, 0xf1, 0x1d, 0x23,
	0x11, 0x4e, 0x98, 0xfc, 0xaa, 0xbb, 0x4c, 0x39, 0x85, 0xd6, 0x37, 0x8c, 0xa6, 0xe3, 0xab, 0xe2,
	0x6f, 0xf5, 0x78, 0xcb, 0x83, 0xf3, 0xef, 0xaf, 0x9b, 0xaa, 0x42, 0x0d, 0x32, 0x00, 0xf7, 0x13,
	0x8a, 0xdb, 0xd1, 0xb7, 0x35, 0xa9, 0x41, 0x7a, 0xf0, 0x50, 0x7e, 0x4a, 0x2c, 0xd8, 0x7c, 0x3d,
	0x70, 0xbf, 0xca, 0xfa, 0x7e, 0x65, 0x95, 0x1a, 0xe4, 0x1d, 0xb4, 0x02, 0xbc, 0x4c, 0x43, 0x8e,
	0xd2, 0x78, 0x1a, 0x5b, 0x33, 0xa2, 0x5f, 0x59, 0xa5, 0x06, 0xf9, 0x00, 0xed, 0x92, 0xad, 0x5c,
	0x46, 0x0e, 0x36, 0x48, 0xdd, 0x77, 0xfe, 0x0d, 0x75, 0x6a, 0x9c, 0x37, 0xe4, 0xef, 0xff, 0xf8,
	0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x04, 0x5f, 0xc4, 0x06, 0x32, 0x06, 0x00, 0x00,
}