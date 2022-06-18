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
	"context"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"google.golang.org/grpc"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmtls"
	"github.com/tjfoc/gmtls/gmcredentials"
	"github.com/tjfoc/tjfoc/core/consensus/raft"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/store/chain"
	"github.com/tjfoc/tjfoc/proto"
	ppr "github.com/tjfoc/tjfoc/protos/peer"
)

func mCallback(usrData interface{}, id []byte, typ int) int {
	return proto.PROTO_UNACCEPTABLE_PEER
}

func getConn(address string) (*grpc.ClientConn, error) {
	if Config.Rpc.UseTLS {
		cert, err := gmtls.LoadX509KeyPair(Config.Rpc.TLSCertPath, Config.Rpc.TLSKeyPath)
		if err != nil {
			return nil, err
		}
		certPool := sm2.NewCertPool()
		cacert, err := ioutil.ReadFile(Config.Rpc.TLSCaPath)
		if err != nil {
			return nil, err
		}
		certPool.AppendCertsFromPEM(cacert)
		creds := gmcredentials.NewTLS(&gmtls.Config{
			ServerName:   Config.Rpc.TLSServerName,
			Certificates: []gmtls.Certificate{cert},
			RootCAs:      certPool,
		})
		return grpc.Dial(address, grpc.WithTransportCredentials(creds))
	} else {
		return grpc.Dial(address, grpc.WithInsecure())
	}
}

func getMemberList() (*ppr.MemberListInfo, error) {
	var addr string

	for _, v := range Config.Members.Peers {
		if v.Id != Config.Self.Id {
			s := strings.Split(v.Addr, ":")
			addr = fmt.Sprintf("%s:%v", s[0], Config.Rpc.Port)
		}
	}
	conn, err := getConn(addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	c := ppr.NewPeerClient(conn)
	members, err := c.GetMemberList(context.Background(), &ppr.BlockchainBool{})
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (p *peer) setMemberList(typ string) error {
	switch typ {
	case "start":
		if members, err := getMemberList(); err == nil {
			for _, v := range members.MemberList {
				if v.Id != Config.Self.Id {
					state := 0
					if v.State != raft.Join {
						state = int(v.State)
					}
					p.memberList[string(v.Id)] = &chain.PeerInfo{
						Typ:  uint32(state),
						Addr: string(miscellaneous.Dup([]byte(v.Addr))),
					}
				}
			}
			return nil
		}
		members, err := p.blockChain.GetMemberList()
		if err != nil {
			return err
		}
		for k, v := range members {
			p.memberList[k] = v
		}
		return nil
	case "join", "observer":
		members, err := getMemberList()
		if err != nil {
			return err
		}
		for _, v := range members.MemberList {
			if v.Id != Config.Self.Id {
				state := 0
				if v.State != raft.Join {
					state = int(v.State)
				}
				p.memberList[string(v.Id)] = &chain.PeerInfo{
					Typ:  uint32(state),
					Addr: string(miscellaneous.Dup([]byte(v.Addr))),
				}
			}
		}
		return nil
	default:
		for _, v := range Config.Members.Peers {
			if v.Id != Config.Self.Id {
				p.memberList[string(v.Id)] = &chain.PeerInfo{
					Typ:  uint32(v.Typ),
					Addr: string(miscellaneous.Dup([]byte(v.Addr))),
				}
			}
		}
		return nil
	}

}

func (p *peer) memberListInit() (proto.SP, error) {
	addr, err := net.ResolveTCPAddr("tcp", Config.Self.Addr)
	if err != nil {
		return nil, err
	}
	id, _ := miscellaneous.GenHash(md5.New(), []byte(Config.Self.Id))
	sp, err := proto.TcpNew(id, addr, p, mCallback)
	if err != nil {
		return nil, err
	}
	p.blockChain.SaveMemberList(p.memberList)
	return sp, nil
}
