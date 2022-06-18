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
	"fmt"
	"net"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/spf13/cobra"
	"github.com/tjfoc/tjfoc/core/blockchain"
	"github.com/tjfoc/tjfoc/core/chaincode"
	"github.com/tjfoc/tjfoc/core/consensus/raft"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/store/chain"
	"github.com/tjfoc/tjfoc/core/worldstate"
	"github.com/tjfoc/tjfoc/proto/p2p"
)

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "节点加入",
	Run: func(cmd *cobra.Command, args []string) {
		join()
	},
}

func init() {
	RootCmd.AddCommand(joinCmd)
}

// var logger = flogging.MustGetLogger("join")

func join() {
	logInit()
	p := new(peer)
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
	if err := p.setMemberList("join"); err != nil {
		logger.Fatal(err)
	}
	sp, err := p.memberListInit()
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println("new worldstate")
	worldstate.New(uint64(17), Config.StorePath.WorldStatePath, sp)
	chaincode.GetInstance()

	p.consensusAPI = raft.NewRaft(raft.Join, p, p.blockChain, sp, packCallback, Config)
	p.p2pServer = p2p.New(sp, time.Duration(Config.Members.P2P.Cycle)*time.Millisecond, p.consensusAPI, p.blockChain)

	for k, v := range p.memberList {
		id, _ := miscellaneous.GenHash(md5.New(), []byte(k))
		if addr, err := net.ResolveTCPAddr("tcp", v.Addr); err != nil {
			logger.Fatal(err)
		} else {
			p.p2pServer.RegisterPeer(id, addr)
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
