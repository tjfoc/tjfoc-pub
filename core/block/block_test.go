package block

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/tjfoc/tjfoc/core/crypt"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/transaction"
)

func BlockPrint(a *Block) {
	fmt.Printf("++++++++version %d, height = %d, timestamp = %d\n",
		a.header.version, a.header.height, a.header.timestamp)
	fmt.Printf("++++++++++prev = %s\n", miscellaneous.Bytes2HexString(a.header.prevBlock))
	fmt.Printf("++++++++++stateRoot = %s\n", miscellaneous.Bytes2HexString(a.header.stateRoot))
	fmt.Printf("++++++++++transactionsRoot = %s\n", miscellaneous.Bytes2HexString(a.header.transactionsRoot))
	if a.transactions != nil {
		TransactionsPrint(a.transactions)
	}
}

func TransactionsPrint(a *transaction.Transactions) {
	for i, v := range *a {
		fmt.Printf("********transactions********\n")
		fmt.Printf("++++%d\n", i)
		TransactionPrint(&v)
	}
}

func TransactionPrint(a *transaction.Transaction) {
	fmt.Printf("\t********transaction********\n")
	fmt.Printf("\t\tversion = %v, timestamp = %v\n", a.Version(), a.Timestamp())
	fmt.Printf("\t\tsign = %v\n", a.SignData())
	fmt.Printf("\t\tsmartcontract = %s\n", string(a.SmartContract()))

	fmt.Printf("\t\targs:\n")
	args := a.SmartContractArgs()
	for i, v := range args {
		fmt.Printf("\t\t\t%d = %s\n", i, string(v))
	}

	fmt.Printf("\t\trecords:\n")
	records := a.Records()
	for i, v := range records {
		fmt.Printf("\t\t\t%d = %s\n", i, string(v))
	}
}

func TestBlock(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	c := crypt.New(sha256.New(), priv, crypto.SHA256)

	txs := transaction.News()
	a := transaction.New([]byte("set a 1; set a 2; set b 3"), [][]byte{}, 0, uint64(time.Now().Unix()))
	a.AddRecords([][]byte{
		[]byte("a = 2"),
		[]byte("b = 3"),
	})

	b := transaction.New([]byte("set c 2; set c 5; set d 3; set d 6"), [][]byte{}, 0, uint64(time.Now().Unix()))
	b.AddRecords([][]byte{
		[]byte("c = 5"),
		[]byte("d = 6"),
	})
	txs.AddTransaction(a)
	txs.AddTransaction(b)

	header := new(BlockHeader)
	if mktree := txs.MkMerkleTree(c); mktree != nil {
		header = NewHeader(0, 1, make([]byte, 32), make([]byte, 32), mktree.GetMtHash())
	} else {
		header = NewHeader(0, 1, make([]byte, 32), make([]byte, 32), make([]byte, 32))
	}

	b0 := New(header, txs)

	BlockPrint(b0)

	data, _ := b0.Show()

	b1 := new(Block)
	b1.Read(data)

	BlockPrint(b1)

	b2 := New(header, nil)
	BlockPrint(b2)
}
