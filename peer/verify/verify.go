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
package verify

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"math/big"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/peer/cmd"
)

var logger = flogging.MustGetLogger("sign verify")

type Signature struct {
	R, S *big.Int
}

func Verify(x, y *big.Int, msg, signature []byte) bool {

	var sig Signature
	_, err := asn1.Unmarshal(signature, &sig)
	if err != nil {
		logger.Error(err)
		return false
	}
	switch cmd.Config.Crypt.KeyTyp {
	case "sm2":
		return sm2.Verify(&sm2.PublicKey{sm2.P256Sm2(), x, y}, msg, sig.R, sig.S)
	case "ecdsa":
		return ecdsa.Verify(&ecdsa.PublicKey{elliptic.P256(), x, y}, msg, sig.R, sig.S)
	default:
		return false
	}

	return false
}
