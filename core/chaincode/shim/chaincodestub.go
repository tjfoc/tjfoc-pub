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
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"hash"
	"math/big"
	"os"

	"github.com/tjfoc/gmsm/sm2"
	pb "github.com/tjfoc/tjfoc/protos/chaincode"
)

type ChaincodeStub struct {
	TxID           string
	chaincodeEvent *pb.ChaincodeEvent
	args           [][]byte
	handler        *Handler
	// signedProposal *pb.SignedProposal
	// proposal       *pb.Proposal

	// Additional fields extracted from the signedProposal
	creator   []byte
	transient map[string][]byte
	binding   []byte
}

// -- init stub ---
// ChaincodeInvocation functionality

func (stub *ChaincodeStub) init(handler *Handler, txid string, input *pb.ChaincodeInput) error {
	stub.TxID = txid
	stub.args = input.Args
	stub.handler = handler
	return nil
}

// GetTxID returns the transaction ID
func (stub *ChaincodeStub) GetTxID() string {
	return stub.TxID
}

// --------- Security functions ----------
//CHAINCODE SEC INTERFACE FUNCS TOBE IMPLEMENTED BY ANGELO

// ------------- Call Chaincode functions ---------------

// --------- State functions ----------

// GetArgs documentation can be found in interfaces.go
func (stub *ChaincodeStub) GetArgs() [][]byte {
	return stub.args
}

// GetStringArgs documentation can be found in interfaces.go
func (stub *ChaincodeStub) GetStringArgs() []string {
	args := stub.GetArgs()
	strargs := make([]string, 0, len(args))
	for _, barg := range args {
		strargs = append(strargs, string(barg))
	}
	return strargs
}

// GetFunctionAndParameters documentation can be found in interfaces.go
func (stub *ChaincodeStub) GetFunctionAndParameters() (function string, params []string) {
	allargs := stub.GetStringArgs()
	function = ""
	params = []string{}
	if len(allargs) >= 1 {
		function = allargs[0]
		params = allargs[1:]
	}
	return
}
func dealKey(key string) string {
	prefix := os.Args[3] + "-"
	return prefix + key
}

// GetState documentation can be found in interfaces.go
func (stub *ChaincodeStub) GetState(key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("key must not be an empty string")
	}
	key = dealKey(key)
	return stub.handler.handleGetState(key, stub.TxID)
}

// DelState documentation can be found in interfaces.go
func (stub *ChaincodeStub) DelState(key string) error {
	if key == "" {
		return fmt.Errorf("key must not be an empty string")
	}
	key = dealKey(key)
	return stub.handler.handleDelState(key, stub.TxID)
}

// GetStaten documentation can be found in interfaces.go
func (stub *ChaincodeStub) GetStaten(keyn []string) (map[string]string, error) {
	a := make([]string, 0)
	for _, v := range keyn {
		if v != "" {
			a = append(a, v)
		}
	}
	if len(a) == 0 {
		return nil, fmt.Errorf("key must not be empty string")
	}
	keyn = a
	for i, _ := range keyn {
		keyn[i] = dealKey(keyn[i])
	}
	return stub.handler.handleGetStaten(keyn, stub.TxID)
}

// DelStaten documentation can be found in interfaces.go
func (stub *ChaincodeStub) DelStaten(keyn []string) error {
	a := make([]string, 0)
	for _, v := range keyn {
		if v != "" {
			a = append(a, v)
		}
	}
	if len(a) == 0 {
		return fmt.Errorf("key must not be empty string")
	}
	keyn = a
	for i, _ := range keyn {
		keyn[i] = dealKey(keyn[i])
	}
	return stub.handler.handleDelStaten(keyn, stub.TxID)
}

// PutState documentation can be found in interfaces.go
func (stub *ChaincodeStub) PutState(key string, value []byte) error {
	if key == "" {
		return fmt.Errorf("key must not be an empty string")
	}
	key = dealKey(key)
	return stub.handler.handlePutState(key, value, stub.TxID)
}

// InvokeChaincode documentation can be found in interfaces.go
func (stub *ChaincodeStub) InvokeChaincode(chaincodeName string, chaincodeVersion string, args [][]byte) pb.Response {
	return stub.handler.handleInvokeChaincode(chaincodeName, chaincodeVersion, args, stub.TxID)
}

func (stub *ChaincodeStub) GetStateByPrefix(key string) (map[string]string, error) {
	key = dealKey(key)
	return stub.handler.handleGetStateByPrefix(key, stub.TxID)
}

// CommonIterator documentation can be found in interfaces.go
type CommonIterator struct {
	handler *Handler
	uuid    string
	// response   *pb.QueryResponse
	currentLoc int
}

// StateQueryIterator documentation can be found in interfaces.go
type StateQueryIterator struct {
	*CommonIterator
}

// HistoryQueryIterator documentation can be found in interfaces.go
type HistoryQueryIterator struct {
	*CommonIterator
}

type resultType uint8

const (
	// SMDSA 国密加密
	SMDSA = iota
	// ECDSA EC加密
	ECDSA
)

// VerifySign 验证签名,公钥 数据 签名  哈希 加密方法
func VerifySign(pubKey, data, sign []byte, hash hash.Hash, dsa int) bool {
	hash.Write(data)
	data = hash.Sum(nil)
	sign, err := base64.StdEncoding.DecodeString(string(sign))
	if err != nil {
		return false
	}
	switch dsa {
	case SMDSA:
		return verifySM(pubKey, data, sign)
	case ECDSA:
		return verifySignEcdsa(pubKey, data, sign)
	}
	return false
}
func verifySM(pub, data, sign []byte) bool {
	pubKey, err := sm2.ReadPublicKeyFromMem(pub, nil)
	if err != nil {
		return false
	}
	return pubKey.Verify(data, sign)
}
func verifySignEcdsa(pub, data, sign []byte) bool {
	var ecdsasign struct {
		R, S *big.Int
	}
	_, err := asn1.Unmarshal(sign, &ecdsasign)
	if err != nil {
		return false
	}

	block, _ := pem.Decode(pub)
	pubkey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}
	ecdsaPub := pubkey.(*ecdsa.PublicKey)

	return ecdsa.Verify(ecdsaPub, data, ecdsasign.R, ecdsasign.S)
}
