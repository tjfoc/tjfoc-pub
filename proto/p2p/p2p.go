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
		probeQueue: &p2pQueue{
			tranList: []*tranP2p{},
			notice:   make(chan bool, 1024),
		},
	}
	for _, v := range p2pRegistry_0 {
		P2P.sp.RegisterFunc(VERSION, v.id, v.spFunc)
	}
	return P2P
}

func (a *p2pProto) Run() error {
	go a.probeRun()
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
			a.addProbe(PROBE, buf, id)
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
			a.addProbe(BLOCK_PROBE, buf, id)
		}
		txs := []byte{}
		hashSize := a.chain.Size()
		txCount := 0
		for j := BLOCK_HEIGHT_SIZE; j < len(data); j += hashSize {
			if !a.chain.TransactionElem(data[j:j+hashSize]) && a.consensusAPI.IsLeader() {
				txs = append(txs, data[j:j+hashSize]...)
				txCount++
				//a.addProbe(TRANSACTION_PROBE, data[j:j+hashSize], id)
			}
		}
		if len(txs) > 0 {
			//logger.Infof("addProbe TRANSACTION_PROBE txCount:%d peer:%x", txCount, id)
			a.addProbe(TRANSACTION_PROBE, txs, id)
		}
	}
	return 0
}

func (a *p2pProto) addProbe(typ uint32, data, id []byte) {
	a.probeQueue.lock.Lock()
	a.probeQueue.tranList = append(a.probeQueue.tranList, &tranP2p{
		typ:  typ,
		id:   miscellaneous.Dup(id),
		data: miscellaneous.Dup(data),
	})
	a.probeQueue.lock.Unlock()

	if len(a.probeQueue.notice) < 1024 {
		a.probeQueue.notice <- true
	}
}

func (a *p2pProto) probeRun() {
	for {
		select {
		case <-a.probeQueue.notice:
			for len(a.probeQueue.notice) > 0 {
				<-a.probeQueue.notice
			}
			a.probeQueue.lock.Lock()
			j := len(a.probeQueue.tranList)
			a.probeQueue.lock.Unlock()

			for i := 0; i < j; i++ {
				a.sp.SendInstruction(VERSION, a.probeQueue.tranList[i].typ, a.probeQueue.tranList[i].data, a.probeQueue.tranList[i].id)
				// if err != nil {
				// 	logger.Errorf("send 0x%x err %v", a.probeQueue.tranList[i].typ, err)
				// }
				// time.Sleep(10 * time.Millisecond)
			}

			a.probeQueue.lock.Lock()
			a.probeQueue.tranList = a.probeQueue.tranList[j:]
			length := len(a.probeQueue.tranList)
			a.probeQueue.lock.Unlock()

			if length > 0 && len(a.probeQueue.notice) == 0 {
				a.probeQueue.notice <- true
			}
		}
	}
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
			a.addProbe(BLOCK_INFO, tmp, id)
		}
		return 0
	}
}

func transactionInfo(usrData interface{}, id, data []byte) int {
	a := P2P
	var okc, errc int
	for {
		if len(data) < 32 {
			break
		}
		tx := new(transaction.Transaction)
		tmp, err := tx.Read(data)
		data = tmp
		if err != nil {
			return -1
		}
		err = a.chain.TransactionAdd(tx)
		if err != nil {
			errc++
		} else {
			okc++
		}
	}
	logger.Infof("pull transaction %d (repeat:%d) from %x", okc, errc, id)
	return 0
}

func transactionProbe(usrData interface{}, id, data []byte) int {
	a := P2P
	buf := []byte{}
	size := a.chain.Size()
	txCount := 0
	for ; len(data) >= size; data = data[size:] {
		if tx := a.chain.GetTransactionFromBuffer(data[:size]); tx != nil {
			if tmp, err := tx.Show(); err != nil {
				return -1
			} else {
				buf = append(buf, tmp...)
				txCount++
			}
		} else {
			return -1
		}
	}
	logger.Infof("send transaction %d to peer %x", txCount, id)
	a.addProbe(TRANSACTION_INFO, buf, id)
	return 0
	/*
		if len(data) != a.chain.Size() {
			return -1
		}
		if tx := a.chain.GetTransactionFromBuffer(data); tx == nil {
			return -1
		} else {
			if tmp, err := tx.Show(); err != nil {
				return -1
			} else {
				a.addProbe(TRANSACTION_INFO, tmp, id)
			}
			return 0
		}
	*/
}
