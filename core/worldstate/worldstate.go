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
package worldstate

//全局说明
//数据库存储方式：
//				保存每个bucket对应的hash:						"0":hash "1":hash "2":hash ...... "(1<<treetall - 1)":hash
//				保存每个bucket中所有的key（按照sort进行排列）:	"0_keys":[]byte([]string) "1_keys":[]byte([]string) ...... "(1<<treetall - 1)_keys":[]byte([]string)
//				保存每个需要存储的数据:							key:value
//				保存该节点自身同步到多少个block:				"currentBIndex":0 "currentBIndex":1 ...... "currentBIndex":index_for_block
import (
	"crypto/sha256"
	//"fmt"
	"errors"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/worldstate/buckettree"
	tjProto "github.com/tjfoc/tjfoc/proto"
)

type transportMes struct {
	id   string
	data string
}

//type checkMes struct {
//bIndex uint64
//bHash  string
//}
type pushMes struct {
	bIndex  uint64
	bHash   string
	bData   WorldStateData
	bAction map[string]int32
}

type WorldState struct {
	bTree         buckettree.BucketTree //buckettree
	db            *leveldb.DB           //数据库
	batch         *leveldb.Batch        //批量写入数据
	sp            tjProto.SP            //tcp连接
	currentBIndex uint64                //当前块高度
	initstate     bool                  //是否是初始化状态，初始化状态不能给别人同步数据
	datamis       bool                  //是否是数据缺失状态
	inCh          chan *transportMes    //同步数据
	pushCh        chan *pushMes         //blockchain发送的要求真正插入数据库
	updateCh      chan string           //block发送的要求与某个节点进行数据同步
	//checkCh       chan *checkMes        //chaincode发送的检查是否能将结果插入数据库的函数

	//clker      sync.Mutex
	//isChecking bool
	//checkRetCh map[uint64]chan bool
	healthCh chan bool

	plker     sync.Mutex
	isPushing bool
	pushRetCh chan int

	ulker       sync.Mutex
	isUpdating  bool
	updateRetCh chan int
}

var wdLogger = flogging.MustGetLogger("worldstate")

func (w *WorldState) initwd(treeTall uint64, path string, sp tjProto.SP) {
	wdLogger.Info("wd:Init WS!")
	w.bTree.Init(treeTall)
	var err error
	w.db, err = leveldb.OpenFile(path, nil)
	if err != nil {
		wdLogger.Error("Open leveldb failed!")
		os.Exit(1)
	}
	w.batch = new(leveldb.Batch)
	w.sp = sp
	w.currentBIndex = 0
	w.initstate = true
	w.datamis = false

	w.healthCh = make(chan bool, 5)
	w.inCh = make(chan *transportMes, 1)
	//w.checkCh = make(chan *checkMes, 1)
	w.pushCh = make(chan *pushMes, 1)
	w.updateCh = make(chan string, 1)

	//w.checkRetCh = make(map[uint64]chan bool, 0)
	w.pushRetCh = make(chan int, 1)
	w.updateRetCh = make(chan int, 1)

	isNew := true
	temp := make([]string, 0)
	for i := uint64(0); i < uint64(1<<treeTall); i++ {
		value, err := w.db.Get([]byte(strconv.FormatUint(i, 10)), nil)
		if err != nil {
			if err != leveldb.ErrNotFound {
				wdLogger.Error("wd:Read bucket hash failed!Read error!")
				os.Exit(1)
			} else {
				if !isNew {
					wdLogger.Error("wd:Read bucket hash failed!Data miss!")
					encoder := sha256.New()
					encoder.Write([]byte(""))
					value = encoder.Sum(nil)
					w.batch.Put([]byte(strconv.FormatUint(i, 10)), value)
					w.datamis = true
				} else {
					wdLogger.Info("wd:New WS!")
					break
				}
			}
		}
		isNew = false
		temp = append(temp, string(value))
	}
	if !isNew {
		if w.datamis {
			w.db.Write(w.batch, nil)
			w.batch.Reset()
		}
		w.bTree.Rebuild(temp)
		vB, _ := w.db.Get([]byte("currentBIndex"), nil)
		w.currentBIndex, _ = strconv.ParseUint(string(vB), 10, 64)
	} else {
		encoder := sha256.New()
		encoder.Write([]byte(""))
		value := encoder.Sum(nil)
		for i := uint64(0); i < uint64(1<<treeTall); i++ {
			w.batch.Put([]byte(strconv.FormatUint(i, 10)), value)
		}
		w.batch.Put([]byte("currentBIndex"), []byte(strconv.FormatUint(w.currentBIndex, 10)))
		w.db.Write(w.batch, nil)
		w.batch.Reset()
	}
	if w.sp != nil {
		w.sp.RegisterFunc(666, 666, w.recvFunc)
	} else {
		wdLogger.Error("wd:Network is nil!")
		os.Exit(1)
	}
	if !isNew {
		w.checkcomplete(treeTall)
	}
	go w.waitmes()
}
func (w *WorldState) checkcomplete(treeTall uint64) {
	wdLogger.Info("wd:Check WS!")
	for i := uint64(0); i < uint64(1<<treeTall); i++ {
		w.checkbucketcomplete(i)
	}
}
func (w *WorldState) checkbucketcomplete(bucketIndex uint64) bool {
	bucketHash, _ := w.db.Get([]byte(strconv.FormatUint(bucketIndex, 10)), nil)
	keysB, _ := w.db.Get([]byte(strconv.FormatUint(bucketIndex, 10)+"_keys"), nil)
	tempKey := &Keys{}
	proto.Unmarshal(keysB, tempKey)
	lalala := ""
	for _, key := range tempKey.Key {
		valueB, _ := w.db.Get([]byte(key), nil)
		lalala += key
		lalala += string(valueB)
	}
	encoder := sha256.New()
	encoder.Write([]byte(lalala))
	newHash := encoder.Sum(nil)
	if string(bucketHash) == string(newHash) {
		return true
	} else {
		wdLogger.Infof("wd:data miss in bucket:%d\n", bucketIndex)
		w.datamis = true
		w.db.Put([]byte(strconv.FormatUint(bucketIndex, 10)), newHash, nil)
		w.bTree.UpdateSingle(bucketIndex, string(newHash))
		return false
	}
}
func (w *WorldState) waitmes() {
	for {
		select {
		case v := <-w.inCh:
			mes := &WDMes{}
			proto.Unmarshal([]byte(v.data), mes)
			switch mes.Type {
			case MesType_CM:
				w.recvCompare(mes.GetCompare(), v.id)
			case MesType_RM:
				w.recvPull(mes.GetRequire(), v.id)
			case MesType_SM:
				w.recvPush(mes.GetSync(), v.id)
			case MesType_CF:
				w.recvCompareFailed(v.id)
			case MesType_CS:
				w.recvCompareSuccess(v.id)
			}
		//case v := <-w.checkCh:
		//w.compareBlockStatehash(v.bIndex, v.bHash)
		case v := <-w.pushCh:
			w.pushDB(v.bIndex, v.bHash, v.bData, v.bAction)
		case peerid := <-w.updateCh:
			w.requireCompare(peerid)
		}
	}
}

/*
func (w *WorldState) Check(index uint64, compareRoothash string, wd WorldStateData) (status bool) {
	w.plker.Lock()
	if w.isPushing {
		w.plker.Unlock()
		wdLogger.Error("Cant do check when system is pushing block!")
		return false
	}
	w.plker.Lock()
	w.ulker.Lock()
	if w.isUpdating {
		w.ulker.Unlock()
		wdLogger.Error("Cant do check when system is updating data!")
		return false
	}
	w.ulker.Unlock()
	var tempCh chan bool
	w.clker.Lock()
	if w.isChecking {
		w.clker.Unlock()
		wdLogger.Error("Cant check data when system is checking block!")
		return false
	}
	if _, ok := w.checkRetCh[index]; !ok {
		w.checkRetCh[index] = make(chan bool, 1)
		tempCh = w.checkRetCh[index]
		w.checkCh <- &checkMes{
			bIndex: index,
			bHash:  compareRoothash,
		}
		w.isChecking = true
		w.clker.Unlock()
	} else {
		w.clker.Unlock()
		return false
	}
	status = <-tempCh
	close(tempCh)
	w.clker.Lock()
	delete(w.checkRetCh, index)
	w.isChecking = false
	w.clker.Unlock()
	return
}
*/

func (w *WorldState) compareBlockStatehash(bIndex uint64, leaderRoothash string) int {
	wdLogger.Info("wd:Check Push!index:", bIndex)
	if w.datamis == true {
		//w.checkRetCh[bIndex] <- false
		return 1
	}
	if bIndex == w.currentBIndex {
		if w.bTree.GetRootHash() == leaderRoothash {
			wdLogger.Info("wd:check success!")
			//w.checkRetCh[bIndex] <- true
			return 0
		} else {
			wdLogger.Error("wd:same index but different hash!bad node!bad node!")
			w.datamis = true
			//w.checkRetCh[bIndex] <- false
			return 4
		}
	} else {
		if bIndex > w.currentBIndex {
			wdLogger.Info("wd:worldstate is lag!")
			w.datamis = true
			//w.checkRetCh[bIndex] <- false
			return 3
		} else {
			wdLogger.Info("wd:worldstate is newer then the push data")
			w.datamis = true
			//w.checkRetCh[bIndex] <- false
			return 2
		}
	}
}

func (w *WorldState) Push(index uint64, hash string, kv map[string]string, action map[string]int32) (status int) {
	/*
		w.clker.Lock()
		if w.isChecking {
			w.clker.Unlock()
			wdLogger.Error("Cant push data when system is checking block!")
		}
		w.clker.Unlock()
	*/
	w.ulker.Lock()
	if w.isUpdating {
		w.ulker.Unlock()
		wdLogger.Error("Cant push data when system is updating data!")
		return 0
	}
	w.ulker.Unlock()

	w.plker.Lock()
	if w.isPushing {
		w.plker.Unlock()
		wdLogger.Error("Cant push data when system is pushing block!")
		return 0
	}
	w.isPushing = true
	w.pushCh <- &pushMes{
		bIndex:  index,
		bHash:   hash,
		bData:   kv,
		bAction: action,
	}
	w.plker.Unlock()

	status = <-w.pushRetCh

	w.plker.Lock()
	w.isPushing = false
	w.plker.Unlock()
	w.healthCh <- true
	return
}
func (w *WorldState) pushDB(bIndex uint64, bHash string, wd WorldStateData, action map[string]int32) {
	wdLogger.Info("wd:start push!Index:", bIndex)
	r := w.compareBlockStatehash(bIndex, bHash)
	if r != 0 {
		w.pushRetCh <- r
		return
	}
	if w.isPushing {
		if len(wd) == 0 {
			w.currentBIndex++
			w.batch.Put([]byte("currentBIndex"), []byte(strconv.FormatUint(w.currentBIndex, 10)))
			err := w.db.Write(w.batch, nil)
			w.batch.Reset()
			if err != nil {
				wdLogger.Error("wd:push failed!")
				w.pushRetCh <- 5
			} else {
				wdLogger.Info("wd:push success!")
				w.initstate = false
				w.datamis = false
				w.pushRetCh <- 0
			}
			return
		}
		bIndexM := make(map[uint64]WorldStateData, 0)
		for k, v := range wd {
			tempIndex := w.bkdrhash(k)
			v1, ok := bIndexM[tempIndex]
			if !ok {
				v1 = make(WorldStateData, 0)
				bIndexM[tempIndex] = v1
			}
			v1[k] = v
			bIndexM[tempIndex] = v1
		}
		bH := make(map[uint64]string)
		for i, v := range bIndexM {
			KeysB, _ := w.db.Get([]byte(strconv.FormatUint(i, 10)+"_keys"), nil)
			tempKeys := &Keys{}
			proto.Unmarshal(KeysB, tempKeys)
			tempKv := make(map[string]string, 0)
			for _, key := range tempKeys.Key {
				valueB, _ := w.db.Get([]byte(key), nil)
				tempKv[key] = string(valueB)
			}
			for k1, v1 := range v {
				act, _ := action[k1]
				switch act {
				case DELACTION:
					//删除
					w.batch.Delete([]byte(k1))
					for i, ii := range tempKeys.Key {
						if ii == k1 {
							tempKeys.Key = append(tempKeys.Key[:i], tempKeys.Key[i+1:]...)
							break
						}
					}
					delete(tempKv, k1)
				case PUTACTION:
					//修改或增加
					w.batch.Put([]byte(k1), []byte(v1))
					_, ok := tempKv[k1]
					if !ok {
						tempKeys.Key = append(tempKeys.Key, k1)
					}
					tempKv[k1] = v1
				}
			}
			sort.Strings(tempKeys.Key)
			lalala := ""
			for _, key := range tempKeys.Key {
				lalala += key
				lalala += tempKv[key]
			}
			encoder := sha256.New()
			encoder.Write([]byte(lalala))
			newHash := encoder.Sum(nil)
			bH[i] = string(newHash)
			w.batch.Put([]byte(strconv.FormatUint(i, 10)), newHash)
			KeysB, _ = proto.Marshal(tempKeys)
			w.batch.Put([]byte(strconv.FormatUint(i, 10)+"_keys"), KeysB)
		}
		w.currentBIndex++
		w.batch.Put([]byte("currentBIndex"), []byte(strconv.FormatUint(w.currentBIndex, 10)))
		err := w.db.Write(w.batch, nil)
		w.batch.Reset()
		if err != nil {
			wdLogger.Info("wd:push failed!")
			w.pushRetCh <- 5
		} else {
			wdLogger.Info("wd:push success!")
			w.initstate = false
			w.datamis = false
			for i, v := range bH {
				w.bTree.UpdateSingle(i, v)
			}
			w.pushRetCh <- 0
		}
	}
	return
}
func (w *WorldState) Update(peerid string) int {
	/*
		w.clker.Lock()
		if w.isChecking {
			w.clker.Unlock()
			wdLogger.Error("Cant update data when system is checking block!")
		}
		w.clker.Unlock()
	*/
	w.plker.Lock()
	if w.isPushing {
		w.plker.Unlock()
		wdLogger.Error("Cant update data when system is pushing data!")
	}
	w.plker.Lock()

	w.ulker.Lock()
	if w.isUpdating {
		w.ulker.Unlock()
		wdLogger.Error("Cant update data when system is updating data!")
		return 0
	}
	w.isUpdating = true
	w.updateCh <- peerid
	w.ulker.Unlock()

	tker := time.NewTicker(10 * time.Second)
	select {
	case <-tker.C:
		w.ulker.Lock()
		w.isUpdating = false
		w.ulker.Unlock()
		return 3
	case v := <-w.updateRetCh:
		w.ulker.Lock()
		w.isUpdating = false
		w.ulker.Unlock()
		return v
	}
}

func (w *WorldState) updateDB(bIndex uint64, wdM map[uint64]WorldStateData, bHM map[uint64]string, keysB map[uint64]string) {
	w.ulker.Lock()
	if w.isUpdating {
		w.ulker.Unlock()
		wdLogger.Info("wd:start update!")
		if bIndex > w.currentBIndex || (bIndex == w.currentBIndex && w.datamis) {
			for i, v := range wdM {

				old, _ := w.db.Get([]byte(strconv.FormatUint(i, 10)+"_keys"), nil)
				oldKeys := &Keys{}
				proto.Unmarshal(old, oldKeys)
				for _, v := range oldKeys.Key {
					w.batch.Delete([]byte(v))
				}

				for k1, v1 := range v {
					w.batch.Put([]byte(k1), []byte(v1))
				}
				w.batch.Put([]byte(strconv.FormatUint(i, 10)+"_keys"), []byte(keysB[i]))
				w.batch.Put([]byte(strconv.FormatUint(i, 10)), []byte(bHM[i]))
			}
			w.currentBIndex = bIndex
			w.batch.Put([]byte("currentBIndex"), []byte(strconv.FormatUint(w.currentBIndex, 10)))
			err := w.db.Write(w.batch, nil)
			w.batch.Reset()
			if err != nil {
				wdLogger.Info("wd:update failed!")
				w.updateRetCh <- 4
			} else {
				wdLogger.Info("wd:update success!")
				w.datamis = false
				w.initstate = false
				for i, v := range bHM {
					w.bTree.UpdateSingle(i, v)
				}
				w.updateRetCh <- 1
			}
		} else {
			wdLogger.Error("wd:update failed!")
			w.updateRetCh <- 0
		}
	} else {
		w.ulker.Unlock()
	}
}

func (w *WorldState) GetHealthCh() chan bool {
	return w.healthCh
}

func (w *WorldState) makeCompareMes() *CompareMes {
	mes := &CompareMes{}
	mes.Isinit = w.initstate
	mes.Isdatamis = w.datamis
	mes.Roothash = w.bTree.GetRootHash()
	mes.Currentindex = w.currentBIndex
	mes.Buckethash = w.bTree.Export()
	return mes
}

//发送比较请求
func (w *WorldState) requireCompare(id string) {
	wdLogger.Info("wd:require compare!")
	tempMes := w.makeCompareMes()
	mes := &WDMes{}
	mes.Type = MesType_CM
	mes.Realmes = &WDMes_Compare{tempMes}
	data, _ := proto.Marshal(mes)
	e := w.sp.SendInstruction(666, 666, data, []byte(id))
	if e != nil {
		wdLogger.Errorf("wd:Send RequireMes error:%s\n", e)
		w.updateRetCh <- 0
	}
}

//收到比较请求
func (w *WorldState) recvCompare(mes *CompareMes, id string) {
	wdLogger.Info("wd:recv compare mes!")
	if mes.Currentindex > w.currentBIndex {
		//对端index比本地大
		if mes.Isdatamis || mes.Isinit {
			//对端有数据缺失或者对端处于初始化状态，直接忽略
			w.sendCompareFailed(id)
			return
		}
		//对端没有数据缺失，也不是初始化状态，那么请求数据覆盖本地
		//differentBucketIndex := w.bTree.SearchDifferent(mes.Buckethash)
		//w.sendPull(differentBucketIndex, id)
	} else if mes.Currentindex < w.currentBIndex {
		//对端index比本地小
		if w.datamis || w.initstate {
			//如果本地有数据缺失，或者本地是初始化状态，忽略
			w.sendCompareFailed(id)
			return
		}
		//本地没有数据缺失，又不是初始化状态，那么推送覆盖对端
		differentBucketIndex := w.bTree.SearchDifferent(mes.Buckethash)
		w.sendPush(differentBucketIndex, id)
	} else {
		if w.bTree.CompareRoot(mes.Roothash) {
			//如果两个节点index一样，hash又一样，那么两者数据库是一样的
			w.sendCompareSuccess(id)
			return
		}
		//一下都是hash不一样的情况
		if mes.Isdatamis && w.datamis {
			//两个比较节点都有数据缺失，那么双方都不能给对方同步数据
			w.sendCompareFailed(id)
			return
		} else if !mes.Isdatamis && !w.datamis {
			//两个节点都没有数据缺失，index又一样，但是hash却不一样，其中肯定有人出错，有人恶意修改数据库
			wdLogger.Error("wd:same index but different hash.bad node!bad node!")
			w.sendCompareFailed(id)
			return
		}
		differentBucketIndex := w.bTree.SearchDifferent(mes.Buckethash)
		if mes.Isdatamis {
			if w.initstate {
				//本地是初始化状态，不能向别人同步数据
				w.sendCompareFailed(id)
			} else {
				//对端有数据缺失，而本地没有数据缺失，index一样，本地也不是初始化状态，那么给对端推送，覆盖对端
				w.sendPush(differentBucketIndex, id)
			}
		} else if w.datamis {
			w.sendCompareFailed(id)
			//if mes.Isinit {
			//对端是初始化状态，不能向它请求同步数据
			//w.sendCompareFailed(id)
			//} else {
			//如果本地有数据缺失，而对端没有数据缺失，index一样，对端不是初始化状态，那么向对端请求，覆盖本地
			//w.sendPull(differentBucketIndex, id)
			//}
		}
	}
}

func (w *WorldState) makeSyncDataMes(bucketIndex []uint64) *SyncDataMes {
	mes := &SyncDataMes{}
	mes.Bhash = make(map[uint64]string, 0)
	mes.Bkeys = make(map[uint64]string, 0)
	mes.Bwd = make(map[uint64]*Datas, 0)
	mes.Currentindex = w.currentBIndex
	for _, v := range bucketIndex {
		mes.Bhash[v] = w.bTree.GetIndexHash(v)
		keysB, _ := w.db.Get([]byte(strconv.FormatUint(v, 10)+"_keys"), nil)
		tempKey := &Keys{}
		proto.Unmarshal(keysB, tempKey)
		mes.Bkeys[v] = string(keysB)
		tempData := &Datas{}
		tempData.Data = make(map[string]string, 0)
		for _, k := range tempKey.Key {
			vB, _ := w.db.Get([]byte(k), nil)
			tempData.Data[k] = string(vB)
		}
		mes.Bwd[v] = tempData
	}
	return mes
}

//给对方同步数据
func (w *WorldState) sendPush(bucketIndex []uint64, id string) {
	wdLogger.Info("wd:sendPush")
	tempMes := w.makeSyncDataMes(bucketIndex)
	mes := &WDMes{}
	mes.Type = MesType_SM
	mes.Realmes = &WDMes_Sync{tempMes}
	data, _ := proto.Marshal(mes)
	e := w.sp.SendInstruction(666, 666, data, []byte(id))
	if e != nil {
		wdLogger.Errorf("wd:Send PushMes error:%s\n", e)
	}
}

//对方给本地同步数据
func (w *WorldState) recvPush(mes *SyncDataMes, id string) {
	wdLogger.Info("wd:recvPush")
	wdM := make(map[uint64]WorldStateData, 0)
	for i, v := range mes.Bwd {
		wdM[i] = v.Data
	}
	w.updateDB(mes.Currentindex, wdM, mes.Bhash, mes.Bkeys)
}

func (w *WorldState) makeRequireMes(bucketIndex []uint64) *RequireMes {
	mes := &RequireMes{}
	mes.Indexs = bucketIndex
	return mes
}

//请求对方给本地同步数据
func (w *WorldState) sendPull(bucketIndex []uint64, id string) {
	wdLogger.Info("wd:sendPull")
	tempMes := w.makeRequireMes(bucketIndex)
	mes := &WDMes{}
	mes.Type = MesType_RM
	mes.Realmes = &WDMes_Require{tempMes}
	data, _ := proto.Marshal(mes)
	e := w.sp.SendInstruction(666, 666, data, []byte(id))
	if e != nil {
		wdLogger.Errorf("wd:Send PullMes error:%s\n", e)
	}
}

//收到对方请求同步数据
func (w *WorldState) recvPull(mes *RequireMes, id string) {
	wdLogger.Info("wd:recvPull")
	w.sendPush(mes.Indexs, id)
}

//两个节点间数据一样
func (w *WorldState) recvCompareSuccess(id string) {
	wdLogger.Info("wd:recvCompareSuccess")
	w.updateRetCh <- 2
}
func (w *WorldState) sendCompareSuccess(id string) {
	wdLogger.Info("wd:sendCompareSuccess")
	mes := &WDMes{}
	mes.Type = MesType_CS
	data, _ := proto.Marshal(mes)
	e := w.sp.SendInstruction(666, 666, data, []byte(id))
	if e != nil {
		wdLogger.Errorf("wd:Send CompareSuccess error:%s\n", e)
	}
}

//出现异常情况，无法同步数据
func (w *WorldState) recvCompareFailed(id string) {
	wdLogger.Info("wd:recvCompareFailed")
	w.updateRetCh <- 0
}

func (w *WorldState) sendCompareFailed(id string) {
	wdLogger.Info("wd:sendCompareFailed")
	mes := &WDMes{}
	mes.Type = MesType_CF
	data, _ := proto.Marshal(mes)
	e := w.sp.SendInstruction(666, 666, data, []byte(id))
	if e != nil {
		wdLogger.Errorf("wd:Send CompareFailed error:%s\n", e)
	}
}

//接受同步数据
func (w *WorldState) recvFunc(userData interface{}, rId []byte, rData []byte) int {
	w.inCh <- &transportMes{
		id:   string(rId),
		data: string(rData),
	}
	return 0
}

//功能函数
func (w *WorldState) bkdrhash(key string) uint64 {
	seed := uint64(131313)
	hash := uint64(0)
	for _, v := range key {
		hash = hash*seed + uint64(v)
	}
	return uint64(hash % uint64(1<<w.bTree.GetTreeTall()))
}
func (w *WorldState) GetRootHash() string {
	return w.bTree.GetRootHash()
}
func (w *WorldState) Search(k string) SearchResult {
	i := w.bkdrhash(k)
	KeysB, _ := w.db.Get([]byte(strconv.FormatUint(i, 10)+"_keys"), nil)
	tempKeys := &Keys{}
	proto.Unmarshal(KeysB, tempKeys)
	isFind := false
	for _, kk := range tempKeys.Key {
		if k == kk {
			isFind = true
			break
		}
	}
	if !isFind {
		return SearchResult{
			Key:   k,
			Value: "",
			Err:   errors.New("Key does not exist!"),
		}
	}
	data, err := w.db.Get([]byte(k), nil)
	v := string(data)
	return SearchResult{
		Key:   k,
		Value: v,
		Err:   err,
	}
}
func (w *WorldState) Searchn(key []string) (res []SearchResult) {
	for _, k := range key {
		resSingle := w.Search(k)
		res = append(res, resSingle)
	}
	return
}
func (w *WorldState) SearchPrefix(prefix string) (map[string]string, error) {
	iter := w.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	if iter.Error() != nil {
		return nil, iter.Error()
	}
	tempKv := make(map[string]string, 0)
	for iter.Next() {
		tempKv[string(iter.Key())] = string(iter.Value())
	}
	iter.Release()
	return tempKv, nil
}
func (w *WorldState) SearchRange(start string, stop string) (map[string]string, error) {
	iter := w.db.NewIterator(&util.Range{Start: []byte(start), Limit: []byte(stop)}, nil)
	if iter.Error() != nil {
		return nil, iter.Error()
	}
	tempKv := make(map[string]string, 0)
	for iter.Next() {
		tempKv[string(iter.Key())] = string(iter.Value())
	}
	iter.Release()
	return tempKv, nil
}
