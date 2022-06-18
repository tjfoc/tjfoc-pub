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
	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	pb "github.com/tjfoc/tjfoc/protos/chaincode"
	"github.com/tjfoc/tjfoc/protos/monitor"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type allop struct {
	inputccNames [][]byte
	inputArgs    [][][]byte
	inputTxHash  [][]byte
}

type deliever struct {
	recvCH       chan *allop
	returnResult map[string]string
	returnKv     map[string]string
	returnAction map[string]int32
	returnCH     chan bool
	isdealing    bool
	dealker      sync.Mutex
}

var dlv *deliever

var chaincodeDelieverLogger = flogging.MustGetLogger("chaincode_deliever")

func (d *deliever) RecvOP(ccNames [][]byte, args [][][]byte, txhash [][]byte) (map[string]string, map[string]string, map[string]int32, bool) {
	d.dealker.Lock()
	if d.isdealing {
		d.dealker.Unlock()
		chaincodeDelieverLogger.Error("chaincode system is busy!")
		return nil, nil, nil, false
	}
	//清除上一次recvop的执行结果
	d.returnResult = make(map[string]string, 0)
	d.returnKv = make(map[string]string, 0)
	d.returnAction = make(map[string]int32, 0)
	d.recvCH <- &allop{
		inputccNames: ccNames,
		inputArgs:    args,
		inputTxHash:  txhash,
	}
	d.isdealing = true
	d.dealker.Unlock()

	<-d.returnCH

	d.dealker.Lock()
	d.isdealing = false
	d.dealker.Unlock()

	return d.returnResult, d.returnKv, d.returnAction, true
}

func (d *deliever) waitmes() {
	enable := viper.GetBool("Docker.Enable")
	for {
		select {
		case v := <-d.recvCH:
			for i := 0; i < len(v.inputccNames); i++ {
				ssa := v.inputccNames[i]
				ssb := v.inputArgs[i]
				ssc := v.inputTxHash[i]
				//非隐私交易或者是隐私交易并且本节点是相关方
				if len(ssa) != 0 && len(ssc) != 0 {
					type AAA struct {
						Name    string
						Content []byte
						Version string
						Type    int32  //0-delete chaincode 1-create chaincode 2-normal tx  3-fate  4-store data on chain 5-secret tx
						Sign    string //创建该交易者的签名，用于docker合约的权限管理（是否有权限删除该合约，是否有权限升级该合约的版本等）
					}
					var zz AAA
					if err := json.Unmarshal(ssa, &zz); err == nil {
						if zz.Type == 5 {
							//隐私交易并且该节点是相关方
							chaincodeDelieverLogger.Info("this is a secret tx,and im the realted part!")
							if enable {
								name := zz.Name
								version := zz.Version
								res.resLocker.Lock()
								res.currentTxIsSecret = true
								res.currentTxId = string(ssc)
								res.resLocker.Unlock()

								chaincodeSecretTx(name, version, ssb, string(ssc))

								//将交易结果发送给监督节点
								rrrr := &monitor.TransactionResults{
									Kv:     make(map[string][]byte),
									Action: make(map[string]int32),
									TxId:   ssc,
								}
								res.resLocker.Lock()
								for k, v := range res.singleTxResult {
									rrrr.Kv[k] = v.value.Value
									rrrr.Action[k] = v.action
								}
								//chaincodeDelieverLogger.Infof("send secret tx result to monitor!KV:%v,Action:%v\n", rrrr.Kv, rrrr.Action)
								res.resLocker.Unlock()
								for {
									conn, err := grpc.Dial(res.monitorAddr, grpc.WithInsecure())
									if err == nil {
										client := monitor.NewMonitorClient(conn)
										response, err := client.SendTxResult(context.Background(), rrrr)
										if err != nil || response.Req != true {
											if conn != nil {
												conn.Close()
											}
											chaincodeDelieverLogger.Error("send secret tx result to monitor error!Error:", err)
											time.Sleep(100 * time.Millisecond)
											continue
										}
										//chaincodeDelieverLogger.Info("send secret tx result to monitor success!")
										if conn != nil {
											conn.Close()
										}
										break
									} else {
										if conn != nil {
											conn.Close()
										}
										chaincodeDelieverLogger.Error("connect to monitor error!Error:", err)
										time.Sleep(100 * time.Millisecond)
										continue
									}
								}

								//拉取隐私交易结果
								for {
									conn, err := grpc.Dial(res.monitorAddr, grpc.WithInsecure())
									if err == nil {
										client := monitor.NewMonitorClient(conn)
										response, err := client.GetTxResult(context.Background(), &monitor.TransactionResults{TxId: ssc})
										if err != nil {
											if conn != nil {
												conn.Close()
											}
											chaincodeDelieverLogger.Error("get secret tx result from monitor error!Error:", err)
											time.Sleep(100 * time.Millisecond)
											continue
										}
										if len(response.TxId) == 0 || len(response.Kv) == 0 || len(response.Action) == 0 {
											//隐私交易双方得到的交易结果不一样，交易失败
											chaincodeDelieverLogger.Info("secret tx merge result by monitor error!related parts have different result!")
											temp := &ccResult{
												Status:    CCError,
												Message:   fmt.Sprintf("secret tx [%s] excute failed!Related parties get different results!\n", response.TxId),
												ChangedKv: nil,
												Response:  nil,
											}
											resB, _ := json.Marshal(temp)
											d.returnResult[string(ssc)] = string(resB)
										} else {
											//隐私交易双方得到的交易结果一样，交易成功
											//chaincodeDelieverLogger.Infof("secret tx excute success!KV:%v,Action:%v\n", response.Kv, response.Action)
											chaincodeDelieverLogger.Info("secret tx merge result by monitor success!")
											res.resLocker.Lock()
											for k, v := range response.Kv {
												oldstate := checkKeyIsSecret(k)
												res.tempResult[k] = &singleResult{
													value: singleValue{
														Value:    v,
														IsSecret: oldstate,
													},
													duiChen: true,
													tongTai: oldstate,
													action:  response.Action[k],
												}
											}
											res.resLocker.Unlock()
											simulateResponse := pb.Response{
												Status: 200,
											}
											tempB, _ := proto.Marshal(&simulateResponse)
											temp := &ccResult{Status: CCSuccess, Message: fmt.Sprintf("secret tx [%s] excute success!\n", response.TxId), ChangedKv: nil, Response: tempB}
											resB, _ := json.Marshal(temp)
											d.returnResult[string(ssc)] = string(resB)
										}
										if conn != nil {
											conn.Close()
										}
										break
									} else {
										if conn != nil {
											conn.Close()
										}
										chaincodeDelieverLogger.Error("connect to monitor error!Error:", err)
										time.Sleep(100 * time.Millisecond)
										continue
									}
								}

								res.resLocker.Lock()
								res.currentTxIsSecret = false
								res.currentTxId = ""
								res.singleTxResult = make(map[string]*singleResult)
								res.resLocker.Unlock()
							} else {
								chaincodeDelieverLogger.Info("docker is disable,u can enable it by conf file and reboot the peer!")
								temp := &ccResult{Status: CCError, Message: fmt.Sprintf("Docker is disable,can't deal transaction which is based on docker container!\n"), ChangedKv: nil, Response: nil}
								resB, _ := json.Marshal(temp)
								d.returnResult[string(ssc)] = string(resB)
							}
						} else if zz.Type == 4 {
							chaincodeDelieverLogger.Info("this is a store data tx!")
							//存证
							temp := ccResult{Status: CCSuccess, Message: "Store evidence operation!Always Success!", ChangedKv: nil, Response: nil}
							tempB, _ := json.Marshal(temp)
							d.returnResult[string(ssc)] = string(tempB)
						} else if zz.Type == 3 {
							//fate
							chaincodeDelieverLogger.Info("this is a fate tx!")
						} else if zz.Type == 2 {
							chaincodeDelieverLogger.Info("this is a normal tx!")
							if enable {
								name := zz.Name
								version := zz.Version
								res.resLocker.Lock()
								res.currentTxIsSecret = false
								res.currentTxId = string(ssc)
								res.resLocker.Unlock()
								//交易结果在函数内处理
								resultS := chaincodeNormalTx(name, version, ssb, string(ssc))
								d.returnResult[string(ssc)] = resultS
								res.resLocker.Lock()
								res.currentTxId = ""
								res.currentTxIsSecret = false
								for k, v := range res.singleTxResult {
									res.tempResult[k] = v
								}
								res.singleTxResult = make(map[string]*singleResult)
								res.resLocker.Unlock()
							} else {
								//配置文件不使用docker但是却收到了docker才能处理的消息
								chaincodeDelieverLogger.Info("docker is disable,u can enable it by conf file and reboot the peer!")
								temp := &ccResult{Status: CCError, Message: fmt.Sprintf("Docker is disable,can't deal transaction which is based on docker container!"), ChangedKv: nil, Response: nil}
								resB, _ := json.Marshal(temp)
								d.returnResult[string(ssc)] = string(resB)
							}
						} else if zz.Type == 1 {
							fmt.Println("this is a create docker tx!")
							if enable {
								name := zz.Name
								version := zz.Version
								content := zz.Content
								sign := zz.Sign
								//交易结果在函数内处理
								//if v := worldstate.GetWorldState().Search(name).Value; v == nil || v == sign {
								fmt.Printf("create docker,name:%s,version:%s,code:%s\n", name, version, content)
								resultS := chaincodeSpecialTxCreate(name, version, string(ssc), content, sign)
								d.returnResult[string(ssc)] = resultS
								//res.resLocker.Lock()
								//res.singleTxResult = make(map[string]*singleResult)
								//res.resLocker.Unlock()
								//} else {
								//chaincodeDelieverLogger.Errorf("This chaincode name [] has been used!", name)
								//temp := &ccResult{Status: CCError, Message: fmt.Sprintf("This chaincode name [%s] has been used!", name), ChangedKv: nil, Response: nil}
								//resB, _ := json.Marshal(temp)
								//d.returnResult[string(ssc)] = string(resB)
								//}
							} else {
								chaincodeDelieverLogger.Info("docker is disable,u can enable it by conf file and reboot the peer!")
								//配置文件不使用docker但是却收到了docker才能处理的消息
								temp := &ccResult{Status: CCError, Message: fmt.Sprintf("Docker is disable,can't deal install docker container transaction!"), ChangedKv: nil, Response: nil}
								resB, _ := json.Marshal(temp)
								d.returnResult[string(ssc)] = string(resB)
							}
						} else if zz.Type == 0 {
							fmt.Println("this is a delete docker tx!")
							if enable {
								name := zz.Name
								version := zz.Version
								sign := zz.Sign
								//if v := worldstate.GetWorldState().Search(name).Value; v == sign {
								deleteDockerImage := true //true 删除镜像
								//交易结果在函数内处理
								resultS := chaincodeSpecialTxDelete(name, version, string(ssc), deleteDockerImage, sign)
								d.returnResult[string(ssc)] = resultS
								//res.resLocker.Lock()
								//res.singleTxResult = make(map[string]*singleResult)
								//res.resLocker.Unlock()
								//} else {
								//chaincodeDelieverLogger.Errorf("You have no access to delete the chaincode named [%s]", name)
								//temp := &ccResult{Status: CCError, Message: fmt.Sprintf("You hace no access to delete the chaincode named [%s]!", name), ChangedKv: nil, Response: nil}
								//resB, _ := json.Marshal(temp)
								//d.returnResult[string(ssc)] = string(resB)
								//}
							} else {
								//配置文件不使用docker但是却收到了docker才能处理的消息
								fmt.Println("peer disable docker!u can change it in conf")
								temp := &ccResult{Status: CCError, Message: fmt.Sprintf("Docker is disable,can't deal delete docker container transaction!"), ChangedKv: nil, Response: nil}
								resB, _ := json.Marshal(temp)
								d.returnResult[string(ssc)] = string(resB)
							}
						}
					} else {
						chaincodeDelieverLogger.Errorf("ummarshal json error!:%s\n", err)
					}
				} else if len(ssa) == 0 && len(ssc) != 0 {
					//隐私交易并且该节点是非相关方
					//向监督节点请求结果
					chaincodeDelieverLogger.Info("this is a secret tx,and im not related part!")
					monitorAddr := viper.GetString("Monitor.Addr")
					for {
						conn, err := grpc.Dial(monitorAddr, grpc.WithInsecure())
						if err == nil {
							client := monitor.NewMonitorClient(conn)
							response, err := client.GetTxResult(context.Background(), &monitor.TransactionResults{TxId: ssc})
							if err != nil {
								if conn != nil {
									conn.Close()
								}
								chaincodeDelieverLogger.Error("get secret tx result from monitor error!Error:", err)
								time.Sleep(100 * time.Millisecond)
								continue
							}
							if len(response.TxId) == 0 || len(response.Kv) == 0 || len(response.Action) == 0 {
								//隐私交易双方得到的交易结果不一样，交易失败
								chaincodeDelieverLogger.Info("secret tx merge result by monitor error!related parts have different result!")
								temp := &ccResult{
									Status:    CCError,
									Message:   fmt.Sprintf("secret tx [%s] excute failed!Related parties get different results!\n", response.TxId),
									ChangedKv: nil,
									Response:  nil,
								}
								resB, _ := json.Marshal(temp)
								d.returnResult[string(ssc)] = string(resB)
							} else {
								//隐私交易双方得到的交易结果一样，交易成功
								//chaincodeDelieverLogger.Infof("secret tx excute success!KV:%v,Action:%v\n", response.Kv, response.Action)
								chaincodeDelieverLogger.Info("secret tx merge result by monitor success!")
								res.resLocker.Lock()
								for k, v := range response.Kv {
									oldstate := checkKeyIsSecret(k)
									res.tempResult[k] = &singleResult{
										value: singleValue{
											Value:    v,
											IsSecret: oldstate,
										},
										duiChen: true,
										tongTai: oldstate,
										action:  response.Action[k],
									}
								}
								res.resLocker.Unlock()
								simulateResponse := pb.Response{
									Status: 200,
								}
								tempB, _ := proto.Marshal(&simulateResponse)
								temp := &ccResult{Status: CCSuccess, Message: fmt.Sprintf("secret tx [%s] excute success!\n", response.TxId), ChangedKv: nil, Response: tempB}
								resB, _ := json.Marshal(temp)
								d.returnResult[string(ssc)] = string(resB)
							}
							if conn != nil {
								conn.Close()
							}
							break
						} else {
							if conn != nil {
								conn.Close()
							}
							chaincodeDelieverLogger.Error("connect to monitor error!Error:", err)
							time.Sleep(100 * time.Millisecond)
							continue
						}
					}
				} else {
					chaincodeDelieverLogger.Error("neither a secret tx,nor a normal tx!smartcontract or args or txhash wrong!")
					chaincodeDelieverLogger.Errorf("smartcontract:%v", ssa)
					chaincodeDelieverLogger.Errorf("args:%v", ssb)
					chaincodeDelieverLogger.Errorf("txhash:%v", ssc)
				}
			}
			//所有该block中的交易结果都处理完之后，将需要进行加密的key发送给监督节点处理
			rrrr := &monitor.DealResults{
				Kv:      make(map[string][]byte),
				Tongtai: make(map[string]bool),
				PeerID:  []byte(res.peerID),
			}
			res.resLocker.Lock()
			for k, v := range res.tempResult {
				if v.duiChen {
					rrrr.Kv[k] = v.value.Value
					rrrr.Tongtai[k] = v.tongTai
				}
			}
			res.resLocker.Unlock()
			if len(rrrr.Kv) > 0 {
				for {
					conn, err := grpc.Dial(res.monitorAddr, grpc.WithInsecure())
					if err == nil {
						client := monitor.NewMonitorClient(conn)
						response, err := client.SendTotalTxResult(context.Background(), rrrr)
						if err != nil {
							if conn != nil {
								conn.Close()
							}
							chaincodeDelieverLogger.Error("send total kv to monitor error!Error:", err)
							time.Sleep(100 * time.Millisecond)
							continue
						}
						chaincodeStartLogger.Info("monitor deal total results success!")
						res.resLocker.Lock()
						for k, v := range response.Kv {
							res.tempResult[k].value.Value = v
							res.tempResult[k].value.IsSecret = true
						}
						res.resLocker.Unlock()
						if conn != nil {
							conn.Close()
						}
						break
					} else {
						if conn != nil {
							conn.Close()
						}
						chaincodeDelieverLogger.Error("send total kv to monitor,connect monitor error!Error:", err)
						time.Sleep(100 * time.Millisecond)
						continue
					}
				}
			}
			res.resLocker.Lock()
			for k, v := range res.tempResult {
				lalala, _ := json.Marshal(&v.value)
				d.returnKv[k] = string(lalala)
				d.returnAction[k] = v.action
			}
			res.tempResult = make(map[string]*singleResult)
			res.resLocker.Unlock()
			d.returnCH <- true
		}
	}
}
