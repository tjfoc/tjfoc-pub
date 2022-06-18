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
	"crypto/rand"

	"github.com/tjfoc/tjfoc/core/crypt"
	"github.com/tjfoc/tjfoc/core/merkle"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
)

func newheader(version uint32, timestamp uint64) *transactionHeader {
	return &transactionHeader{
		version:   version,
		timestamp: timestamp,
	}
}

func newSmartContract(smartContract []byte, smartContractArgs [][]byte) *transactionSmartContract {
	return &transactionSmartContract{
		smartContract:     miscellaneous.Dup(smartContract),
		smartContractArgs: newSmartContractArgs(smartContractArgs),
	}
}

func newSmartContractArgs(smartContractArgs [][]byte) *transactionSmartContractArgs {
	args := transactionSmartContractArgs{}
	for _, v := range smartContractArgs {
		args = append(args, transactionSmartContractArg{
			smartContractArg: miscellaneous.Dup(v),
		})
	}
	return &args
}

func News() *Transactions {
	return &Transactions{}
}

func New(smartContract []byte, smartContractArgs [][]byte, version uint32, timestamp uint64) *Transaction {
	return &Transaction{
		header:        newheader(version, timestamp),
		smartContract: newSmartContract(smartContract, smartContractArgs),
	}
}

func (a *Transactions) AddTransaction(b *Transaction) error {
	*a = append(*a, *b)
	return nil
}

func (a *Transactions) MkMerkleTree(c crypt.Crypto) *merkle.MerkleTree {
	buf := [][]byte{}
	for _, v := range *a {
		if tmp, err := v.Hash(c); err != nil {
			return nil
		} else {
			buf = append(buf, tmp)
		}
	}
	return merkle.MkMerkleTree(c.HashFunc().New(), buf)
}

func (a *Transaction) AddSign(b []byte) error {
	if len(b) > 0 {
		a.sign = miscellaneous.Dup(b)
	}
	return nil
}

func (a *Transaction) AddRecords(b [][]byte) error {
	records := transactionRecords{}
	for _, v := range b {
		records = append(records, transactionRecord{
			record: miscellaneous.Dup(v),
		})
	}
	a.records = &records
	return nil
}

func (a *Transaction) Hash(c crypt.Crypto) ([]byte, error) {
	buf := []byte{}
	if tmp, err := a.header.Show(); err != nil {
		return []byte{}, err
	} else {
		buf = append(buf, tmp...)
	}
	if tmp, err := a.smartContract.Show(); err != nil {
		return []byte{}, err
	} else {
		buf = append(buf, tmp...)
	}
	return miscellaneous.GenHash(c.HashFunc().New(), buf)
}

func (a *Transaction) Sign(c crypt.Crypto) error {
	buf := []byte{}
	if tmp, err := a.header.Show(); err != nil {
		return err
	} else {
		buf = append(buf, tmp...)
	}
	if tmp, err := a.smartContract.Show(); err != nil {
		return err
	} else {
		buf = append(buf, tmp...)
	}
	if sig, err := c.Sign(rand.Reader, buf, c.HashFunc()); err != nil {
		return err
	} else {
		a.sign = miscellaneous.Dup(sig)
	}
	return nil
}

func (a *Transaction) Verify(c crypt.Crypto) bool {
	buf := []byte{}
	if tmp, err := a.header.Show(); err != nil {
		return false
	} else {
		buf = append(buf, tmp...)
	}
	if tmp, err := a.smartContract.Show(); err != nil {
		return false
	} else {
		buf = append(buf, tmp...)
	}
	return crypt.Verify(c.Public(), buf, a.sign, c.HashFunc())
}
