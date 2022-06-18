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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/tjfoc/tjfoc/core/block"
	"github.com/tjfoc/tjfoc/core/consensus"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	pb "github.com/tjfoc/tjfoc/protos/raft"
)

var lastMsgType MsgType

type message struct {
	msgtype MsgType
	peerid  string
	data    []byte
}

type MsgType uint32

const (
	//vote
	vote MsgType = iota
	voteRequest
	voteResponse
	//log
	heartBeat
	appendLogRequest
	appendLogResponse
	// commitLog
	leaderCheck
	//snapshoot
	snapshootRequest
)

func (r *raft) messageProcess() {
	for msg := range r.msgCh {
		switch msg.msgtype {
		case vote:
			r.doVote(msg)
		case voteRequest:
			r.doVoteRequest(msg)
		case voteResponse:
			r.doVoteResponse(msg)
		case leaderCheck:
			r.doLeaderCheck()
		case heartBeat:
			r.doHeartBeat()
		case appendLogRequest:
			r.doAppendLogRequest(msg)
		case appendLogResponse:
			r.doAppendLogResponse(msg)
		case snapshootRequest:
			r.doSnapRequest(msg)
		}
	}
}

func (r *raft) notifyMessage(p *peer, msgtype MsgType, data []byte) {
	msg := &message{msgtype: msgtype}
	if p != nil {
		msg.peerid = p.pid
	}
	if data != nil {
		msg.data = data
	}
	if p != nil {
		logger.Debugf("add messageCh len:%d, msgtype:%s, lastMsgType:%s, peer:%s", len(r.msgCh), parseMsgType(msgtype), parseMsgType(lastMsgType), p.name)
	} else {
		logger.Debugf("add messageCh len:%d, msgtype:%s, lastMsgType:%s", len(r.msgCh), parseMsgType(msgtype), parseMsgType(lastMsgType))
	}
	if len(r.msgCh) > 1 {
		logger.Infof("add messageCh len:%d, msgtype:%s, lastMsgType:%s", len(r.msgCh), parseMsgType(msgtype), parseMsgType(lastMsgType))
	}
	if len(r.msgCh) == maxMsgChan {
		logger.Warningf("lastMsgType:%s, msgCh:%d is full, remove top", parseMsgType(msgtype), len(r.msgCh))
		<-r.msgCh
	}
	lastMsgType = msg.msgtype
	r.msgCh <- msg
}

func (r *raft) doLeaderCheck() {
	logger.Debugf("doLeaderCheck")
	r.checkSnapshoot()
	r.checkCommit()
}

func (r *raft) doSnapshoot(msg *message) {
	if p, ok := r.peers[msg.peerid]; ok {
		p.notifyPeer(snapshootRequest, msg.data)
	}
}

func (r *raft) doSnapRequest(msg *message) {
	p, ok := r.peers[msg.peerid]
	if !ok {
		logger.Errorf("could not found peer id:%s", msg.peerid)
		return
	}
	logger.Infof("in doSnapRequest rev peer:%s", p.name)
	snapReq := &pb.SnapshootRequest{}
	err := proto.Unmarshal(msg.data, snapReq)
	if err != nil {
		logger.Errorf("proto.Unmarshal err %s", err)
		return
	}
	logger.Infof("snapshoot begin req[term:%d,index:%d,height:%d], cur[term:%d,index:%d,logLen:%d]",
		snapReq.Term, snapReq.LastIncludedIndex, snapReq.BlockHeight, r.log[0].term, r.log[0].index, len(r.log))

	defer func() {
		respLog := &pb.AppendEntriesResponse{
			Term:           r.currentTerm,
			LastLog:        r.getLastIndex(),
			Success:        true,
			NoRetryBackoff: false,
		}
		bytes, err := proto.Marshal(respLog)
		if err != nil {
			logger.Errorf("proto.Marshal(voteResp) err %s", err)
			return
		}
		if p, ok := r.peers[msg.peerid]; ok {
			p.notifyPeer(appendLogResponse, bytes)
		}
	}()

	//任期
	if snapReq.Term < r.currentTerm {
		logger.Infof("return false,LastLog:%d. peer:%s, req.Term:%d < cur.Term: %d ",
			r.getLastIndex(), p.name, snapReq.Term, r.currentTerm)
		return
	}

	// term >currentTerm tobe follower
	if snapReq.Term > r.currentTerm {
		logger.Infof("snapreq.Term:%d, cur.Term:%d, cur.state:%s", snapReq.Term, r.currentTerm, r.parseRaftState(r.getState()))
		r.leaderID = ""
		r.currentTerm = snapReq.Term
		if r.getState() != follower && r.getState() != suspended {
			logger.Infof("set new state:follower, term:%d", r.currentTerm)
			r.setState(follower)
			r.breakCh <- 1
		}
	}
	//更新leaderID
	if (r.getState() == follower || r.getState() == suspended) && r.leaderID == "" {
		r.leaderID = snapReq.LeaderId
		logger.Infof("update new leader:%s", p.name)
	}
	//领导人不一致
	if snapReq.LeaderId != r.leaderID {
		logger.Infof("snapreq.Term [%d] = r.currentTerm but reqLog.LeaderId[%x] != cur.leaderID[%x]", snapReq.Term, snapReq.LeaderId, r.leaderID)
		return
	}
	// follower 接收一次心跳
	if r.getState() == follower {
		r.heartbeatCh <- true
	}
	//快照是否已经安装过
	if r.log[0].index >= snapReq.LastIncludedIndex {
		logger.Infof("return, cur.baseIndex:%d >= snapReq.index:%d, have installed snapshoot", r.log[0].index, snapReq.LastIncludedIndex)
		return
	}

	//节点更新
	var snapPeers []consensus.PeerInfo
	err = json.Unmarshal(snapReq.Peers[8:], &snapPeers)
	if err != nil {
		logger.Errorf("json.Unmarshal(snapReq.Peers, pi) err :%s", err)
		return
	}
	curPeers := r.Peers()
	var updatePeers []map[string]string
	//新增
	flag := true
	for _, v := range snapPeers {
		flag = true
		for _, v2 := range curPeers {
			if v.ID == v2.ID {
				flag = false
				break
			}
		}
		if flag {
			logger.Infof("add new peer [id:%s,addr:%s]", v.ID, v.Addr)
			log := make(map[string]string)
			log["ID"] = v.ID
			log["Addr"] = v.Addr
			if v.Model == 0 {
				log["Typ"] = fmt.Sprintf("%d", PeerJoin)
			} else {
				log["Typ"] = fmt.Sprintf("%d", PeerObserver)
			}
			updatePeers = append(updatePeers, log)
		}
	}
	//退出
	for _, v := range curPeers {
		flag = true
		for _, v2 := range snapPeers {
			if v.ID == v2.ID {
				flag = false
				break
			}
		}
		if flag {
			logger.Infof("delete peer [id:%s,addr:%s]", v.ID, v.Addr)
			log := make(map[string]string)
			log["ID"] = v.ID
			log["Addr"] = v.Addr
			log["Typ"] = fmt.Sprintf("%d", PeerQuit)
			updatePeers = append(updatePeers, log)
		}
	}
	for _, updpeer := range updatePeers {
		log, err := json.Marshal(updpeer)
		if err != nil {
			logger.Errorf("json.Marshal(updpeer) err :%s", err)
			return
		}
		log = append(miscellaneous.E32func(2), log...)
		logger.Infof("begin commit log, peer update")
		r.commitLog(2, 0, log)
		logger.Infof("end commit log, peer update")
	}

	//清除现有日志，生成新快照
	r.log = r.log[:1]
	r.log[0] = logEntry{index: snapReq.LastIncludedIndex, term: snapReq.LastIncludedTerm, data: snapReq.Peers}
	logger.Infof("snapshoot end snapLog term:%d, index:%d", snapReq.LastIncludedTerm, snapReq.LastIncludedIndex)

	//提交快照
	log := append(miscellaneous.E32func(1), miscellaneous.E64func(snapReq.BlockHeight)...)
	logger.Infof("begin commit snaplog,height:%d", snapReq.BlockHeight)
	r.commitLog(1, 0, log)
	logger.Infof("end commit snaplog")
}

func (r *raft) doVote(msg *message) {
	logger.Debugf("in doVote")
	if r.getState() != candidate {
		logger.Infof("cur state is %s not candidate, exit doVote", r.parseRaftState(r.getState()))
		return
	}
	if p, ok := r.peers[msg.peerid]; ok {
		if p.model == 0 {
			p.notifyPeer(voteRequest, msg.data)
		}
	}
}

func (r *raft) getBlockHeight(index uint64) uint64 {
	logger.Infof("begin getBlockHeight")
	var bh uint64
	t1 := time.Now()
	if r.getLastIndex() == 0 {
		bh = r.blockChain.Height()
		logger.Infof("end getBlockHeight, bh:%d", bh)
		return bh
	}
	//从提交的日志中获取高度
	if index > 0 {
		b := new(block.Block)
		for i := index; i > 0; i-- {
			typ, err := bytesToUint32(r.log[i].data[:4])
			if err == nil && typ == 0 {
				b.Read(r.log[index].data[4:])
				bh = b.Height()
				logger.Infof("end getBlockHeight from logs, bh:%d", bh)
				return bh
			}
		}
	}
	bh = r.blockChain.Height()
	t := time.Now().Sub(t1)
	if t.Seconds() > 1 {
		logger.Warningf("getBlockHeight take long time %s", t)
	}
	logger.Infof("end getBlockHeight, bh:%d", bh)
	return bh
}

func (r *raft) doVoteRequest(msg *message) {
	p, ok := r.peers[msg.peerid]
	if !ok {
		logger.Errorf("unknow peer:%s", msg.peerid)
		return
	}
	logger.Infof("doVoteRequest rev peer:%s", p.name)
	//返回投票结果
	voteResp := &pb.RequestVoteResponse{
		Term:        r.currentTerm,
		VoteGranted: false,
	}
	defer func() {
		logger.Infof("back voteResponse [term:%d, blockHeight:%d, voteGranted:%v]", voteResp.Term, voteResp.BlockHeight, voteResp.VoteGranted)
		bytes, err := proto.Marshal(voteResp)
		if err != nil {
			logger.Errorf("proto.Marshal(voteResp) err %s", err)
			return
		}
		if p, ok := r.peers[msg.peerid]; ok {
			p.notifyPeer(voteResponse, bytes)
		}
	}()

	//处理投票请求
	voteReq := &pb.RequestVoteRequest{}
	err := proto.Unmarshal(msg.data, voteReq)
	if err != nil {
		logger.Errorf("proto.Unmarshal err %s", err)
		return
	}
	// bh := r.getBlockHeight(r.commitIndex - r.log[0].index)
	bh := r.getBlockHeight(0)
	voteResp.BlockHeight = bh
	logger.Infof("req,term:%d, blockHeight:%d; cur,term:%d, blockHeight:%d ", voteReq.Term, voteReq.BlockHeight, r.currentTerm, bh)

	//如果term < currentTerm返回 false
	if voteReq.Term < r.currentTerm {
		logger.Infof("req.Term:%d < cur.Term :%d not vote return false.", voteReq.Term, r.currentTerm)
		return
	}
	//任期号T > currentTerm，那么就令 currentTerm 等于 T，并切换状态为跟随者
	if voteReq.Term > r.currentTerm {
		logger.Infof("update new term %d -> %d", r.currentTerm, voteReq.Term)
		r.currentTerm = voteReq.Term
		r.leaderID = ""
		if r.getState() != follower {
			logger.Infof("update state %s -> follower", r.parseRaftState(r.getState()))
			r.setState(follower)
			r.breakCh <- 1
		}
		//
		r.cleanPeer()
	}
	voteResp.Term = r.currentTerm
	if voteReq.BlockHeight < bh {
		logger.Infof("req.BlockHeight:%d < cur.BlockHeight:%d not vote,return", voteReq.BlockHeight, bh)
		return
	}
	//一个任期号投出一张选票
	if strings.Compare(r.leaderID, "") == 0 || strings.Compare(r.leaderID, voteReq.CandidateId) == 0 {
		logger.Infof("req (LastLogTerm :%d,LastLogIndex:%d), cur (lastLogTerm:%d,lastLogIndex:%d)",
			voteReq.LastLogTerm, voteReq.LastLogIndex, r.getLastTerm(), r.getLastIndex())
		//如果两份日志最后的条目的任期号不同，那么任期号大的日志更加新。如果两份日志最后的条目任期号相同，那么日志比较长的那个就更加新
		if voteReq.LastLogTerm > r.getLastTerm() || (voteReq.LastLogTerm == r.getLastTerm() && voteReq.LastLogIndex >= r.getLastIndex()) {
			voteResp.VoteGranted = true
			logger.Infof("vote to %s", p.name)
			r.leaderID = voteReq.CandidateId
			if r.getState() != follower {
				r.setState(follower)
				r.breakCh <- 1
			}
		} else {
			logger.Infof("req peer %s not newer then cur, not vote", p.name)
		}
	} else {
		logger.Infof("cur peer have voted.")
	}
	logger.Infof("exit doVoteRequest")
}

func (r *raft) doVoteResponse(msg *message) {
	//如果一个候选人或者领导者发现自己的任期号过期了，那么他会立即恢复成跟随者状态
	//如果一个服务器的当前任期号比其他人小，那么他会更新自己的编号到较大的编号值
	//候选人从整个集群的大多数服务器节点获得了针对同一个任期号的选票
	if p, ok := r.peers[msg.peerid]; ok {
		logger.Infof("doVoteResponse rev peer:%s", p.name)
	}
	if r.getState() != candidate {
		logger.Infof("cur state is %s not candidate, exit doVoteResponse", r.parseRaftState(r.getState()))
		return
	}
	voteResp := &pb.RequestVoteResponse{}
	err := proto.Unmarshal(msg.data, voteResp)
	if err != nil {
		logger.Infof("proto.Unmarshal err %s", err)
		return
	}
	// bh := r.getBlockHeight(r.commitIndex - r.log[0].index)
	bh := r.getBlockHeight(0)
	logger.Infof("vote resp[term:%d, voteGranted:%v, blockHeight:%d] cur[term:%d, blockHeight:%d]",
		voteResp.Term, voteResp.VoteGranted, voteResp.BlockHeight, r.currentTerm, bh)
	if r.currentTerm > voteResp.Term {
		logger.Infof("cur.Term > resp.Term, exit doVoteResponse")
		return
	}
	if r.currentTerm < voteResp.Term || bh < voteResp.BlockHeight {
		logger.Infof("cur.Term < resp.Term or < cur.blockHeigh < resp.blockHeight")
		r.currentTerm = voteResp.Term
		logger.Infof("update state candidate -> follower, term:%d", r.currentTerm)
		r.setState(follower)
		r.leaderID = ""
		r.breakCh <- 1
		return
	}
	if voteResp.VoteGranted {
		r.grantedVotes++
	}
	logger.Infof("get grantedVotes:%d", r.grantedVotes)
	if r.grantedVotes > r.quorumSize() {
		logger.Infof("win leader term:%d", r.currentTerm)
		logger.Infof("update all follower peers nextIndex=%d", r.getLastIndex()+1)
		for _, peer := range r.peers {
			peer.clean() //clean message
			if peer.model == 0 {
				r.matchIndex[peer.pid] = 0
			}
			r.nextIndex[peer.pid] = r.getLastIndex() + 1
		}
		r.leaderID = r.localID
		r.setState(leader)
		r.leaderCh <- true
	}
	logger.Info("exit doVoteResponse")
}

func (r *raft) doHeartBeat() {
	for _, peer := range r.peers {
		//r.notifyMessage(peer, heartBeat, nil)
		r.heartBeatToPeer(peer)
	}
}

func (r *raft) heartBeatToPeer(p *peer) {
	// func (r *raft) doHeartBeat(msg *message) {
	if r.localID != r.leaderID {
		logger.Infof("cur state:%s, not leader, cancled send heartBeat", r.parseRaftState(r.getState()))
		if ld, ok := r.peers[r.leaderID]; ok {
			logger.Debugf("leader is %s", ld.name)
		}
		return
	}
	// p, ok := r.peers[msg.peerid]
	// if !ok {
	// 	logger.Errorf("not found peer peerid: %x", msg.peerid)
	// 	return
	// }
	baseIndex := r.log[0].index
	// nextIndex := r.nextIndex[msg.peerid]
	nextIndex := r.nextIndex[p.pid]
	//发送日志，否则通知follower安装快照
	if nextIndex > baseIndex {
		if nextIndex-baseIndex > uint64(len(r.log)) {
			logger.Errorf("peer:%s, nextIndex:%d, baseIndex:%d, len log:%d, commitIndex:%d", p.name, nextIndex, baseIndex, len(r.log), r.commitIndex)
			nextIndex = baseIndex + uint64(len(r.log)) - 1
			if nextIndex < 1 {
				nextIndex = 1
			}
			r.nextIndex[p.pid] = nextIndex
			logger.Warningf("set nextIndex:%d", nextIndex)
		}
		reqLog := &pb.AppendEntriesRequest{
			Term:         r.currentTerm,
			LeaderId:     r.localID,
			LeaderCommit: r.commitIndex,
			PrevLogIndex: r.log[nextIndex-baseIndex-1].index,
			PrevLogTerm:  r.log[nextIndex-baseIndex-1].term,
			Entries:      []*pb.LogEntry{},
		}
		if r.getLastIndex() >= nextIndex {
			reqLog.Entries = append(reqLog.Entries, &pb.LogEntry{
				Index:   r.log[nextIndex-baseIndex].index,
				Term:    r.log[nextIndex-baseIndex].term,
				Command: r.log[nextIndex-baseIndex].data,
			})
			logger.Debugf("send appendLogRequest %s, log index:%d", p.name, nextIndex)
		} else {
			logger.Debugf("send heartBeat %s", p.name)
		}
		bytes, err := proto.Marshal(reqLog)
		if err != nil {
			logger.Errorf("proto.Marshal err %s", err)
			return
		}
		p.notifyPeer(appendLogRequest, bytes)
	} else {
		snapReq := &pb.SnapshootRequest{
			Term:              r.currentTerm,
			LeaderId:          r.localID,
			LastIncludedIndex: r.log[0].index,
			LastIncludedTerm:  r.log[0].term,
			BlockHeight:       r.log[0].height,
			Peers:             r.log[0].data,
		}
		bytes, err := proto.Marshal(snapReq)
		if err != nil {
			logger.Errorf("proto.Marshal err %s", err)
			return
		}
		logger.Infof("peer:%s nextIndex:%d < cur.baseIndex:%d, install snapshoot [Term:%d, Index:%d，Height:%d]",
			p.name, nextIndex, baseIndex, snapReq.LastIncludedTerm, snapReq.LastIncludedIndex, snapReq.BlockHeight)
		p.notifyPeer(snapshootRequest, bytes)
	}
}

func (r *raft) doAppendLogRequest(msg *message) {
	p, ok := r.peers[msg.peerid]
	if !ok {
		logger.Errorf("could not found peer id:%s exit", msg.peerid)
		return
	}
	logger.Debugf("peer:%s doAppendLogRequest", p.name)
	reqLog := &pb.AppendEntriesRequest{}
	respLog := &pb.AppendEntriesResponse{
		Term:           r.currentTerm,
		LastLog:        r.getLastIndex(),
		Success:        false,
		NoRetryBackoff: false,
	}

	//返回附加日志结果
	defer func() {
		respLog.Term = r.currentTerm
		bytes, err := proto.Marshal(respLog)
		if err != nil {
			logger.Errorf("proto.Marshal(voteResp) err %s", err)
			return
		}
		if p, ok := r.peers[msg.peerid]; ok {
			logger.Debugf("back respLog(Term:%d,LastLog:%d,Success:%v,NoRetryBackoff:%v)",
				respLog.Term, respLog.LastLog, respLog.Success, respLog.NoRetryBackoff)
			p.notifyPeer(appendLogResponse, bytes)
		}
	}()

	err := proto.Unmarshal(msg.data, reqLog)
	if err != nil {
		logger.Errorf("proto.Unmarshal err %s", err)
		respLog.NoRetryBackoff = true
		return
	}
	//如果 term < currentTerm 就返回 false
	if reqLog.Term < r.currentTerm {
		logger.Infof("peer:%s req.Term:%d < cur.Term: %d, exit doAppendLogRequest",
			p.name, reqLog.Term, r.currentTerm)
		return
	}

	// term >currentTerm tobe follower
	if reqLog.Term > r.currentTerm {
		logger.Infof("req.Term:%d > cur.Term:%d, cur.state:%s", reqLog.Term, r.currentTerm, r.parseRaftState(r.getState()))
		r.leaderID = ""
		r.currentTerm = reqLog.Term
		respLog.Term = r.currentTerm
		//
		r.cleanPeer()

		if r.getState() != follower && r.getState() != suspended {
			logger.Infof("set new state:follower, term:%d", r.currentTerm)
			r.setState(follower)
			r.breakCh <- 1
		}
	}
	//更新leaderID
	if (r.getState() == follower || r.getState() == suspended) && r.leaderID == "" {
		r.leaderID = reqLog.LeaderId
		logger.Infof("update new leader:%s", p.name)
	}
	//领导人不一致
	if reqLog.LeaderId != r.leaderID {
		logger.Infof("peer:%s. reqLog.Term [%d] = r.currentTerm but reqLog.LeaderId[%x] != cur.leaderID[%x] return false",
			p.name, reqLog.Term, reqLog.LeaderId, r.leaderID)
		respLog.NoRetryBackoff = true
		return
	}

	//follower 接收一次心跳
	if r.getState() == follower {
		r.heartbeatCh <- true
	}
	baseIndex := r.log[0].index

	if reqLog.PrevLogIndex > r.getLastIndex() || reqLog.PrevLogIndex < baseIndex {
		logger.Warningf("peer:%s. does not match cur.baseIndex:%d < req.PrevLogIndex:%d < cur.lastIndex:%d, exit",
			p.name, baseIndex, reqLog.PrevLogIndex, r.getLastIndex())
		if r.getLastIndex() < baseIndex {
			respLog.LastLog = baseIndex
		}
		return
	}
	//如果已经存在的日志条目和新的产生冲突（索引值相同但是任期号不同），删除这一条和之后所有的
	if reqLog.PrevLogIndex > baseIndex {
		term := r.log[reqLog.PrevLogIndex-baseIndex].term
		if reqLog.PrevLogTerm != term {
			logger.Infof("reqLog.PrevLogIndex:%d, reqLog.PrevLogTerm:%d, baseIndex:%d, term:%d", reqLog.PrevLogIndex, reqLog.PrevLogTerm, baseIndex, term)
			logger.Warningf("reqLog.PrevLogTerm %d != exist log term %d, exit", reqLog.PrevLogTerm, term)
			r.log = r.log[:reqLog.PrevLogIndex-baseIndex]
			return
		}
	}
	//附加日志
	if len(reqLog.Entries) > 0 {
		logger.Infof("reqLog[term:%d,preTerm:%d,preIdx:%d,leaderCommit:%d] ", reqLog.Term, reqLog.PrevLogTerm, reqLog.PrevLogIndex, reqLog.LeaderCommit)
		logger.Debug("curLog[baseIndex:%d,logLen:%d,lastLogTerm:%d,lastLogIdx:%d,commitIdex:%d]", baseIndex, len(r.log), r.getLastTerm(), r.getLastIndex(), r.commitIndex)
		if uint64(len(r.log)) > reqLog.PrevLogIndex+1-baseIndex {
			logger.Debugf("cur.logLen:%d, baseIndex:%d, reqLog.PrevLogIndex:%d", len(r.log), baseIndex, reqLog.PrevLogIndex)
			r.log = r.log[:reqLog.PrevLogIndex+1-baseIndex]
			logger.Infof("cut off logs, lastIndex:%d", r.getLastIndex())
		}
		newLog := reqLog.Entries[0]
		if newLog.Index == r.getLastIndex()+1 {
			r.log = append(r.log, logEntry{index: newLog.Index, term: newLog.Term, data: newLog.Command})
			logger.Infof("follower add log, term:%d, lastIndex:%d", newLog.Term, newLog.Index)
			//quit self
			typ, _ := miscellaneous.D32func(newLog.Command[:4])
			if typ == 2 {
				r.preQuitPeer(newLog.Command[4:])
			}
			respLog.Success = true
		} else {
			logger.Infof("follower does not add the log, newLog.Index:%d, need index:%d", newLog.Index, r.getLastIndex()+1)
		}
		respLog.LastLog = r.getLastIndex()
	}

	//commit
	if reqLog.LeaderCommit != r.commitIndex {
		logger.Infof("req.LeaderCommit:%d, cur.commitIndex:%d, cur.lastIndex:%d", reqLog.LeaderCommit, r.commitIndex, r.getLastIndex())
		commitindex := r.commitIndex
		for i := r.commitIndex + 1; i <= reqLog.LeaderCommit && i <= r.getLastIndex(); i++ {
			if i > baseIndex {
				logger.Infof("follower commit term:%d, commitIndex:%d", r.currentTerm, i)
				r.commitLog(0, i, r.log[i-baseIndex].data)
			}
			commitindex++
		}
		r.commitIndex = commitindex
	}

	//snaplog
	r.checkSnapshoot()
}

func (r *raft) doAppendLogResponse(msg *message) {
	peerid := msg.peerid
	p, ok := r.peers[peerid]
	if !ok {
		logger.Errorf("could not found peer id:%s", peerid)
		return
	}
	if r.leaderID != r.localID {
		if ld, ok := r.peers[r.leaderID]; ok {
			logger.Warningf("peer %s this does not leader, the new leader is %s", p.name, ld.name)
		}
		return
	}
	logResp := &pb.AppendEntriesResponse{}
	err := proto.Unmarshal(msg.data, logResp)
	if err != nil {
		logger.Errorf("proto.Unmarshal err %s", err)
		return
	}

	logger.Debugf("peer %s appendLogResponse[term:%d,succ:%v,lastLog:%d]",
		p.name, logResp.Term, logResp.Success, logResp.LastLog)
	if logResp.NoRetryBackoff {
		return
	}
	if logResp.Term < r.currentTerm {
		logger.Infof("peer:%s, resp.term: %d < cur.term: %d exit", p.name, logResp.Term, r.currentTerm)
		return
	}
	if logResp.Term > r.currentTerm {
		logger.Infof("peer:%s, resp.term[%d] > cur.term[%d], set new state:follower", p.name, logResp.Term, r.currentTerm)
		r.setState(follower)
		r.currentTerm = logResp.Term
		r.leaderID = ""
		//
		r.cleanPeer()

		r.breakCh <- 1
		return
	}
	if r.nextIndex[peerid] != logResp.LastLog+1 {
		r.nextIndex[peerid] = logResp.LastLog + 1
		logger.Debugf("peer:%s set nextIndex:%d", p.name, logResp.LastLog+1)
	}
	if logResp.Success {
		if r.matchIndex[peerid] != logResp.LastLog {
			r.matchIndex[peerid] = logResp.LastLog
			logger.Debugf("peer:%s set matchIndex:%d", p.name, logResp.LastLog)
		}
	}
}

func (r *raft) checkCommit() {
	last := r.getLastIndex()
	baseIndex := r.log[0].index
	commitIndex := r.commitIndex
	for i := r.commitIndex + 1; i <= last; i++ {
		n := uint64(1)
		for _, idx := range r.matchIndex {
			if idx >= i {
				n++
			}
		}
		if n > r.quorumSize() {
			commitIndex = i
		}
	}
	if commitIndex != r.commitIndex {
		for i := r.commitIndex + 1; i <= commitIndex; i++ {
			if i > baseIndex {
				logger.Infof("leader commit term:%d, commitIndex:%d", r.currentTerm, i)
				r.commitLog(0, i, r.log[i-baseIndex].data)
			}
		}
		r.commitIndex = commitIndex
	}
}

func (r *raft) checkSnapshoot() {
	if r.commitIndex <= r.log[0].index+snapshootLimit {
		return
	}

	//节点信息列表
	peerBytes, err := json.Marshal(r.Peers())
	if err != nil {
		logger.Errorf("json.Marshal(r.Peers()) err : %v", err)
		return
	}
	logger.Infof("snapshoot begin commitIndex:%d, logLen:%d, old [term:%d, index:%d]", r.commitIndex, len(r.log), r.log[0].term, r.log[0].index)
	bh := r.getBlockHeight(snapshootLimit)
	//生成快照
	snapIndex := r.log[snapshootLimit].index
	snapTerm := r.log[snapshootLimit].term
	logData := append(miscellaneous.E64func(bh), peerBytes...)
	r.log = r.log[snapshootLimit:]
	r.log[0] = logEntry{index: snapIndex, term: snapTerm, data: logData, height: bh}

	if snapIndex%snapshootLimit != 0 {
		logger.Errorf("err snapIndex:%d,snapIndex / napshootLimit != 0", snapIndex)
		//os.Exit(0)
		return
	}
	logger.Infof("snapshoot end logLen:%d, new [term:%d, index:%d]", len(r.log), snapTerm, snapIndex)

	//提交快照
	logger.Infof("begin commit snaplog, block height:%d", bh)
	commitData := miscellaneous.E32func(1)                        //类型 1
	commitData = append(commitData, miscellaneous.E64func(bh)...) //高度
	r.commitLog(1, 0, commitData)
	logger.Infof("end commit snaplog")
}

func parseMsgType(mt MsgType) string {
	switch mt {
	case vote:
		return "vote"
	case voteRequest:
		return "voteRequest"
	case voteResponse:
		return "voteResponse"
	case heartBeat:
		return "heartBeat"
	case appendLogRequest:
		return "appendLogRequest"
	case appendLogResponse:
		return "appendLogResponse"
	case leaderCheck:
		return "leaderCheck"
	case snapshootRequest:
		return "snapshootRequest"
	default:
		return string(mt)
	}
}

func (r *raft) quorumSize() uint64 {
	voters := 1
	for _, peer := range r.peers {
		if peer.model == 0 {
			voters++
		}
	}
	return uint64(voters / 2)
}

func (r *raft) cleanPeer() {
	for _, peer := range r.peers {
		peer.clean()
	}
}
