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

package chaincode

import "github.com/spf13/viper"
import "github.com/tjfoc/tjfoc/core/crypt"

const (
	CCSuccess = uint8(1)
	CCError   = uint8(2)
	CCTimeout = uint8(3)
)

type ccResult struct {
	Status    uint8
	Message   string
	ChangedKv map[string]string
	Response  []byte
}

// Block传入的交易合集
type Deliever interface {
	// 接收block中打包的所有交易
	// 参数一: 对应合约的名字
	// 参数二: 合约参数
	// 参数三: 交易哈希

	// 返回值一:交易结果 key为交易id，value为交易结果
	// 返回值二:交易改变的kv对，需要插入到worldstate中
	// 返回值三:每一个key对应的action
	// 返回值四:是否成功执行
	RecvOP([][]byte, [][][]byte, [][]byte) (map[string]string, map[string]string, map[string]int32, bool)
}

// chaincode执行后的交易结果
type Result interface {
	SetAttribute(id string, c crypt.Crypto)
}

func GetDelieverInstance() Deliever {
	if dlv == nil {
		dlv = new(deliever)
		dlv.recvCH = make(chan *allop, 10)
		dlv.returnCH = make(chan bool, 1)
		go dlv.waitmes()
	}
	return dlv
}

func GetResultInstance() Result {
	if res == nil {
		res = new(result)
		res.monitorAddr = viper.GetString("Monitor.Addr")
		res.tempResult = make(map[string]*singleResult)
		res.singleTxResult = make(map[string]*singleResult)
		res.currentTxIsSecret = false
		//启动docker
		startDocker()
	}
	return res
}
