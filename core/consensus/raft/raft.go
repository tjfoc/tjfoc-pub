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
	"crypto/rand"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/viper"
	"github.com/tjfoc/tjfoc/core/blockchain"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/config"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	p "github.com/tjfoc/tjfoc/proto"
	pb "github.com/tjfoc/tjfoc/protos/raft"
)

var logger = flogging.MustGetLogger("raft")

var (
	randbetween int64 = 10 //选举随机种子
	// randbase       int64  = 1000 //选举最低间隔
	votetimeout    int64  = 1000 //选举超时间隔
	heartBeatout   int64  = 200  //心跳周期
	snapshootLimit uint64 = 100  //日志快照
	maxPeerChan    int    = 1024 //消息通道最大值
	maxMsgChan     int
)

const (
	raftTCP = iota
)

type raftState uint32

const (
	follower raftState = iota
	candidate
	leader
	suspended
)

//NewRaft . Create raft instance
func NewRaft(model int, block interface{}, blockchain blockchain.BlockChain, sp p.SP, commitFun commitLogFunc, peerConofig config.PeerConfig) *raft {
	raft := &raft{
		breakCh:     make(chan int),
		heartbeatCh: make(chan bool),
		joinMemCh:   make(chan bool),
		leaderCh:    make(chan bool),
		quitCh:      make(chan int),
		peers:       make(map[string]*peer),
		nextIndex:   make(map[string]uint64),
		matchIndex:  make(map[string]uint64),
		blockChain:  blockchain,
		state:       follower,
	}
	memberList := peerConofig.Members.Peers
	if model != Init && model != Start {
		logger.Infof("peer start unNormal model")
		raft.state = suspended
		memberList = []config.IdConfig{}
		members, err := blockchain.GetMemberList()
		if err != nil {
			return nil
		}
		for k, v := range members {
			memberList = append(memberList, config.IdConfig{
				Id:   k,
				Addr: v.Addr,
				Typ:  int(v.Typ),
			})
		}
	}

	raft.commit = newCommit(block, commitFun)
	raft.log = append(raft.log, logEntry{index: 0, term: 0})
	sp.RegisterFunc(0, raftTCP, raft.callbackFunc)
	localid, _ := miscellaneous.GenHash(md5.New(), []byte(peerConofig.Self.Id))
	raft.localID = string(localid)
	raft.localPlainID = peerConofig.Self.Id
	raft.localAddr = peerConofig.Self.Addr
	raft.localModel = peerConofig.Self.Typ
	raft.sp = sp

	logger.Infof("localAddr:%s, localModel:%d ", raft.localAddr, raft.localModel)
	maxMsgChan = maxPeerChan * len(peerConofig.Members.Peers)
	for _, v := range memberList {
		if strings.Compare(v.Id, peerConofig.Self.Id) != 0 {
			hashid, _ := miscellaneous.GenHash(md5.New(), []byte(v.Id))
			peerid := string(hashid)
			raft.peers[peerid] = &peer{
				name:    "p_" + v.Id,
				id:      v.Id,
				pid:     peerid,
				hashid:  hashid,
				sp:      sp,
				address: v.Addr,
				model:   v.Typ,
				//peerCh:   make(chan *PeerMessage, maxPeerChan),
				exitCh:   make(chan bool),
				noticeCh: make(chan bool, maxPeerChan),
				raft:     raft,
			}
		}
	}
	raft.msgCh = make(chan *message, maxMsgChan)
	// raft.msgCh = make(chan *message, 1024)
	return raft
}

//日志条目
type logEntry struct {
	index  uint64 //索引
	term   uint64 //任期
	data   []byte //内容,如果为快照，值为节点信息
	height uint64 //高度，快照使用
}

type raft struct {
	state        raftState
	currentTerm  uint64     //服务器最后一次知道的任期号（初始化为 0，持续递增）
	votedFor     string     //在当前获得选票的候选人的 Id
	log          []logEntry //日志条目集；每一个条目包含一个用户状态机执行的指令，和收到时的任期号
	grantedVotes uint64     //最近一次发起选举得票数

	//所有服务器上经常变的
	commitIndex uint64 //区块ID，已知的最大的已经被提交的日志条目的索引值

	//在领导人里经常改变的 （选举后重新初始化）
	nextIndex  map[string]uint64 //对于每一个服务器，需要发送给他的下一个日志条目的索引值（初始化为领导人最后索引值加一）
	matchIndex map[string]uint64 //对于每一个服务器，已经复制给他的日志的最高索引值

	//lastLogTerm  uint64 //最后日志条目的任期号
	// lastLogIndex uint64 //最后日志条目的索引值

	leaderID     string //leader的ID
	localID      string //当前服务器的ID hash
	localPlainID string //当前服务器id，明文
	localAddr    string //当前服务器地址
	localModel   int    //当前服务器启动类型：0-核心节点参与共识，1-非核心节点不参与共识，只同步数据

	sp         p.SP
	blockChain blockchain.BlockChain
	commit     *commit //提交日志

	heartbeatCh chan bool        //接收心跳
	breakCh     chan int         //退出通道
	leaderCh    chan bool        //leader
	joinMemCh   chan bool        //加入共识节点
	quitCh      chan int         //节点退出
	msgCh       chan *message    //消息通道
	peers       map[string]*peer //节点列表
}

func (r *raft) initConfig() {
	if hb := viper.GetInt64("Raft.HeartBeat"); hb > 0 {
		heartBeatout = hb
	}
	if vt := viper.GetInt64("Raft.VoteTimeout"); vt > 0 {
		votetimeout = vt
		randbetween = votetimeout / 100
		if randbetween == 0 {
			randbetween = 3
		}
	}
	logger.Infof("heartBeatout:%d,votetimeout:%d,randbetween:%d,snapshootLimit:%d ", heartBeatout, votetimeout, randbetween, snapshootLimit)
}

func (r *raft) run() {
	logger.Infof("raft startup")
	r.initConfig()
	go r.messageProcess()
	go r.commit.commitProcess()

	logger.Infof("===== peers list =====")
	for _, peer := range r.peers {
		logger.Infof("run peerProcess %s id:%x", peer.name, peer.hashid)
		go peer.peerProcess()
	}
	for {
		switch r.getState() {
		case suspended:
			r.runSuspended()
		case follower:
			r.runFollower()
		case candidate:
			r.runCandidate()
		case leader:
			r.runLeader()
		}
	}
}

func (r *raft) runSuspended() {
	logger.Infof("runSuspended")
	for {
		select {
		case <-r.heartbeatCh:
			logger.Debugf("rev heart beat")
		case <-r.breakCh:
		case <-r.quitCh:
		case <-r.joinMemCh:
			logger.Infof("peer join raft membership, exit suspended")
			return
		}
	}
}

func (r *raft) createRand(base int64) int64 {
	//随机数
	for {
		buf := make([]byte, 8)
		io.ReadFull(rand.Reader, buf)
		v, _ := miscellaneous.D64func(buf)
		rand := int64(v) % base
		if rand <= base && rand > 0 {
			return rand
		}
	}
}

func (r *raft) runFollower() {
	logger.Infof("runFollower term %d", r.currentTerm)
	for {
		tout := r.createRand(votetimeout) + votetimeout
		select {
		case <-r.heartbeatCh:
			logger.Debugf("rev heart beat")
		case <-r.breakCh:
			logger.Infof("breaked, exit follower")
			return
		case <-r.quitCh:
			logger.Infof("quit raft membership, exit follower")
			return
		case <-time.After(time.Duration(tout) * time.Millisecond):
			logger.Infof("time out [%dms], exit follower", tout)
			r.setState(candidate)
			return
		}
	}
}

func (r *raft) runCandidate() {
	logger.Infof("run candidate")
	bytes, err := r.newVoteReq()
	if err != nil {
		logger.Infof("newVoteReq err %s", err)
		return
	}
	logger.Infof("vote self term:%d", r.currentTerm)
	for _, peer := range r.peers {
		peer.clean()
		r.notifyMessage(peer, vote, bytes)
	}
	for {
		select {
		case <-r.leaderCh:
			return
		case <-r.quitCh:
			logger.Infof("quit raft membership")
			return
		case <-r.breakCh:
			logger.Infof("breaked, exit candidate")
			return
		case <-r.heartbeatCh:
			logger.Infof("ignore heartbeat")
		case <-time.After(time.Duration(votetimeout) * time.Millisecond):
			logger.Infof("timeout [%dms], exit candidate", votetimeout)
			return
		}
	}
}

func (r *raft) runLeader() {
	logger.Infof("runLeader term %d", r.currentTerm)
	//更新 follow的高度
	if r.getLastIndex() == 0 {
		bh := r.getBlockHeight(0)
		if bh > 1 {
			logdata := append(miscellaneous.E32func(3), miscellaneous.E64func(bh-1)...)
			r.log = append(r.log, logEntry{index: r.getLastIndex() + 1, term: r.currentTerm, data: logdata})
			logger.Infof("leader append update threshold log term:%d, lastIndex:%d,threshold:%d", r.currentTerm, r.getLastIndex(), (bh - 1))
		}
	}
	// r.heartBeat()
	r.notifyMessage(nil, heartBeat, nil)
	for {
		select {
		case <-r.heartbeatCh:
			logger.Infof("ignore heartbeat")
		case <-r.breakCh:
			logger.Infof("breaked, exit leader, term %d", r.currentTerm)
			return
		case <-r.quitCh:
			logger.Infof("quit raft membership, exit leader")
			return
		case <-time.After(time.Duration(heartBeatout) * time.Millisecond):
			logger.Debugf("-----------------hbeat--------------------------")
			logger.Debugf("heartBeat timeout [%dms] begin", heartBeatout)
			r.notifyMessage(nil, leaderCheck, nil)
			r.notifyMessage(nil, heartBeat, nil)
			logger.Debugf("heartBeat timeout end")
		}
	}
}

//Receive tcp message
func (r *raft) callbackFunc(ifc interface{}, serverID []byte, data []byte) int {
	if p, ok := r.peers[string(serverID)]; ok {
		if peerMsg, err := show(data); err == nil {
			logger.Debugf("----------------cback---------------------------")
			logger.Debugf("callbackFunc %s %s ", p.name, parseMsgType(peerMsg.MSGType))
			//当前节点状态若为不可用，只接收附加日志跟快照请求
			if r.getState() == suspended {
				if peerMsg.MSGType != appendLogRequest && peerMsg.MSGType != snapshootRequest {
					logger.Debugf("ignore message.cur peer state is suspended,but the msgType is %d,only receiv appendLogRequest(%d) or snapshootRequest(%d)",
						peerMsg.MSGType, appendLogRequest, snapshootRequest)
					return 0
				}
			}
			r.notifyMessage(p, peerMsg.MSGType, peerMsg.Data)
		}
	}
	return 0
}

//pre quitPeer ,only follow quit self
func (r *raft) preQuitPeer(log []byte) {
	pi := make(map[string]string)
	err := json.Unmarshal(log, &pi)
	if err != nil {
		logger.Errorf("json.Unmarshal log err:%v", err)
		return
	}
	id := pi["ID"]
	addr := pi["Addr"]
	typStr := pi["Typ"]
	if id == "" || addr == "" || typStr == "" {
		logger.Errorf("get map para err")
		return
	}
	typ, err := strconv.Atoi(typStr)
	if err != nil {
		logger.Errorf("strconv.Atoi(typStr) err:%v", err)
		return
	}
	hashid, _ := miscellaneous.GenHash(md5.New(), []byte(id))
	peerid := string(hashid)
	if typ == PeerQuit && peerid == r.localID && r.getState() == follower {
		logger.Infof("pre peerQuit self")
		r.setState(suspended)
		r.quitCh <- 1
	}

}

func (r *raft) updatePeer(log []byte) {
	logger.Infof("in updatePeer")
	defer logger.Infof("exit updatePeer")
	pi := make(map[string]string)
	err := json.Unmarshal(log, &pi)
	if err != nil {
		logger.Errorf("json.Unmarshal log err:%v", err)
		return
	}
	id := pi["ID"]
	addr := pi["Addr"]
	typStr := pi["Typ"]
	logger.Infof("change peer typ:%s, id:%s, addr:%s", typStr, id, addr)
	if id == "" || addr == "" || typStr == "" {
		logger.Errorf("get map para err")
		return
	}
	typ, err := strconv.Atoi(typStr)
	if err != nil {
		logger.Errorf("strconv.Atoi(typStr) err:%v", err)
		return
	}
	hashid, _ := miscellaneous.GenHash(md5.New(), []byte(id))
	peerid := string(hashid)
	//节点退出 OR 节点加入
	if typ == PeerQuit {
		if peer, ok := r.peers[peerid]; ok {
			logger.Infof("peer [id:%s, addr:%s] remove", id, addr)
			peer.exitCh <- true
			delete(r.peers, peer.pid)
		} else if peerid == r.localID {
			logger.Infof("peerQuit self")
			r.setState(suspended)
			r.quitCh <- 1
		} else {
			logger.Warningf("not found peer in raft membership")
		}
	} else {
		//新加节点
		if r.localID != peerid {
			if _, ok := r.peers[peerid]; !ok {
				newPeer := &peer{
					name:    "p_" + id,
					pid:     peerid,
					id:      id,
					hashid:  hashid,
					sp:      r.sp,
					address: addr,
					//peerCh:  make(chan *PeerMessage, maxPeerChan),
					noticeCh: make(chan bool, maxPeerChan),
					exitCh:   make(chan bool),
					model:    typ,
					raft:     r,
				}
				r.peers[peerid] = newPeer
				r.matchIndex[newPeer.pid] = 0
				r.nextIndex[newPeer.pid] = r.log[0].index + 1
				go newPeer.peerProcess()
				logger.Infof("join new peer[id:%s, addr:%s] and start a new peerProcess", id, addr)
			}
			return
		}
		//新节点提交自己
		logger.Infof("peer commit self, id:%s, addr:%s", id, addr)

		//共识节点
		if typ == PeerJoin && r.getState() == suspended {
			logger.Infof("cur peer join membership, update state suspended -> follower")
			r.setState(follower)
			r.localModel = 0
			r.joinMemCh <- true
		}
	}
}

func (r *raft) commitLog(typ uint32, commitIndex uint64, log []byte) {
	//typ, _ := miscellaneous.D32func(log[:4])
	//节点管理日志
	if typ == 2 {
		r.updatePeer(log[4:])
		//return
	}
	cmtlog := commitLog{
		commitIndex: commitIndex,
		typ:         typ,
		log:         log,
	}
	r.commit.execCommitLog(cmtlog)
}

func (r *raft) newVoteReq() ([]byte, error) {
	logger.Infof("begin create voteReq")
	// ht := r.getBlockHeight(r.commitIndex - r.log[0].index)
	ht := r.getBlockHeight(0)
	r.currentTerm++    //当前任期自增
	r.grantedVotes = 1 //给自己投票
	//选举投票请求
	req := &pb.RequestVoteRequest{
		Term:         r.currentTerm,
		CandidateId:  r.localID,
		LastLogIndex: r.getLastIndex(),
		LastLogTerm:  r.getLastTerm(),
		BlockHeight:  ht,
	}
	logger.Infof("end create voteReq ht:%d", ht)
	return proto.Marshal(req)
}

func (r *raft) getState() raftState {
	stateAddr := (*uint32)(&r.state)
	return raftState(atomic.LoadUint32(stateAddr))
}

func (r *raft) setState(s raftState) {
	stateAddr := (*uint32)(&r.state)
	atomic.StoreUint32(stateAddr, uint32(s))
}

func (r *raft) getLastIndex() uint64 {
	return r.log[len(r.log)-1].index
}

func (r *raft) getLastTerm() uint64 {
	return r.log[len(r.log)-1].term
}

func (r *raft) parseRaftState(state raftState) string {
	switch state {
	case suspended:
		return "suspended"
	case follower:
		return "follower"
	case candidate:
		return "candidate"
	case leader:
		return "leader"
	default:
		return string(state)
	}
}
