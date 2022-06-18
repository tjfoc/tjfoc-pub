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

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/worldstate"
)

var delieverLogger = flogging.MustGetLogger("deliever")

type allop struct {
	inputccNames    [][]byte
	inputArgs       [][][]byte
	inputTxHash     [][]byte
	inputWSHash     string
	inputBlockIndex uint64
}

type deliever struct {
	recvCH       chan *allop
	returnResult map[string]string
	returnCH     chan bool
}

var dlv *deliever

func (d *deliever) RecvOP(ccNames [][]byte, args [][][]byte, txhash [][]byte, stateHash string, index uint64) (map[string]string, bool) {
	delieverLogger.Infof("begin RecvOP h-index:%d===", index)
	t1 := time.Now()
	d.recvCH <- &allop{
		inputccNames:    ccNames,
		inputArgs:       args,
		inputTxHash:     txhash,
		inputWSHash:     stateHash,
		inputBlockIndex: index,
	}
	status := <-d.returnCH
	t := time.Now().Sub(t1)
	delieverLogger.Infof("end RecvOP h-index:%d take:%s===", index, t)
	return d.returnResult, status
}

func (d *deliever) waitmes() {
	enable := viper.GetBool("Docker.Enable")
	for {
		select {
		case v := <-d.recvCH:
			select {
			case <-res.lastBlockTxFinished:
				d.returnResult = make(map[string]string, 0)
				res.index = v.inputBlockIndex - 1
				res.wshash = v.inputWSHash
				for i := 0; i < len(v.inputccNames); i++ {
					ssa := v.inputccNames[i]
					ssb := v.inputArgs[i]
					ssc := v.inputTxHash[i]
					type AAA struct {
						Name    string
						Content []byte
						Version string
						Type    int32 //0-delete chaincode 1-create chaincode 2-normal tx  3-fate  4-store data on chain
						sign    string
					}
					var zz AAA
					if err := json.Unmarshal(ssa, &zz); err == nil {
						if zz.Type == 4 {
							//存证
							temp := ccResult{Status: CCSuccess, Message: "Store evidence operation!Always Success!", ChangedKv: nil}
							tempB, _ := json.Marshal(temp)
							d.returnResult[string(ssc)] = string(tempB)
						} else if zz.Type == 3 {
							//fate
						} else if zz.Type == 2 {
							//判断是否使用docker
							if enable {
								name := zz.Name
								version := zz.Version
								//交易结果在函数内处理
								resultS := chaincodeNormalTx(name, version, ssb, string(ssc))
								d.returnResult[string(ssc)] = resultS
								res.resLocker.Lock()
								res.singleTxResult = make(map[string]string)
								res.resLocker.Unlock()
							} else {
								//配置文件不使用docker但是却收到了docker才能处理的消息
								temp := &ccResult{Status: CCError, Message: fmt.Sprintf("Docker is disable,can't deal transaction which is based on docker container!"), ChangedKv: nil, Response: nil}
								resB, _ := json.Marshal(temp)
								d.returnResult[string(ssc)] = string(resB)
							}
						} else if zz.Type == 1 {
							//判断是否使用docker
							if enable {
								name := zz.Name
								version := zz.Version
								content := zz.Content
								sign := zz.sign
								//交易结果在函数内处理
								//if v := worldstate.GetWorldState().Search(name).Value; v == nil || v == sign {
								resultS := chaincodeSpecialTxCreate(name, version, string(ssc), content, sign)
								d.returnResult[string(ssc)] = resultS
								//} else {
								//delieverLogger.Errorf("This chaincode name [] has been used!", name)
								//temp := &ccResult{Status: CCError, Message: fmt.Sprintf("This chaincode name [%s] has been used!", name), ChangedKv: nil, Response: nil}
								//resB, _ := json.Marshal(temp)
								//d.returnResult[string(ssc)] = string(resB)
								//}
							} else {
								//配置文件不使用docker但是却收到了docker才能处理的消息
								temp := &ccResult{Status: CCError, Message: fmt.Sprintf("Docker is disable,can't deal install docker container transaction!"), ChangedKv: nil, Response: nil}
								resB, _ := json.Marshal(temp)
								d.returnResult[string(ssc)] = string(resB)
							}
						} else if zz.Type == 0 {
							//判断是否使用docker
							if enable {
								name := zz.Name
								version := zz.Version
								sign := zz.sign
								//if v := worldstate.GetWorldState().Search(name).Value; v == sign {
								deleteDockerImage := true //true 删除镜像
								//交易结果在函数内处理
								resultS := chaincodeSpecialTxDelete(name, version, string(ssc), deleteDockerImage, sign)
								d.returnResult[string(ssc)] = resultS
								//} else {
								//delieverLogger.Errorf("You have no access to delete the chaincode named [%s]", name)
								//temp := &ccResult{Status: CCError, Message: fmt.Sprintf("You hace no access to delete the chaincode named [%s]!", name), ChangedKv: nil, Response: nil}
								//resB, _ := json.Marshal(temp)
								//d.returnResult[string(ssc)] = string(resB)
								//}
							} else {
								//配置文件不使用docker但是却收到了docker才能处理的消息
								temp := &ccResult{Status: CCError, Message: fmt.Sprintf("Docker is disable,can't deal delete docker container transaction!"), ChangedKv: nil, Response: nil}
								resB, _ := json.Marshal(temp)
								d.returnResult[string(ssc)] = string(resB)
							}
						}
					} else {
						delieverLogger.Errorf("ummarshal json error!:%s", err)
					}
				}
				ret := worldstate.GetWorldState().Check(v.inputBlockIndex-1, v.inputWSHash, res.tempResult)
				res.resLocker.Lock()
				res.tempResult = make(map[string]string, 0)
				res.index = 0
				res.wshash = ""
				res.lastBlockTxFinished <- true
				res.resLocker.Unlock()
				d.returnCH <- ret
			}
		}
	}
}
