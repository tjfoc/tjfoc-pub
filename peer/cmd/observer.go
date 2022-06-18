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
	"crypto/md5"
	"net"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/spf13/cobra"
	"github.com/tjfoc/tjfoc/core/blockchain"
	"github.com/tjfoc/tjfoc/core/chaincode"
	"github.com/tjfoc/tjfoc/core/consensus/raft"
	"github.com/tjfoc/tjfoc/core/health"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/store/chain"
	"github.com/tjfoc/tjfoc/core/worldstate"
	"github.com/tjfoc/tjfoc/proto/p2p"
)

// observerCmd represents the observer command
var observerCmd = &cobra.Command{
	Use:   "observer",
	Short: "Peer join as observer node",
	Run: func(cmd *cobra.Command, args []string) {
		observer()
	},
}

func init() {
	RootCmd.AddCommand(observerCmd)
}

// var logger = flogging.MustGetLogger("observer")

func observer() {
	logInit()
	p := new(peer)
	p.peerid = []byte(Config.Self.Id)
	p.memberList = make(map[string]*chain.PeerInfo)
	if c, err := newCryptPlug(); err != nil {
		logger.Fatal(err)
	} else {
		p.cryptPlug = c
	}
	if chn, err := chain.New(p.cryptPlug, Config.StorePath.Path); err != nil {
		logger.Fatal(err)
	} else {

		p.blockChain = blockchain.New(chn, p.cryptPlug,
			blockchain.GenesisBlock(0, []byte("test")), p, tCallBack)
	}
	if err := p.setMemberList("observer"); err != nil {
		logger.Fatal(err)
	}
	sp, err := p.memberListInit()
	if err != nil {
		logger.Fatal(err)
	}

	health.GetInstance()
	worldstate.New(uint64(17), "worldstate.db", sp)
	chaincode.GetDelieverInstance()
	chaincode.GetResultInstance().SetAttribute(Config.Self.Id, p.cryptPlug)

	p.consensusAPI = raft.NewRaft(raft.Observer, p, p.blockChain, sp, packCallback, Config)
	p.p2pServer = p2p.New(sp, time.Duration(Config.P2P.Cycle)*time.Millisecond, p.consensusAPI, p.blockChain)

	for k, v := range p.memberList {
		id, _ := miscellaneous.GenHash(md5.New(), []byte(k))
		if addr, err := net.ResolveTCPAddr("tcp", v.Addr); err != nil {
			logger.Fatal(err)
		} else {
			p.p2pServer.RegisterPeer(id, addr)
			logger.Infof("p2p registerPeer [name:%s, addr:%s]", k, v.Addr)
		}
	}

	//CA
	//ls
	//	err := ca.CertLocalVerify()
	p.grpcServer, p.grpcOptions = newGrpcServer()

	go p.serverRun()
	for {
		time.Sleep(10 * time.Second)
		runtime.GC()
		debug.FreeOSMemory()
	}
}
