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
package symcryption

import (
	"crypto/aes"
	"crypto/cipher"
	"math/rand"
	"time"

	"github.com/tjfoc/tjfoc/core/common/errors"
	"github.com/tjfoc/tjfoc/core/common/flogging"
)

var logger = flogging.MustGetLogger("symcryption.aes")

type AesEncrypt struct {
}

//	=====================================================================
//	function name: Genkeys
//	function type: public
//	function receiver: AesEncrypt
//  generate a key, return byte array
//	=====================================================================
func (a *AesEncrypt) Genkeys() []byte {
	return randFieldElement()
}

//	=====================================================================
//	function name: randFieldElement
//	function type: public
//	function receiver: AesEncrypt
//  generate 32 bytes rand array, return byte array
//	=====================================================================
func randFieldElement() []byte {
	var a int
	k := make([]byte, 32)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i <= 34; i++ {
		b := int(r.Intn(9))
		a = a + b
		a <<= 1
		if i >= 3 {
			k[i-3] = byte(a)
		}

	}

	return k
}

//	=====================================================================
//	function name: Encrypt
//	function type: public
//	function receiver: AesEncrypt
//  encrypt message, return byte array or error
//	=====================================================================
func (a *AesEncrypt) Encrypt(mes, key []byte) ([]byte, errors.CallStackError) {
	var iv = []byte(key)[:aes.BlockSize]
	encrypted := make([]byte, len(mes))
	aesBlockEncrypter, err := aes.NewCipher(key)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000006", string(mes))
		logger.Error(e.Message())
		return nil, e
	}
	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	aesEncrypter.XORKeyStream(encrypted, []byte(mes))

	return encrypted, nil
}

//	=====================================================================
//	function name: Decrypt
//	function type: public
//	function receiver: AesEncrypt
//  decrypt message, return byte array or error
//	=====================================================================
func (a *AesEncrypt) Decrypt(mes, key []byte) ([]byte, errors.CallStackError) {
	var iv = []byte(key)[:aes.BlockSize]
	decrypted := make([]byte, len(mes))
	aesBlockDecrypter, e := aes.NewCipher([]byte(key))
	if e != nil {
		err := errors.ErrorWithCallstack("0006", "000004", string(mes))
		logger.Error(err.Message())
		return nil, err
	}
	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	aesDecrypter.XORKeyStream(decrypted, mes)

	return decrypted, nil
}

//	=====================================================================
//	function name: New
//	function type: public
//	function receiver: AesEncrypt
//  instantiate AesEncrypt
//	=====================================================================
func New() *AesEncrypt {
	return &AesEncrypt{}
}
