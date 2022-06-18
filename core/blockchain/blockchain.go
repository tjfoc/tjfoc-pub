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
	"bytes"
	"errors"
	"time"

	"github.com/tjfoc/tjfoc/core/block"
	"github.com/tjfoc/tjfoc/core/chaincode"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/crypt"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/store/chain"
	"github.com/tjfoc/tjfoc/core/transaction"
	"github.com/tjfoc/tjfoc/core/worldstate"
)

var logger = flogging.MustGetLogger("blockchain")

func GenesisBlock(timestamp uint64, des []byte) *block.Block {
	txs := transaction.News()
	tx, _ := transaction.New(nil, des, 0, 0, 0)
	// txs.AddTransaction(transaction.New(des, [][]byte{}, 0, 0))
	txs.AddTransaction(tx)
	return block.New(block.NewHeader(0, timestamp, make([]byte, block.BLOCK_HASH_SIZE),
		make([]byte, block.BLOCK_HASH_SIZE), make([]byte, block.BLOCK_HASH_SIZE)), txs)
}

func New(bc chain.Chain, c crypt.Crypto, b *block.Block, usrData interface{},
	callBack BcCallBackFunc) *blockChain {
	var last *block.Block

	if bc.Height() == 0 {
		if b == nil || b.Height() != 0 {
			return nil
		}
		if err := bc.AddBlock(b); err != nil {
			return nil
		}
	}
	if last = bc.GetBlockByHeight(bc.Height() - 1); last == nil {
		return nil
	}
	if hashData, err := last.Hash(c); err != nil {
		return nil
	} else {
		bc := &blockChain{
			cryptoPlug: c,
			blockchain: bc,
			last:       last,
			isPack:     true,
			isSync:     false,
			usrData:    usrData,
			lastHash:   hashData,
			callBack:   callBack,
			threshold:  last.Height(),
			packBlocks: &blockQueue{
				blockList: []*BlockTxs{},
				notice:    make(chan bool, 1024),
			},
			syncBlocks: &blockQueue{
				blockList: []*BlockTxs{},
				notice:    make(chan bool, 1024),
			},
			transactionBuffer: make(map[string]*transaction.Transaction),
		}
		go bc.syncRun()
		return bc
	}
}

func (a *blockChain) GetTransactionNumber() int {
	return len(a.transactionBuffer)
}

func (a *blockChain) TransactionElem(k []byte) bool {
	a.Lock()
	defer a.Unlock()
	if _, ok := a.transactionBuffer[string(k)]; ok {
		return true
	}
	if _, err := a.GetTransactionIndex(k); err == nil {
		return true
	}
	return false
}

func (a *blockChain) GetTransactionFromBuffer(k []byte) *transaction.Transaction {
	a.Lock()
	defer a.Unlock()
	if v, ok := a.transactionBuffer[string(k)]; ok {
		return v
	} else {
		return nil
	}
}

func (a *blockChain) GetTransactionList() []byte {
	a.Lock()
	defer a.Unlock()
	buf := []byte{}
	for k, _ := range a.transactionBuffer {
		if _, err := a.GetTransactionIndex([]byte(k)); err == nil {
			delete(a.transactionBuffer, k)
			continue
		}
		buf = append(buf, []byte(k)...)
		if len(buf) > BLOCK_SIZE {
			break
		}
	}
	return buf
}

func (a *blockChain) Pack(timestamp uint64) *block.Block {
	a.Lock()
	defer a.Unlock()
	if !a.isPack {
		return nil
	}
	lim := BLOCK_SIZE
	txs := transaction.News()

	//
	if len(a.transactionBuffer) > 0 {
		logger.Infof("begin pack,buffer len:%d", len(a.transactionBuffer))
	}

	for k, v := range a.transactionBuffer {
		if _, err := a.GetTransactionIndex([]byte(k)); err == nil {
			delete(a.transactionBuffer, string(k))
			continue
		}
		if d, err := v.Show(); err == nil {
			if lim = lim - len(d); lim > 0 {
				*txs = append(*txs, *v)
			} else {
				break
			}
		}
	}
	if len(*txs) > 0 {
		logger.Infof("end pack,txs:%d", len(*txs))
	}
	if len(*txs) == 0 {
		return nil
	}
	mkt := txs.MkMerkleTree(a.cryptoPlug)
	if mkt == nil {
		return nil
	}
	prevBlock := miscellaneous.Dup(a.lastHash)
	stateRoot := worldstate.GetWorldState().GetRootHash()
	transactionsRoot := miscellaneous.Dup(mkt.GetMtHash())
	header := block.NewHeader(a.last.Height()+1, timestamp, prevBlock, []byte(stateRoot), transactionsRoot)
	return block.New(header, txs)
}

func (a *blockChain) PackAddBlock(b *block.Block) error {
	if b == nil {
		return errors.New("AddBlock: null pointer")
	}

	a.Lock()
	threshold := a.threshold
	a.Unlock()

	if b.Height() <= threshold {
		logger.Infof("blockHeight:%d <= threshold:%d return", b.Height(), threshold)
		return nil
	}
	btxs := &BlockTxs{
		UsrData: a.usrData,
	}

	txs := *b.TransactionList()
	for _, v := range txs {
		a.callBack(btxs, &v) // 处理交易
	}
	btxs.bk = b
	btxs.txs = txs

	a.Lock()
	a.isPack = false
	a.Unlock()
	defer func() {
		a.Lock()
		a.isPack = true
		a.Unlock()
	}()

	for {
		a.Lock()
		last := a.last
		a.Unlock()
		if btxs.bk.Height() <= threshold {
			logger.Infof("blockHeight:%d <= threshold:%d", btxs.bk.Height(), threshold)
			return nil
		} else if btxs.bk.Height() > last.Height()+1 {
			a.Lock()
			a.isSync = true
			a.Unlock()
			time.Sleep(500 * time.Millisecond)
			logger.Infof("pack waiting syncpack lastHeight:%d need:%d", last.Height()+1, btxs.bk.Height())
		} else {
			break
		}
	}

	a.Lock()
	last := a.last
	lastHash := miscellaneous.Dup(a.lastHash)
	a.Unlock()
	if last.Height()+1 != btxs.bk.Height() {
		logger.Infof("blochHeight:%d need:%d return", btxs.bk.Height(), last.Height()+1)
		return nil
	}
	if bytes.Compare(btxs.bk.PreviousBlock(), lastHash) != 0 {
		logger.Infof("block previousBlock:%x != lastHash:%x", btxs.bk.PreviousBlock(), lastHash)
		return nil
	}

	a.wlock.Lock()
	t := time.Now()
	logger.Infof("begin pack block h-index:%d txs:%d", last.Height()+1, len(btxs.Hashs))
	rst, kv, action, ok := chaincode.GetDelieverInstance().RecvOP(btxs.SmartContract, btxs.SmartContractArgs, btxs.Hashs)
	logger.Infof("RecvOP txs:%d take:%v ok:%v", len(btxs.Hashs), time.Now().Sub(t), ok)

	if ok {
		a.Lock()
		for _, t := range btxs.txs {
			h, _ := t.Hash(a.cryptoPlug)
			for k, v := range rst {
				if bytes.Compare([]byte(k), h) == 0 {
					t.AddRecords([][]byte{miscellaneous.Dup([]byte(v))})
					break
				}
			}
			delete(a.transactionBuffer, string(h))
		}
		//a.Unlock()
		hashData, _ := btxs.bk.Hash(a.cryptoPlug)

		//a.Lock()
		a.last = btxs.bk
		miscellaneous.Memmove(a.lastHash, hashData)
		a.blockchain.AddBlock(btxs.bk)
		a.Unlock()

		logger.Infof("pack left transactionBuffer:%v; packBlocks:%v", len(a.transactionBuffer), len(a.packBlocks.blockList))
		//注意判断push状态
		pushstatus := worldstate.GetWorldState().Push(btxs.bk.Height()-1, string(btxs.bk.WorldStateRoot()), kv, action)
		logger.Infof("pack pushstatus:%d", pushstatus)
	}
	logger.Infof("end pack block h-index:%d", last.Height()+1)
	a.wlock.Unlock()
	return nil
}

func (a *blockChain) SyncAddBlock(b *block.Block) error {
	if b == nil {
		return errors.New("AddBlock: null pointer")
	}

	a.Lock()
	defer a.Unlock()
	if !a.isSync || b.Height() > a.threshold {
		return nil
	}
	btxs := &BlockTxs{
		UsrData: a.usrData,
	}
	txs := *b.TransactionList()
	for _, v := range txs {
		a.callBack(btxs, &v) // 处理交易
	}

	btxs.bk = b
	btxs.txs = txs
	a.syncBlocks.blockList = append(a.syncBlocks.blockList, btxs)
	a.syncBlocks.notice <- true
	return nil
}

func (a *blockChain) syncRun() {
	for {
		select {
		case <-a.syncBlocks.notice:
			for len(a.syncBlocks.notice) > 0 {
				<-a.syncBlocks.notice
			}
			a.Lock()
			i := 0
			j := len(a.syncBlocks.blockList)
			a.Unlock()
			for ; i < j; i++ {
				btxs := a.syncBlocks.blockList[i]
				a.Lock()
				last := a.last
				lastHash := miscellaneous.Dup(a.lastHash)
				a.Unlock()

				if a.isSync && last.Height()+1 == btxs.bk.Height() && bytes.Compare(btxs.bk.PreviousBlock(), lastHash) == 0 {
					a.wlock.Lock()
					t := time.Now()
					logger.Infof("begin sync block h-index:%d", last.Height()+1)
					rst, kv, action, ok := chaincode.GetDelieverInstance().RecvOP(btxs.SmartContract, btxs.SmartContractArgs, btxs.Hashs)
					logger.Infof("sync RecvOP txs:%d take:%v ok:%v h-index:%d", len(btxs.Hashs), time.Now().Sub(t), ok, last.Height()+1)
					if ok {
						for idx := 0; idx < len(btxs.txs); idx++ {
							h, _ := btxs.txs[idx].Hash(a.cryptoPlug)
							for k, v := range rst {
								if bytes.Compare([]byte(k), h) == 0 {
									btxs.txs[idx].AddRecords([][]byte{miscellaneous.Dup([]byte(v))})
									break
								}
							}
						}

						a.Lock()
						for _, v := range btxs.txs {
							k, _ := v.Hash(a.cryptoPlug)
							delete(a.transactionBuffer, string(k))
						}
						//a.Unlock()
						hashData, _ := btxs.bk.Hash(a.cryptoPlug)

						//a.Lock()
						a.last = btxs.bk
						miscellaneous.Memmove(a.lastHash, hashData)
						a.blockchain.AddBlock(btxs.bk)
						a.Unlock()

						logger.Infof("sync pack left transactionBuffer:%d; syncBlocks:%d", len(a.transactionBuffer), len(a.syncBlocks.blockList))
						//注意判断push状态
						pushstatus := worldstate.GetWorldState().Push(btxs.bk.Height()-1, string(btxs.bk.WorldStateRoot()), kv, action)
						logger.Infof("sync pushstatus:%d", pushstatus)
					}
					logger.Infof("end sync block h-index:%d", last.Height()+1)
					a.wlock.Unlock()
				}
			}
			a.Lock()
			a.syncBlocks.blockList = a.syncBlocks.blockList[j:]
			if a.last.Height() == a.threshold {
				logger.Errorf("******************* sync set false*******************\n")
				a.isSync = false
			}
			length := len(a.syncBlocks.blockList)
			a.Unlock()
			if length > 0 && len(a.syncBlocks.notice) == 0 {
				a.syncBlocks.notice <- true
			}
		}
	}
}

func (a *blockChain) TransactionAdd(b *transaction.Transaction) error {
	a.Lock()
	defer a.Unlock()
	if k, err := b.Hash(a.cryptoPlug); err != nil {
		return err
	} else {
		if _, ok := a.transactionBuffer[string(k)]; ok {
			return errors.New("AddTransaction: exist")
		}
		if _, err := a.GetTransactionIndex(k); err == nil {
			return errors.New("AddTransaction: exist")
		}
		a.transactionBuffer[string(k)] = b
		return nil
	}
}
