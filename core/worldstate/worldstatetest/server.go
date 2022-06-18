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
package main

import (
	"crypto/md5"
	//"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/worldstate"
	"github.com/tjfoc/tjfoc/proto"
)

type p2pProto struct {
	addr net.Addr
	sp   proto.SP
}

func New(id []byte, addr net.Addr) *p2pProto {
	p := &p2pProto{
		addr: addr,
	}
	sp, err := proto.TcpNew(id, addr, p, p2pCallback)
	if err != nil {
		return nil
	}
	p.sp = sp
	return p
}

func (p *p2pProto) Run() error {
	return p.sp.Run()
}

func (p *p2pProto) RegisterPeer(id []byte, addr net.Addr) bool {
	pid := make([]byte, len(id))
	copy(pid, id)
	ids = append(ids, string(pid))
	p.sp.RegisterPeer(pid, addr)
	go p.p2pRun(pid)
	return true
}

func (p *p2pProto) p2pRun(id []byte) {
	time.Sleep(10 * time.Second)
}

func p2pCallback(usrData interface{}, id []byte, typ int) int {
	return proto.PROTO_UNACCEPTABLE_PEER
}

var ids []string

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: ./test src id dst0 id0 dst1 id1 ...")
	}
	ch := make(chan bool)
	src, _ := net.ResolveTCPAddr("tcp", os.Args[1])
	id, _ := miscellaneous.GenHash(md5.New(), []byte(os.Args[2]))
	p := New(id, src)
	ids = make([]string, 0)
	for i := 3; i < len(os.Args); i++ {
		dst, _ := net.ResolveTCPAddr("tcp", os.Args[i])
		i++
		id, _ := miscellaneous.GenHash(md5.New(), []byte(os.Args[i]))
		p.RegisterPeer(id, dst)
	}
	worldstate.New(uint64(17), "db1", p.sp)
	/*
		temp_hash := worldstate.GetWorldState().GetRootHash()
		kv := make(worldstate.WorldStateData, 0)
		kv["k1"] = "aaa"
		kv["k2"] = "bbb"
		kv["k3"] = "ccc"
		kv["k4"] = "ddd"
		kv["k5"] = "eee"
		kv["k6"] = "fff"
		kv["k7"] = "ggg"
		kv["k8"] = "hhh"
		kv["k9"] = "iii"
		kv["k10"] = "jjj"
		kv["k11"] = "kkk"
		kv["k12"] = "lll"
		kv["k13"] = "mmm"
		kv["k14"] = "nnn"
		kv["k15"] = "ooo"
		kv["k16"] = "ppp"
		kv["k17"] = "qqq"
		kv["k18"] = "rrr"
		kv["k19"] = "sss"
		kv["k20"] = "ttt"
		kv["k21"] = "uuu"
		kv["k22"] = "vvv"
		kv["k23"] = "www"
		kv["k24"] = "xxx"
		kv["k25"] = "yyy"
		kv["k26"] = "zzz"
		worldstate.GetWorldState().PushDB(uint64(0), temp_hash, kv)
	*/
	/*
		a := worldstate.GetWorldState().Search("k1")
		fmt.Println(a)
		temp := make([]string, 0)
		temp = append(temp, "k1")
		temp = append(temp, "k2")
		temp = append(temp, "k3")
		b := worldstate.GetWorldState().Searchn(temp)
		fmt.Println(b)
	*/
	p.Run()
	<-ch
}
