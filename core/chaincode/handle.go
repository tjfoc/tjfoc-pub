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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/looplab/fsm"
	logging "github.com/op/go-logging"

	"github.com/tjfoc/tjfoc/core/common/ccprovider"
	"github.com/tjfoc/tjfoc/core/worldstate"
	pb "github.com/tjfoc/tjfoc/protos/chaincode"
)

const (
	createdstate     = "created"     //start state
	establishedstate = "established" //in: CREATED, rcv:  REGISTER, send: REGISTERED, INIT
	readystate       = "ready"       //in:ESTABLISHED,TRANSACTION, rcv:COMPLETED
	endstate         = "end"         //in:INIT,ESTABLISHED, rcv: error, terminate container
)

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
	ccInstance       *ChaincodeInstance
	txCtxs           map[string]*transactionContext
	chaincodeSupport *ChaincodeSupport
	registered       bool
	readyNotify      chan bool
	// Map of tx txid to either invoke tx. Each tx will be
	// added prior to execute and remove when done execute

	txidMap map[string]bool

	// used to do Send after making sure the state transition is complete
	nextState chan *nextStateInfo
}

//ChaincodeInstance 合约的相关参数
type ChaincodeInstance struct {
	ChainID          string
	ChaincodeName    string
	ChaincodeVersion string
}

func getChaincodeInstance(ccName string) *ChaincodeInstance {
	b := []byte(ccName)
	ci := &ChaincodeInstance{}

	//compute suffix (ie, chain name)
	i := bytes.IndexByte(b, '/')
	if i >= 0 {
		if i < len(b)-1 {
			ci.ChainID = string(b[i+1:])
		}
		b = b[:i]
	}

	//compute version
	i = bytes.IndexByte(b, ':')
	if i >= 0 {
		if i < len(b)-1 {
			ci.ChaincodeVersion = string(b[i+1:])
		}
		b = b[:i]
	}
	// remaining is the chaincode name
	ci.ChaincodeName = string(b)

	return ci
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
	chaincodeLogger.Debug("entry HandleChaincodeStream")
	handler := newChaincodeSupportHandler(chaincodeSupport, stream)
	return handler.processStream()
}

func newChaincodeSupportHandler(chaincodeSupport *ChaincodeSupport, peerChatStream ChaincodeStream) *Handler {
	v := &Handler{
		ChatStream: peerChatStream,
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
		},
		fsm.Callbacks{
			"before_" + pb.ChaincodeMessage_REGISTER.String():  func(e *fsm.Event) { v.beforeRegisterEvent(e, v.FSM.Current()) },
			"before_" + pb.ChaincodeMessage_COMPLETED.String(): func(e *fsm.Event) { v.beforeCompletedEvent(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_GET_STATE.String():  func(e *fsm.Event) { v.afterGetState(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_GET_STATEN.String(): func(e *fsm.Event) { v.afterGetStaten(e, v.FSM.Current()) },
			// "after_" + pb.ChaincodeMessage_GET_STATE_BY_RANGE.String():  func(e *fsm.Event) { v.afterGetStateByRange(e, v.FSM.Current()) },
			// "after_" + pb.ChaincodeMessage_GET_QUERY_RESULT.String():    func(e *fsm.Event) { v.afterGetQueryResult(e, v.FSM.Current()) },
			// "after_" + pb.ChaincodeMessage_GET_HISTORY_FOR_KEY.String(): func(e *fsm.Event) { v.afterGetHistoryForKey(e, v.FSM.Current()) },
			// "after_" + pb.ChaincodeMessage_QUERY_STATE_NEXT.String():    func(e *fsm.Event) { v.afterQueryStateNext(e, v.FSM.Current()) },
			// "after_" + pb.ChaincodeMessage_QUERY_STATE_CLOSE.String():   func(e *fsm.Event) { v.afterQueryStateClose(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_PUT_STATE.String():           func(e *fsm.Event) { v.enterBusyState(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_DEL_STATE.String():           func(e *fsm.Event) { v.enterBusyState(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_DEL_STATEN.String():          func(e *fsm.Event) { v.enterBusyState(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_INVOKE_CHAINCODE.String():    func(e *fsm.Event) { v.enterBusyState(e, v.FSM.Current()) },
			"enter_" + establishedstate:                                 func(e *fsm.Event) { v.enterEstablishedState(e, v.FSM.Current()) },
			"enter_" + readystate:                                       func(e *fsm.Event) { v.enterReadyState(e, v.FSM.Current()) },
			"after_" + pb.ChaincodeMessage_GET_STATE_BY_PREFIX.String(): func(e *fsm.Event) { v.enterGetStateByPrefix(e, v.FSM.Current()) },
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
	if chaincodeLogger.IsEnabledFor(logging.DEBUG) {
		chaincodeLogger.Debugf("[%x]Inside sendExecuteMessage. Message %s", shorttxid(msg.Txid), msg.Type.String())
	}

	//if security is disabled the context elements will just be nil
	// if err = handler.setChaincodeProposal(signedProp, prop, msg); err != nil {
	// 	return nil, err
	// }

	chaincodeLogger.Debugf("[%x]sendExecuteMsg trigger event %s", shorttxid(msg.Txid), msg.Type)
	handler.triggerNextState(msg, true)

	return txctx.responseNotifier, nil
}

func (handler *Handler) triggerNextState(msg *pb.ChaincodeMessage, send bool) {
	//this will send Async
	defer func() {
		if err := recover(); err != nil {
			chaincodeLogger.Warning("panic in triggerNextState!")
			name := handler.ChaincodeID.Name
			index := strings.Index(name, "_")
			chaincodeName := name[:index]
			name = name[index+1:]
			index = strings.Index(name, "_")
			chaincodeVersion := name[:index]
			cccid := ccprovider.NewCCContext("tjfoc", chaincodeName, chaincodeVersion, "", theChaincodeSupport.ip, theChaincodeSupport.port, false)
			//if !theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].ispanic {
			//重启docker
			theChaincodeSupport.rebootingChaincodes.Lock()
			theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].needReboot = true
			theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].nextstate = &nextStateInfo{msg: msg, sendToCC: send, sendSync: false}
			theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].txid = msg.Txid
			theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].txctx = handler.txCtxs[msg.Txid]
			theChaincodeSupport.rebootingChaincodes.Unlock()
			if _, _, err := theChaincodeSupport.Launch(context.Background(), cccid, []byte("")); err != nil {
				chaincodeLogger.Errorf("Reboot chaincode [%s] failed!Delete it!", cccid.Name+"_"+cccid.Version)
				theChaincodeSupport.Stop(context.Background(), cccid, true)
				//theChaincodeSupport.rebootingChaincodes.Lock()
				//delete(theChaincodeSupport.rebootingChaincodes.chaincodeMap, handler.ChaincodeID.Name)
				//theChaincodeSupport.rebootingChaincodes.Unlock()
			} else {
				chaincodeLogger.Warningf("Reboot chaincode [%s] success!", handler.ChaincodeID.Name)
			}
			//} else {
			//删除chaincode
			//chaincodeLogger.Errorf("Delete chaincode [%s] because of panic!", cccid.Name+"_"+cccid.Version)
			//theChaincodeSupport.Stop(context.Background(), cccid, true)
			//theChaincodeSupport.rebootingChaincodes.Lock()
			//delete(theChaincodeSupport.rebootingChaincodes.chaincodeMap, handler.ChaincodeID.Name)
			//theChaincodeSupport.rebootingChaincodes.Unlock()
			//}
		}
	}()
	handler.nextState <- &nextStateInfo{msg: msg, sendToCC: send, sendSync: false}
}

func (handler *Handler) processStream() error {
	chaincodeLogger.Debug("entry handler processStream")
	//defer handler.deregister()
	msgAvail := make(chan *pb.ChaincodeMessage)
	var nsInfo *nextStateInfo
	var in *pb.ChaincodeMessage
	var err error

	//recv is used to spin Recv routine after previous received msg
	//has been processed
	recv := true

	//catch send errors and bail now that sends aren't synchronous
	errc := make(chan error, 1)
	for {
		in = nil
		err = nil
		nsInfo = nil
		if recv {
			recv = false
			go func() {
				var in2 *pb.ChaincodeMessage
				in2, err = handler.ChatStream.Recv()
				chaincodeLogger.Debug("/***************************  Recv a ChaincodeMessage  **********************************/")
				msgAvail <- in2
			}()
		}
		select {
		case sendErr := <-errc:
			//chaincodeLogger.Error("case sendErr")
			if sendErr != nil {
				return sendErr
			}
			//send was successful, just continue
			continue
		case in = <-msgAvail:
			chaincodeLogger.Debug("case msgAvail")
			if err == io.EOF {
				chaincodeLogger.Error("close chan EOF")
				close(handler.nextState)
				return err
			} else if err != nil {
				chaincodeLogger.Error("close chan Unknow err!")
				close(handler.nextState)
				return err
			} else if in == nil {
				chaincodeLogger.Error("close chan nil message!")
				close(handler.nextState)
				return fmt.Errorf("Received nil message, ending chaincode support stream")
			}
			// we can spin off another Recv again
			recv = true
			if in.Type == pb.ChaincodeMessage_KEEPALIVE {
				chaincodeLogger.Debug("Received KEEPALIVE Response")
				continue
			}
		case nsInfo = <-handler.nextState:
			chaincodeLogger.Debug("case nextState")
			in = nsInfo.msg
			if in == nil {
				chaincodeLogger.Debug("Next state nil message, ending chaincode support stream")
				return fmt.Errorf("Next state nil message, ending chaincode support stream")
			}
		case <-handler.waitForKeepaliveTimer():
			chaincodeLogger.Infof("case waitForKeepaliveTimer")
			if handler.chaincodeSupport.keepalive <= 0 {
				chaincodeLogger.Errorf("Invalid select: keepalive not on (keepalive=%d)", handler.chaincodeSupport.keepalive)
				continue
			}
			handler.serialSendAsync(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_KEEPALIVE}, nil)
			continue
		}
		err = handler.HandleMessage(in)
		if err != nil {
			chaincodeLogger.Errorf("[%v]Error handling message, ending stream: %s", shorttxid(in.Txid), err)
			return fmt.Errorf("Error handling message, ending stream: %s", err)
		}
		if nsInfo != nil && nsInfo.sendToCC {
			chaincodeLogger.Debugf("[%v]sending state message %s", shorttxid(in.Txid), in.Type.String())
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
	chaincodeLogger.Debug("sending READY")
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
	chaincodeLogger.Debug("enter HandleMessage")
	if msg.Type == pb.ChaincodeMessage_CC_PANIC {
		chaincodeLogger.Error("Panic Message")
		//theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].ispanic = true
		handler.notify(msg)
		return nil
	}
	if (msg.Type == pb.ChaincodeMessage_COMPLETED || msg.Type == pb.ChaincodeMessage_ERROR) && handler.FSM.Current() == "ready" {
		chaincodeLogger.Debug("Complete Message")
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
		chaincodeLogger.Errorf("[%x]Failed to trigger FSM event %s: %s", shorttxid(msg.Txid), msg.Type.String(), filteredErr)
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
			chaincodeLogger.Debugf("Ignoring NoTransitionError: %s", noTransitionErr)
		}
		if canceledErr, ok := errFromFSMEvent.(*fsm.CanceledError); ok {
			if canceledErr.Err != nil {
				// Squash the CanceledError
				return canceledErr
			}
			chaincodeLogger.Debugf("Ignoring CanceledError: %s", canceledErr)
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
	//txctx.txsimulator = getTxSimulator(ctxt)
	//txctx.historyQueryExecutor = getHistoryQueryExecutor(ctxt)
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
	defer func() {
		if err := recover(); err != nil {
			chaincodeLogger.Warning("panic in triggerNextStateSync!")
			name := handler.ChaincodeID.Name
			index := strings.Index(name, "_")
			chaincodeName := name[:index]
			name = name[index+1:]
			index = strings.Index(name, "_")
			chaincodeVersion := name[:index]
			cccid := ccprovider.NewCCContext("tjfoc", chaincodeName, chaincodeVersion, "", theChaincodeSupport.ip, theChaincodeSupport.port, false)
			//if !theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].ispanic {
			//重启docker
			theChaincodeSupport.rebootingChaincodes.Lock()
			theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].needReboot = true
			theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].nextstate = &nextStateInfo{msg: msg, sendToCC: true, sendSync: true}
			theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].txid = msg.Txid
			theChaincodeSupport.rebootingChaincodes.chaincodeMap[handler.ChaincodeID.Name].txctx = handler.txCtxs[msg.Txid]
			theChaincodeSupport.rebootingChaincodes.Unlock()
			if _, _, err := theChaincodeSupport.Launch(context.Background(), cccid, []byte("")); err != nil {
				chaincodeLogger.Errorf("sync Reboot chaincode [%s] failed!Delete it!", cccid.Name+"_"+cccid.Version)
				theChaincodeSupport.Stop(context.Background(), cccid, true)
				//theChaincodeSupport.rebootingChaincodes.Lock()
				//delete(theChaincodeSupport.rebootingChaincodes.chaincodeMap, handler.ChaincodeID.Name)
				//theChaincodeSupport.rebootingChaincodes.Unlock()
			} else {
				chaincodeLogger.Warningf("sync Reboot chaincode [%s] success!", handler.ChaincodeID.Name)
			}
			//} else {
			//删除chaincode
			//chaincodeLogger.Errorf("sync Delete chaincode [%s] because of panic!", cccid.Name+"_"+cccid.Version)
			//theChaincodeSupport.Stop(context.Background(), cccid, true)
			//theChaincodeSupport.rebootingChaincodes.Lock()
			//delete(theChaincodeSupport.rebootingChaincodes.chaincodeMap, handler.ChaincodeID.Name)
			//theChaincodeSupport.rebootingChaincodes.Unlock()
			//}
		}
	}()
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
		chaincodeLogger.Errorf("%s", err)
		return err
	}
	chaincodeLogger.Debugf("Transaction [%s] type [%s] send to docker", shorttxid(msg.Txid), msg.Type.String())
	return nil
}

func (handler *Handler) waitForKeepaliveTimer() <-chan time.Time {
	if handler.chaincodeSupport.keepalive > 0 {
		c := time.After(handler.chaincodeSupport.keepalive)
		return c
	}
	//no one will signal this channel, listner blocks forever
	c := make(chan time.Time, 1)
	return c
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
	chaincodeLogger.Debugf("enter beforeRegisterEvent")
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
		chaincodeLogger.Debugf("notifier Txid:%x does not exist", shorttxid(msg.Txid))
	} else {
		chaincodeLogger.Debugf("notifying Txid:%x", shorttxid(msg.Txid))
		tctx.responseNotifier <- msg

		// clean up queryIteratorMap
		// for _, v := range tctx.queryIteratorMap {
		// 	v.Close()
		// }
	}
}

func (handler *Handler) notifyDuringStartup(val bool) {
	//if USER_RUNS_CC readyNotify will be nil
	if handler.readyNotify != nil {
		chaincodeLogger.Debug("Notifying during startup")
		handler.readyNotify <- val
	} else {
		chaincodeLogger.Debug("nothing to notify (dev mode ?)")
		//In theory, we don't even need a devmode flag in the peer anymore
		//as the chaincode is brought up without any context (ledger context
		//in particular). What this means is we can have - in theory - a nondev
		//environment where we can attach a chaincode manually. This could be
		//useful .... but for now lets just be conservative and allow manual
		//chaincode only in dev mode (ie, peer started with --peer-chaincodedev=true)
		if handler.chaincodeSupport.userRunsCC {
			if val {
				chaincodeLogger.Debug("sending READY")
				ccMsg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_READY}
				go handler.triggerNextState(ccMsg, true)
			} else {
				chaincodeLogger.Errorf("Error during startup .. not sending READY")
			}
		} else {
			chaincodeLogger.Warningf("trying to manually run chaincode when not in devmode ?")
		}
	}
}

func (handler *Handler) enterBusyState(e *fsm.Event, state string) {
	chaincodeLogger.Debugf("enterBusyState")
	go func() {
		msg, _ := e.Args[0].(*pb.ChaincodeMessage)
		if chaincodeLogger.IsEnabledFor(logging.DEBUG) {
			chaincodeLogger.Debugf("[%x]state is %s", shorttxid(msg.Txid), state)
		}
		var triggerNextStateMsg *pb.ChaincodeMessage
		var resTemp []byte
		var err error
		defer func() {
			handler.triggerNextState(triggerNextStateMsg, true)
		}()

		errHandler := func(payload []byte, errFmt string, errArgs ...interface{}) {
			chaincodeLogger.Errorf(errFmt, errArgs)
			triggerNextStateMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: payload, Txid: msg.Txid}
		}

		if msg.Type.String() == pb.ChaincodeMessage_PUT_STATE.String() {
			putStateInfo := &pb.PutStateInfo{}
			unmarshalErr := proto.Unmarshal(msg.Payload, putStateInfo)
			if unmarshalErr != nil {
				errHandler([]byte(unmarshalErr.Error()), "[%x]Unable to decipher payload. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR)
				return
			}
			res.resLocker.Lock()
			res.tempResult[putStateInfo.Key] = string(putStateInfo.Value)
			res.singleTxResult[putStateInfo.Key] = string(putStateInfo.Value)
			res.resLocker.Unlock()
		} else if msg.Type.String() == pb.ChaincodeMessage_DEL_STATE.String() {
			// Invoke ledger to delete state
			key := string(msg.Payload)
			res.resLocker.Lock()
			res.tempResult[key] = ""
			res.singleTxResult[key] = ""
			res.resLocker.Unlock()
		} else if msg.Type.String() == pb.ChaincodeMessage_DEL_STATEN.String() {
			keyn := strings.Split(string(msg.Payload), "\n")
			res.resLocker.Lock()
			for _, y := range keyn {
				res.tempResult[y] = ""
				res.singleTxResult[y] = ""
			}
			res.resLocker.Unlock()
		} else if msg.Type.String() == pb.ChaincodeMessage_INVOKE_CHAINCODE.String() {
			if chaincodeLogger.IsEnabledFor(logging.DEBUG) {
				chaincodeLogger.Debugf("[%x] C-call-C", shorttxid(msg.Txid))
			}
			type InvokeChaincode struct {
				ChaincodeName    string
				ChaincodeVersion string
				Args             [][]byte
			}
			var chaincodeSpec InvokeChaincode
			unmarshalErr := json.Unmarshal(msg.Payload, &chaincodeSpec)
			if unmarshalErr != nil {
				errHandler([]byte(unmarshalErr.Error()), "[%x]Unable to decipher payload. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR)
				return
			}

			// Get the chaincodeID to invoke. The chaincodeID to be called may
			// contain composite info like "chaincode-name:version/channel-name"
			// We are not using version now but default to the latest

			cccid := ccprovider.NewCCContext("tjfoc", chaincodeSpec.ChaincodeName, chaincodeSpec.ChaincodeVersion, msg.Txid, theChaincodeSupport.ip, theChaincodeSupport.port, false)
			canName := cccid.GetCanonicalName()
			if _, ok := theChaincodeSupport.runningChaincodes.chaincodeMap[canName]; !ok {
				errHandler([]byte("chaincode is not exist"), "[%x]chaincode is not exist. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR)
				return
			}
			// TODO: Need to handle timeout correctly
			timeout := time.Duration(30000) * time.Millisecond

			cMsg := &pb.ChaincodeInput{Args: chaincodeSpec.Args}
			payload, err := proto.Marshal(cMsg)
			if err != nil {
				chaincodeStartLogger.Error(err)
			}
			var ccMsg *pb.ChaincodeMessage
			if len(chaincodeSpec.Args) != 0 {
				ccMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: msg.Txid}
			} else {
				errHandler([]byte("Invokechaincode args error"), "[%x]Invokechaincode args error. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR)
				return
			}
			response, execErr := theChaincodeSupport.Execute(context.Background(), cccid, ccMsg, timeout)
			//payload is marshalled and send to the calling chaincode's shim which unmarshals and
			//sends it to chaincode
			resTemp = nil
			if execErr != nil {
				err = execErr
			} else {
				resTemp, err = proto.Marshal(response)
			}
		}

		if err != nil {
			errHandler([]byte(err.Error()), "[%x]Failed to handle %s. Sending %s", shorttxid(msg.Txid), msg.Type.String(), pb.ChaincodeMessage_ERROR)
			return
		}

		// Send response msg back to chaincode.
		if chaincodeLogger.IsEnabledFor(logging.DEBUG) {
			chaincodeLogger.Debugf("[%x]Completed %s. Sending %s", shorttxid(msg.Txid), msg.Type.String(), pb.ChaincodeMessage_RESPONSE)
		}
		triggerNextStateMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: resTemp, Txid: msg.Txid}
	}()
}

//afterGetState 处理来自chaincode的GET_STATE消息
func (handler *Handler) afterGetState(e *fsm.Event, state string) {
	msg, ok := e.Args[0].(*pb.ChaincodeMessage)
	if !ok {
		e.Cancel(fmt.Errorf("Received unexpected message type"))
		return
	}
	chaincodeLogger.Debugf("[%x]Received %s, invoking get state from ledger", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_STATE)

	// Query ledger for state
	handler.handleGetState(msg)
}

func (handler *Handler) handleGetState(msg *pb.ChaincodeMessage) {
	// The defer followed by triggering a go routine dance is needed to ensure that the previous state transition
	// is completed before the next one is triggered. The previous state transition is deemed complete only when
	// the afterGetState function is exited. Interesting bug fix!!
	go func() {
		// Check if this is the unique state request from this chaincode txid
		// uniqueReq := handler.createTXIDEntry(msg.Txid)
		// if !uniqueReq {
		// 	// Drop this request
		// 	chaincodeLogger.Error("Another state request pending for this Txid. Cannot process.")
		// 	return
		// }

		var serialSendMsg *pb.ChaincodeMessage

		defer func() {
			// handler.deleteTXIDEntry(msg.Txid)
			if chaincodeLogger.IsEnabledFor(logging.DEBUG) {
				chaincodeLogger.Debugf("[%x]handleGetState serial send %s",
					shorttxid(serialSendMsg.Txid), serialSendMsg.Type)
			}
			handler.serialSendAsync(serialSendMsg, nil)
		}()

		key := string(msg.Payload)
		var resTemp []byte
		res.resLocker.Lock()
		if v, ok := res.tempResult[key]; ok {
			resTemp = []byte(v)
		} else {
			result := worldstate.GetWorldState().Search(key)
			if result.Err == nil {
				resTemp = []byte(result.Value)
			} else {
				chaincodeLogger.Errorf("can't search %s ,error %s", key, result.Err)
			}
		}
		res.resLocker.Unlock()
		if resTemp == nil {
			//The state object being requested does not exist
			chaincodeLogger.Debugf("[%x]No state associated with key: %s. Sending %s with an empty payload",
				shorttxid(msg.Txid), key, pb.ChaincodeMessage_RESPONSE)
			serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: resTemp, Txid: msg.Txid}
		} else {
			// Send response msg back to chaincode. GetState will not trigger event
			if chaincodeLogger.IsEnabledFor(logging.DEBUG) {
				chaincodeLogger.Debugf("[%x]Got state. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_RESPONSE)
			}
			serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: resTemp, Txid: msg.Txid}
		}

	}()
}

//afterGetStaten 处理来自chaincode的GET_STATEN消息
func (handler *Handler) afterGetStaten(e *fsm.Event, state string) {
	msg, ok := e.Args[0].(*pb.ChaincodeMessage)
	if !ok {
		e.Cancel(fmt.Errorf("Received unexpected message type"))
		return
	}
	chaincodeLogger.Debugf("[%x]Received %s, invoking get state from ledger", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_STATE)

	// Query ledger for state
	handler.handleGetStaten(msg)
}

func (handler *Handler) handleGetStaten(msg *pb.ChaincodeMessage) {
	go func() {
		var serialSendMsg *pb.ChaincodeMessage
		defer func() {
			if chaincodeLogger.IsEnabledFor(logging.DEBUG) {
				chaincodeLogger.Debugf("[%x]handleGetState serial send %s",
					shorttxid(serialSendMsg.Txid), serialSendMsg.Type)
			}
			handler.serialSendAsync(serialSendMsg, nil)
		}()
		if msg.Type.String() == pb.ChaincodeMessage_GET_STATEN.String() {
			keyn := strings.Split(string(msg.Payload), "\n")
			//首先读取worldstate,之后拿缓存中数据更新读取到的map
			type ReturnKV struct {
				Kv map[string]string
			}
			tempRes := &ReturnKV{Kv: make(map[string]string)}
			for _, v := range worldstate.GetWorldState().Searchn(keyn) {
				tempRes.Kv[v.Key] = v.Value
			}
			//读取缓存
			res.resLocker.Lock()
			for _, y := range keyn {
				v, ok := res.tempResult[y]
				if ok {
					tempRes.Kv[y] = v
				}
			}
			res.resLocker.Unlock()
			tempdata, _ := json.Marshal(tempRes)
			chaincodeLogger.Debugf("[%x]No state associated with keyn: %s. Sending %s with an empty payload", shorttxid(msg.Txid), keyn, pb.ChaincodeMessage_RESPONSE)
			serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: tempdata, Txid: msg.Txid}
		}
	}()
}

func (handler *Handler) enterGetStateByPrefix(e *fsm.Event, state string) {

	chaincodeLogger.Debugf("enterGetStateByPrefix++++++++")
	go func() {
		var serialSendMsg *pb.ChaincodeMessage
		defer func() {
			if chaincodeLogger.IsEnabledFor(logging.DEBUG) {
				chaincodeLogger.Debugf("[%x]handleGetState serial send %s",
					shorttxid(serialSendMsg.Txid), serialSendMsg.Type)
			}
			handler.serialSendAsync(serialSendMsg, nil)
		}()
		msg, _ := e.Args[0].(*pb.ChaincodeMessage)
		if msg.Type.String() == pb.ChaincodeMessage_GET_STATE_BY_PREFIX.String() {
			//先从worldstate查询,之后在缓存中得到最新的结果
			key := string(msg.Payload)
			var err error
			type ReturnKV struct {
				Kv map[string]string
			}
			tempRes := &ReturnKV{Kv: make(map[string]string)}

			tempRes.Kv, err = worldstate.GetWorldState().SearchPrefix(key)

			if err == nil {
				//The state object being requested does not exist
				regKey := key + "*"
				res.resLocker.Lock()
				for k, v := range res.tempResult {
					if ok, _ := regexp.Match(regKey, []byte(k)); ok {
						tempRes.Kv[k] = v
					}
				}
				res.resLocker.Unlock()
				tempData, _ := json.Marshal(tempRes)
				chaincodeLogger.Debugf("[%x]No state associated with key: %s. Sending %s with an empty payload",
					shorttxid(msg.Txid), key, pb.ChaincodeMessage_RESPONSE)
				serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: tempData, Txid: msg.Txid}
			} else {
				// Send response msg back to chaincode. GetState will not trigger event
				if chaincodeLogger.IsEnabledFor(logging.DEBUG) {
					chaincodeLogger.Debugf("[%x]Got state. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_RESPONSE)
				}
				serialSendMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: []byte(fmt.Sprintf("%s\n", err)), Txid: msg.Txid}
			}
		}

	}()
}
