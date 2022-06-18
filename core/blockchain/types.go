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
package blockchain

import (
	"sync"

	"github.com/tjfoc/tjfoc/core/block"
	"github.com/tjfoc/tjfoc/core/crypt"
	"github.com/tjfoc/tjfoc/core/store/chain"
	"github.com/tjfoc/tjfoc/core/transaction"
)

const (
	BLOCK_SIZE = 100 * 1024 // 区块大小限制为10KB
)

type BcCallBackFunc (func(interface{}, *transaction.Transaction) int)

type blockQueue struct {
	notice    chan bool
	blockList []*BlockTxs
}

type blockChain struct {
	isPack            bool
	isSync            bool
	lastHash          []byte
	threshold         uint64 // 启用数据同步的阀值
	syncBlocks        *blockQueue
	packBlocks        *blockQueue
	lock              sync.Mutex
	usrData           interface{}
	blockchain        chain.Chain
	last              *block.Block
	cryptoPlug        crypt.Crypto
	callBack          BcCallBackFunc
	transactionBuffer map[string]*transaction.Transaction // 交易缓冲区

}

type BlockTxs struct {
	SmartContract     [][]byte
	Hashs             [][]byte
	SmartContractArgs [][][]byte
	UsrData           interface{}
	bk                *block.Block
	txs               transaction.Transactions
}

type BlockChain interface {
	Lock()
	Unlock()
	Size() int
	chain.Chain
	LastHash() []byte
	Last() *block.Block

	PackAddBlock(*block.Block) error

	SyncAddBlock(*block.Block) error

	SaveThreshold(uint64)

	Pack(uint64) *block.Block

	GetTransactionList() []byte

	GetTransactionNumber() int

	TransactionElem([]byte) bool
	TransactionAdd(*transaction.Transaction) error
	GetTransactionFromBuffer([]byte) *transaction.Transaction
}

func (a *blockChain) Lock() {
	a.lock.Lock()
}

func (a *blockChain) Unlock() {
	a.lock.Unlock()
}

func (a *blockChain) Size() int {
	return a.cryptoPlug.Size()
}

func (a *blockChain) Height() uint64 {
	a.Lock()
	defer a.Unlock()
	return a.last.Height() + 1
}

func (a *blockChain) SaveThreshold(b uint64) {
	a.Lock()
	defer a.Unlock()
	a.threshold = b
}

func (a *blockChain) LastHash() []byte {
	a.Lock()
	defer a.Unlock()
	return a.lastHash
}

func (a *blockChain) Last() *block.Block {
	a.Lock()
	defer a.Unlock()
	return a.last
}

func (a *blockChain) AddBlock(b *block.Block) error {
	a.Lock()
	defer a.Unlock()
	return a.blockchain.AddBlock(b)
}

func (a *blockChain) SaveMemberList(b map[string]*chain.PeerInfo) error {
	return a.blockchain.SaveMemberList(b)
}

func (a *blockChain) GetMemberList() (map[string]*chain.PeerInfo, error) {
	return a.blockchain.GetMemberList()
}

func (a *blockChain) GetBlockByHash(h []byte) *block.Block {
	return a.blockchain.GetBlockByHash(h)
}

func (a *blockChain) GetBlockByHeight(h uint64) *block.Block {
	return a.blockchain.GetBlockByHeight(h)
}

func (a *blockChain) GetTransaction(h []byte) *transaction.Transaction {
	return a.blockchain.GetTransaction(h)
}

func (a *blockChain) GetTransactionIndex(h []byte) (uint32, error) {
	return a.blockchain.GetTransactionIndex(h)
}

func (a *blockChain) GetTransactionBlock(h []byte) (uint64, error) {
	return a.blockchain.GetTransactionBlock(h)
}
