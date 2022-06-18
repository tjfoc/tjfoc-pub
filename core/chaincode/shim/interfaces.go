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

package shim

import (
	pb "github.com/tjfoc/tjfoc/protos/chaincode"
)

// Chaincode interface must be implemented by all chaincodes. The Wutong runs
// the transactions by calling these functions as specified.
type Chaincode interface {
	// Init is called during Instantiate transaction after the chaincode container
	// has been established for the first time, allowing the chaincode to
	// initialize its internal data
	Init(stub ChaincodeStubInterface) pb.Response

	// Invoke is called to update or query the ledger in a proposal transaction.
	// Updated state variables are not committed to the ledger until the
	// transaction is committed.
	Invoke(stub ChaincodeStubInterface) pb.Response
}

// ChaincodeStubInterface is used by deployable chaincode apps to access and
// modify their ledgers
type ChaincodeStubInterface interface {
	// GetArgs returns the arguments intended for the chaincode Init and Invoke
	// as an array of byte arrays.
	GetArgs() [][]byte

	GetStringArgs() []string

	GetFunctionAndParameters() (string, []string)

	// If the key does not exist in the state database, (nil, nil) is returned.
	GetState(key string) ([]byte, error)
	GetStaten(keyn []string) (map[string]string, error)

	//The key and its value will be deleted from the ledger
	DelState(key string) error
	DelStaten(keyn []string) error

	//模糊匹配:key前缀查询
	GetStateByPrefix(key string) (map[string]string, error)
	//GetStateByRange

	// 存储key-value
	PutState(key string, value []byte) error
	//合约调用合约
	InvokeChaincode(chaincodeName string, chaincodeVersion string, args [][]byte) pb.Response

	// GetHistoryForKey随时间返回键值的历史记录。
	// 对于每个历史的key更新，都会返回历史值和关联的交易ID和时间戳。时间戳是客户端在提案标题中提供的时间戳。
	// GetHistoryForKey需要peer配置core.ledger.history.enableHistoryDatabase为true。
	// 在验证阶段不会重新执行该查询，不会检测到幻像读取。也就是说，其他提交的事务可能会同时更新密钥，影响结果集，并且这不会在验证/提交时检测到。
	// 因此，易受此影响的应用程序不应使用GetHistoryForKey作为更新分类帐的交易的一部分，并且应将使用限制为只读链式代码操作
	// GetHistoryForKey(key string) (HistoryQueryIteratorInterface, error)
}

// CommonIteratorInterface allows a chaincode to check whether any more result
// to be fetched from an iterator and close it when done.
type CommonIteratorInterface interface {
	// HasNext returns true if the range query iterator contains additional keys
	// and values.
	HasNext() bool

	// Close closes the iterator. This should be called when done
	// reading from the iterator to free up resources.
	Close() error
}

// StateQueryIteratorInterface allows a chaincode to iterate over a set of
// key/value pairs returned by range and execute query.
type StateQueryIteratorInterface interface {
	// Inherit HasNext() and Close()
	CommonIteratorInterface

	// Next returns the next key and value in the range and execute query iterator.
	//Next() (*queryresult.KV, error)
}

// HistoryQueryIteratorInterface allows a chaincode to iterate over a set of
// key/value pairs returned by a history query.
type HistoryQueryIteratorInterface interface {
	// Inherit HasNext() and Close()
	CommonIteratorInterface

	// Next returns the next key and value in the history query iterator.
	//Next() (*queryresult.KeyModification, error)
}

// MockQueryIteratorInterface allows a chaincode to iterate over a set of
// key/value pairs returned by range query.
// TODO: Once the execute query and history query are implemented in MockStub,
// we need to update this interface
type MockQueryIteratorInterface interface {
	StateQueryIteratorInterface
}
