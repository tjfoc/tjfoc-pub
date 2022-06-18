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
package raft

import (
	"errors"
	"sync"
	"time"

	"github.com/tjfoc/tjfoc/proto"
)

var lastSendMsgType MsgType

type PeerMessage struct {
	MSGType MsgType
	Data    []byte
}

type peer struct {
	//name
	name string
	//string ID
	pid string
	//origin ID
	id string
	//hashid
	hashid []byte
	//ip:port
	address string
	sp      proto.SP
	raft    *raft
	//启动模式是否核心节点，只有核心节点参与选举与共识
	model int
	//消息通道
	//peerCh chan *PeerMessage
	//退出通道
	exitCh   chan bool
	lock     sync.Mutex
	messages []*PeerMessage
	noticeCh chan bool

	//清理缓存的消息
	cleanNum int
}

func (p *peer) peerProcess() {
	for {
		select {
		case <-p.exitCh:
			logger.Warningf("peer %s remove raft membership.", p.name)
			return
		case <-p.noticeCh:
			for len(p.noticeCh) > 0 {
				<-p.noticeCh
			}
			p.lock.Lock()
			j := len(p.messages)
			p.lock.Unlock()

			for i := 0; i < j; i++ {
				if i < p.cleanNum {
					logger.Infof("peer %s msgLen:%d next:%d < cleanNum:%d break", p.name, j, i, p.cleanNum)
					break
				}
				msg := p.messages[i]
				p.sendPeerMessage(msg)
			}

			p.lock.Lock()
			if p.cleanNum > 0 {
				logger.Infof("peer %s clean message, cleanNum:%d  messagesLen:%d", p.name, p.cleanNum, len(p.messages))
				p.messages = p.messages[p.cleanNum:]
				p.cleanNum = 0
			} else {
				p.messages = p.messages[j:]
			}
			length := len(p.messages)
			p.lock.Unlock()

			if length > 0 && len(p.noticeCh) == 0 {
				p.noticeCh <- true
			}
		}
	}
}

func (p *peer) clean() {
	p.lock.Lock()
	if p.cleanNum == 0 {
		p.cleanNum = len(p.messages)
		if p.cleanNum > 0 {
			logger.Infof("peer %s clean message,set cleanNum:%d", p.name, p.cleanNum)
		}
	}
	p.lock.Unlock()
}

func (p *peer) notifyPeer(msgtype MsgType, data []byte) {
	p.lock.Lock()
	p.messages = append(p.messages,
		&PeerMessage{
			MSGType: msgtype,
			Data:    data,
		})
	p.lock.Unlock()
	if len(p.noticeCh) < maxPeerChan {
		p.noticeCh <- true
	}
}

func (p *peer) sendPeerMessage(msg *PeerMessage) {
	data := write(msg)
	t1 := time.Now()
	err := p.sp.SendInstruction(0, raftTCP, data, p.hashid)
	t := time.Now().Sub(t1)
	if t.Seconds() > 3 {
		logger.Warningf("SendInstruction %s %s take %s err:%v", p.name, parseMsgType(msg.MSGType), t, err)
	}
	if err != nil {
		logger.Debugf("SendInstruction %s %s error:%v", p.name, parseMsgType(msg.MSGType), err)
	} else {
		logger.Debugf("SendInstruction %s %s ok", p.name, parseMsgType(msg.MSGType))
	}
}

func show(buf []byte) (*PeerMessage, error) {
	msg := &PeerMessage{}
	t, err := bytesToUint32(buf[:4])
	if err != nil {
		logger.Errorf("bytesToUint32 err %s ", err)
		return msg, err
	}
	msg.MSGType = MsgType(t)
	msg.Data = buf[4:]
	return msg, nil
}

func write(msg *PeerMessage) []byte {
	buf := []byte{}
	d1 := uint32ToBytes(uint32(msg.MSGType))
	buf = append(buf, d1...)
	buf = append(buf, msg.Data...)
	return buf
}

func uint32ToBytes(a uint32) []byte {
	buf := make([]byte, 4)
	buf[0] = byte(a & 0xFF)
	buf[1] = byte((a >> 8) & 0xFF)
	buf[2] = byte((a >> 16) & 0xFF)
	buf[3] = byte((a >> 24) & 0xFF)
	return buf
}

func bytesToUint32(a []byte) (uint32, error) {
	if len(a) != 4 {
		return 0, errors.New("bytesToUint32: Illegal slice length")
	}
	b := uint32(0)
	for i, v := range a {
		b += uint32(v) << (8 * uint32(i))
	}
	return b, nil
}
