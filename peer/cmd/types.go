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
package cmd

import (
	"math/big"
	"time"

	"github.com/tjfoc/tjfoc/core/blockchain"
	"github.com/tjfoc/tjfoc/core/consensus"
	"github.com/tjfoc/tjfoc/core/crypt"
	"github.com/tjfoc/tjfoc/core/store/chain"
	"github.com/tjfoc/tjfoc/proto/p2p"
	"google.golang.org/grpc"
)

type peer struct {
	conNum       uint32
	t            time.Time
	peerid       []byte
	cryptPlug    crypt.Crypto
	grpcServer   *grpc.Server
	p2pServer    p2p.P2pServer
	grpcOptions  []grpc.DialOption
	consensusAPI consensus.Consensus
	blockChain   blockchain.BlockChain
	memberList   map[string]*chain.PeerInfo
}

type AdminSignature struct {
	R, S *big.Int
}
