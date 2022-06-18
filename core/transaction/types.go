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
	"errors"

	"github.com/tjfoc/tjfoc/core/miscellaneous"
)

const (
	TRANSACTION_NUM_SIZE         = 4
	TRANSACTION_VERSION_OFFSET   = 4
	TRANSACTION_TIMESTAMP_OFFSET = 12
	TRANSACTION_HEADER_SIZE      = 12
	TRANSACTION_HASH_SIZE        = 32 // hash算法固定32字节
)

type Transactions []Transaction
type transactionRecords []transactionRecord
type transactionSmartContractArgs []transactionSmartContractArg

// 12 bytes
type transactionHeader struct {
	version   uint32 // 4 bytes
	timestamp uint64 // UTC time -- 8 bytes
}

type transactionSmartContract struct {
	smartContract     []byte                        // 合约脚本 - utf-8
	smartContractArgs *transactionSmartContractArgs // 合约参数
}

// 合约参数
type transactionSmartContractArg struct {
	smartContractArg []byte
}

// 合约操作记录
type transactionRecord struct {
	record []byte
}

type Transaction struct {
	sign          []byte // 签名是对header, smartContract的签名
	header        *transactionHeader
	records       *transactionRecords       // 合约操作的结果，比如a=10, b=20...
	smartContract *transactionSmartContract // 合约
}

func (a *Transaction) Version() uint32 {
	return a.header.version
}

func (a *Transaction) Timestamp() uint64 {
	return a.header.timestamp
}

func (a *Transaction) SignData() []byte {
	return a.sign
}

func (a *Transaction) Records() [][]byte {
	buf := [][]byte{}
	for _, v := range *a.records {
		buf = append(buf, v.record)
	}
	return buf
}

func (a *Transaction) SmartContract() []byte {
	return a.smartContract.smartContract
}

func (a *Transaction) SmartContractArgs() [][]byte {
	buf := [][]byte{}
	for _, v := range *a.smartContract.smartContractArgs {
		buf = append(buf, v.smartContractArg)
	}
	return buf
}

/*
交易信息存储方式如下:
	交易头(12 bytes) - 智能合约 - 智能合约状态列表 - 签名
*/
func (a *Transaction) Show() ([]byte, error) {
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
	if a.records != nil {
		if tmp, err := a.records.Show(); err != nil {
			return []byte{}, err
		} else {
			buf = append(buf, tmp...)
		}
	} else {
		buf = append(buf, miscellaneous.E32func(0)...)
	}
	buf = append(buf, miscellaneous.E32func(uint32(len(a.sign)))...)
	buf = append(buf, a.sign...)
	return buf, nil
}

func (a *Transaction) Read(b []byte) ([]byte, error) {
	var err error

	if len(b) < TRANSACTION_HEADER_SIZE {
		return []byte{}, errors.New("transaction Read: Illegal slice length")
	}
	if a.header == nil {
		a.header = new(transactionHeader)
	}
	if b, err = a.header.Read(b); err != nil {
		return []byte{}, err
	}
	if a.smartContract == nil {
		a.smartContract = new(transactionSmartContract)
	}
	if b, err = a.smartContract.Read(b); err != nil {
		return []byte{}, err
	}
	a.records = new(transactionRecords)
	if b, err = a.records.Read(b); err != nil {
		return []byte{}, err
	}
	if len(b) < TRANSACTION_NUM_SIZE {
		return []byte{}, errors.New("transaction Read: Illegal slice length")
	}
	signSize, _ := miscellaneous.D32func(b[:TRANSACTION_NUM_SIZE])
	b = b[TRANSACTION_NUM_SIZE:]
	if uint32(len(b)) < signSize {
		return []byte{}, errors.New("transaction Read: Illegal sign length")
	}
	a.sign = miscellaneous.Dup(b[:int(signSize)])
	return b[int(signSize):], nil
}

/*
Transactions存储格式:
	0 ~ 3 	byte: transaction count
	transaction 	0
	transaction 	1
	transaction 	2
	transaction 	3
	...
	transaction 	n
*/
func (a *Transactions) Show() ([]byte, error) {
	buf := []byte{}
	count := uint32(0)
	for _, v := range *a {
		if tmp, err := v.Show(); err != nil {
			return []byte{}, nil
		} else {
			buf = append(buf, tmp...)
		}
		count++
	}
	buf = append(miscellaneous.E32func(count), buf...)
	return buf, nil
}

func (a *Transactions) Read(b []byte) ([]byte, error) {
	var err error

	if len(b) < TRANSACTION_NUM_SIZE {
		return []byte{}, errors.New("transactions Read: Illegal slice length")
	}
	count, _ := miscellaneous.D32func(b[:TRANSACTION_NUM_SIZE])
	b = b[TRANSACTION_NUM_SIZE:]
	for i := 0; uint32(i) < count; i++ {
		v := new(Transaction)
		if b, err = v.Read(b); err != nil {
			return []byte{}, err
		}
		*a = append(*a, *v)
	}
	return b, nil
}

/*
transactionHeader存储格式:
	0 ~ 3  byte: 版本号
	4 ~ 11 byte: 时间戳
*/
func (a *transactionHeader) Show() ([]byte, error) {
	buf := []byte{}
	buf = append(buf, miscellaneous.E32func(a.version)...)
	buf = append(buf, miscellaneous.E64func(a.timestamp)...)
	return buf, nil
}

func (a *transactionHeader) Read(b []byte) ([]byte, error) {
	if len(b) < TRANSACTION_HEADER_SIZE {
		return []byte{}, errors.New("transaction header Read: Illegal slice length")
	}
	a.version, _ = miscellaneous.D32func(b[:TRANSACTION_VERSION_OFFSET])
	a.timestamp, _ = miscellaneous.D64func(b[TRANSACTION_VERSION_OFFSET:TRANSACTION_HEADER_SIZE])
	return b[TRANSACTION_HEADER_SIZE:], nil
}

/*
transactionRecord存储格式:
	0 ~ 3 	byte: record长度
	4 ~ ?   byte: record
*/
func (a *transactionRecord) Show() ([]byte, error) {
	buf := []byte{}
	buf = append(buf, miscellaneous.E32func(uint32(len(a.record)))...)
	buf = append(buf, a.record...)
	return buf, nil
}

func (a *transactionRecord) Read(b []byte) ([]byte, error) {
	if len(b) < TRANSACTION_NUM_SIZE {
		return []byte{}, errors.New("transaction record Read: Illegal slice length")
	}
	recordSize, _ := miscellaneous.D32func(b[:TRANSACTION_NUM_SIZE])
	b = b[TRANSACTION_NUM_SIZE:]
	if uint32(len(b)) < recordSize {
		return []byte{}, errors.New("transaction record Read: Illegal record length")
	}
	a.record = miscellaneous.Dup(b[:int(recordSize)])
	return b[int(recordSize):], nil
}

/*
transactionRecords存储格式:
	transactionRecord count - 4 bytes
	transactionRecord 0
	transactionRecord 1
	transactionRecord 2
	transactionRecord 3
	...
	transactionRecord n
*/
func (a *transactionRecords) Show() ([]byte, error) {
	buf := []byte{}
	count := uint32(0)
	for _, v := range *a {
		if tmp, err := v.Show(); err != nil {
			return nil, err
		} else {
			buf = append(buf, tmp...)
		}
		count++
	}
	buf = append(miscellaneous.E32func(count), buf...)
	return buf, nil
}

func (a *transactionRecords) Read(b []byte) ([]byte, error) {
	var err error

	if len(b) < TRANSACTION_NUM_SIZE {
		return []byte{}, errors.New("transaction records Read: Illegal slice length")
	}
	count, _ := miscellaneous.D32func(b[:TRANSACTION_NUM_SIZE])
	b = b[TRANSACTION_NUM_SIZE:]
	for i := 0; uint32(i) < count; i++ {
		v := new(transactionRecord)
		if b, err = v.Read(b); err != nil {
			return []byte{}, err
		}
		*a = append(*a, *v)
	}
	return b, nil
}

/*
transactionSmartContractArg存储格式:
	0 ~ 3 	byte: arg长度
	4 ~ ? 	byte: arg
*/
func (a *transactionSmartContractArg) Show() ([]byte, error) {
	buf := []byte{}
	buf = append(buf, miscellaneous.E32func(uint32(len(a.smartContractArg)))...)
	buf = append(buf, a.smartContractArg...)
	return buf, nil
}

func (a *transactionSmartContractArg) Read(b []byte) ([]byte, error) {
	if len(b) < TRANSACTION_NUM_SIZE {
		return []byte{}, errors.New("transaction smartContract's arg Read: Illegal slice length")
	}
	argSize, _ := miscellaneous.D32func(b[:TRANSACTION_NUM_SIZE])
	b = b[TRANSACTION_NUM_SIZE:]
	if uint32(len(b)) < argSize {
		return []byte{}, errors.New("transaction smartContract Read: Illegal smartContract's arg length")
	}
	a.smartContractArg = miscellaneous.Dup(b[:int(argSize)])
	return b[int(argSize):], nil
}

/*
transactionSmartContractArgs存储格式:
	transactionSmartContractArg count - 4 bytes
	transactionSmartContractArg 0
	transactionSmartContractArg 1
	transactionSmartContractArg 2
	...
	transactionSmartContractArg n
*/
func (a *transactionSmartContractArgs) Show() ([]byte, error) {
	buf := []byte{}
	count := uint32(0)
	for _, v := range *a {
		if tmp, err := v.Show(); err != nil {
			return nil, err
		} else {
			buf = append(buf, tmp...)
		}
		count++
	}
	buf = append(miscellaneous.E32func(count), buf...)
	return buf, nil
}

func (a *transactionSmartContractArgs) Read(b []byte) ([]byte, error) {
	var err error

	if len(b) < TRANSACTION_NUM_SIZE {
		return []byte{}, errors.New("transaction smartContract's args Read: Illegal slice length")
	}
	count, _ := miscellaneous.D32func(b[:TRANSACTION_NUM_SIZE])
	b = b[TRANSACTION_NUM_SIZE:]
	for i := 0; uint32(i) < count; i++ {
		v := new(transactionSmartContractArg)
		if b, err = v.Read(b); err != nil {
			return []byte{}, err
		}
		*a = append(*a, *v)
	}
	return b, nil
}

/*
transactionSmartContract存储格式:
	0 ~ 3 	byte: smartContract长度
	4 ~ ?   byte: smartContract
	? ~ ?   byte: args
*/
func (a *transactionSmartContract) Show() ([]byte, error) {
	buf := []byte{}
	buf = append(buf, miscellaneous.E32func(uint32(len(a.smartContract)))...)
	buf = append(buf, a.smartContract...)
	if tmp, err := a.smartContractArgs.Show(); err != nil {
		return nil, err
	} else {
		buf = append(buf, tmp...)
	}
	return buf, nil
}

func (a *transactionSmartContract) Read(b []byte) ([]byte, error) {
	var err error

	if len(b) < TRANSACTION_NUM_SIZE {
		return []byte{}, errors.New("transaction smartContract Read: Illegal slice length")
	}
	smartContractSize, _ := miscellaneous.D32func(b[:TRANSACTION_NUM_SIZE])
	b = b[TRANSACTION_NUM_SIZE:]
	if uint32(len(b)) < smartContractSize {
		return []byte{}, errors.New("transaction smartContract Read: Illegal smartContract length")
	}
	a.smartContract = miscellaneous.Dup(b[:int(smartContractSize)])
	b = b[int(smartContractSize):]
	a.smartContractArgs = new(transactionSmartContractArgs)
	if b, err = a.smartContractArgs.Read(b); err != nil {
		return []byte{}, err
	}
	return b, nil
}
