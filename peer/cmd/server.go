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
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/tjfoc/tjfoc/core/store/chain"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmtls"
	"github.com/tjfoc/gmtls/gmcredentials"
	"github.com/tjfoc/tjfoc/core/block"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	ppr "github.com/tjfoc/tjfoc/protos/peer"
	"google.golang.org/grpc"
)

func newGrpcServer() (*grpc.Server, []grpc.DialOption) {
	if Config.Rpc.UseTLS {
		cert, err := gmtls.LoadX509KeyPair(Config.Rpc.TLSCertPath, Config.Rpc.TLSKeyPath)
		if err != nil {
			logger.Fatal(err)
		}
		certPool := sm2.NewCertPool()
		cacert, err := ioutil.ReadFile(Config.Rpc.TLSCaPath)
		if err != nil {
			logger.Fatal(err)
		}
		certPool.AppendCertsFromPEM(cacert)
		creds := gmcredentials.NewTLS(&gmtls.Config{
			ServerName:   Config.Rpc.TLSServerName,
			Certificates: []gmtls.Certificate{cert},
			RootCAs:      certPool,
		})
		return grpc.NewServer(grpc.Creds(gmcredentials.NewTLS(&gmtls.Config{
				Certificates: []gmtls.Certificate{cert},
				ClientCAs:    certPool,
			}))), []grpc.DialOption{grpc.WithTimeout(time.Duration(Config.Rpc.Timeout) * time.Millisecond),
				grpc.WithTransportCredentials(creds)}

	} else {
		return grpc.NewServer(), []grpc.DialOption{grpc.WithInsecure()}
	}
}

func (p *peer) serverRun() {
	ppr.RegisterPeerServer(p.grpcServer, p)
	if lis, err := net.Listen("tcp", fmt.Sprintf(":%d", Config.Rpc.Port)); err != nil {
		logger.Fatal(err)
	} else {
		go p.packProcess()
		go p.p2pServer.Run()
		go p.consensusAPI.Start()
	retry:
		p.grpcServer.Serve(lis)
		for {
			lis, err = net.Listen("tcp", fmt.Sprintf(":%d", Config.Rpc.Port))
			if err == nil {
				goto retry
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func (p *peer) packProcess() {
	for {
		select {
		case <-time.After(time.Duration(Config.PackTime) * time.Millisecond):
			if p.consensusAPI.IsLeader() {
				if b := p.blockChain.Pack(uint64(time.Now().Unix())); b != nil {
					if data, err := b.Show(); err == nil {
						p.consensusAPI.AppendEntries(data)
					}
				}
			}
		}
	}
}

func packCallback(a interface{}, d []byte) error {
	p, _ := a.(*peer)
	typ, _ := miscellaneous.D32func(d[:4])

	logger.Infof("*************packCallback begin typ:%v***********", typ)
	defer logger.Infof("*************packCallback end typ:%v***********", typ)

	switch typ {
	case 0:

		b := new(block.Block)

		if _, err := b.Read(d[4:]); err != nil {
			logger.Errorf("b.Read Err: %s", err)
			return err
		}
		if err := p.blockChain.PackAddBlock(b); err != nil {
			return err
		}
	case 1:
		h, _ := miscellaneous.D64func(d[4:12])
		p.blockChain.SaveThreshold(h)
		logger.Warningf("+++++++++++packcall pack snapshot %v\n", h)
	case 2:
		v := make(map[string]string)

		miscellaneous.Unmarshal(d[4:], &v)
		id, _ := miscellaneous.GenHash(md5.New(), []byte(v["ID"]))
		sid, _ := miscellaneous.GenHash(md5.New(), []byte(Config.Self.Id))
		switch v["Typ"] {
		case "1":
			if bytes.Compare(id, sid) == 0 {
				logger.Warning("Peer exit")
			} else {
				delete(p.memberList, v["ID"])
				p.blockChain.SaveMemberList(p.memberList)
				p.p2pServer.UnregisterPeer(id)
			}
		case "0":
			if addr, err := net.ResolveTCPAddr("tcp", v["Addr"]); err != nil {
				logger.Fatal(err)
			} else {
				p.memberList[v["ID"]] = &chain.PeerInfo{
					Typ:  0,
					Addr: v["Addr"],
				}
				p.blockChain.SaveMemberList(p.memberList)
				p.p2pServer.RegisterPeer(id, addr)
			}
		case "2":
			if addr, err := net.ResolveTCPAddr("tcp", v["Addr"]); err != nil {
				logger.Fatal(err)
			} else {
				p.memberList[v["ID"]] = &chain.PeerInfo{
					Typ:  2,
					Addr: v["Addr"],
				}
				p.blockChain.SaveMemberList(p.memberList)
				p.p2pServer.RegisterPeer(id, addr)
			}
		}
	}
	return nil
}
