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
package proto

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/tjfoc/tjfoc/core/miscellaneous"
)

/*
Package:
    type        2 byte {ack, data, query, probe}
	data        0 ~ 1048 byte

Data Package: headerSize = 12 type        2 byte
	index       4 byte
	time        4 byte
	size        2 byte
	data        0 ~ 1024 byte

Ack Package: headerSize = 6
   	type        2 byte
   	index       4 byte

Query Package: headerSize = 6
   	type        2 byte
   	index       4 byte

Reset package: headerSize = 4
	type 		2 byte

Probe Package:
   	type        2 byte
   	count       2 count
		pkg0 index  4 byte
   		pkg0 time   4 byte
		pkg1 index  4 byte
		pkg1 time   4 byte
		...
*/

const (
	SP_UDP_DATA_PACKAGE_HEADER_SIZE = 12
	SP_UDP_PROBE_PACKAGES_LIMIT     = 128 // 128 * 8 = 1024
	SP_UDP_DATA_PACKAGE_BODY_SIZE   = 1024
	SP_UDP_PACKAGE_POOL_SIZE        = 4096
	SP_UDP_PROBE_CYCLE              = 1000 * time.Millisecond
	SP_UDP_SLEEP_CYCLE              = 1000 * time.Millisecond
)

const (
	SP_UDP_RESERVED = iota
	SP_UDP_ACK
	SP_UDP_DATA
	SP_UDP_QUERY
	SP_UDP_PROBE
	SP_UDP_RESET
)

type udpSpPeerInfo struct {
	sendCurr   uint32
	recvCurr   uint32
	sendTime   uint32 // 最新的发送时间戳
	recvTime   uint32 // 最新的接受时间戳
	peerId     []byte
	startTime  time.Time
	pkgChannel chan []byte
	address    *net.UDPAddr
	sendLock   sync.RWMutex      // sendCurr, sendTime, sendPool
	recvLock   sync.RWMutex      // recvCurr, recvTime, recvPool
	sendPool   map[uint32][]byte // key = index, value = pkg
	recvPool   map[uint32][]byte // key = index, value = pkg
	spChannel  chan *spInstruction
}

type UdpSp struct {
	sp               *sp
	reset            bool
	peerId           []byte
	address          *net.UDPAddr
	connection       *net.UDPConn
	memberLock       sync.RWMutex
	memberListById   map[string]*udpSpPeerInfo
	memberListByAddr map[string]*udpSpPeerInfo
}

var spUdpAck []byte
var spUdpData []byte
var spUdpQuery []byte
var spUdpProbe []byte
var spUdpReset []byte

func init() {
	spUdpAck = miscellaneous.E16func(SP_UDP_ACK)
	spUdpData = miscellaneous.E16func(SP_UDP_DATA)
	spUdpQuery = miscellaneous.E16func(SP_UDP_QUERY)
	spUdpProbe = miscellaneous.E16func(SP_UDP_PROBE)
	spUdpReset = miscellaneous.E16func(SP_UDP_RESET)
}

func UdpNew(id []byte, addr net.Addr, usrData interface{}, callback SpCallback) (*UdpSp, error) {
	udpAddr, err := net.ResolveUDPAddr(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}
	con, err := net.ListenUDP(udpAddr.Network(), udpAddr)
	if err != nil {
		return nil, err
	}
	return &UdpSp{
		connection:       con,
		reset:            true,
		address:          udpAddr,
		peerId:           miscellaneous.Dup(id),
		sp:               spNew(usrData, callback),
		memberListById:   make(map[string]*udpSpPeerInfo),
		memberListByAddr: make(map[string]*udpSpPeerInfo),
	}, nil
}

func (u *UdpSp) GetMemberList() map[string]string {
	b := make(map[string]string)
	u.memberLock.Lock()
	for _, v := range u.memberListById {
		b[string(v.peerId)] = v.address.String()
	}
	u.memberLock.Unlock()
	return b
}

func (u *UdpSp) RegisterPeer(id []byte, addr net.Addr) error {
	if strings.Compare(addr.Network(), u.address.Network()) != 0 {
		return errors.New(fmt.Sprintf("RegisterPeer: only support: %s", u.address.Network()))
	}
	pid := miscellaneous.Dup(id)
	udpAddr, err := net.ResolveUDPAddr(addr.Network(), addr.String())
	if err != nil {
		return err
	}
	pi := &udpSpPeerInfo{
		peerId:     pid,
		address:    udpAddr,
		startTime:  time.Now(),
		pkgChannel: make(chan []byte),
		sendPool:   make(map[uint32][]byte),
		recvPool:   make(map[uint32][]byte),
		spChannel:  make(chan *spInstruction),
	}
	u.memberLock.Lock()
	u.memberListByAddr[addr.String()] = pi
	u.memberListById[miscellaneous.Bytes2HexString(pid)] = pi
	u.memberLock.Unlock()
	go u.sp.protoParse(pi.spChannel)
	go u.routineProcess(pi)
	return nil
}

func (u *UdpSp) UnregisterPeer(id []byte) error {
	key := miscellaneous.Bytes2HexString(id)
	u.memberLock.Lock()
	pi, ok := u.memberListById[key]
	if ok {
		delete(u.memberListByAddr, pi.address.String())
		pi.pkgChannel <- []byte{} // exit
		delete(u.memberListById, key)
	}
	u.memberLock.Unlock()
	return nil
}

func (u *UdpSp) UnregisterFunc(version uint32, class uint32) error {
	return u.sp.UnregisterFunc(version, class)
}

func (u *UdpSp) RegisterFunc(version uint32, class uint32, callback SpFunc) error {
	return u.sp.RegisterFunc(version, class, callback)
}

func (u *UdpSp) SendInstruction(version uint32, class uint32, data []byte, id []byte) error {
	u.memberLock.Lock()
	pi, ok := u.memberListById[miscellaneous.Bytes2HexString(id)]
	u.memberLock.Unlock()
	if !ok {
		return errors.New(fmt.Sprintf("SendInstruction: unregister peer %s", miscellaneous.Bytes2HexString(id)))
	}
	for {
		pi.sendLock.RLock()
		poolSize := len(pi.sendPool)
		pi.sendLock.RUnlock()
		if poolSize > SP_UDP_PACKAGE_POOL_SIZE {
			return nil
			time.Sleep(SP_UDP_SLEEP_CYCLE)
		} else {
			break
		}
	}
	u.sendSpPackages(&spProto{
		class:   class,
		version: version,
		size:    uint32(len(data)),
	}, pi, data)
	return nil
}

func (u *UdpSp) Run() error {
	buf := make([]byte, 1048)
	for {
		n, addr, err := u.connection.ReadFromUDP(buf)
		if err != nil || n == 0 {
			return err
		}
		u.memberLock.RLock()
		pi, ok := u.memberListByAddr[addr.String()]
		u.memberLock.RUnlock()
		if !ok {
			switch u.sp.callback(u.sp.usrData, []byte(addr.String()), PROTO_ADDR) {
			case PROTO_ACCEPTABLE_PEER:
				u.memberLock.RLock()
				pi, ok = u.memberListByAddr[addr.String()]
				u.memberLock.RUnlock()
			case PROTO_UNACCEPTABLE_PEER:
			}
		}
		if ok {
			pi.pkgChannel <- miscellaneous.Dup(buf[:n])
		}
	}
}

func (u *UdpSp) probeRun(pi *udpSpPeerInfo, ch chan bool) {
	sendPool := make(map[uint32][]byte)
	for {
		select {
		case <-ch:
			return
		case <-time.After(SP_UDP_PROBE_CYCLE):
			for k, _ := range sendPool {
				delete(sendPool, k)
			}
			pi.sendLock.RLock()
			sendCurr := pi.sendCurr
			for k, v := range pi.sendPool {
				sendPool[k] = v
			}
			pi.sendLock.RUnlock()
			v, ok := sendPool[sendCurr-1]
			if ok {
				u.sendData(v, pi.address)
			}
			currTime := (time.Now().UnixNano() - pi.startTime.UnixNano()) / 1000000
			if len(sendPool) >= SP_UDP_PACKAGE_POOL_SIZE/4 {
				buf := []byte{}
				count := uint16(0)
				pkgBuf := []byte{}
				buf = append(buf, spUdpProbe...)
				for _, v := range sendPool {
					t, _ := miscellaneous.D32func(v[6:10])
					if uint32(currTime)-t > 100 {
						count++
						pkgBuf = append(pkgBuf, v[2:10]...)
						if count >= SP_UDP_PROBE_PACKAGES_LIMIT {
							break
						}
					}
				}
				if count > 0 {
					buf = append(buf, miscellaneous.E16func(count)...)
					buf = append(buf, pkgBuf...)
					u.sendData(buf, pi.address)
				}
			}
		}
	}
}

func (u *UdpSp) routineProcess(pi *udpSpPeerInfo) {
	ch := make(chan bool)
	go u.probeRun(pi, ch)
	for {
		pkg := <-pi.pkgChannel
		if len(pkg) == 0 {
			ch <- true
			pi.spChannel <- &spInstruction{
				len: -1,
			}
			return
		}
		switch {
		case bytes.Compare(pkg[:2], spUdpAck) == 0:
			u.ackProcess(pkg, pi)
		case bytes.Compare(pkg[:2], spUdpData) == 0:
			u.spProtoProcess(pkg, pi)
		case bytes.Compare(pkg[:2], spUdpQuery) == 0:
			u.queryProcess(pkg, pi)
		case bytes.Compare(pkg[:2], spUdpProbe) == 0:
			u.probeProcess(pkg, pi)
		case bytes.Compare(pkg[:2], spUdpReset) == 0:
			u.resetProcess(pkg, pi)
		}
	}
}

func (u *UdpSp) ackProcess(pkg []byte, pi *udpSpPeerInfo) {
	idx, _ := miscellaneous.D32func(pkg[2:6])
	pi.sendLock.Lock()
	_, ok := pi.sendPool[idx]
	if ok {
		delete(pi.sendPool, idx)
	}
	pi.sendLock.Unlock()
}

func (u *UdpSp) queryProcess(pkg []byte, pi *udpSpPeerInfo) {
	idx, _ := miscellaneous.D32func(pkg[2:6])
	pi.sendLock.RLock()
	v, ok := pi.sendPool[idx]
	pi.sendLock.RUnlock()
	if ok {
		u.reset = false
		u.sendData(v, pi.address)
	} else {
		if pi.sendCurr+1 < idx && u.reset {
			u.sendPackage(SP_UDP_RESET, []byte{}, pi.address)
		}
	}
}

func (u *UdpSp) resetProcess(pkg []byte, pi *udpSpPeerInfo) {
	pi.recvLock.Lock()
	pi.recvCurr = 0
	for k, _ := range pi.recvPool {
		delete(pi.recvPool, k)
	}
	pi.recvLock.Unlock()
	pi.sendLock.Lock()
	pi.sendCurr = 0
	for k, _ := range pi.sendPool { //reset
		delete(pi.sendPool, k)
	}
	pi.startTime = time.Now()
	pi.sendLock.Unlock()
	pi.spChannel <- &spInstruction{
		len: 0,
	}
}

func (u *UdpSp) probeProcess(pkg []byte, pi *udpSpPeerInfo) {
	count, _ := miscellaneous.D16func(pkg[2:4])
	pi.recvLock.RLock()
	recvTime := pi.recvTime
	pi.recvLock.RUnlock()
	for i := 0; i < int(count); i++ {
		t, _ := miscellaneous.D32func(pkg[8+8*i : 12+8*i])
		if t < recvTime {
			u.sendPackage(SP_UDP_ACK, pkg[4+8*i:8+8*i], pi.address)
		}
	}
}

func (u *UdpSp) spProtoProcess(pkg []byte, pi *udpSpPeerInfo) {
	idx, _ := miscellaneous.D32func(pkg[2:6])
	pi.recvLock.Lock()
	recvCurr := pi.recvCurr
	pi.recvLock.Unlock()
	if idx == recvCurr {
		u.sendPackage(SP_UDP_ACK, miscellaneous.E32func(idx), pi.address)
		pi.recvLock.Lock()
		pi.recvCurr++
		pi.recvTime, _ = miscellaneous.D32func(pkg[6:10])
		pi.recvLock.Unlock()
		pi.spChannel <- &spInstruction{
			pid:  pi.peerId,
			data: pkg[SP_UDP_DATA_PACKAGE_HEADER_SIZE:],
			len:  len(pkg) - SP_UDP_DATA_PACKAGE_HEADER_SIZE,
		}
		for {
			pi.recvLock.RLock()
			v, ok := pi.recvPool[pi.recvCurr]
			pi.recvLock.RUnlock()
			if !ok {
				return
			}
			u.sendPackage(SP_UDP_ACK, miscellaneous.E32func(pi.recvCurr), pi.address)
			pi.spChannel <- &spInstruction{
				pid:  pi.peerId,
				data: v[SP_UDP_DATA_PACKAGE_HEADER_SIZE:],
				len:  len(v) - SP_UDP_DATA_PACKAGE_HEADER_SIZE,
			}
			pi.recvLock.Lock()
			delete(pi.recvPool, pi.recvCurr)
			pi.recvCurr++
			pi.recvTime, _ = miscellaneous.D32func(v[6:10])
			pi.recvLock.Unlock()
		}
	} else {
		pi.recvLock.Lock()
		if len(pi.recvPool) < SP_UDP_PACKAGE_POOL_SIZE {
			pi.recvPool[idx] = pkg
		}
		pi.recvLock.Unlock()
		u.sendPackage(SP_UDP_QUERY, miscellaneous.E32func(pi.recvCurr), pi.address)
	}
}

func (u *UdpSp) sendData(data []byte, addr *net.UDPAddr) {
	for { // udp数据包写入超时是很少见的
		_, err := u.connection.WriteToUDP(data, addr)
		if err == nil {
			return
		}
	}
}

func (u *UdpSp) sendPackage(typ uint32, data []byte, addr *net.UDPAddr) []byte {
	buf := []byte{}
	switch typ {
	case SP_UDP_ACK:
		buf = append(buf, spUdpAck...)
	case SP_UDP_DATA:
		buf = append(buf, spUdpData...)
	case SP_UDP_QUERY:
		buf = append(buf, spUdpQuery...)
	case SP_UDP_PROBE:
		buf = append(buf, spUdpProbe...)
	case SP_UDP_RESET:
		buf = append(buf, spUdpReset...)
	}
	buf = append(buf, data...)
	u.sendData(buf, addr)
	return buf
}

// 发送udp报文
func (u *UdpSp) sendSpUdpPackage(data []byte, pi *udpSpPeerInfo) []byte {
	buf := []byte{}
	buf = append(buf, miscellaneous.E32func(pi.sendCurr)...) // index
	pi.sendTime = uint32((time.Now().UnixNano() - pi.startTime.UnixNano()) / 1000000)
	buf = append(buf, miscellaneous.E32func(pi.sendTime)...)       // time
	buf = append(buf, miscellaneous.E16func(uint16(len(data)))...) // size
	buf = append(buf, data...)                                     // data
	pi.sendCurr++
	return u.sendPackage(SP_UDP_DATA, buf, pi.address)
}

// 发送SP协议报文
func (u *UdpSp) sendSpPackage(sp *spProto, pi *udpSpPeerInfo, data []byte) {
	pi.sendLock.Lock()
	defer pi.sendLock.Unlock()
	pkg := u.sendSpUdpPackage(data, pi)
	pkgIndex, _ := miscellaneous.D32func(pkg[2:6])
	pi.sendPool[pkgIndex] = pkg
}

func (u *UdpSp) sendSpPackages(sp *spProto, pi *udpSpPeerInfo, data []byte) {
	buf, _ := spProtoToBytes(sp)
	buf = append(buf, data...)
	for {
		switch {
		case len(buf) <= SP_UDP_DATA_PACKAGE_BODY_SIZE:
			u.sendSpPackage(sp, pi, buf)
			return
		default:
			u.sendSpPackage(sp, pi, buf[:SP_UDP_DATA_PACKAGE_BODY_SIZE])
			buf = buf[SP_UDP_DATA_PACKAGE_BODY_SIZE:]
		}
	}
}
