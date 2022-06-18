package chain

import (
	"errors"

	"github.com/tjfoc/tjfoc/core/block"
	"github.com/tjfoc/tjfoc/core/crypt"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/queue"
	"github.com/tjfoc/tjfoc/core/store"
	"github.com/tjfoc/tjfoc/core/transaction"
)

/*
约定俗成的一些东西:
	height - 区块链高度 -- 最后一个区块的高度+1
	区块height - 区块hash
	b + 区块hash - 区块数据
	t + 交易hash - 区块高度 - 交易编号 - 交易数据

*/
var chainHeight = []byte("height")
var memberList = []byte("memberList")

type chainStore struct {
	db store.Store
	c  crypt.Crypto
}

func New(c crypt.Crypto, path string) (*chainStore, error) {
	if db, err := store.NewDb(path); err != nil {
		return nil, err
	} else {
		return &chainStore{
			c:  c,
			db: db,
		}, nil
	}
}

func (a *chainStore) GetMemberList() (map[string]*PeerInfo, error) {
	b := make(map[string]*PeerInfo)
	data, err := a.db.Get(memberList)
	if err != nil {
		return nil, err
	}
	if err := miscellaneous.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	return b, nil
}

func (a *chainStore) SaveMemberList(b map[string]*PeerInfo) error {
	data, err := miscellaneous.Marshal(b)
	if err != nil {
		return err
	}
	return a.db.Set(memberList, data)
}

// 返回链的高度
func (a *chainStore) Height() uint64 {
	if v, err := a.db.Get(chainHeight); err != nil {
		return 0
	} else {
		if h, err := miscellaneous.D64func(v[:8]); err != nil {
			return 0
		} else {
			return h
		}
	}
}

func (a *chainStore) GetBlockByHeight(b uint64) *block.Block {
	if k, err := a.db.Get(miscellaneous.E64func(b)); err != nil {
		return nil
	} else {
		return a.GetBlockByHash(k)
	}
}

func (a *chainStore) GetBlockByHash(k []byte) *block.Block {
	if len(k) != a.c.Size() {
		return nil
	}
	if v, err := a.db.Get(append([]byte{'b'}, k...)); err != nil {
		return nil
	} else {
		b := new(block.Block)
		if _, err := b.Read(v); err != nil {
			return nil
		} else {
			return b
		}
	}
}

func (a *chainStore) GetTransaction(k []byte) *transaction.Transaction {
	if len(k) != a.c.Size() {
		return nil
	}
	if v, err := a.db.Get(append([]byte{'t'}, k...)); err != nil {
		return nil
	} else {
		t := new(transaction.Transaction)
		if _, err := t.Read(v[12:]); err != nil {
			return nil
		} else {
			return t
		}
	}
}

// 返回交易所在的区块高度
func (a *chainStore) GetTransactionBlock(k []byte) (uint64, error) {
	if len(k) != a.c.Size() {
		return 0, errors.New("GetTransactionBlock: unsupport hash function")
	}
	if v, err := a.db.Get(append([]byte{'t'}, k...)); err != nil {
		return 0, err
	} else {
		return miscellaneous.D64func(v[:8])
	}
}

// 返回交易在区块中的编号
func (a *chainStore) GetTransactionIndex(k []byte) (uint32, error) {
	if len(k) != a.c.Size() {
		return 0, errors.New("GetTransactionIndex: unsupport hash function")
	}
	if v, err := a.db.Get(append([]byte{'t'}, k...)); err != nil {
		return 0, err
	} else {
		return miscellaneous.D32func(v[8:12])
	}
}

/*
区块存储格式:
	height 		-- hash data
	hash data   -- block data
*/
func (a *chainStore) AddBlock(b *block.Block) error {
	if k, err := b.Hash(a.c); err != nil {
		return err
	} else if v, err := b.Show(); err != nil {
		return err
	} else {
		bc := queue.New()
		txs := b.TransactionList()
		for i, v := range *txs {
			if k0, v0, err := a.addTransaction(&v, b.Height(), uint32(i)); err != nil {
				return nil
			} else {
				bc.PushBack(store.NewFactor(k0, v0))
			}
		}
		bc.PushBack(store.NewFactor(append([]byte{'b'}, k...), v))
		bc.PushBack(store.NewFactor(miscellaneous.E64func(b.Height()), k))
		bc.PushBack(store.NewFactor(chainHeight, miscellaneous.E64func(b.Height()+1)))
		return a.db.BatchWrite(bc)
	}
}

/*
交易存储格式:
	hash 			-- 区块高度(8 byte) + 交易编号(4 byte) + 交易数据
func (a *ChainStore) AddTransaction(b *transaction.Transaction, height uint64, index uint32) error {
	if k, err := b.Hash(); err != nil {
		return err
	} else if v, err := b.Show(); err != nil {
		return err
	} else {
		buf = []byte{}
		buf = append(buf, miscellaneous.D64func(height)...)
		buf = append(buf, miscellaneous.D32func(index)...)
		buf = append(buf, v...)
		a.Set(append([]byte{'t'}, k...), v)
		return nil
	}
}
*/

func (a *chainStore) addTransaction(b *transaction.Transaction, height uint64, index uint32) ([]byte, []byte, error) {
	if k, err := b.Hash(a.c); err != nil {
		return []byte{}, []byte{}, err
	} else if v, err := b.Show(); err != nil {
		return []byte{}, []byte{}, err
	} else {
		buf := []byte{}
		buf = append(buf, miscellaneous.E64func(height)...)
		buf = append(buf, miscellaneous.E32func(index)...)
		buf = append(buf, v...)
		return append([]byte{'t'}, k...), buf, nil
	}
}
