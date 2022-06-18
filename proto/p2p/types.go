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
package p2p

import (
	"net"
	"sync"
	"time"

	"github.com/tjfoc/tjfoc/core/blockchain"
	"github.com/tjfoc/tjfoc/core/consensus"
	"github.com/tjfoc/tjfoc/proto"
)

const (
	VERSION           = 0
	BLOCK_HEIGHT_SIZE = 8
)

const (
	PROBE             = 0x10
	BLOCK_INFO        = 0x20
	BLOCK_PROBE       = 0x40
	TRANSACTION_INFO  = 0x80
	TRANSACTION_PROBE = 0x100
)

type P2pServer interface {
	Run() error

	UnregisterPeer([]byte) error
	RegisterPeer([]byte, net.Addr) bool

	UnregisterFunc(uint32, uint32) error
	RegisterFunc(uint32, uint32, proto.SpFunc) error

	SendInstruction(uint32, uint32, []byte, []byte) error

	GetMemberList() map[string]string
}

type p2pProto struct {
	sp           proto.SP
	lock         sync.Mutex
	cycle        time.Duration
	consensusAPI consensus.Consensus
	exitCh       map[string]chan bool
	chain        blockchain.BlockChain
}

type p2pRegistry struct {
	id     uint32
	spFunc proto.SpFunc
}

// version 0
var p2pRegistry_0 = []p2pRegistry{
	{PROBE, probe},
	{BLOCK_INFO, blockInfo},
	{BLOCK_PROBE, blockProbe},
	{TRANSACTION_INFO, transactionInfo},
	{TRANSACTION_PROBE, transactionProbe},
}
