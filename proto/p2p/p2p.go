package p2p

import (
	"fmt"
	"net"
	"time"

	"github.com/tjfoc/tjfoc/core/block"
	"github.com/tjfoc/tjfoc/core/blockchain"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/consensus"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/transaction"
	"github.com/tjfoc/tjfoc/proto"
)

var P2P *p2pProto

var logger = flogging.MustGetLogger("p2p")

func New(sp proto.SP, cycle time.Duration, consensusAPI consensus.Consensus, chain blockchain.BlockChain) *p2pProto {
	P2P = &p2pProto{
		sp:           sp,
		cycle:        cycle,
		chain:        chain,
		consensusAPI: consensusAPI,
		exitCh:       make(map[string]chan bool),
	}
	for _, v := range p2pRegistry_0 {
		P2P.sp.RegisterFunc(VERSION, v.id, v.spFunc)
	}
	return P2P
}

func (a *p2pProto) Run() error {
	return a.sp.Run()
}

func (a *p2pProto) UnregisterPeer(id []byte) error {
	a.lock.Lock()
	ch, ok := a.exitCh[string(id)]
	a.lock.Unlock()
	if !ok {
		return fmt.Errorf("UnregisterPeer: %x not exist\n", id)
	}
	delete(a.exitCh, string(id))
	ch <- true
	return nil
}

func (a *p2pProto) RegisterPeer(id []byte, addr net.Addr) bool {
	ch := make(chan bool)
	remoteId := miscellaneous.Dup(id)
	a.sp.RegisterPeer(id, addr)
	a.lock.Lock()
	a.exitCh[string(id)] = ch
	a.lock.Unlock()
	go a.p2pRun(remoteId)
	return true
}

func (a *p2pProto) UnregisterFunc(version uint32, class uint32) error {
	return a.sp.UnregisterFunc(version, class)
}

func (a *p2pProto) RegisterFunc(version uint32, class uint32, callback proto.SpFunc) error {
	return a.sp.RegisterFunc(version, class, callback)
}

func (a *p2pProto) SendInstruction(version uint32, class uint32, data []byte, id []byte) error {
	return a.sp.SendInstruction(version, class, data, id)
}

func (a *p2pProto) GetMemberList() map[string]string {
	return a.sp.GetMemberList()
}

func (a *p2pProto) p2pRun(id []byte) {
	a.lock.Lock()
	ch := a.exitCh[string(id)]
	a.lock.Unlock()
	count := uint32(0)
	for {
		select {
		case <-ch:
			return
		case <-time.After(a.cycle):
			buf := []byte{}
			buf = append(buf, miscellaneous.E64func(a.chain.Height())...) // 区块高度
			if !a.consensusAPI.IsLeader() {
				buf = append(buf, a.chain.GetTransactionList()...) // 交易缓冲区的交易hash集合
			}
			a.sp.SendInstruction(VERSION, PROBE, buf, id)
			count++
		}
	}
}

func probe(usrData interface{}, id, data []byte) int {
	if len(data) < BLOCK_HEIGHT_SIZE {
		return -1
	}
	if h0, err := miscellaneous.D64func(data[:BLOCK_HEIGHT_SIZE]); err != nil { // h0 = peer height
		return -1
	} else {
		a := P2P
		if h1 := a.chain.Height(); h1 < h0 {
			buf := []byte{}
			buf = append(buf, miscellaneous.E64func(h1)...)
			a.sp.SendInstruction(VERSION, BLOCK_PROBE, buf, id)
		}
		hashSize := a.chain.Size()
		for j := BLOCK_HEIGHT_SIZE; j < len(data); j += hashSize {
			if !a.chain.TransactionElem(data[j:j+hashSize]) && a.consensusAPI.IsLeader() {
				a.sp.SendInstruction(VERSION, TRANSACTION_PROBE, data[j:j+hashSize], id)
			}
		}
	}
	return 0
}

func blockInfo(usrData interface{}, id, data []byte) int {
	b := new(block.Block)
	if _, err := b.Read(data); err != nil {
		logger.Errorf("********************recv block %x: %v*****************\n", id, err)
		return -1
	} else {
		a := P2P
		if err := a.chain.SyncAddBlock(b); err != nil {
			logger.Errorf("*******************add block height %v: %v*****************", b.Height(), err)
			return -1
		}
		return 0
	}
}

func blockProbe(usrData interface{}, id, data []byte) int {
	if len(data) < BLOCK_HEIGHT_SIZE {
		return -1
	}
	if height, err := miscellaneous.D64func(data[:BLOCK_HEIGHT_SIZE]); err != nil {
		return -1
	} else {
		a := P2P
		if height >= a.chain.Height() {
			return -1
		}
		if b := a.chain.GetBlockByHeight(height); b == nil {
			return -1
		} else if tmp, err := b.Show(); err == nil {
			a.sp.SendInstruction(VERSION, BLOCK_INFO, tmp, id)
		}
		return 0
	}
}

func transactionInfo(usrData interface{}, id, data []byte) int {
	tx := new(transaction.Transaction)
	if _, err := tx.Read(data); err != nil {
		return -1
	}
	a := P2P
	a.chain.TransactionAdd(tx)
	return 0
}

func transactionProbe(usrData interface{}, id, data []byte) int {
	a := P2P
	if len(data) != a.chain.Size() {
		return -1
	}
	if tx := a.chain.GetTransactionFromBuffer(data); tx == nil {
		return -1
	} else {
		if tmp, err := tx.Show(); err != nil {
			return -1
		} else {
			a.sp.SendInstruction(VERSION, TRANSACTION_INFO, tmp, id)
		}
		return 0
	}
}
