package proto

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
)

/*
 TCP报文格式
 	4 bytes: type{ 0: LOGIN, 1: PROTO_DATA}
 	16 bytes: peer id
*/
var logger = flogging.MustGetLogger("proto")

const SP_TCP_KEEP_LIVE_TIME = 2 * time.Second

const SP_TCP_PROTO_SIZE = 20

const (
	SP_TCP_LOGIN = iota
	SP_TCP_PROTO_DATA
)

type tcpSpPeerInfo struct {
	id   []byte
	addr *net.TCPAddr
	ch   chan *spInstruction
}

type TcpSp struct {
	sp         *sp
	id         []byte
	addr       *net.TCPAddr
	memberLock sync.RWMutex
	memberList map[string]*tcpSpPeerInfo
}

func TcpNew(id []byte, addr net.Addr, usrData interface{}, callback SpCallback) (*TcpSp, error) {
	tcpAddr, err := net.ResolveTCPAddr(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}
	return &TcpSp{
		addr:       tcpAddr,
		id:         miscellaneous.Dup(id),
		sp:         spNew(usrData, callback),
		memberList: make(map[string]*tcpSpPeerInfo),
	}, nil
}

func (t *TcpSp) GetMemberList() map[string]string {
	b := make(map[string]string)
	t.memberLock.Lock()
	for _, v := range t.memberList {
		b[string(v.id)] = v.addr.String()
	}
	t.memberLock.Unlock()
	return b
}

func (t *TcpSp) RegisterPeer(id []byte, addr net.Addr) error {
	pid := miscellaneous.Dup(id)
	if strings.Compare(addr.Network(), t.addr.Network()) != 0 {
		return errors.New(fmt.Sprintf("RegisterPeer: only support: %s", t.addr.Network()))
	}
	tcpAddr, err := net.ResolveTCPAddr(addr.Network(), addr.String())
	if err != nil {
		return err
	}
	pi := &tcpSpPeerInfo{
		id:   pid,
		addr: tcpAddr,
		ch:   make(chan *spInstruction),
	}
	t.memberLock.Lock()
	t.memberList[miscellaneous.Bytes2HexString(pid)] = pi
	t.memberLock.Unlock()
	return nil
}

func (t *TcpSp) UnregisterPeer(id []byte) error {
	key := miscellaneous.Bytes2HexString(id)
	t.memberLock.Lock()
	_, ok := t.memberList[key]
	if ok {
		delete(t.memberList, key)
	}
	t.memberLock.Unlock()
	return nil
}

func (t *TcpSp) UnregisterFunc(version uint32, class uint32) error {
	return t.sp.UnregisterFunc(version, class)
}

func (t *TcpSp) RegisterFunc(version uint32, class uint32, callback SpFunc) error {
	return t.sp.RegisterFunc(version, class, callback)
}

func (t *TcpSp) SendInstruction(version uint32, class uint32, data []byte, id []byte) error {
	t.memberLock.RLock()
	pi, ok := t.memberList[miscellaneous.Bytes2HexString(id)]
	t.memberLock.RUnlock()
	if !ok {
		return errors.New(fmt.Sprintf("SendInstruction: unregister peer %s", miscellaneous.Bytes2HexString(id)))
	}
	con, err := net.DialTimeout(t.addr.Network(), pi.addr.String(), 15*time.Second)
	if err != nil {
		return err
	}
	defer con.Close()
	buf := []byte{}
	buf = append(buf, miscellaneous.E32func(SP_TCP_LOGIN)...) // login pkg
	buf = append(buf, t.id...)
	sp := &spProto{
		class:   class,
		version: version,
		size:    uint32(len(data)),
	}
	spBuf, _ := spProtoToBytes(sp)
	buf = append(buf, spBuf...)
	buf = append(buf, data...) // sp proto pkg
	con.Write(buf)
	return nil
}

func (t *TcpSp) Run() error {
	lis, err := net.ListenTCP(t.addr.Network(), t.addr)
	if err != nil {
		return err
	}
	logger.Infof("tcp server startup %s", t.addr)
	for {
		if con, err := lis.Accept(); err != nil {
			lis.Close()
			lis, err = net.ListenTCP(t.addr.Network(), t.addr)
			if err != nil {
				time.Sleep(3 * time.Second)
			}
			continue
		} else {
			go t.connectionProcess(con)
		}
	}
}

func (t *TcpSp) connectionProcess(con net.Conn) {
	var count int
	var pid []byte
	defer con.Close()
	buf := make([]byte, 1024)
	con.SetReadDeadline(time.Now().Add(SP_TCP_KEEP_LIVE_TIME))
	ch := make(chan *spInstruction)
	go t.sp.protoParse(ch)
	defer func() {
		ch <- &spInstruction{
			len: -1, // exit
		}
	}()
	for { // login
		n, err := con.Read(buf)
		if err != nil || n == 0 {
			ch <- &spInstruction{
				len: 0, // reset
			}
			return
		}
		if n >= SP_TCP_PROTO_SIZE {
			pid = miscellaneous.Dup(buf[4:SP_TCP_PROTO_SIZE])
			t.memberLock.RLock()
			_, ok := t.memberList[miscellaneous.Bytes2HexString(pid)]
			t.memberLock.RUnlock()
			if !ok {
				switch t.sp.callback(t.sp.usrData, pid, PROTO_ID) {
				case PROTO_ACCEPTABLE_PEER:
					t.memberLock.RLock()
					_, ok = t.memberList[miscellaneous.Bytes2HexString(pid)]
					t.memberLock.RUnlock()
				case PROTO_UNACCEPTABLE_PEER:
					return
				}
				if !ok {
					return
				}
			}
			count = n - SP_TCP_PROTO_SIZE
			if n > SP_TCP_PROTO_SIZE {
				ch <- &spInstruction{
					pid:  pid,
					len:  n - SP_TCP_PROTO_SIZE,
					data: miscellaneous.Dup(buf[SP_TCP_PROTO_SIZE:n]),
				}
			}
			break
		}
	}
	for { // sp proto process
		n, err := con.Read(buf)
		if err != nil || n == 0 {
			ch <- &spInstruction{
				len: 0, // reset
			}
			return
		}
		count += n
		ch <- &spInstruction{
			len:  n,
			pid:  pid,
			data: miscellaneous.Dup(buf[:n]),
		}
	}
}
