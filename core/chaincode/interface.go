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

// Deliever Block传入的交易合集
type Deliever interface {
	// 接收block中打包的所有交易
	// 参数一: 对应合约的名字
	// 参数二: 合约参数
	// 参数三: 交易哈希
	// 参数四: worldstate哈希
	// 参数五: block的Index
	RecvOP([][]byte, [][][]byte, [][]byte, string, uint64) (map[string]string, bool)
}

// GetInstance 初始化相关参数
func GetInstance() Deliever {
	if dlv == nil {
		dlv = new(deliever)
		dlv.recvCH = make(chan *allop, 10)
		dlv.returnCH = make(chan bool, 1)
		go dlv.waitmes()
		initRes()
	}
	return dlv
}
