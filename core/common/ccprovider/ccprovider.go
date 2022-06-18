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

package ccprovider

import (
	"fmt"

	"github.com/tjfoc/tjfoc/core/common/flogging"
)

var ccproviderLogger = flogging.MustGetLogger("ccprovider")

type CCContext struct {
	ChainID string

	Name string

	Version string

	TxID string

	PeerIp string

	PeerPort string

	Syscc bool

	canonicalName string
}

func NewCCContext(cid, name, version, txid, peerid, peerport string, syscc bool) *CCContext {
	if version == "" {
		panic(fmt.Sprintf("---empty version---(chain=%s,chaincode=%s,version=%s,txid=%s,syscc=%t", cid, name, version, txid, syscc))
	}

	canName := name + "_" + version + "_" + peerid + "_" + peerport

	cccid := &CCContext{cid, name, version, txid, peerid, peerport, syscc, canName}

	return cccid
}

func (cccid *CCContext) GetCanonicalName() string {
	if cccid.canonicalName == "" {
		panic(fmt.Sprintf("cccid not constructed using NewCCContext(chain=%s,chaincode=%s,version=%s,txid=%s,syscc=%t)", cccid.ChainID, cccid.Name, cccid.Version, cccid.TxID, cccid.Syscc))
	}

	return cccid.canonicalName
}
