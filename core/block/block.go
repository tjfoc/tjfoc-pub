package block

import (
	"github.com/tjfoc/tjfoc/core/crypt"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/transaction"
)

func New(header *BlockHeader, transactions *transaction.Transactions) *Block {
	return &Block{
		header:       header,
		transactions: transactions,
	}
}

func NewHeader(height uint64, timestamp uint64, prevBlock, stateRoot, transactionsRoot []byte) *BlockHeader {
	if len(prevBlock) != BLOCK_HASH_SIZE || len(stateRoot) != BLOCK_HASH_SIZE ||
		len(transactionsRoot) != BLOCK_HASH_SIZE {
		return nil
	}
	return &BlockHeader{
		version:          0, // 默认版本号0
		height:           height,
		timestamp:        timestamp,
		prevBlock:        miscellaneous.Dup(prevBlock),
		stateRoot:        miscellaneous.Dup(stateRoot),
		transactionsRoot: miscellaneous.Dup(transactionsRoot),
	}
}

func (a *Block) Hash(c crypt.Crypto) ([]byte, error) {
	if buf, err := a.Show(); err != nil {
		return []byte{}, nil
	} else {
		return miscellaneous.GenHash(c.HashFunc().New(), buf)
	}
}
