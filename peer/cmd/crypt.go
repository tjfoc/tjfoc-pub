// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"hash"
	"io/ioutil"
	"strings"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm3"
	"github.com/tjfoc/tjfoc/core/crypt"
)

const (
	SM3 crypto.Hash = crypto.BLAKE2b_256 // 将BLAKE2b_256算法替换成SM3
)

func init() {
	crypto.RegisterHash(SM3, func() hash.Hash {
		return sm3.New()
	})
}

func readPrivateKeyFromMem(data []byte) (*ecdsa.PrivateKey, error) {
	var block *pem.Block

	if block, _ = pem.Decode(data); block == nil {
		return nil, errors.New("failed to decode private key")
	}
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	privKey, _ := priv.(*ecdsa.PrivateKey)
	return privKey, err
}

func readPrivateKeyFromPem(fileName string) (*ecdsa.PrivateKey, error) {
	if data, err := ioutil.ReadFile(fileName); err != nil {
		return nil, err
	} else {
		return readPrivateKeyFromMem(data)
	}
}

func newCryptPlug() (crypt.Crypto, error) {
	var h hash.Hash
	var v crypto.Hash
	var s crypto.Signer

	switch {
	case strings.Compare(Config.Crypt.KeyTyp, "sm2") == 0:
		s, _ = sm2.ReadPrivateKeyFromPem(Config.Crypt.KeyPath, nil)
	case strings.Compare(Config.Crypt.KeyTyp, "ecc") == 0:
		s, _ = readPrivateKeyFromPem(Config.Crypt.KeyPath)
	default:
		return nil, errors.New("newCryptPlug: unsupport sign algorithm")
	}
	switch {
	case strings.Compare(Config.Crypt.HashTyp, "sm3") == 0:
		v = SM3
		h = sm3.New()
	case strings.Compare(Config.Crypt.HashTyp, "sha256") == 0:
		h = sha256.New()
		v = crypto.SHA256
	default:
		return nil, errors.New("newCryptPlug: unsupport hash algorithm")
	}
	return crypt.New(h, s, v), nil
}
