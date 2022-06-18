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
	"strconv"
	"strings"
	"testing"

	pb "github.com/tjfoc/tjfoc/protos/chaincode"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/tjfoc/tjfoc/core/common/flogging"
)

// shimTestCC example simple Chaincode implementation
type shimTestCC struct {
}

func (t *shimTestCC) Init(stub ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error

	if len(args) != 4 {
		return Error("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return Error("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return Error("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return Error(err.Error())
	}

	return Success(nil)
}

func (t *shimTestCC) Invoke(stub ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// Make payment of X units from A to B
		return t.invoke(stub, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		// /return t.delete(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke
		return t.query(stub, args)
	}

	return Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

// Transaction makes payment of X units from A to B
func (t *shimTestCC) invoke(stub ChaincodeStubInterface, args []string) pb.Response {
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return Error("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return Error("Failed to get state")
	}
	if Avalbytes == nil {
		return Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return Error(err.Error())
	}

	return Success(nil)
}

// query callback representing the query of a chaincode
func (t *shimTestCC) query(stub ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 1 {
		return Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return Error(jsonResp)
	}

	return Success(Avalbytes)
}

// Test Go shim functionality that can be tested outside of a real chaincode
// context.

// TestShimLogging simply tests that the APIs are working. These tests test
// for correct control over the shim's logging object and the LogLevel
// // function.
// func TestShimLogging(t *testing.T) {
// 	if chaincodeLogger.IsEnabledFor(logging.DEBUG) {
// 		t.Errorf("The chaincodeLogger should not be enabled for DEBUG")
// 	}
// 	if !chaincodeLogger.IsEnabledFor(logging.CRITICAL) {
// 		t.Errorf("The chaincodeLogger should be enabled for CRITICAL")
// 	}
// 	var level LoggingLevel
// 	var err error
// 	level, err = LogLevel("debug")
// 	if err != nil {
// 		t.Errorf("LogLevel(debug) failed")
// 	}
// 	if level != LogDebug {
// 		t.Errorf("LogLevel(debug) did not return LogDebug")
// 	}
// 	level, err = LogLevel("INFO")
// 	if err != nil {
// 		t.Errorf("LogLevel(INFO) failed")
// 	}
// 	if level != LogInfo {
// 		t.Errorf("LogLevel(INFO) did not return LogInfo")
// 	}
// 	level, err = LogLevel("Notice")
// 	if err != nil {
// 		t.Errorf("LogLevel(Notice) failed")
// 	}
// 	if level != LogNotice {
// 		t.Errorf("LogLevel(Notice) did not return LogNotice")
// 	}
// 	level, err = LogLevel("WaRnInG")
// 	if err != nil {
// 		t.Errorf("LogLevel(WaRnInG) failed")
// 	}
// 	if level != LogWarning {
// 		t.Errorf("LogLevel(WaRnInG) did not return LogWarning")
// 	}
// 	level, err = LogLevel("ERRor")
// 	if err != nil {
// 		t.Errorf("LogLevel(ERRor) failed")
// 	}
// 	if level != LogError {
// 		t.Errorf("LogLevel(ERRor) did not return LogError")
// 	}
// 	level, err = LogLevel("critiCAL")
// 	if err != nil {
// 		t.Errorf("LogLevel(critiCAL) failed")
// 	}
// 	if level != LogCritical {
// 		t.Errorf("LogLevel(critiCAL) did not return LogCritical")
// 	}
// 	level, err = LogLevel("foo")
// 	if err == nil {
// 		t.Errorf("LogLevel(foo) did not fail")
// 	}
// 	if level != LogError {
// 		t.Errorf("LogLevel(foo) did not return LogError")
// 	}
// }

// TestChaincodeLogging tests the logging APIs for chaincodes.
// func TestChaincodeLogging(t *testing.T) {

// 	// From start() - We can't call start() from this test
// 	format := logging.MustStringFormatter("%{time:15:04:05.000} [%{module}] %{level:.4s} : %{message}")
// 	backend := logging.NewLogBackend(os.Stderr, "", 0)
// 	backendFormatter := logging.NewBackendFormatter(backend, format)
// 	logging.SetBackend(backendFormatter).SetLevel(logging.Level(shimLoggingLevel), "shim")

// 	foo := NewLogger("foo")
// 	bar := NewLogger("bar")

// 	foo.Debugf("Foo is debugging: %d", 10)
// 	bar.Infof("Bar is informational? %s.", "Yes")
// 	foo.Noticef("NOTE NOTE NOTE")
// 	bar.Warningf("Danger, Danger %s %s", "Will", "Robinson!")
// 	foo.Errorf("I'm sorry Dave, I'm afraid I can't do that.")
// 	bar.Criticalf("PI is not equal to 3.14, we computed it as %.2f", 4.13)

// 	bar.Debug("Foo is debugging:", 10)
// 	foo.Info("Bar is informational?", "Yes.")
// 	bar.Notice("NOTE NOTE NOTE")
// 	foo.Warning("Danger, Danger", "Will", "Robinson!")
// 	bar.Error("I'm sorry Dave, I'm afraid I can't do that.")
// 	foo.Critical("PI is not equal to", 3.14, ", we computed it as", 4.13)

// 	foo.SetLevel(LogWarning)
// 	if foo.IsEnabledFor(LogDebug) {
// 		t.Errorf("'foo' should not be enabled for LogDebug")
// 	}
// 	if !foo.IsEnabledFor(LogCritical) {
// 		t.Errorf("'foo' should be enabled for LogCritical")
// 	}
// 	bar.SetLevel(LogCritical)
// 	if bar.IsEnabledFor(LogDebug) {
// 		t.Errorf("'bar' should not be enabled for LogDebug")
// 	}
// 	if !bar.IsEnabledFor(LogCritical) {
// 		t.Errorf("'bar' should be enabled for LogCritical")
// 	}
// }

type testCase struct {
	name         string
	ccLogLevel   string
	shimLogLevel string
}

func TestSetupChaincodeLogging_shim(t *testing.T) {
	var tc []testCase

	tc = append(tc,
		testCase{"ValidLevels", "debug", "warning"},
		testCase{"EmptyLevels", "", ""},
		testCase{"BadShimLevel", "debug", "war"},
		testCase{"BadCCLevel", "deb", "notice"},
		testCase{"EmptyShimLevel", "error", ""},
		testCase{"EmptyCCLevel", "", "critical"},
	)

	assert := assert.New(t)

	for i := 0; i < len(tc); i++ {
		t.Run(tc[i].name, func(t *testing.T) {
			viper.Set("chaincode.logging.level", tc[i].ccLogLevel)
			viper.Set("chaincode.logging.shim", tc[i].shimLogLevel)

			SetupChaincodeLogging()

			_, ccErr := logging.LogLevel(tc[i].ccLogLevel)
			_, shimErr := logging.LogLevel(tc[i].shimLogLevel)
			if ccErr == nil {
				assert.Equal(strings.ToUpper(tc[i].ccLogLevel), flogging.GetModuleLevel("ccLogger"), "Test case '%s' failed", tc[i].name)
				if shimErr == nil {
					assert.Equal(strings.ToUpper(tc[i].shimLogLevel), flogging.GetModuleLevel("shim"), "Test case '%s' failed", tc[i].name)
				} else {
					assert.Equal(strings.ToUpper(tc[i].ccLogLevel), flogging.GetModuleLevel("shim"), "Test case '%s' failed", tc[i].name)
				}
			} else {
				assert.Equal(flogging.DefaultLevel(), flogging.GetModuleLevel("ccLogger"), "Test case '%s' failed", tc[i].name)
				if shimErr == nil {
					assert.Equal(strings.ToUpper(tc[i].shimLogLevel), flogging.GetModuleLevel("shim"), "Test case '%s' failed", tc[i].name)
				} else {
					assert.Equal(flogging.DefaultLevel(), flogging.GetModuleLevel("shim"), "Test case '%s' failed", tc[i].name)
				}
			}
		})
	}
}

//TestInvoke tests init and invoke along with many of the stub functions
//such as get/put/del/range...
func TestInvoke(t *testing.T) {
	//streamGetter = mockChaincodeStreamGetter
	cc := &shimTestCC{}
	//viper.Set("chaincode.logging.shim", "debug")
	// var err error
	// ccname := "shimTestCC"
	// peerSide := setupcc(ccname, cc)
	// defer mockPeerCCSupport.RemoveCC(ccname)
	//start the shim+chaincode
	go func() {
		// err = Start(cc)
		Start(cc)
	}()

	// done := setuperror()

	// errorFunc := func(ind int, err error) {
	// 	done <- err
	// }

	// //start the mock peer
	// go func() {
	// 	respSet := &mockpeer.MockResponseSet{errorFunc, nil, []*mockpeer.MockResponse{
	// 		&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_REGISTER}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_REGISTERED}}}}
	// 	peerSide.SetResponses(respSet)
	// 	peerSide.SetKeepAlive(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_KEEPALIVE})
	// 	err = peerSide.Run()
	// }()

	// //wait for init
	// processDone(t, done, false)

	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_READY, Txid: "1"})

	// ci := &pb.ChaincodeInput{[][]byte{[]byte("init"), []byte("A"), []byte("100"), []byte("B"), []byte("200")}}
	// payload := utils.MarshalOrPanic(ci)
	// respSet := &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_PUT_STATE, Txid: "2"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: "2"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_PUT_STATE, Txid: "2"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: "2"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "2"}, nil}}}
	// peerSide.SetResponses(respSet)

	// //use the payload computed from prev init
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_INIT, Payload: payload, Txid: "2"})

	// //wait for done
	// processDone(t, done, false)

	// //good invoke
	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE, Txid: "3"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: []byte("100"), Txid: "3"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE, Txid: "3"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: []byte("200"), Txid: "3"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_PUT_STATE, Txid: "3"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: "3"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_PUT_STATE, Txid: "3"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: "3"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "3"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("invoke"), []byte("A"), []byte("B"), []byte("10")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "3"})

	// //wait for done
	// processDone(t, done, false)

	// //bad put
	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE, Txid: "3a"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: []byte("100"), Txid: "3a"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE, Txid: "3a"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: []byte("200"), Txid: "3a"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_PUT_STATE, Txid: "3a"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: "3a"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "3a"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("invoke"), []byte("A"), []byte("B"), []byte("10")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "3a"})

	// //wait for done
	// processDone(t, done, false)

	// //bad get
	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE, Txid: "3b"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: "3b"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "3b"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("invoke"), []byte("A"), []byte("B"), []byte("10")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "3b"})

	// //wait for done
	// processDone(t, done, false)

	// //bad delete
	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_DEL_STATE, Txid: "4"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: "4"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "4"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("delete"), []byte("A")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "4"})

	// //wait for done
	// processDone(t, done, false)

	// //good delete
	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_DEL_STATE, Txid: "4a"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: "4a"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "4a"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("delete"), []byte("A")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "4a"})

	// //wait for done
	// processDone(t, done, false)

	// //bad invoke
	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "5"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("badinvoke")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "5"})

	// //wait for done
	// processDone(t, done, false)

	// //range query

	// //create the response
	// rangeQueryResponse := &pb.QueryResponse{Results: []*pb.QueryResultBytes{
	// 	&pb.QueryResultBytes{ResultBytes: utils.MarshalOrPanic(&lproto.KV{"getputcc", "A", []byte("100")})},
	// 	&pb.QueryResultBytes{ResultBytes: utils.MarshalOrPanic(&lproto.KV{"getputcc", "B", []byte("200")})}},
	// 	HasMore: true}
	// rangeQPayload := utils.MarshalOrPanic(rangeQueryResponse)

	// //create the next response
	// rangeQueryNext := &pb.QueryResponse{Results: nil, HasMore: false}

	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE_BY_RANGE, Txid: "6"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: rangeQPayload, Txid: "6"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_QUERY_STATE_NEXT, Txid: "6"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: utils.MarshalOrPanic(rangeQueryNext), Txid: "6"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_QUERY_STATE_CLOSE, Txid: "6"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: "6"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "6"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("rangeq"), []byte("A"), []byte("B")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "6"})

	// //wait for done
	// processDone(t, done, false)

	// //error range query

	// //create the response
	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE_BY_RANGE, Txid: "6a"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: payload, Txid: "6a"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "6a"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("rangeq"), []byte("A"), []byte("B")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "6a"})

	// //wait for done
	// processDone(t, done, false)

	// //error range query next

	// //create the response
	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE_BY_RANGE, Txid: "6b"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: rangeQPayload, Txid: "6b"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_QUERY_STATE_NEXT, Txid: "6b"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: "6b"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_QUERY_STATE_CLOSE, Txid: "6b"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: "6b"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "6b"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("rangeq"), []byte("A"), []byte("B")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "6b"})

	// //wait for done
	// processDone(t, done, false)

	// //error range query close

	// //create the response
	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_STATE_BY_RANGE, Txid: "6c"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: rangeQPayload, Txid: "6c"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_QUERY_STATE_NEXT, Txid: "6c"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: "6c"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_QUERY_STATE_CLOSE, Txid: "6c"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Txid: "6c"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "6c"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("rangeq"), []byte("A"), []byte("B")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "6c"})

	// //wait for done
	// processDone(t, done, false)

	// //history query

	// //create the response
	// historyQueryResponse := &pb.QueryResponse{Results: []*pb.QueryResultBytes{
	// 	&pb.QueryResultBytes{ResultBytes: utils.MarshalOrPanic(&lproto.KeyModification{TxId: "6", Value: []byte("100")})}},
	// 	HasMore: true}
	// payload = utils.MarshalOrPanic(historyQueryResponse)

	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_HISTORY_FOR_KEY, Txid: "7"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: payload, Txid: "7"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_QUERY_STATE_NEXT, Txid: "7"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: utils.MarshalOrPanic(rangeQueryNext), Txid: "7"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_QUERY_STATE_CLOSE, Txid: "7"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: "7"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "7"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("historyq"), []byte("A")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "7"})

	// //wait for done
	// processDone(t, done, false)

	// //error history query

	// //create the response
	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_HISTORY_FOR_KEY, Txid: "7a"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: payload, Txid: "7a"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "7a"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("historyq"), []byte("A")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "7a"})

	// //wait for done
	// processDone(t, done, false)

	// //query result

	// //create the response
	// getQRResp := &pb.QueryResponse{Results: []*pb.QueryResultBytes{
	// 	&pb.QueryResultBytes{ResultBytes: utils.MarshalOrPanic(&lproto.KV{"getputcc", "A", []byte("100")})},
	// 	&pb.QueryResultBytes{ResultBytes: utils.MarshalOrPanic(&lproto.KV{"getputcc", "B", []byte("200")})}},
	// 	HasMore: true}
	// getQRRespPayload := utils.MarshalOrPanic(getQRResp)

	// //create the next response
	// rangeQueryNext = &pb.QueryResponse{Results: nil, HasMore: false}

	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_QUERY_RESULT, Txid: "8"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: getQRRespPayload, Txid: "8"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_QUERY_STATE_NEXT, Txid: "8"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Payload: utils.MarshalOrPanic(rangeQueryNext), Txid: "8"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_QUERY_STATE_CLOSE, Txid: "8"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_RESPONSE, Txid: "8"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "8"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("richq"), []byte("A")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "8"})

	// //wait for done
	// processDone(t, done, false)

	// //query result error

	// respSet = &mockpeer.MockResponseSet{errorFunc, errorFunc, []*mockpeer.MockResponse{
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_GET_QUERY_RESULT, Txid: "8a"}, &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_ERROR, Payload: nil, Txid: "8a"}},
	// 	&mockpeer.MockResponse{&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_COMPLETED, Txid: "8a"}, nil}}}
	// peerSide.SetResponses(respSet)

	// ci = &pb.ChaincodeInput{[][]byte{[]byte("richq"), []byte("A")}}
	// payload = utils.MarshalOrPanic(ci)
	// peerSide.Send(&pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: "8a"})

	// //wait for done
	// processDone(t, done, false)

	// time.Sleep(1 * time.Second)
	// peerSide.Quit()
}

func TestRealPeerStream(t *testing.T) {
	viper.Set("peer.address", "127.0.0.1:12345")
	_, err := userChaincodeStreamGetter("fake")
	assert.Error(t, err)
}
