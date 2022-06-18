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
package block

import (
	"errors"
	"reflect"

	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/transaction"
	"github.com/tjfoc/tjfoc/core/typeclass"
)

const (
	BLOCK_HASH_SIZE   = 32
	BLOCK_HEADER_SIZE = 116
)

/*
区块信息存储方式如下:
	区块高度 -- 区块头 - 区块hash值(32bytes) - 交易信息
*/

// 116 bytes
type BlockHeader struct {
	version          uint32 // 4 bytes
	height           uint64 // 8 bytes
	timestamp        uint64 // UTC time -- 8 bytes
	prevBlock        []byte // prevous block hash -- 32 bytes
	stateRoot        []byte // root of world state -- 32 bytes
	transactionsRoot []byte // root of all transactions -- 32 bytes
}

type Block struct {
	header       *BlockHeader
	transactions *transaction.Transactions
}

func (a *Block) Eq(b typeclass.Eq) bool {
	v, ok := b.(*Block)
	if !ok {
		return false
	}
	return reflect.DeepEqual(*a.header, *v.header)
}

func (a *Block) NotEq(b typeclass.Eq) bool {
	return !a.Eq(b)
}
func (a *Block) Header() *BlockHeader {
	return a.header
}

func (a *Block) Height() uint64 {
	return a.header.height
}

func (a *Block) Version() uint32 {
	return a.header.version
}

func (a *Block) Timestamp() uint64 {
	return a.header.timestamp
}

func (a *Block) PreviousBlock() []byte {
	return a.header.prevBlock
}

func (a *Block) WorldStateRoot() []byte {
	return a.header.stateRoot
}

func (a *Block) TransactionsRoot() []byte {
	return a.header.transactionsRoot
}

func (a *Block) TransactionList() *transaction.Transactions {
	return a.transactions
}

func (a *BlockHeader) Show() ([]byte, error) {
	buf := []byte{}
	if len(a.prevBlock) != BLOCK_HASH_SIZE || len(a.stateRoot) != BLOCK_HASH_SIZE || len(a.transactionsRoot) != BLOCK_HASH_SIZE {
		return nil, errors.New("Block Header Show: Illegal hash function")
	}
	buf = append(buf, miscellaneous.E32func(a.version)...)
	buf = append(buf, miscellaneous.E64func(a.height)...)
	buf = append(buf, miscellaneous.E64func(a.timestamp)...)
	buf = append(buf, a.prevBlock...)
	buf = append(buf, a.stateRoot...)
	buf = append(buf, a.transactionsRoot...)
	return buf, nil
}

func (a *BlockHeader) Read(b []byte) ([]byte, error) {
	if len(b) < BLOCK_HEADER_SIZE {
		return nil, errors.New("Block Header Read: Illegal slice length")
	}
	a.version, _ = miscellaneous.D32func(b[:4])
	a.height, _ = miscellaneous.D64func(b[4:12])
	a.timestamp, _ = miscellaneous.D64func(b[12:20])
	a.prevBlock = miscellaneous.Dup(b[20:52])
	a.stateRoot = miscellaneous.Dup(b[52:84])
	a.transactionsRoot = miscellaneous.Dup(b[84:116])
	return b[BLOCK_HEADER_SIZE:], nil
}

func (a *Block) Show() ([]byte, error) {
	buf := []byte{}
	if tmp, err := a.header.Show(); err != nil {
		return []byte{}, err
	} else {
		buf = append(buf, tmp...)
	}
	if tmp, err := a.transactions.Show(); err != nil {
		return []byte{}, err
	} else {
		buf = append(buf, tmp...)
	}
	return buf, nil
}

func (a *Block) Read(b []byte) ([]byte, error) {
	var err error

	if a.header == nil {
		a.header = new(BlockHeader)
	}
	if b, err = a.header.Read(b); err != nil {
		return []byte{}, err
	}
	a.transactions = new(transaction.Transactions)
	if b, err = a.transactions.Read(b); err != nil {
		return []byte{}, err
	}
	return b, nil
}
