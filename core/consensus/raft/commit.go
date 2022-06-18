package raft

import (
	"sync"
	"time"
)

type commit struct {
	commitCh chan bool
	logs     []commitLog
	lock     sync.Mutex

	commitFun commitLogFunc
	block     interface{}
}

type commitLog struct {
	commitIndex uint64
	typ         uint32 //0-普通日志；1-快照；2-节点变更
	log         []byte
}

func newCommit(block interface{}, commitFun commitLogFunc) *commit {
	cmtlog := &commit{
		commitCh:  make(chan bool, 1024),
		block:     block,
		commitFun: commitFun,
	}
	return cmtlog
}

func (c *commit) commitProcess() {
	for {
		select {
		case <-c.commitCh:
			for len(c.commitCh) > 0 {
				<-c.commitCh
			}
			c.lock.Lock()
			j := len(c.logs)
			c.lock.Unlock()

			for i := 0; i < j; i++ {
				cmtlog := c.logs[i]
				t1 := time.Now()
				c.commitFun(c.block, cmtlog.log)
				t := time.Now().Sub(t1)
				switch cmtlog.typ {
				case 0:
					logger.Infof("commitFun take %v commitIndex:%d ", t, cmtlog.commitIndex)
				case 1:
					logger.Infof("commitFun snapshootlog take %v", t)
				case 2:
					logger.Infof("commitFun updatePeer take %v", t)
				}
				logger.Infof("---------------------------[%d/%d]-------------------------------", (i + 1), j)
			}

			c.lock.Lock()
			c.logs = c.logs[j:]
			length := len(c.logs)
			c.lock.Unlock()

			if length > 0 && len(c.commitCh) == 0 {
				c.commitCh <- true
			}
		}
	}
}

func (c *commit) execCommitLog(log commitLog) {
	c.lock.Lock()
	c.logs = append(c.logs, log)
	logger.Infof("add commit notice, waiting for commit:%d", len(c.logs))
	c.lock.Unlock()
	if len(c.commitCh) < 1024 {
		c.commitCh <- true
	}
}
