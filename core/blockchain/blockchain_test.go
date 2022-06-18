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
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/tjfoc/tjfoc/core/block"
	"github.com/tjfoc/tjfoc/core/crypt"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/store/chain"
	"github.com/tjfoc/tjfoc/core/transaction"
)

func BlockPrint(a *block.Block) {
	fmt.Printf("++++++++version %d, height = %d, timestamp = %d\n",
		a.Version(), a.Height(), a.Timestamp())
	fmt.Printf("++++++++++prev = %s\n", miscellaneous.Bytes2HexString(a.PreviousBlock()))
	fmt.Printf("++++++++++stateRoot = %s\n", miscellaneous.Bytes2HexString(a.WorldStateRoot()))
	fmt.Printf("++++++++++transactionsRoot = %s\n", miscellaneous.Bytes2HexString(a.TransactionsRoot()))
	txs := a.TransactionList()
	if txs != nil {
		TransactionsPrint(txs)
	}
}

func TransactionsPrint(a *transaction.Transactions) {
	for i, v := range *a {
		fmt.Printf("\t********transactions %d********\n", i)
		TransactionPrint(&v)
	}
}

func TransactionPrint(a *transaction.Transaction) {
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

func blockchainPrint(a BlockChain) {
	for i := uint64(0); i < a.Height(); i++ {
		BlockPrint(a.GetBlockByHeight(i))
	}
}

func testCallBack(usrdata interface{}, tx *transaction.Transaction) int {
	return 0
}

func TestBlockchain(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	c := crypt.New(sha256.New(), priv, crypto.SHA256)
	ch, _ := chain.New(c, "test")
	b := GenesisBlock(uint64(time.Now().Unix()), []byte("test"))
	bc := New(ch, c, b, b, testCallBack)
	count := 0
	bcount := 0

	tmp := bc.Pack(uint64(time.Now().Unix()))
	bc.BlockAdd(tmp)
	bcount++

	for {
		select {
		case <-time.After(1 * time.Millisecond):
			a := transaction.New([]byte(fmt.Sprintf("id_%d = %d", count, count)), [][]byte{}, 0, uint64(time.Now().Unix()))
			a.AddRecords([][]byte{
				[]byte(fmt.Sprintf("id_%d = %d", count, count)),
			})
			a.Sign(c)
			count++
			bc.TransactionAdd(a)
			if count%1000 == 0 {
				b := bc.Pack(uint64(time.Now().Unix()))
				bc.BlockAdd(b)
				bcount++
				if bcount%10 == 0 {
					blockchainPrint(bc)
				}
			}
		}
	}
}
