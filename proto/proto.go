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
	"errors"
	"sync"

	"github.com/tjfoc/tjfoc/core/miscellaneous"
)

const PROTO_SIZE = 12

type spRegistry map[uint32]SpFunc

// little endian
type spProto struct {
	size    uint32
	class   uint32
	version uint32
}

type sp struct {
	callback     SpCallback
	usrData      interface{}
	registryLock sync.RWMutex
	registry     map[uint32]spRegistry
}

type spInstruction struct {
	len  int    // 报文大小
	pid  []byte // 节点标识
	data []byte // 报文数据
}

func spProtoToBytes(sp *spProto) ([]byte, error) {
	if sp == nil {
		return nil, errors.New("spProtoToBytes: Null pointer")
	}
	buf := []byte{}
	buf = append(buf, miscellaneous.E32func(sp.size)...)
	buf = append(buf, miscellaneous.E32func(sp.class)...)
	buf = append(buf, miscellaneous.E32func(sp.version)...)
	return buf, nil
}

func bytesToSpProto(buf []byte) (*spProto, error) {
	if len(buf) != PROTO_SIZE {
		return nil, errors.New("bytesTospProto: Illegal slice length")
	}
	sp := &spProto{}
	sp.size, _ = miscellaneous.D32func(buf[:4])
	sp.class, _ = miscellaneous.D32func(buf[4:8])
	sp.version, _ = miscellaneous.D32func(buf[8:12])
	return sp, nil
}

func spNew(usrData interface{}, callback SpCallback) *sp {
	return &sp{
		usrData:  usrData,
		callback: callback,
		registry: make(map[uint32]spRegistry),
	}
}

func (s *sp) UnregisterFunc(version uint32, class uint32) error {
	s.registryLock.Lock()
	r, ok := s.registry[version]
	if ok {
		_, ok = r[class]
		if ok {
			delete(r, class)
		}
	}
	s.registryLock.Unlock()
	return nil
}

func (s *sp) RegisterFunc(version uint32, class uint32, callback SpFunc) error {
	s.registryLock.Lock()
	r, ok := s.registry[version]
	if !ok {
		r = make(spRegistry)
		s.registry[version] = r
	}
	_, ok = r[class]
	if !ok { // no cover
		r[class] = callback
	}
	s.registryLock.Unlock()
	return nil
}

func (s *sp) instructionDo(data, pid []byte, sp *spProto) {
	s.registryLock.RLock()
	defer s.registryLock.RUnlock()

	r, ok := s.registry[sp.version]
	if ok {
		f, ok := r[sp.class]
		if ok {
			f(s.usrData, pid, data)
		}
	}
}

func (s *sp) protoParse(ch chan *spInstruction) {
	tmp, sp := []byte{}, (*spProto)(nil)
	for {
		ins := <-ch
		n := ins.len
		switch n {
		case 0: // reset
			if sp != nil && uint32(len(tmp)) == sp.size {
				go s.instructionDo(miscellaneous.Dup(tmp), ins.pid, sp)
			}
			sp = nil
			tmp = []byte{}
		case -1: // exit
			close(ch)
			return
		}
		for i := 0; i < n; i++ {
			tmp = append(tmp, ins.data[i])
			switch sp {
			case nil:
				if len(tmp) == PROTO_SIZE {
					sp, _ = bytesToSpProto(tmp)
					tmp = []byte{}
				}
			default:
				if uint32(len(tmp)) == sp.size {
					go s.instructionDo(miscellaneous.Dup(tmp), ins.pid, sp)
					sp = nil
					tmp = []byte{}
				}
			}
		}
	}
}
