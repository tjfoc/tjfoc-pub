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
	"crypto/md5"
	"encoding/json"
	"fmt"

	"github.com/tjfoc/tjfoc/core/consensus"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
)

const (
	Init = iota
	Start
	Join
	Observer
)
const (
	// join = 0, quit = 1, observer = 2
	PeerJoin     = iota //节点加入
	PeerQuit            //节点退出
	PeerObserver        //节点加入 非共识节点
)

type commitLogFunc func(interface{}, []byte) error

func (r *raft) LeaderID() string {
	return r.leaderID
}

func (r *raft) Peers() []consensus.PeerInfo {
	logger.Infof("list Peers")
	var peers []consensus.PeerInfo
	localPeer := consensus.PeerInfo{ID: r.localPlainID, Addr: r.localAddr, Model: r.localModel, IsLeader: r.IsLeader()}
	peers = append(peers, localPeer)
	logger.Infof("PeerInfo [ID:%s, Addr:%s, Model:%d, IsLeader:%v]", localPeer.ID, localPeer.Addr, localPeer.Model, localPeer.IsLeader)
	for _, p := range r.peers {
		peer := consensus.PeerInfo{
			ID:    string(p.id),
			Addr:  p.address,
			Model: p.model,
		}
		if r.leaderID == p.pid {
			peer.IsLeader = true
		}
		peers = append(peers, peer)
		logger.Infof("PeerInfo [ID:%s, Addr:%s, Model:%d, IsLeader:%v]", peer.ID, peer.Addr, peer.Model, peer.IsLeader)
	}
	return peers
}

//typ 0-节点加入（PeerJoin）；1-节点退出（PeerQuit）；2-节点加入非共识PeerObserver
func (r *raft) UpdatePeer(typ int, id, addr string) bool {
	flag := false
	if r.IsLeader() {
		pid, _ := miscellaneous.GenHash(md5.New(), []byte(id))
		logger.Infof("update peer id:%s, addr:%s, typ:%d", id, addr, typ)
		_, ok := r.peers[string(pid)]
		//quite or join
		logger.Infof("r.localID:%x, para pid:%x, ok:%v", r.localID, pid, ok)
		if typ == PeerQuit {
			if r.localID != string(pid) && !ok {
				logger.Errorf("membership does not have peer:%s", id)
				return false
			}
		} else {
			if r.localID == string(pid) || ok {
				logger.Errorf("membership have exist peer:%s", id)
				return false
			}
		}
		log := make(map[string]string)
		log["ID"] = id
		log["Addr"] = addr
		log["Typ"] = fmt.Sprintf("%d", typ)
		peerLog, _ := json.Marshal(log)

		peerLog = append(miscellaneous.E32func(2), peerLog...)
		r.log = append(r.log, logEntry{index: r.getLastIndex() + 1, term: r.currentTerm, data: peerLog})
		logger.Infof("leader add peer log, term:%d, lastIndex:%d", r.currentTerm, r.getLastIndex())
		flag = true
	} else {
		logger.Infof("this not leader ,the leader peer: %s", r.leaderID)
	}
	return flag
}

func (r *raft) Start() {
	r.run()
}

func (r *raft) AppendEntries(log []byte) {
	if r.IsLeader() {
		log = append(miscellaneous.E32func(0), log...)
		r.log = append(r.log, logEntry{index: r.getLastIndex() + 1, term: r.currentTerm, data: log})
		logger.Infof("leader add log, term:%d, lastIndex:%d", r.currentTerm, r.getLastIndex())
	} else {
		logger.Infof("this not leader, the new leader is %x", r.leaderID)
	}
}

func (r *raft) IsLeader() bool {
	return r.getState() == leader
}
