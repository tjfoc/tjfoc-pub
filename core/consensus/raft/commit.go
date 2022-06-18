package raft

import (
	"sync"
	"time"
)

type commitLog struct {
	commitCh chan bool
	logs     [][]byte
	lock     sync.Mutex

	commitFun commitLogFunc
	block     interface{}
}

func newCommitLog(block interface{}, commitFun commitLogFunc) *commitLog {
	cmtlog := &commitLog{
		commitCh:  make(chan bool, 1024),
		block:     block,
		commitFun: commitFun,
	}
	return cmtlog
}

func (c *commitLog) commitProcess() {
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
				log := c.logs[i]
				t1 := time.Now()
				c.commitFun(c.block, log)
				t := time.Now().Sub(t1)
				if t.Seconds() > 1 {
					logger.Warningf("exec commitFun take long time %v", t)
				}
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

func (c *commitLog) execCommitLog(log []byte) {
	c.lock.Lock()
	c.logs = append(c.logs, log)
	c.lock.Unlock()
	if len(c.commitCh) < 1024 {
		c.commitCh <- true
	}
}
