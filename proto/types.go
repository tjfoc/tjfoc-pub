/*
Copyright Suzhou Tongji Fintech Research Institute 2018 All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
/*
 一个简单的通讯协议的实现，协议头如下: 协议头采用小端序
	0 ~ 3: 报文大小
	4 ~ 7: 报文类型
	8 ~ 11: 版本号
*/

package proto

import (
	"net"
)

// callback的返回值列表
const (
	PROTO_ACCEPTABLE_PEER = iota
	PROTO_UNACCEPTABLE_PEER
)

const (
	PROTO_ID = iota
	PROTO_ADDR
)

/*
 回调函数，在调用TcpNew和UdpNew的时候传入，该函数会在收到没有注册的节点的请求时调用，用户可以有以下操作:
 	0: 直接返回PROTO_UNACCEPTABLE_PEER, 拒绝该节点的所有报文
	1: 调用RegisterPeer后返回PROTO_ACCEPTABLE_PEER接受该节点的报文
*/
type SpCallback (func(interface{}, []byte, int) int) // int == 0时[]byte为用户id, int == 1时[]byte为addr(ip:port)

// 报文处理函数，第一个[]byte为对端标识，第二个[]byte为报文数据，interface{}为用户传入的值
type SpFunc (func(interface{}, []byte, []byte) int)

/*
 协议接口: tcp udp共用
 	Run运行服务
	RegisterPeer 注册节点
	UnregisterPeer 注销节点
	UnregisterFunc 注销报文处理函数, 第一个参数为版本号， 第二个参数为类型
	RegisterFunc 注册报文处理函数，第一个参数为版本号，第二个参数为类型，第三个参数为报文处理函数
	SendInstruction 报文发送指令， 第一个参数为版本号， 第二个参数为类型， 第三个参数为报文，第四个参数为节点标识
*/
type SP interface {
	Run() error

	UnregisterPeer([]byte) error
	RegisterPeer([]byte, net.Addr) error

	UnregisterFunc(uint32, uint32) error
	RegisterFunc(uint32, uint32, SpFunc) error

	SendInstruction(uint32, uint32, []byte, []byte) error

	GetMemberList() map[string]string
}
