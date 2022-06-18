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
package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	"github.com/tjfoc/tjfoc/core/common/crypto/idea/hashcryption"
	"github.com/tjfoc/tjfoc/core/common/errors"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/common/util/file"
)

type PKCS1v15 struct {
	PrivateKey     *rsa.PrivateKey
	PublicKey      *rsa.PublicKey
	PriKeyFilePath string
	PubKeyFilePath string
}

var logger = flogging.MustGetLogger("crypto.rsaAlgorithm")

const PrFilePath string = "D:\\GOPATH\\src\\github.com/tjfoc/tjfoc\\crypto\\tools\\RSAprivatekey.pem"
const PuFilePath string = "D:\\GOPATH\\src\\github.com/tjfoc/tjfoc\\crypto\\tools\\RSApublickey.pem"

//	=====================================================================
//	function name: GenerateKeys
//	function type: public
//	function receiver: PKCS1v15
//  generate a pair of public and private key, return error
//	=====================================================================
func (pClient *PKCS1v15) GenerateKeys() errors.CallStackError {
	pClient.PriKeyFilePath = PrFilePath
	pClient.PubKeyFilePath = PuFilePath

	var bits = 2048
	var err error
	pClient.PrivateKey, err = rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000010", "Create RSA keys fail")
		logger.Error(e.Message())
		return e
	}
	pClient.PublicKey = &pClient.PrivateKey.PublicKey

	return nil
}

//	=====================================================================
//	function name: Sign
//	function type: public
//	function receiver: PKCS1v15
//	generate signature according to the private key,
//	return byte array or error
//	=====================================================================
func (pClient *PKCS1v15) Sign(src []byte) ([]byte, errors.CallStackError) {
	hash := crypto.SHA256
	hashed := hashcryption.Sha256Hash(src)
	sign, err := rsa.SignPKCS1v15(rand.Reader, pClient.PrivateKey, hash, hashed)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000011", string(src))
		logger.Error(e.Message())
		return nil, e
	}

	return sign, nil
}

//	=====================================================================
//	function name: Verify
//	function type: public
//	function receiver: PKCS1v15
//	verify signature, return bool
//	=====================================================================
func (pClient *PKCS1v15) Verify(src, sign, pubKey []byte) bool {
	hash := crypto.SHA256
	hashed := hashcryption.Sha256Hash(src)
	e := pClient.GetPuKeyByByte(pubKey)
	if e != nil {
		e := errors.ErrorWithCallstack("0006", "000003", pubKey)
		logger.Error(e.Message())
		return false
	}
	err := rsa.VerifyPKCS1v15(pClient.PublicKey, hash, hashed, sign)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000007", string(src))
		logger.Error(e.Message())
		return false
	}

	return true
}

//	=====================================================================
//	function name: WriteKeyFile
//	function type: public
//	function receiver: PKCS1v15
//	write a pair of public and private key into two files, return error
//	=====================================================================
func (pClient *PKCS1v15) WriteKeyFile() errors.CallStackError {
	buf := x509.MarshalPKCS1PrivateKey(pClient.PrivateKey)
	e := file.WriteFile(pClient.PriKeyFilePath, "PRIVATE KEY", buf)
	if e != nil {
		logger.Error(e.Message())
		return e
	}

	puk, err := x509.MarshalPKIXPublicKey(pClient.PublicKey)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000002", puk)
		logger.Error(e.Message())
		return e
	}
	e = file.WriteFile(pClient.PubKeyFilePath, "PUBLIC KEY", puk)
	if e != nil {
		logger.Error(e.Message())
		return e
	}

	return nil
}

//	=====================================================================
//	function name: GetPriKey
//	function type: public
//	function receiver: PKCS1v15
//	get private key from underlying file, return error
//	=====================================================================
func (pClient *PKCS1v15) GetPriKey(pivKeyFilePath string) errors.CallStackError {
	buf, err := ioutil.ReadFile(pivKeyFilePath)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000013", pivKeyFilePath)
		logger.Error(e.Message())
		return e
	}

	p, rest := pem.Decode(buf)
	if p == nil || len(p.Bytes) == 0 {
		pClient.PrivateKey, err = x509.ParsePKCS1PrivateKey(rest)
		if err != nil {
			e := errors.ErrorWithCallstack("0006", "000003", rest)
			logger.Error(e.Message())
			return e
		}
	} else {
		pClient.PrivateKey, err = x509.ParsePKCS1PrivateKey(p.Bytes)
		if err != nil {
			e := errors.ErrorWithCallstack("0006", "000003", p.Bytes)
			logger.Error(e.Message())
			return e
		}
	}

	return nil
}

//	=====================================================================
//	function name: GetPriKeyByByte
//	function type: public
//	function receiver: PKCS1v15
//	byte array decode into the private key struct, return error
//	=====================================================================
func (pClient *PKCS1v15) GetPriKeyByByte(pivKey []byte) errors.CallStackError {
	var err error
	p, rest := pem.Decode(pivKey)
	if p == nil || len(p.Bytes) == 0 {
		pClient.PrivateKey, err = x509.ParsePKCS1PrivateKey(rest)
	} else {
		pClient.PrivateKey, err = x509.ParsePKCS1PrivateKey(p.Bytes)
	}
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000003", pivKey)
		logger.Error(e.Message())
		return e
	}
	return nil
}

//	=====================================================================
//	function name: GetPubKey
//	function type: public
//	function receiver: PKCS1v15
//	get public key from underlying file, return error
//	=====================================================================
func (pClient *PKCS1v15) GetPubKey(PubKeyFilePath string) errors.CallStackError {
	buf, err := ioutil.ReadFile(PubKeyFilePath)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000013", PubKeyFilePath)
		logger.Error(e.Message())
		return e
	}

	p, rest := pem.Decode(buf)
	if p == nil || len(p.Bytes) == 0 {
		pubInterface, err := x509.ParsePKIXPublicKey(rest)
		if err != nil {
			e := errors.ErrorWithCallstack("0006", "000003", rest)
			logger.Error(e.Message())
			return e
		}
		pClient.PublicKey = pubInterface.(*rsa.PublicKey)
	} else {
		pubInterface, err := x509.ParsePKIXPublicKey(p.Bytes)
		if err != nil {
			e := errors.ErrorWithCallstack("0006", "000003", p.Bytes)
			logger.Error(e.Message())
			return e
		}
		pClient.PublicKey = pubInterface.(*rsa.PublicKey)
	}

	return nil
}

//	=====================================================================
//	function name: GetPuKeyByByte
//	function type: public
//	function receiver: PKCS1v15
//	byte array decode into the public key struct, return error
//	=====================================================================
func (pClient *PKCS1v15) GetPuKeyByByte(PubKey []byte) errors.CallStackError {
	p, rest := pem.Decode(PubKey)
	var err errors.CallStackError
	if p == nil || len(p.Bytes) == 0 {
		err = pClient.ParsePub(rest)
	} else {
		err = pClient.ParsePub(p.Bytes)
	}
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000003", PubKey)
		logger.Error(e.Message())
		return e
	}

	return nil
}

//	=====================================================================
//	function name: MessageEncrypt
//	function type: public
//	function receiver: PKCS1v15
//  encrypt message, return byte array or error
//	=====================================================================
func (pClient *PKCS1v15) MessageEncrypt(mes_hash []byte) ([]byte, errors.CallStackError) {
	out, err := rsa.EncryptPKCS1v15(rand.Reader, pClient.PublicKey, mes_hash)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000006", string(mes_hash))
		logger.Error(e.Message())
		return nil, e
	}
	return out, nil
}

//	=====================================================================
//	function name: MessageDecrypt
//	function type: public
//	function receiver: PKCS1v15
//	decrypt message, return byte array or error
//	=====================================================================
func (pClient *PKCS1v15) MessageDecrypt(mes_hash []byte) ([]byte, errors.CallStackError) {
	out, err := rsa.DecryptPKCS1v15(rand.Reader, pClient.PrivateKey, mes_hash)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000004", string(mes_hash))
		logger.Error(e.Message())
		return nil, e
	}

	return out, nil
}

//	=====================================================================
//	function name: NewRSA
//	function type: public
//	instantiate via a pair of public and private key file,
//	return PKCS1v15 struct or error
//	=====================================================================
func NewRSA(priKeyFilePath, publickKeyPath string) (*PKCS1v15, errors.CallStackError) {
	rsa := PKCS1v15{}
	if len(priKeyFilePath) != 0 {
		rsa.PriKeyFilePath = priKeyFilePath
		if e := rsa.GetPriKey(rsa.PriKeyFilePath); e != nil {
			return nil, e
		}
		rsa.PublicKey = &rsa.PrivateKey.PublicKey
	}
	if len(publickKeyPath) != 0 {
		rsa.PubKeyFilePath = publickKeyPath
		if e := rsa.GetPubKey(rsa.PubKeyFilePath); e != nil {
			return nil, e
		}
	}

	return &rsa, nil
}

//	=====================================================================
//	function name: New
//	function type: public
//	instantiate via a pair of public and private key byte array,
//	return PKCS1v15 struct or error
//	=====================================================================
func New(priKey, publickKey []byte) (*PKCS1v15, errors.CallStackError) {
	pClient := PKCS1v15{}
	if len(priKey) != 0 {
		out, err := x509.ParsePKCS1PrivateKey(priKey)
		if err != nil {
			e := errors.ErrorWithCallstack("0006", "000003", priKey)
			logger.Error(e.Message())
			return nil, e
		}
		pClient.PrivateKey = out
		pClient.PublicKey = &pClient.PrivateKey.PublicKey
	}

	if len(publickKey) != 0 {
		err := pClient.ParsePub(publickKey)
		if err != nil {
			logger.Error(err.Message())
			return nil, err
		}
	}

	return &pClient, nil
}

//	=====================================================================
//	function name: GetPublickKey
//	function type: public
//	function receiver: PKCS1v15
//	marshal public key, return byte array or error
//	=====================================================================
func (pClient *PKCS1v15) GetPublickKey() ([]byte, errors.CallStackError) {
	pu := &pClient.PrivateKey.PublicKey
	puk, err := x509.MarshalPKIXPublicKey(pu)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000002", puk)
		logger.Error(e.Message())
		return nil, e
	}

	return puk, nil
}

//	=====================================================================
//	function name: ParsePub
//	function type: public
//	function receiver: PKCS1v15
// 	unmarshal public key, return error
//	=====================================================================
func (pClient *PKCS1v15) ParsePub(pub []byte) errors.CallStackError {
	pubk, err := x509.ParsePKIXPublicKey(pub)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000003", pub)
		logger.Error(e.Message())
		return e
	}

	switch pub := pubk.(type) {
	case *rsa.PublicKey:
		pClient.PublicKey = pub
	}
	return nil
}
