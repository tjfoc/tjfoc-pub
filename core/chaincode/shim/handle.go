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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/looplab/fsm"
	"strconv"
	"strings"
	"sync"

	pb "github.com/tjfoc/tjfoc/protos/chaincode"
)

type PeerChaincodeStream interface {
	Send(*pb.ChaincodeMessage) error
	Recv() (*pb.ChaincodeMessage, error)
	CloseSend() error
}

type nextStateInfo struct {
	msg      *pb.ChaincodeMessage
	sendToCC bool
}

type Handler struct {
	sync.RWMutex
	//shim to peer grpc serializer. User only in serialSend
	serialLock sync.Mutex
	To         string
	ChatStream PeerChaincodeStream
	FSM        *fsm.FSM
	cc         Chaincode
	// Multiple queries (and one transaction) with different txids can be executing in parallel for this chaincode
	// responseChannel is the channel on which responses are communicated by the shim to the chaincodeStub.
	responseChannel map[string]chan pb.ChaincodeMessage
	nextState       chan *nextStateInfo
}

// NewChaincodeHandler returns a new instance of the shim side handler.
func newChaincodeHandler(peerChatStream PeerChaincodeStream, chaincode Chaincode) *Handler {
	v := &Handler{
		ChatStream: peerChatStream,
		cc:         chaincode,
	}
	v.responseChannel = make(map[string]chan pb.ChaincodeMessage)
	v.nextState = make(chan *nextStateInfo)

	// Create the shim side FSM
	v.FSM = fsm.NewFSM(
		"created",
		fsm.Events{
			{Name: pb.ChaincodeMessage_REGISTERED.String(), Src: []string{"created"}, Dst: "established"},
			{Name: pb.ChaincodeMessage_READY.String(), Src: []string{"established"}, Dst: "ready"},
			{Name: pb.ChaincodeMessage_ERROR.String(), Src: []string{"init"}, Dst: "established"},
			{Name: pb.ChaincodeMessage_RESPONSE.String(), Src: []string{"init"}, Dst: "init"},
			{Name: pb.ChaincodeMessage_INIT.String(), Src: []string{"ready"}, Dst: "ready"},
			{Name: pb.ChaincodeMessage_TRANSACTION.String(), Src: []string{"ready"}, Dst: "ready"},
			{Name: pb.ChaincodeMessage_RESPONSE.String(), Src: []string{"ready"}, Dst: "ready"},
			{Name: pb.ChaincodeMessage_ERROR.String(), Src: []string{"ready"}, Dst: "ready"},
			{Name: pb.ChaincodeMessage_COMPLETED.String(), Src: []string{"init"}, Dst: "ready"},
			{Name: pb.ChaincodeMessage_COMPLETED.String(), Src: []string{"ready"}, Dst: "ready"},
		},
		fsm.Callbacks{
			"before_" + pb.ChaincodeMessage_REGISTERED.String():  func(e *fsm.Event) { v.beforeRegistered(e) },
			"after_" + pb.ChaincodeMessage_RESPONSE.String():     func(e *fsm.Event) { v.afterResponse(e) },
			"after_" + pb.ChaincodeMessage_ERROR.String():        func(e *fsm.Event) { v.afterError(e) },
			"before_" + pb.ChaincodeMessage_INIT.String():        func(e *fsm.Event) { v.beforeInit(e) },
			"before_" + pb.ChaincodeMessage_TRANSACTION.String(): func(e *fsm.Event) { v.beforeTransaction(e) },
		},
	)
	return v
}

func (handler *Handler) afterResponse(e *fsm.Event) {
	msg, ok := e.Args[0].(*pb.ChaincodeMessage)
	if !ok {
		e.Cancel(fmt.Errorf("Received unexpected message type"))
		return
	}

	if err := handler.sendChannel(msg); err != nil {
		chaincodeLogger.Errorf("[%s]error sending %s (state:%s): %s", shorttxid(msg.Txid), msg.Type, handler.FSM.Current(), err)
	} else {
		chaincodeLogger.Debugf("[%s]Received %s, communicated (state:%s)", shorttxid(msg.Txid), msg.Type, handler.FSM.Current())
	}
}

func (handler *Handler) sendChannel(msg *pb.ChaincodeMessage) error {
	handler.Lock()
	defer handler.Unlock()
	if handler.responseChannel == nil {
		return fmt.Errorf("[%s]Cannot send message response channel", shorttxid(msg.Txid))
	}
	if handler.responseChannel[msg.Txid] == nil {
		return fmt.Errorf("[%s]sendChannel does not exist", shorttxid(msg.Txid))
	}

	chaincodeLogger.Debugf("[%s]before send", shorttxid(msg.Txid))
	handler.responseChannel[msg.Txid] <- *msg
	chaincodeLogger.Debugf("[%s]after send", shorttxid(msg.Txid))

	return nil
}

//更改本地状态机状态，触发chartWithPeer for循环， 如过 send=true，则会通过grpc把msg发给Peer
func (handler *Handler) triggerNextState(msg *pb.ChaincodeMessage, send bool) {
	handler.nextState <- &nextStateInfo{msg, send}
}

// chaincode的初始化.
func (handler *Handler) handleInit(msg *pb.ChaincodeMessage) {
	chaincodeLogger.Info("entry handleInit")
	go func() {
		var nextStateMsg *pb.ChaincodeMessage
		defer func() {
			handler.triggerNextState(nextStateMsg, true)
		}()
		errFunc := func(err error, payload []byte, ce *pb.ChaincodeEvent, errFmt string, args ...string) *pb.ChaincodeMessage {
			if err != nil {
				chaincodeLogger.Errorf("in errFunc ,err:%s\n", err)
				// Send ERROR message to chaincode support and change state
				if payload == nil {
					payload = []byte(err.Error())
				}
				chaincodeLogger.Errorf(errFmt, args)
				return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: payload, Txid: msg.Txid, ChaincodeEvent: ce}
			}
			return nil
		}
		// Get the function and args from Payload
		input := &pb.ChaincodeInput{}
		unmarshalErr := proto.Unmarshal(msg.Payload, input)
		if unmarshalErr != nil {
			chaincodeLogger.Infof("proto.Unmarshal unmarshalErr: %s\n", unmarshalErr)
		}
		if nextStateMsg = errFunc(unmarshalErr, nil, nil, "[%s]Incorrect payload format. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR.String()); nextStateMsg != nil {
			return
		}

		// Call chaincode's Run
		// Create the ChaincodeStub which the chaincode can use to callback
		stub := new(ChaincodeStub)
		err := stub.init(handler, msg.Txid, input)
		if nextStateMsg = errFunc(err, nil, stub.chaincodeEvent, "[%s]Init get error response [%s]. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR.String()); nextStateMsg != nil {
			return
		}

		res := handler.cc.Init(stub)

		resBytes, err := proto.Marshal(&res)
		if nextStateMsg = errFunc(err, nil, stub.chaincodeEvent, "[%s]Init marshal response error [%s]. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR.String()); nextStateMsg != nil {
			return
		}

		// Send COMPLETED message to chaincode support and change state
		nextStateMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Payload: resBytes, Txid: msg.Txid, ChaincodeEvent: stub.chaincodeEvent}
		chaincodeLogger.Debugf("[%s]Init succeeded. Sending %s to peer", shorttxid(msg.Txid), pb.ChaincodeMessage_COMPLETED)
	}()
}

// chaincode 执行交易
func (handler *Handler) handleTransaction(msg *pb.ChaincodeMessage) {
	// The defer followed by triggering a go routine dance is needed to ensure that the previous state transition
	// is completed before the next one is triggered. The previous state transition is deemed complete only when
	// the beforeInit function is exited. Interesting bug fix!!
	go func() {
		//better not be nil
		var nextStateMsg *pb.ChaincodeMessage
		defer func() {
			handler.triggerNextState(nextStateMsg, true)
		}()
		errFunc := func(err error, ce *pb.ChaincodeEvent, errStr string, args ...string) *pb.ChaincodeMessage {
			if err != nil {
				payload := []byte(err.Error())
				chaincodeLogger.Errorf(errStr, args)
				return &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: payload, Txid: msg.Txid, ChaincodeEvent: ce}
			}
			return nil
		}

		// Get the function and args from Payload
		input := &pb.ChaincodeInput{}
		unmarshalErr := proto.Unmarshal(msg.Payload, input)
		if nextStateMsg = errFunc(unmarshalErr, nil, "[%s]Incorrect payload format. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR.String()); nextStateMsg != nil {
			return
		}
		chaincodeLogger.Debugf("proto.Unmarshal input: %v\n", input)
		// Call chaincode's Run
		// Create the ChaincodeStub which the chaincode can use to callback
		stub := new(ChaincodeStub)
		err := stub.init(handler, msg.Txid, input)
		if nextStateMsg = errFunc(err, stub.chaincodeEvent, "[%s]Transaction execution failed. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR.String()); nextStateMsg != nil {
			return
		}

		res := handler.cc.Invoke(stub)

		// Endorser will handle error contained in Response.
		resBytes, err := proto.Marshal(&res)
		if nextStateMsg = errFunc(err, stub.chaincodeEvent, "[%s]Transaction execution failed. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_ERROR.String()); nextStateMsg != nil {
			return
		}

		// Send COMPLETED message to chaincode support and change state
		chaincodeLogger.Debugf("[%s]Transaction completed. Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_COMPLETED)
		nextStateMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Payload: resBytes, Txid: msg.Txid, ChaincodeEvent: stub.chaincodeEvent}
	}()
}

// beforeRegistered is called to handle the REGISTERED message.
func (handler *Handler) beforeRegistered(e *fsm.Event) {
	if _, ok := e.Args[0].(*pb.ChaincodeMessage); !ok {
		e.Cancel(fmt.Errorf("Received unexpected message type"))
		return
	}
	chaincodeLogger.Debugf("Received %s, ready for invocations", pb.ChaincodeMessage_REGISTERED)
}

// beforeInit will initialize the chaincode if entering init from established.
func (handler *Handler) beforeInit(e *fsm.Event) {
	chaincodeLogger.Debugf("Entered state %s", handler.FSM.Current())
	msg, ok := e.Args[0].(*pb.ChaincodeMessage)
	if !ok {
		e.Cancel(fmt.Errorf("Received unexpected message type"))
		return
	}
	chaincodeLogger.Debugf("[%s]Received %s, initializing chaincode", shorttxid(msg.Txid), msg.Type.String())
	if msg.Type.String() == pb.ChaincodeMessage_INIT.String() {
		// Call the chaincode's Run function to initialize
		handler.handleInit(msg)
	}
}

func (handler *Handler) createChannel(txid string) (chan pb.ChaincodeMessage, error) {
	handler.Lock()
	defer handler.Unlock()
	if handler.responseChannel == nil {
		return nil, fmt.Errorf("[%s]Cannot create response channel", shorttxid(txid))
	}
	if handler.responseChannel[txid] != nil {
		return nil, fmt.Errorf("[%s]Channel exists", shorttxid(txid))
	}
	c := make(chan pb.ChaincodeMessage)
	handler.responseChannel[txid] = c
	return c, nil
}
func (handler *Handler) deleteChannel(txid string) {
	handler.Lock()
	defer handler.Unlock()
	if handler.responseChannel != nil {
		delete(handler.responseChannel, txid)
	}
}

// TODO: Implement method to get and put entire state map and not one key at a time?
// handleGetState communicates with the validator to fetch the requested state information from the ledger.
func (handler *Handler) handleGetState(key string, txid string) ([]byte, error) {
	// Create the channel on which to communicate the response from validating peer
	var respChan chan pb.ChaincodeMessage
	var err error
	if respChan, err = handler.createChannel(txid); err != nil {
		return nil, err
	}

	defer handler.deleteChannel(txid)

	// Send GET_STATE message to validator chaincode support
	payload := []byte(key)
	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE, Payload: payload, Txid: txid}
	chaincodeLogger.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_STATE)

	var responseMsg pb.ChaincodeMessage

	if responseMsg, err = handler.sendReceive(msg, respChan); err != nil {
		return nil, errors.New(fmt.Sprintf("[%s]error sending GET_STATE %s", shorttxid(txid), err))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		chaincodeLogger.Debugf("[%s]GetState received payload %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_RESPONSE)
		return responseMsg.Payload, nil
	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		chaincodeLogger.Errorf("[%s]GetState received error %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR)
		return nil, errors.New(fmt.Sprintf("peer return error when getstate [%s]", key))
	}

	// Incorrect chaincode message received
	return nil, errors.New(fmt.Sprintf("[%s]Incorrect chaincode message %s received. Expecting %s or %s", shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR))
}

// handleDelState communicates with the validator to delete a key from the state in the ledger.
func (handler *Handler) handleDelState(key string, txid string) error {
	// Create the channel on which to communicate the response from validating peer
	var respChan chan pb.ChaincodeMessage
	var err error
	if respChan, err = handler.createChannel(txid); err != nil {
		return err
	}

	defer handler.deleteChannel(txid)

	// Send DEL_STATE message to validator chaincode support
	payload := []byte(key)
	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_DEL_STATE, Payload: payload, Txid: txid}
	chaincodeLogger.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_DEL_STATE)

	var responseMsg pb.ChaincodeMessage

	if responseMsg, err = handler.sendReceive(msg, respChan); err != nil {
		return errors.New(fmt.Sprintf("[%s]error sending DEL_STATE %s", shorttxid(msg.Txid), pb.ChaincodeMessage_DEL_STATE))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		chaincodeLogger.Debugf("[%s]Received %s. Successfully deleted state", msg.Txid, pb.ChaincodeMessage_RESPONSE)
		return nil
	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		chaincodeLogger.Errorf("[%s]Received %s. Payload: %s", msg.Txid, pb.ChaincodeMessage_ERROR, responseMsg.Payload)
		return errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return errors.New(fmt.Sprintf("[%s]Incorrect chaincode message %s received. Expecting %s or %s", shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR))
}

type ReturnKv struct {
	Kv map[string][]byte
}

//前缀查询
func (handler *Handler) handleGetStateByPrefix(key string, txid string) (map[string][]byte, error) {
	var respChan chan pb.ChaincodeMessage
	var err error
	if respChan, err = handler.createChannel(txid); err != nil {
		return nil, err
	}
	defer handler.deleteChannel(txid)

	// Send DEL_STATE message to validator chaincode support
	payload := []byte(key)
	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE_BY_PREFIX, Payload: payload, Txid: txid}
	chaincodeLogger.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_STATE_BY_PREFIX)

	var responseMsg pb.ChaincodeMessage

	if responseMsg, err = handler.sendReceive(msg, respChan); err != nil {
		return nil, errors.New(fmt.Sprintf("[%s]error sending GetStateByPrefix %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_STATE_BY_PREFIX))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		var tempRes ReturnKv
		json.Unmarshal(responseMsg.Payload, &tempRes)
		return tempRes.Kv, nil
	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		chaincodeLogger.Errorf("[%s]Received %s. Payload: %s", msg.Txid, pb.ChaincodeMessage_ERROR, responseMsg.Payload)
		return nil, errors.New(string(responseMsg.Payload[:]))
	}
	// Incorrect chaincode message received
	return nil, errors.New(fmt.Sprintf("[%s]Incorrect chaincode message %s received. Expecting %s or %s", shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR))
}

func (handler *Handler) handleGetStaten(keyn []string, txid string) (map[string][]byte, error) {
	var respChan chan pb.ChaincodeMessage
	var err error
	if respChan, err = handler.createChannel(txid); err != nil {
		return nil, err
	}
	defer handler.deleteChannel(txid)

	var payload []byte
	var payloadval []string
	for i := 0; i < len(keyn); i++ {
		payloadval = append(payloadval, keyn[i])
	}
	res := strings.Join(payloadval, "\n")
	payload = []byte(res)
	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATEN, Payload: payload, Txid: txid}
	chaincodeLogger.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_GET_STATEN)

	var responseMsg pb.ChaincodeMessage
	if responseMsg, err = handler.sendReceive(msg, respChan); err != nil {
		return nil, errors.New(fmt.Sprintf("[%s]error sending GET_STATEN %s", shorttxid(txid), err))
	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		chaincodeLogger.Debugf("[%s]GetStaten received payload %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_RESPONSE)
		var tempRes ReturnKv
		json.Unmarshal(responseMsg.Payload, &tempRes)
		return tempRes.Kv, nil
	}
	return nil, errors.New(fmt.Sprintf("[%s]Incorrect chaincode message %s received. Expecting %s or %s", shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR))
}

// handleDelStaten communicates with the validator to delete a key from the state in the ledger.
func (handler *Handler) handleDelStaten(keyn []string, txid string) error {
	var respChan chan pb.ChaincodeMessage
	var err error
	if respChan, err = handler.createChannel(txid); err != nil {
		return err
	}
	defer handler.deleteChannel(txid)

	var payload []byte
	var payloadval []string
	for i := 0; i < len(keyn); i++ {
		payloadval = append(payloadval, keyn[i])
	}
	res := strings.Join(payloadval, "\n")
	payload = []byte(res)
	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_DEL_STATEN, Payload: payload, Txid: txid}
	chaincodeLogger.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_DEL_STATEN)

	var responseMsg pb.ChaincodeMessage

	if responseMsg, err = handler.sendReceive(msg, respChan); err != nil {
		return errors.New(fmt.Sprintf("[%s]error sending DEL_STATEN %s", shorttxid(msg.Txid), pb.ChaincodeMessage_DEL_STATEN))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		chaincodeLogger.Debugf("[%s]Received %s. Successfully deleted state", msg.Txid, pb.ChaincodeMessage_RESPONSE)
		return nil
	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		chaincodeLogger.Errorf("[%s]Received %s. Payload: %s", msg.Txid, pb.ChaincodeMessage_ERROR, responseMsg.Payload)
		return errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return errors.New(fmt.Sprintf("[%s]Incorrect chaincode message %s received. Expecting %s or %s", shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR))
}

// handlePutState communicates with the validator to put state information into the ledger.
func (handler *Handler) handlePutState(key string, value []byte, txid string) error {
	// Check if this is a transaction
	chaincodeLogger.Debugf("[%s]Inside putstate", shorttxid(txid))

	//we constructed a valid object. No need to check for error
	payloadBytes, _ := proto.Marshal(&pb.PutStateInfo{Key: key, Value: value})

	// Create the channel on which to communicate the response from validating peer
	var respChan chan pb.ChaincodeMessage
	var err error
	if respChan, err = handler.createChannel(txid); err != nil {
		return err
	}

	defer handler.deleteChannel(txid)

	// Send PUT_STATE message to validator chaincode support
	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_PUT_STATE, Payload: payloadBytes, Txid: txid}
	chaincodeLogger.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_PUT_STATE)

	var responseMsg pb.ChaincodeMessage

	if responseMsg, err = handler.sendReceive(msg, respChan); err != nil {
		return errors.New(fmt.Sprintf("[%s]error sending PUT_STATE %s", msg.Txid, err))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		chaincodeLogger.Debugf("[%s]Received %s. Successfully updated state", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_RESPONSE)
		return nil
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		chaincodeLogger.Errorf("[%s]Received %s. Payload: %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR, responseMsg.Payload)
		return errors.New(string(responseMsg.Payload[:]))
	}

	// Incorrect chaincode message received
	return errors.New(fmt.Sprintf("[%s]Incorrect chaincode message %s received. Expecting %s or %s", shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR))
}
func (handler *Handler) handleRequireCompare(originvalue []byte, dealvalue int64, txid string) (int64, error) {
	chaincodeLogger.Debugf("[%s]Inside RequireCompare", shorttxid(txid))
	var respChan chan pb.ChaincodeMessage
	var err error
	if respChan, err = handler.createChannel(txid); err != nil {
		return int64(0), err
	}

	defer handler.deleteChannel(txid)
	type AAA struct {
		Originvalue []byte
		Dealvalue   int64
	}
	aaa := &AAA{
		Originvalue: originvalue,
		Dealvalue:   dealvalue,
	}
	d, _ := json.Marshal(aaa)
	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_REQUIRE_COMPARE, Payload: d, Txid: txid}

	var responseMsg pb.ChaincodeMessage
	if responseMsg, err = handler.sendReceive(msg, respChan); err != nil {
		return int64(0), errors.New(fmt.Sprintf("[%s]error sending PUT_STATE %s", msg.Txid, err))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		//success response
		chaincodeLogger.Debugf("[%s]Received %s. Successfully updated state", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_RESPONSE)
		i, _ := strconv.ParseInt(string(responseMsg.Payload), 10, 64)
		return i, nil
	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		chaincodeLogger.Errorf("[%s]Received %s. Payload: %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR, responseMsg.Payload)
		return int64(0), errors.New(string(responseMsg.Payload[:]))
	}
	return int64(0), errors.New(fmt.Sprintf("[%s]Incorrect chaincode message %s received. Expecting %s or %s", shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR))
}
func (handler *Handler) handleRequireCrypt(value int64, txid string) ([]byte, error) {
	chaincodeLogger.Debugf("[%s]Inside RequireCrypt", shorttxid(txid))

	var respChan chan pb.ChaincodeMessage
	var err error
	if respChan, err = handler.createChannel(txid); err != nil {
		return nil, err
	}

	defer handler.deleteChannel(txid)

	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_REQUIRE_CRYPT, Payload: []byte(strconv.FormatInt(value, 10)), Txid: txid}

	var responseMsg pb.ChaincodeMessage
	if responseMsg, err = handler.sendReceive(msg, respChan); err != nil {
		return nil, errors.New(fmt.Sprintf("[%s]error sending PUT_STATE %s", msg.Txid, err))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		//success response
		chaincodeLogger.Debugf("[%s]Received %s. Successfully updated state", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_RESPONSE)
		return responseMsg.Payload, nil
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		chaincodeLogger.Errorf("[%s]Received %s. Payload: %s", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR, responseMsg.Payload)
		return nil, errors.New(string(responseMsg.Payload[:]))
	}
	return nil, errors.New(fmt.Sprintf("[%s]Incorrect chaincode message %s received. Expecting %s or %s", shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR))
}
func (handler *Handler) afterError(e *fsm.Event) {
	msg, ok := e.Args[0].(*pb.ChaincodeMessage)
	if !ok {
		e.Cancel(fmt.Errorf("Received unexpected message type"))
		return
	}

	/* TODO- revisit. This may no longer be needed with the serialized/streamlined messaging model
	 * There are two situations in which the ERROR event can be triggered:
	 * 1. When an error is encountered within handleInit or handleTransaction - some issue at the chaincode side; In this case there will be no responseChannel and the message has been sent to the validator.
	 * 2. The chaincode has initiated a request (get/put/del state) to the validator and is expecting a response on the responseChannel; If ERROR is received from validator, this needs to be notified on the responseChannel.
	 */
	if err := handler.sendChannel(msg); err == nil {
		chaincodeLogger.Debugf("[%s]Error received from validator %s, communicated(state:%s)", shorttxid(msg.Txid), msg.Type, handler.FSM.Current())
	}
}
func (handler *Handler) createResponse(status int32, payload []byte) pb.Response {
	return pb.Response{Status: status, Payload: payload}
}

/*
// handleInvokeChaincode communicates with the validator to invoke another chaincode.
func (handler *Handler) handleInvokeChaincode(chaincodeName string, chaincodeVersion string, args [][]byte, txid string) pb.Response {
	//we constructed a valid object. No need to check for error
	type InvokeChaincode struct {
		ChaincodeName    string
		ChaincodeVersion string
		Args             [][]byte
	}
	tempSpec := InvokeChaincode{ChaincodeName: chaincodeName, ChaincodeVersion: chaincodeVersion, Args: args}
	payloadBytes, _ := json.Marshal(tempSpec)
	// payloadBytes, _ := proto.Marshal(&pb.ChaincodeSpec{ChaincodeId: &pb.ChaincodeID{Name: chaincodeName}, Input: &pb.ChaincodeInput{Args: args}})

	// Create the channel on which to communicate the response from validating peer
	var respChan chan pb.ChaincodeMessage
	var err error
	if respChan, err = handler.createChannel(txid); err != nil {
		return handler.createResponse(ERROR, []byte(err.Error()))
	}

	defer handler.deleteChannel(txid)

	// Send INVOKE_CHAINCODE message to validator chaincode support
	msg := &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_INVOKE_CHAINCODE, Payload: payloadBytes, Txid: txid}
	chaincodeLogger.Debugf("[%s]Sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_INVOKE_CHAINCODE)

	var responseMsg pb.ChaincodeMessage

	if responseMsg, err = handler.sendReceive(msg, respChan); err != nil {
		return handler.createResponse(ERROR, []byte(fmt.Sprintf("[%s]error sending %s", shorttxid(msg.Txid), pb.ChaincodeMessage_INVOKE_CHAINCODE)))
	}

	if responseMsg.Type.String() == pb.ChaincodeMessage_RESPONSE.String() {
		// Success response
		chaincodeLogger.Debugf("[%s]Received %s. Successfully invoked chaincode", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_RESPONSE)
		respMsg := &pb.ChaincodeMessage{}
		if err = proto.Unmarshal(responseMsg.Payload, respMsg); err != nil {
			return handler.createResponse(ERROR, []byte(err.Error()))
		}
		if respMsg.Type == pb.ChaincodeMessage_COMPLETED {
			// Success response
			chaincodeLogger.Debugf("[%s]Received %s. Successfully invoed chaincode", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_RESPONSE)
			res := &pb.Response{}
			if err = proto.Unmarshal(respMsg.Payload, res); err != nil {
				return handler.createResponse(ERROR, []byte(err.Error()))
			}
			return *res
		}
		chaincodeLogger.Errorf("[%s]Received %s. Error from chaincode", shorttxid(responseMsg.Txid), respMsg.Type.String())
		return handler.createResponse(ERROR, responseMsg.Payload)
	}
	if responseMsg.Type.String() == pb.ChaincodeMessage_ERROR.String() {
		// Error response
		chaincodeLogger.Errorf("[%s]Received %s.", shorttxid(responseMsg.Txid), pb.ChaincodeMessage_ERROR)
		return handler.createResponse(ERROR, responseMsg.Payload)
	}

	// Incorrect chaincode message received
	return handler.createResponse(ERROR, []byte(fmt.Sprintf("[%s]Incorrect chaincode message %s received. Expecting %s or %s", shorttxid(responseMsg.Txid), responseMsg.Type, pb.ChaincodeMessage_RESPONSE, pb.ChaincodeMessage_ERROR)))
}
*/
//sends a message and selects
func (handler *Handler) sendReceive(msg *pb.ChaincodeMessage, c chan pb.ChaincodeMessage) (pb.ChaincodeMessage, error) {
	errc := make(chan error, 1)
	handler.serialSendAsync(msg, errc)

	//the serialsend above will send an err or nil
	//the select filters that first error(or nil)
	//and continues to wait for the response
	//it is possible that the response triggers first
	//in which case the errc obviously worked and is
	//ignored
	for {
		select {
		case err := <-errc:
			if err == nil {
				continue
			}
			//would have been logged, return false
			return pb.ChaincodeMessage{}, err
		case outmsg, val := <-c:
			if !val {
				return pb.ChaincodeMessage{}, fmt.Errorf("unexpected failure on receive")
			}
			return outmsg, nil
		}
	}
}

func (handler *Handler) beforeTransaction(e *fsm.Event) {
	msg, ok := e.Args[0].(*pb.ChaincodeMessage)
	if !ok {
		e.Cancel(fmt.Errorf("Received unexpected message type"))
		return
	}
	chaincodeLogger.Debugf("[%s]Received %s, invoking transaction on chaincode(Src:%s, Dst:%s)", shorttxid(msg.Txid), msg.Type.String(), e.Src, e.Dst)
	if msg.Type.String() == pb.ChaincodeMessage_TRANSACTION.String() {
		// Call the chaincode's Run function to invoke transaction
		handler.handleTransaction(msg)
	}
}
