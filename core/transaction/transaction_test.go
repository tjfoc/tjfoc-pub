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
package transaction

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
)

func TransactionsPrint(a *Transactions) {
	for i, v := range *a {
		fmt.Printf("********transactions********\n")
		fmt.Printf("++++%d\n", i)
		TransactionPrint(&v)
	}
}

func TransactionPrint(a *Transaction) {
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

func TestTransaction(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	c := crypt.New(sha256.New(), priv, crypto.SHA256)

	a := New([]byte("set a 1; set a 2; set b 3"), [][]byte{}, 0, uint64(time.Now().Unix()))
	a.AddRecords([][]byte{
		[]byte("a = 2"),
		[]byte("b = 3"),
	})

	a.Sign(c)
	fmt.Printf("++++++verify %v+++++\n", a.Verify(c))
	h, _ := a.Hash(c)
	fmt.Printf("+++++++++hash++++++++++\n")
	fmt.Printf("\t%v\n", h)
	fmt.Printf("+++++++++++++++++++++++\n")

	b := New([]byte("set c 2; set c 5; set d 3; set d 6"), [][]byte{}, 0, uint64(time.Now().Unix()))
	b.AddRecords([][]byte{
		[]byte("c = 5"),
		[]byte("d = 6"),
	})

	s := News()
	s.AddTransaction(a)
	s.AddTransaction(b)

	TransactionsPrint(s)

	data, _ := s.Show()

	s0 := new(Transactions)
	s0.Read(data)

	TransactionsPrint(s0)
}
