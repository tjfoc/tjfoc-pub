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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/looplab/fsm"

	"github.com/tjfoc/tjfoc/core/common/ccprovider"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/worldstate"
	pb "github.com/tjfoc/tjfoc/protos/chaincode"
	"github.com/tjfoc/tjfoc/protos/monitor"
	"google.golang.org/grpc"
)

const (
	createdstate     = "created"     //start state
	establishedstate = "established" //in: CREATED, rcv:  REGISTER, send: REGISTERED, INIT
	readystate       = "ready"       //in:ESTABLISHED,TRANSACTION, rcv:COMPLETED
	endstate         = "end"         //in:INIT,ESTABLISHED, rcv: error, terminate container
)

var chaincodeHandleLogger = flogging.MustGetLogger("chaincode_handle")

//ChaincodeStream 通信
type ChaincodeStream interface {
	Send(*pb.ChaincodeMessage) error
	Recv() (*pb.ChaincodeMessage, error)
}

//Handler 定义Handler的基本结构
type Handler struct {
	sync.RWMutex
	//peer to shim grpc serializer. User only in serialSend
	serialLock       sync.Mutex
	ChatStream       ChaincodeStream
	FSM              *fsm.FSM
	ChaincodeID      *pb.ChaincodeID
	txCtxs           map[string]*transactionContext
	chaincodeSupport *ChaincodeSupport
	registered       bool
	readyNotify      chan bool
	// Map of tx txid to either invoke tx. Each tx will be
	// added prior to execute and remove when done execute

	txidMap map[string]bool

	// used to do Send after making sure the state transition is complete
	nextState chan *nextStateInfo

	getLastHeartResponse bool
}

type transactionContext struct {
	chainID string
	// signedProp       *pb.SignedProposal
	// proposal         *pb.Proposal
	responseNotifier chan *pb.ChaincodeMessage
}

type nextStateInfo struct {
	msg      *pb.ChaincodeMessage
	sendToCC bool
	sendSync bool
}

//HandleChaincodeStream 处理chaincode与peer通信
func HandleChaincodeStream(chaincodeSupport *ChaincodeSupport, ctxt context.Context, stream ChaincodeStream) error {
	chaincodeHandleLogger.Debug("entry HandleChaincodeStream")
	handler := newChaincodeSupportHandler(chaincodeSupport, stream)
	return handler.processStream()
}

func newChaincodeSupportHandler(chaincodeSupport *ChaincodeSupport, peerChatStream ChaincodeStream) *Handler {
	v := &Handler{
		ChatStream:           peerChatStream,
		getLastHeartResponse: true,
	}
	v.chaincodeSupport = chaincodeSupport
	//we want this to block
	v.nextState = make(chan *nextStateInfo)

	v.FSM = fsm.NewFSM(
		createdstate,
		fsm.Events{
			//Send REGISTERED, then, if deploy { trigger INIT(via INIT) } else { trigger READY(via COMPLETED) }
			{Name: pb.ChaincodeMessage_REGISTER.String(), Src: []string{createdstate}, Dst: establishedstate},
			{Name: pb.ChaincodeMessage_READY.String(), Src: []string{establishedstate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_PUT_STATE.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_DEL_STATE.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_DEL_STATEN.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_INVOKE_CHAINCODE.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_COMPLETED.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_GET_STATE.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_GET_STATEN.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_GET_STATE_BY_RANGE.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_GET_QUERY_RESULT.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_GET_HISTORY_FOR_KEY.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_QUERY_STATE_NEXT.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_QUERY_STATE_CLOSE.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_ERROR.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_RESPONSE.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_INIT.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_TRANSACTION.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_GET_STATE_BY_PREFIX.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_REQUIRE_CRYPT.String(), Src: []string{readystate}, Dst: readystate},
			{Name: pb.ChaincodeMessage_REQUIRE_COMPARE.String(), Src: []string{readystate}, Dst: readystate},
		},
		fsm.Callbacks{
			"before_" + pb.ChaincodeMessage_REGISTER.String():           func(e *fsm.Event) { v.beforeRegisterEvent(e, v.FSM.Current()) },
			"before_" + pb.ChaincodeMessage_COMPLETED.String():          func(e *fsm.Event) { v.beforeCompletedEvent(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_GET_STATE.String():           func(e *fsm.Event) { v.enterGetState(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_GET_STATEN.String():          func(e *fsm.Event) { v.enterGetStaten(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_GET_STATE_BY_PREFIX.String(): func(e *fsm.Event) { v.enterGetStateByPrefix(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_PUT_STATE.String():           func(e *fsm.Event) { v.enterPutState(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_DEL_STATE.String():           func(e *fsm.Event) { v.enterDeleteState(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_DEL_STATEN.String():          func(e *fsm.Event) { v.enterDeleteStaten(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_REQUIRE_CRYPT.String():       func(e *fsm.Event) { v.enterRequireCrypt(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_REQUIRE_COMPARE.String():     func(e *fsm.Event) { v.enterRequireCompare(e, v.FSM.Current()) },
			//"after_" + pb.ChaincodeMessage_INVOKE_CHAINCODE.String():    func(e *fsm.Event) { v.enterInvokeChaincode(e, v.FSM.Current()) },
			"enter_" + establishedstate: func(e *fsm.Event) { v.enterEstablishedState(e, v.FSM.Current()) },
			"enter_" + readystate:       func(e *fsm.Event) { v.enterReadyState(e, v.FSM.Current()) },
			// "enter_" + endstate:                                      func(e *fsm.Event) { v.enterEndState(e, v.FSM.Current()) },
		},
	)

	return v
}

func (handler *Handler) createTXIDEntry(txid string) bool {
	if handler.txidMap == nil {
		return false
	}
	handler.Lock()
	defer handler.Unlock()
	if handler.txidMap[txid] {
		return false
	}
	handler.txidMap[txid] = true
	return handler.txidMap[txid]
}

func (handler *Handler) sendExecuteMessage(ctxt context.Context, chainID string, msg *pb.ChaincodeMessage) (chan *pb.ChaincodeMessage, error) {
	txctx, err := handler.createTxContext(ctxt, chainID, msg.Txid)
	if err != nil {
		return nil, err
	}
	//if security is disabled the context elements will just be nil
	// if err = handler.setChaincodeProposal(signedProp, prop, msg); err != nil {
	// 	return nil, err
	// }

	chaincodeHandleLogger.Debugf("[%x]sendExecuteMsg trigger event %s", shorttxid(msg.Txid), msg.Type)
	handler.triggerNextState(msg, true)

	return txctx.responseNotifier, nil
}

func (handler *Handler) triggerNextState(msg *pb.ChaincodeMessage, send bool) {
	//this will send Async
	handler.nextState <- &nextStateInfo{msg: msg, sendToCC: send, sendSync: false}
}

func (handler *Handler) processStream() error {
	chaincodeHandleLogger.Debug("entry handler processStream")

	msgAvail := make(chan *pb.ChaincodeMessage)
	var nsInfo *nextStateInfo
	var in *pb.ChaincodeMessage
	var err error

	//recv is used to spin Recv routine after previous received msg
	//has been processed
	recv := true

	//catch send errors and bail now that sends aren't synchronous
	errc := make(chan error, 1)

	var tker *time.Ticker
	if theChaincodeSupport.keepalive > 0 {
		tker = time.NewTicker(theChaincodeSupport.keepalive * time.Second)
	} else {
		tker = time.NewTicker(1 * time.Hour)
	}

	for {
		in = nil
		err = nil
		nsInfo = nil
		if recv {
			recv = false
			go func() {
				var in2 *pb.ChaincodeMessage
				in2, err = handler.ChatStream.Recv()
				chaincodeHandleLogger.Debug("/***************************  Recv a ChaincodeMessage  **********************************/")
				msgAvail <- in2
			}()
		}
		select {
		case sendErr := <-errc:
			//chaincodeHandleLogger.Error("case sendErr")
			if sendErr != nil {
				return sendErr
			}
			//send was successful, just continue
			continue
		case in = <-msgAvail:
			chaincodeHandleLogger.Debug("case msgAvail")
			if err == io.EOF {
				//对端正常关闭连接
				chaincodeHandleLogger.Error("close chan EOF")
				return err
			} else if err != nil || in == nil {
				//异常的连接断开，尝试重启

				chaincodeHandleLogger.Warning("connection with docker was closed unexpectedly!try to reboot it!")
				name := handler.ChaincodeID.Name
				index := strings.Index(name, "_")
				chaincodeName := name[:index]
				name = name[index+1:]
				index = strings.Index(name, "_")
				chaincodeVersion := name[:index]
				cccid := ccprovider.NewCCContext("tjfoc", chaincodeName, chaincodeVersion, "", theChaincodeSupport.ip, theChaincodeSupport.port, false)

				theChaincodeSupport.rebootingChaincodes.Lock()
				if !theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].isreboot {
					theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].isreboot = true
					if theChaincodeSupport.currentTxContent != nil {
						theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].txid = theChaincodeSupport.currentTxContent.Txid
						theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].txctx = handler.txCtxs[theChaincodeSupport.currentTxContent.Txid]
					}
				} else {
					theChaincodeSupport.rebootingChaincodes.Unlock()
					return nil
				}
				theChaincodeSupport.rebootingChaincodes.Unlock()

				if _, _, err := theChaincodeSupport.Launch(context.Background(), cccid, []byte("")); err != nil {
					//重启失败
					chaincodeHandleLogger.Errorf("Reboot chaincode [%s] failed!Delete it!", cccid.Name+"_"+cccid.Version)
					theChaincodeSupport.Stop(context.Background(), cccid, true)
				} else {
					//重启成功
					chaincodeHandleLogger.Warningf("Reboot chaincode [%s] success!", handler.ChaincodeID.Name)
				}

				//有新的handler（socket）了，退出旧的handler（socket），不需要从runningchaincode中删除该旧handler，由于key使用的是同一个名字，重启之后旧已经覆盖了
				if err != nil {
					chaincodeHandleLogger.Error("close chan Unknow err!")
					return err
				}
				if in == nil {
					chaincodeHandleLogger.Error("close chan nil message!")
					return errors.New("recv nil message!")
				}
			}
			//正常接受消息
			recv = true
			if in.Type == pb.ChaincodeMessage_KEEPALIVE {
				chaincodeHandleLogger.Debug("Received KEEPALIVE Response")
				handler.getLastHeartResponse = true
				continue
			}
		case nsInfo = <-handler.nextState:
			chaincodeHandleLogger.Debug("case nextState")
			in = nsInfo.msg
			if in == nil {
				chaincodeHandleLogger.Debug("Next state nil message, ending chaincode support stream")
				return fmt.Errorf("Next state nil message, ending chaincode support stream")
			}
		case <-tker.C:
			chaincodeHandleLogger.Infof("time to send a heart package!")
			if !handler.getLastHeartResponse {
				chaincodeHandleLogger.Error("Heart timeout!")
				//心跳超时如何处理 有待改善
				continue
			}
			handler.serialSendAsync(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_KEEPALIVE}, nil)
			handler.getLastHeartResponse = false
			continue
		}
		err = handler.HandleMessage(in)
		if err != nil {
			chaincodeHandleLogger.Errorf("[%v]Error handling message, ending stream: %s", shorttxid(in.Txid), err)
			return fmt.Errorf("Error handling message, ending stream: %s", err)
		}
		if nsInfo != nil && nsInfo.sendToCC {
			chaincodeHandleLogger.Debugf("[%v]sending state message %s", shorttxid(in.Txid), in.Type.String())
			//ready messages are sent sync
			if nsInfo.sendSync {
				if in.Type.String() != pb.ChaincodeMessage_READY.String() {
					panic(fmt.Sprintf("[%v]Sync send can only be for READY state,current state: %s\n", shorttxid(in.Txid), in.Type.String()))
				}
				if err = handler.serialSend(in); err != nil {
					return fmt.Errorf("[%v]Error sending ready  message, ending stream: %s", shorttxid(in.Txid), err)
				}
			} else {
				//if error bail in select
				handler.serialSendAsync(in, errc)
			}
		}
	}
}

//向 Chaincode 发送 ready 状态
func (handler *Handler) ready(ctxt context.Context, chainID string, txid string) (chan *pb.ChaincodeMessage, error) {
	chaincodeHandleLogger.Debug("sending READY")
	txctx, funcErr := handler.createTxContext(ctxt, chainID, txid)
	if funcErr != nil {
		return nil, funcErr
	}
	ccMsg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_READY, Txid: txid}
	handler.triggerNextStateSync(ccMsg)
	return txctx.responseNotifier, nil
}

//HandleMessage 处理消息
func (handler *Handler) HandleMessage(msg *pb.ChaincodeMessage) error {
	chaincodeHandleLogger.Debug("enter HandleMessage")
	if (msg.Type == pb.ChaincodeMessage_COMPLETED || msg.Type == pb.ChaincodeMessage_ERROR) && handler.FSM.Current() == "ready" {
		chaincodeHandleLogger.Debug("Complete Message")
		handler.notify(msg)
		return nil
	}
	if handler.FSM.Cannot(msg.Type.String()) {
		// Other errors
		return fmt.Errorf("[%x]Chaincode handler validator FSM cannot handle message (%s) with payload size (%d) while in state: %s", msg.Txid, msg.Type.String(), len(msg.Payload), handler.FSM.Current())
	}
	eventErr := handler.FSM.Event(msg.Type.String(), msg)
	filteredErr := filterError(eventErr)
	if filteredErr != nil {
		chaincodeHandleLogger.Errorf("[%x]Failed to trigger FSM event %s: %s", shorttxid(msg.Txid), msg.Type.String(), filteredErr)
	}

	// return filteredErr
	return nil
}

func filterError(errFromFSMEvent error) error {
	if errFromFSMEvent != nil {
		if noTransitionErr, ok := errFromFSMEvent.(*fsm.NoTransitionError); ok {
			if noTransitionErr.Err != nil {
				// Squash the NoTransitionError
				return errFromFSMEvent
			}
			chaincodeHandleLogger.Debugf("Ignoring NoTransitionError: %s", noTransitionErr)
		}
		if canceledErr, ok := errFromFSMEvent.(*fsm.CanceledError); ok {
			if canceledErr.Err != nil {
				// Squash the CanceledError
				return canceledErr
			}
			chaincodeHandleLogger.Debugf("Ignoring CanceledError: %s", canceledErr)
		}
	}
	return nil
}

func (handler *Handler) createTxContext(ctxt context.Context, chainID string, txid string) (*transactionContext, error) {
	if handler.txCtxs == nil {
		return nil, fmt.Errorf("cannot create notifier for txid:%x", txid)
	}
	handler.Lock()
	defer handler.Unlock()
	if handler.txCtxs[txid] != nil {
		return nil, fmt.Errorf("txid:%x exists", txid)
	}
	txctx := &transactionContext{
		chainID:          chainID,
		responseNotifier: make(chan *pb.ChaincodeMessage, 1),
	}
	handler.txCtxs[txid] = txctx
	return txctx, nil
}

//删除交易数据
func (handler *Handler) deleteTxContext(txid string) {
	handler.Lock()
	defer handler.Unlock()
	if handler.txCtxs != nil {
		delete(handler.txCtxs, txid)
	}
}

func (handler *Handler) triggerNextStateSync(msg *pb.ChaincodeMessage) {
	//this will send sync
	handler.nextState <- &nextStateInfo{msg: msg, sendToCC: true, sendSync: true}
}

func (handler *Handler) serialSendAsync(msg *pb.ChaincodeMessage, errc chan error) {
	go func() {
		err := handler.serialSend(msg)
		if errc != nil {
			errc <- err
		}
	}()
}

func (handler *Handler) serialSend(msg *pb.ChaincodeMessage) error {
	handler.serialLock.Lock()
	defer handler.serialLock.Unlock()

	var err error
	if err = handler.ChatStream.Send(msg); err != nil {
		err = fmt.Errorf("[%x]Error sending %s: %s", shorttxid(msg.Txid), msg.Type.String(), err)
		chaincodeHandleLogger.Errorf("%s", err)
		return err
	}
	chaincodeHandleLogger.Debugf("Transaction [%s] type [%s] send to docker", shorttxid(msg.Txid), msg.Type.String())
	return nil
}

func shorttxid(txid string) string {
	if len(txid) < 8 {
		return txid
	}
	return txid[0:8]
}

//以下为FSM事件

// beforeRegisterEvent is invoked when chaincode tries to register.
func (handler *Handler) beforeRegisterEvent(e *fsm.Event, state string) {
	chaincodeHandleLogger.Debugf("enter beforeRegisterEvent")
	msg, ok := e.Args[0].(*pb.ChaincodeMessage)
	if !ok {
		e.Cancel(fmt.Errorf("Received unexpected message type"))
		return
	}
	chaincodeID := &pb.ChaincodeID{}
	err := proto.Unmarshal(msg.Payload, chaincodeID)
	if err != nil {
		e.Cancel(fmt.Errorf("Error in received %s, could NOT unmarshal registration info: %s", pb.ChaincodeMessage_REGISTER, err))
		return
	}
	// Now register with the chaincodeSupport
	handler.ChaincodeID = chaincodeID
	err = handler.chaincodeSupport.registerHandler(handler)
	if err != nil {
		e.Cancel(err)
		handler.notifyDuringStartup(false)
		return
	}
	if err := handler.serialSend(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_REGISTERED}); err != nil {
		e.Cancel(fmt.Errorf("Error sending %s: %s", pb.ChaincodeMessage_REGISTERED, err))
		handler.notifyDuringStartup(false)
		return
	}
}

func (handler *Handler) enterEstablishedState(e *fsm.Event, state string) {
	handler.notifyDuringStartup(true)
}

func (handler *Handler) enterReadyState(e *fsm.Event, state string) {
	// Now notify
	msg, ok := e.Args[0].(*pb.ChaincodeMessage)
	if !ok {
		e.Cancel(fmt.Errorf("Received unexpected message type"))
		return
	}
	handler.notify(msg)
}

// beforeCompletedEvent is invoked when chaincode has completed execution of init, invoke.
func (handler *Handler) beforeCompletedEvent(e *fsm.Event, state string) {
	_, ok := e.Args[0].(*pb.ChaincodeMessage)
	if !ok {
		e.Cancel(fmt.Errorf("Received unexpected message type"))
		return
	}
	return
}

func (handler *Handler) notify(msg *pb.ChaincodeMessage) {
	handler.Lock()
	defer handler.Unlock()
	tctx := handler.txCtxs[msg.Txid]
	if tctx == nil {
		chaincodeHandleLogger.Debugf("notifier Txid:%x does not exist", shorttxid(msg.Txid))
	} else {
		chaincodeHandleLogger.Debugf("notifying Txid:%x", shorttxid(msg.Txid))
		tctx.responseNotifier <- msg
	}
}

func (handler *Handler) notifyDuringStartup(val bool) {
	if handler.readyNotify != nil {
		chaincodeHandleLogger.Debug("Notifying during startup")
		handler.readyNotify <- val
	} else {
		chaincodeHandleLogger.Debug("nothing to notify (dev mode ?)")
		if handler.chaincodeSupport.userRunsCC {
			if val {
				chaincodeHandleLogger.Debug("sending READY")
				ccMsg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_READY}
				go handler.triggerNextState(ccMsg, true)
			} else {
				chaincodeHandleLogger.Errorf("Error during startup .. not sending READY")
			}
		} else {
			chaincodeHandleLogger.Warningf("trying to manually run chaincode when not in devmode ?")
		}
	}
}

//enterGetState 处理来自chaincode的GET_STATE消息
func (handler *Handler) enterGetState(e *fsm.Event, state string) {
	chaincodeHandleLogger.Debugf("enterGetState++++++++")
	go func() {
		var serialSendMsg *pb.ChaincodeMessage
		defer func() {
			handler.serialSendAsync(serialSendMsg, nil)
		}()
		msg, _ := e.Args[0].(*pb.ChaincodeMessage)
		if msg.Type.String() == pb.ChaincodeMessage_GET_STATE.String() {
			key := string(msg.Payload)
			var resTemp []byte
			res.resLocker.Lock()
			if v, ok := res.singleTxResult[key]; ok {
				res.resLocker.Unlock()
				resTemp, _ := json.Marshal(&v.value)
				serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: resTemp, Txid: msg.Txid}
				return
			} else if v, ok := res.tempResult[key]; ok {
				res.resLocker.Unlock()
				resTemp, _ := json.Marshal(&v.value)
				serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: resTemp, Txid: msg.Txid}
				return
			}
			res.resLocker.Unlock()
			var oldresult string
			r := worldstate.GetWorldState().Search(key)
			if r.Err == nil {
				oldresult = r.Value
			} else {
				serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: []byte(fmt.Sprintf("Can't find the key!\n")), Txid: msg.Txid}
				return
			}
			var temp singleValue
			json.Unmarshal([]byte(oldresult), &temp)
			if temp.IsSecret {
				//执行grpc去监督节点请求
				rrrr := &monitor.TransactionKey{
					Key:         []byte(key),
					TxId:        []byte(res.currentTxId),
					PeerID:      []byte(res.peerID),
					Secretvalue: temp.Value,
				}
				//查询key，第一步获得随机数并且签名
				for {
					conn, err := grpc.Dial(res.monitorAddr, grpc.WithInsecure())
					if err != nil {
						if conn != nil {
							conn.Close()
						}
						chaincodeHandleLogger.Error("QueryKey connect to monitor error!Error:", err)
						time.Sleep(100 * time.Millisecond)
						continue
					}
					client := monitor.NewMonitorClient(conn)
					response, err := client.QueryKey(context.Background(), rrrr)
					if err != nil || response.RandNum == nil {
						if conn != nil {
							conn.Close()
						}
						chaincodeHandleLogger.Error("query key error!Error:", err)
						time.Sleep(100 * time.Millisecond)
						continue
					}
					s, err := res.c.Sign(nil, response.RandNum, nil)
					if err != nil {
						chaincodeHandleLogger.Errorf("use privkey for signing failed,error:%s\n", err)
					}
					rrrr.RandNum = response.RandNum
					rrrr.Sign = s
					if conn != nil {
						conn.Close()
					}
					break
				}
				//验证随机数和签名
				for {
					conn, err := grpc.Dial(res.monitorAddr, grpc.WithInsecure())
					if err != nil {
						if conn != nil {
							conn.Close()
						}
						chaincodeHandleLogger.Error("VerifyKey connect to monitor error!Error:", err)
						time.Sleep(100 * time.Millisecond)
						continue
					}
					client := monitor.NewMonitorClient(conn)
					response, err := client.Authentication(context.Background(), rrrr)
					if err != nil {
						if conn != nil {
							conn.Close()
						}
						chaincodeHandleLogger.Error("verify key error!Error:", err)
						time.Sleep(100 * time.Millisecond)
						continue
					}
					if response.Status != nil {
						chaincodeHandleLogger.Warningf("query key,monitor return error for this key!Key:%s,Error:%s\n", key, response.Status)
						serialSendMsg = &pb.ChaincodeMessage{
							Type:    pb.ChaincodeMessage_ERROR,
							Payload: []byte(fmt.Sprintf("Get secret key from monitor error!Error:%s\n", response.Status)),
							Txid:    msg.Txid,
						}
					} else {
						chaincodeHandleLogger.Infof("QueryKey success!value:%s\n", response.Value)
						resTemp, _ = json.Marshal(&singleValue{Value: response.Value, IsSecret: true})
					}
					if conn != nil {
						conn.Close()
					}
					break
				}
			} else {
				resTemp = []byte(oldresult)
			}
			serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: resTemp, Txid: msg.Txid}
		}
	}()
}

type ReturnKv struct {
	Kv map[string][]byte
}

//enterGetStaten 处理来自chaincode的GET_STATEN消息
func (handler *Handler) enterGetStaten(e *fsm.Event, state string) {
	chaincodeHandleLogger.Debugf("enterGetStaten++++++++")
	go func() {
		var serialSendMsg *pb.ChaincodeMessage
		defer func() {
			handler.serialSendAsync(serialSendMsg, nil)
		}()
		msg, _ := e.Args[0].(*pb.ChaincodeMessage)
		if msg.Type.String() == pb.ChaincodeMessage_GET_STATEN.String() {
			keyn := strings.Split(string(msg.Payload), "\n")
			tempRes := &ReturnKv{
				Kv: make(map[string][]byte),
			}
			for _, key := range keyn {
				res.resLocker.Lock()
				if v, ok := res.singleTxResult[key]; ok {
					res.resLocker.Unlock()
					rrrr, _ := json.Marshal(&v.value)
					tempRes.Kv[key] = rrrr
					continue
				} else if v, ok := res.tempResult[key]; ok {
					res.resLocker.Unlock()
					rrrr, _ := json.Marshal(&v.value)
					tempRes.Kv[key] = rrrr
					continue
				}
				res.resLocker.Unlock()
				var oldresult string
				r := worldstate.GetWorldState().Search(key)
				if r.Err == nil {
					oldresult = r.Value
				} else {
					tempRes.Kv[key] = nil
					continue
				}
				var temp singleValue
				json.Unmarshal([]byte(oldresult), &temp)
				if temp.IsSecret {
					//执行grpc去监督节点请求
					rrrr := &monitor.TransactionKey{
						Key:         []byte(key),
						TxId:        []byte(res.currentTxId),
						PeerID:      []byte(res.peerID),
						Secretvalue: temp.Value,
					}
					//查询key，第一步获得随机数并且签名
					for {
						conn, err := grpc.Dial(res.monitorAddr, grpc.WithInsecure())
						if err != nil {
							if conn != nil {
								conn.Close()
							}
							chaincodeHandleLogger.Error("QueryKeyn connect to monitor error!Error:", err)
							time.Sleep(100 * time.Millisecond)
							continue
						}
						client := monitor.NewMonitorClient(conn)
						response, err := client.QueryKey(context.Background(), rrrr)
						if err != nil || response.RandNum == nil {
							if conn != nil {
								conn.Close()
							}
							chaincodeHandleLogger.Error("query keyn error!Error:", err)
							time.Sleep(100 * time.Millisecond)
							continue
						}
						s, err := res.c.Sign(nil, response.RandNum, nil)
						if err != nil {
							chaincodeHandleLogger.Errorf("use privkey for signing failed,error:%s\n", err)
						}
						rrrr.RandNum = response.RandNum
						rrrr.Sign = s
						if conn != nil {
							conn.Close()
						}
						break
					}
					//验证随机数和签名
					for {
						conn, err := grpc.Dial(res.monitorAddr, grpc.WithInsecure())
						if err != nil {
							if conn != nil {
								conn.Close()
							}
							chaincodeHandleLogger.Error("VerifyKeyn connect to monitor error!Error:", err)
							time.Sleep(100 * time.Millisecond)
							continue
						}
						client := monitor.NewMonitorClient(conn)
						response, err := client.Authentication(context.Background(), rrrr)
						if err != nil {
							if conn != nil {
								conn.Close()
							}
							chaincodeHandleLogger.Error("verify keyn error!Error:", err)
							time.Sleep(100 * time.Millisecond)
							continue
						}
						if response.Status == nil {
							resTemp, _ := json.Marshal(&singleValue{Value: response.Value, IsSecret: true})
							tempRes.Kv[key] = resTemp
						} else {
							chaincodeHandleLogger.Warningf("query keyn,monitor return error for this key!Key:%s,Error:%s", key, response.Status)
						}
						if conn != nil {
							conn.Close()
						}
						break
					}
				} else {
					tempRes.Kv[key] = []byte(oldresult)
				}
			}
			tempdata, _ := json.Marshal(tempRes)
			serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: tempdata, Txid: msg.Txid}
		}
	}()
}

func (handler *Handler) enterGetStateByPrefix(e *fsm.Event, state string) {
	chaincodeHandleLogger.Debugf("enterGetStateByPrefix++++++++")
	go func() {
		var serialSendMsg *pb.ChaincodeMessage
		defer func() {
			handler.serialSendAsync(serialSendMsg, nil)
		}()
		msg, _ := e.Args[0].(*pb.ChaincodeMessage)
		if msg.Type.String() == pb.ChaincodeMessage_GET_STATE_BY_PREFIX.String() {
			//找到所有符合的key，存入keyn中
			keyn := make([]string, 0)
			r, err := worldstate.GetWorldState().SearchPrefix(string(msg.Payload))
			if err == nil {
				for k, _ := range r {
					keyn = append(keyn, k)
				}
			}
			regkey := string(msg.Payload) + "*"
			res.resLocker.Lock()
			for k, _ := range res.tempResult {
				if ok, _ := regexp.Match(regkey, []byte(k)); ok {
					isFind := false
					for _, existkey := range keyn {
						if existkey == k {
							isFind = true
							break
						}
					}
					if !isFind {
						keyn = append(keyn, k)
					}
				}
			}
			for k, _ := range res.singleTxResult {
				if ok, _ := regexp.Match(regkey, []byte(k)); ok {
					isFind := false
					for _, existkey := range keyn {
						if existkey == k {
							isFind = true
							break
						}
					}
					if !isFind {
						keyn = append(keyn, k)
					}
				}
			}
			res.resLocker.Unlock()
			tempRes := &ReturnKv{
				Kv: make(map[string][]byte),
			}
			for _, key := range keyn {
				res.resLocker.Lock()
				if v, ok := res.singleTxResult[key]; ok {
					res.resLocker.Unlock()
					rrrr, _ := json.Marshal(&v.value)
					tempRes.Kv[key] = rrrr
					continue
				} else if v, ok := res.tempResult[key]; ok {
					res.resLocker.Unlock()
					rrrr, _ := json.Marshal(&v.value)
					tempRes.Kv[key] = rrrr
					continue
				}
				res.resLocker.Unlock()
				var oldresult string
				r := worldstate.GetWorldState().Search(key)
				if r.Err == nil {
					oldresult = r.Value
				} else {
					tempRes.Kv[key] = nil
					continue
				}
				var temp singleValue
				json.Unmarshal([]byte(oldresult), &temp)
				if temp.IsSecret {
					//执行grpc去监督节点请求
					rrrr := &monitor.TransactionKey{
						Key:         []byte(key),
						TxId:        []byte(res.currentTxId),
						PeerID:      []byte(res.peerID),
						Secretvalue: temp.Value,
					}
					//查询key，第一步获得随机数并且签名
					for {
						conn, err := grpc.Dial(res.monitorAddr, grpc.WithInsecure())
						if err != nil {
							if conn != nil {
								conn.Close()
							}
							chaincodeHandleLogger.Error("QueryKeyByPrefix connect to monitor error!Error:", err)
							time.Sleep(100 * time.Millisecond)
							continue
						}
						client := monitor.NewMonitorClient(conn)
						response, err := client.QueryKey(context.Background(), rrrr)
						if err != nil {
							if conn != nil {
								conn.Close()
							}
							chaincodeHandleLogger.Error("query key by prefix error!Error:", err)
							time.Sleep(100 * time.Millisecond)
							continue
						}
						s, _ := res.c.Sign(nil, response.RandNum, nil)
						rrrr.RandNum = response.RandNum
						rrrr.Sign = s
						if conn != nil {
							conn.Close()
						}
						break
					}
					//验证随机数和签名
					for {
						conn, err := grpc.Dial(res.monitorAddr, grpc.WithInsecure())
						if err != nil {
							if conn != nil {
								conn.Close()
							}
							chaincodeHandleLogger.Error("VerifyKeyByPrefix connect to monitor error!Error:", err)
							time.Sleep(100 * time.Millisecond)
							continue
						}
						client := monitor.NewMonitorClient(conn)
						response, err := client.Authentication(context.Background(), rrrr)
						if err != nil {
							if conn != nil {
								conn.Close()
							}
							chaincodeHandleLogger.Error("verify key by prefix error!Error:", err)
							time.Sleep(100 * time.Millisecond)
							continue
						}
						if response.Status == nil {
							resTemp, _ := json.Marshal(&singleValue{Value: response.Value, IsSecret: true})
							tempRes.Kv[key] = resTemp
						} else {
							chaincodeHandleLogger.Warningf("query key by prefix,monitor return error for this key!Key:%s,Error:%s", key, response.Status)
						}
						if conn != nil {
							conn.Close()
						}
						break
					}
				} else {
					tempRes.Kv[key] = []byte(oldresult)
				}
			}
			tempdata, _ := json.Marshal(tempRes)
			serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: tempdata, Txid: msg.Txid}
		}
	}()
}

func (handler *Handler) enterPutState(e *fsm.Event, state string) {
	chaincodeHandleLogger.Debugf("enterPutState++++++++")
	go func() {
		var serialSendMsg *pb.ChaincodeMessage
		defer func() {
			handler.serialSendAsync(serialSendMsg, nil)
		}()
		msg, _ := e.Args[0].(*pb.ChaincodeMessage)
		if msg.Type.String() == pb.ChaincodeMessage_PUT_STATE.String() {
			putStateInfo := &pb.PutStateInfo{}
			proto.Unmarshal(msg.Payload, putStateInfo)
			var rrrr singleValue
			json.Unmarshal(putStateInfo.Value, &rrrr)
			res.resLocker.Lock()
			oldstate := checkKeyIsSecret(putStateInfo.Key)
			tempResultDuiChen := false
			v, ok := res.tempResult[putStateInfo.Key]
			if ok {
				tempResultDuiChen = v.duiChen
			}
			res.singleTxResult[putStateInfo.Key] = &singleResult{
				value: singleValue{
					Value:    rrrr.Value,
					IsSecret: oldstate,
				},
				duiChen: (oldstate || res.currentTxIsSecret || tempResultDuiChen),
				tongTai: oldstate,
				action:  worldstate.PUTACTION,
			}
			res.resLocker.Unlock()
			serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: nil, Txid: msg.Txid}
		}
	}()
}

func (handler *Handler) enterDeleteState(e *fsm.Event, state string) {
	chaincodeHandleLogger.Debugf("enterDeleteState++++++++")
	go func() {
		var serialSendMsg *pb.ChaincodeMessage
		defer func() {
			handler.serialSendAsync(serialSendMsg, nil)
		}()
		msg, _ := e.Args[0].(*pb.ChaincodeMessage)
		if msg.Type.String() == pb.ChaincodeMessage_DEL_STATE.String() {
			key := string(msg.Payload)
			res.resLocker.Lock()
			res.singleTxResult[key] = &singleResult{
				value: singleValue{
					Value:    nil,
					IsSecret: false,
				},
				duiChen: false,
				tongTai: false,
				action:  worldstate.DELACTION,
			}
			res.resLocker.Unlock()
			serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: nil, Txid: msg.Txid}
		}
	}()
}

func (handler *Handler) enterDeleteStaten(e *fsm.Event, state string) {
	chaincodeHandleLogger.Debugf("enterDeleteStaten++++++++")
	go func() {
		var serialSendMsg *pb.ChaincodeMessage
		defer func() {
			handler.serialSendAsync(serialSendMsg, nil)
		}()
		msg, _ := e.Args[0].(*pb.ChaincodeMessage)
		if msg.Type.String() == pb.ChaincodeMessage_DEL_STATEN.String() {
			keyn := strings.Split(string(msg.Payload), "\n")
			res.resLocker.Lock()
			for _, key := range keyn {
				res.singleTxResult[key] = &singleResult{
					value: singleValue{
						Value:    nil,
						IsSecret: false,
					},
					duiChen: false,
					tongTai: false,
					action:  worldstate.DELACTION,
				}
			}
			res.resLocker.Unlock()
			serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: nil, Txid: msg.Txid}
		}
	}()
}

func (handler *Handler) enterRequireCrypt(e *fsm.Event, state string) {
	chaincodeHandleLogger.Debugf("enterRequireCrypt++++++++")
	go func() {
		var serialSendMsg *pb.ChaincodeMessage
		defer func() {
			handler.serialSendAsync(serialSendMsg, nil)
		}()
		msg, _ := e.Args[0].(*pb.ChaincodeMessage)
		if msg.Type.String() == pb.ChaincodeMessage_REQUIRE_CRYPT.String() {
			num, e := strconv.ParseInt(string(msg.Payload), 10, 64)
			if e != nil {
				chaincodeHandleLogger.Error("strconv from string to int64 error!error:", e)
				serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: []byte(fmt.Sprintf("strconv from string to int64 error!error:%s\n", e)), Txid: msg.Txid}
				return
			}
			rrrr := &monitor.CryptNumber{
				TxId:  []byte(msg.Txid),
				Value: num,
			}
			for {
				//请求监督节点进行加密
				conn, err := grpc.Dial(res.monitorAddr, grpc.WithInsecure())
				if err != nil {
					if conn != nil {
						conn.Close()
					}
					chaincodeHandleLogger.Error("RequireCrypt connect to monitor error!Error:", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}
				client := monitor.NewMonitorClient(conn)
				response, err := client.RequireCrypt(context.Background(), rrrr)
				if err != nil {
					if conn != nil {
						conn.Close()
					}
					chaincodeHandleLogger.Error("RequireCrypt return response error!error:", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}
				//成功获得加密数据
				serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: response.Svalue, Txid: msg.Txid}
				if conn != nil {
					conn.Close()
				}
				break
			}
		}
	}()
}

func (handler *Handler) enterRequireCompare(e *fsm.Event, state string) {
	chaincodeHandleLogger.Debugf("enterRequireCompare++++++++")
	go func() {
		var serialSendMsg *pb.ChaincodeMessage
		defer func() {
			handler.serialSendAsync(serialSendMsg, nil)
		}()
		msg, _ := e.Args[0].(*pb.ChaincodeMessage)
		if msg.Type.String() == pb.ChaincodeMessage_REQUIRE_COMPARE.String() {
			type AAA struct {
				Originvalue []byte
				Dealvalue   int64
			}
			var aaa AAA
			json.Unmarshal(msg.Payload, &aaa)
			rrrr := &monitor.CompareNumber{
				TxId:   []byte(msg.Txid),
				Svalue: aaa.Originvalue,
				Value:  aaa.Dealvalue,
			}
			for {
				//请求监督节点进行比较
				conn, err := grpc.Dial(res.monitorAddr, grpc.WithInsecure())
				if err != nil {
					if conn != nil {
						conn.Close()
					}
					chaincodeHandleLogger.Error("RequireCompare connect to monitor error!Error:", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}
				client := monitor.NewMonitorClient(conn)
				response, err := client.RequireCompare(context.Background(), rrrr)
				if err != nil {
					if conn != nil {
						conn.Close()
					}
					chaincodeHandleLogger.Error("RequireCompare return response error!error:", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}
				//成功获得比较结果
				serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: []byte(strconv.FormatInt(response.Result, 10)), Txid: msg.Txid}
				if conn != nil {
					conn.Close()
				}
				break
			}
		}
	}()

}

/*
func (handler *Handler) enterInvokeChaincode(e *fsm.Event, state string) {
	chaincodeHandleLogger.Debugf("enterInvokeChaincode++++++++")
	go func() {
		var serialSendMsg *pb.ChaincodeMessage
		defer func() {
			handler.serialSendAsync(serialSendMsg, nil)
		}()
		msg, _ := e.Args[0].(*pb.ChaincodeMessage)
		if msg.Type.String() == pb.ChaincodeMessage_INVOKE_CHAINCODE.String() {
				type InvokeChaincode struct {
					ChaincodeName    string
					ChaincodeVersion string
					Args             [][]byte
				}
				var chaincodeSpec InvokeChaincode
				err := json.Unmarshal(msg.Payload, &chaincodeSpec)
				if err != nil {
					serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: []byte("unmarshal json error!"), Txid: msg.Txid}
					chaincodeHandleLogger.Error("unmarshal json error!")
					return
				}
				//创建一笔新的交易
				cccid := ccprovider.NewCCContext("tjfoc", chaincodeSpec.ChaincodeName, chaincodeSpec.ChaincodeVersion, msg.Txid, theChaincodeSupport.ip, theChaincodeSupport.port, false)
				canName := cccid.GetCanonicalName()
				if _, ok := theChaincodeSupport.runningChaincodes.chaincodeMap[canName]; !ok {
					serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: []byte(fmt.Sprintf("Required chaincode [%s] does not exist!\n", canName)), Txid: msg.Txid}
					chaincodeHandleLogger.Errorf("Required chaincode [%s] does not exist!\n", canName)
					return
				}
				//超时时间
				timeout := time.Duration(30) * time.Second
				cMsg := &pb.ChaincodeInput{Args: chaincodeSpec.Args}
				payload, err := proto.Marshal(cMsg)
				if err != nil {
					chaincodeHandleLogger.Error("marshal pb error!")
					serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: []byte("marshal pb error!"), Txid: msg.Txid}
					return
				}
				var ccMsg *pb.ChaincodeMessage
				if len(chaincodeSpec.Args) != 0 {
					ccMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: msg.Txid}
				} else {
					chaincodeHandleLogger.Error("InvokeChaincode args error!")
					serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: []byte("InvokeChaincode args error!"), Txid: msg.Txid}
					return
				}
				//执行交易
				response, execErr := theChaincodeSupport.Execute(context.Background(), cccid, ccMsg, timeout)
				if response == nil {
					//超时
					chaincodeHandleLogger.Error("InvokeChaincode timeout!")
					serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: []byte("InvokeChaincode timeout!"), Txid: msg.Txid}
					return
				} else if execErr != nil {
					//系统错误
					chaincodeHandleLogger.Error("InvokeChaincode system error!Error:", execErr)
					serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: []byte(fmt.Sprintf("InvokeChaincode system error!Error:%s\n", execErr)), Txid: msg.Txid}
					return
				} else {
					//正常
					resTemp, err := proto.Marshal(response)
					if err != nil {
						chaincodeHandleLogger.Error("marshal pb error!")
						serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: []byte("marshal pb error!"), Txid: msg.Txid}
						return
					}
					serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: resTemp, Txid: msg.Txid}
				}
		}
	}()
}
*/
func checkKeyIsSecret(key string) bool {
	if v, ok := res.tempResult[key]; ok {
		return v.value.IsSecret
	}
	r := worldstate.GetWorldState().Search(key)
	if r.Err == nil {
		var temp singleValue
		json.Unmarshal([]byte(r.Value), &temp)
		return temp.IsSecret
	}
	return false
}
