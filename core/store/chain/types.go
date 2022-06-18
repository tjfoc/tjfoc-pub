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
package chain

import (
	"github.com/tjfoc/tjfoc/core/block"
	"github.com/tjfoc/tjfoc/core/transaction"
)

type PeerInfo struct {
	Typ  uint32
	Addr string
}

type Chain interface {
	Height() uint64
	AddBlock(*block.Block) error
	GetBlockByHash([]byte) *block.Block
	GetBlockByHeight(uint64) *block.Block
	GetTransactionIndex([]byte) (uint32, error)
	GetTransactionBlock([]byte) (uint64, error)
	GetTransaction([]byte) *transaction.Transaction

	SaveMemberList(map[string]*PeerInfo) error
	GetMemberList() (map[string]*PeerInfo, error)
}
