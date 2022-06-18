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
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/proto"
)

const (
	PROBE = iota
)

type p2pProto struct {
	addr net.Addr
	sp   proto.SP
	buf  [10240]byte
}

type P2pProbe struct {
	Number int32
	Buf    []byte
}

type p2pRegistry struct {
	id     uint32
	spFunc proto.SpFunc
}

// version 0
var p2pRegistry_0 = []p2pRegistry{
	{PROBE, p2pProbe},
}

func New(id []byte, addr net.Addr) *p2pProto {
	p := &p2pProto{
		addr: addr,
	}
	sp, err := proto.UdpNew(id, addr, p, p2pCallback)
	if err != nil {
		return nil
	}
	p.sp = sp
	return p
}

func (p *p2pProto) Run() error {
	return p.sp.Run()
}

func (p *p2pProto) RegistryInit() {
	for _, v := range p2pRegistry_0 {
		p.sp.RegisterFunc(0, v.id, v.spFunc)
	}
}

func (p *p2pProto) RegisterPeer(id []byte, addr net.Addr) bool {
	pid := make([]byte, len(id))
	copy(pid, id)
	p.sp.RegisterPeer(pid, addr)
	go p.p2pRun(pid)
	return true
}

func (p *p2pProto) p2pRun(id []byte) {
	var count uint32
	for {
		select {
		//case <-time.After(50 * time.Microsecond):
		case <-time.After(1 * time.Second):
			data := []byte{}
			data = append(data, miscellaneous.E32func(count)...)
			data = append(data, make([]byte, 200*1024)...)
			p.sp.SendInstruction(0, PROBE, data, id)
			count++
		}
	}
}

func p2pCallback(usrData interface{}, id []byte, typ int) int {
	return proto.PROTO_UNACCEPTABLE_PEER
}

func p2pProbe(usrData interface{}, id []byte, data []byte) int {
	n, _ := miscellaneous.D32func(data[:4])
	fmt.Printf("++++++++Recv %s+++++++\n", miscellaneous.Bytes2HexString(id))
	fmt.Printf("++++++++%v++++++++\n", n)
	fmt.Printf("++++++++++++++++++++++++\n")
	return 0
}

func main() {
	id := make([]byte, 16)
	if len(os.Args) < 3 {
		log.Fatalf("Usage: ./test src src_id dst0 id0 dst1 id1 ...")
	}
	ch := make(chan bool)
	src, _ := net.ResolveUDPAddr("udp", os.Args[1])
	copy(id[:], os.Args[2])
	p := New(id[:], src)
	p.RegistryInit()
	for i := 3; i < len(os.Args); i++ {
		dst, _ := net.ResolveUDPAddr("udp", os.Args[i])
		i++
		copy(id[:], os.Args[i])
		p.RegisterPeer(id[:], dst)
	}
	p.Run()
	<-ch
}
