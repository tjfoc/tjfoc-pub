package process

// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"math/rand"
// 	"runtime"
// 	"sync"
// 	"time"

// 	metrics "github.com/rcrowley/go-metrics"
// 	"github.com/tjfoc/gmsm/sm3"
// 	"github.com/tjfoc/tjfoc/core/common/flogging"
// 	"github.com/tjfoc/tjfoc/core/miscellaneous"
// 	"github.com/tjfoc/tjfoc/protos/peer"
// 	"github.com/tjfoc/tjfoc/protos/transaction"
// 	"google.golang.org/grpc"
// )

// var logger = flogging.MustGetLogger("process")

// var DefaultAddress = "10.1.3.150:9000"

// var Address string

// var Times int

// type EnterKey []byte

// var Key = []EnterKey{}

// var TestDB = true

// var Height int

// func connection() (peer.PeerClient, error) { //connect to peer

// 	conn, err := grpc.Dial(Address, grpc.WithInsecure())
// 	if err != nil {
// 		logger.Error(err)
// 		return nil, err
// 	}
// 	return peer.NewPeerClient(conn), nil
// }

// func genOperation(i int, index int, in int) (string, string) {
// 	sA := randUpString(i)
// 	sB := randUpString(index)
// 	sC := randUpString(in)
// 	nD := (i + index + in)

// 	currentKey := fmt.Sprintf("%s%s%s%d", sA, sB, sC, nD)

// 	if TestDB == false {
// 		Key = append(Key, []byte(currentKey))
// 	}
// 	//fmt.Println(fmt.Sprintf("SET %s 12345%d", currentKey))
// 	return currentKey, fmt.Sprintf("SET %s 12345%d", currentKey, index)
// }

// func randUpString(l int) string {
// 	var result bytes.Buffer
// 	var temp byte
// 	for i := 0; i < l; {
// 		if randInt(65, 91) != temp {
// 			temp = randInt(65, 91)
// 			result.WriteByte(temp)
// 			i++
// 		}
// 	}
// 	return result.String()
// }

// func randInt(min int, max int) byte {
// 	rand.Seed(time.Now().UnixNano())
// 	return byte(min + rand.Intn(max-min))
// }

// func genHash(timestamp uint64) ([]byte, string, string) {

// 	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
// 	currentKey, operation := genOperation(rnd.Intn(50), rnd.Intn(50), rnd.Intn(50))
// 	buf := []byte{}
// 	//version =0
// 	buf = append(buf, miscellaneous.E32func(0)...)
// 	//取时间戳
// 	//timestamp := uint64(time.Now().Unix())
// 	//buf = append(buf, []byte(timestamp)...)
// 	buf = append(buf, miscellaneous.E64func(timestamp)...)
// 	buf = append(buf, miscellaneous.E32func(uint32(len([]byte(operation))))...)
// 	buf = append(buf, []byte(operation)...)
// 	buf = append(buf, miscellaneous.E32func(0)...)
// 	hashData, _ := miscellaneous.GenHash(sm3.New(), buf)

// 	//fmt.Println(hashData)   //sam
// 	//fmt.Println(currentKey) //sam
// 	//fmt.Println(operation)  //sam

// 	return hashData, currentKey, operation
// }

// func AddNewTranscation() {
// 	t1 := time.Now()
// 	client, err := connection()
// 	if err != nil {
// 		logger.Error("connection err!")
// 		return
// 	}
// 	time1 := time.Since(t1)
// 	//生成metric对象
// 	cSuccess := metrics.NewCounter()
// 	cfail := metrics.NewCounter()
// 	cSearch := metrics.NewCounter()
// 	cfail.Inc(int64(Times))
// 	historgam1 := metrics.NewHistogram(metrics.NewUniformSample(10000))
// 	historgam2 := metrics.NewHistogram(metrics.NewUniformSample(10000))
// 	historgam1.Update(int64(time1))

// 	var Tx []*transaction.Transaction

// 	for i := 0; i < Times; i++ {

// 		timestamp := uint64(time.Now().Unix())
// 		hashData, _, operation := genHash(timestamp)
// 		txHead := &transaction.TransactionHeader{
// 			Version:         0,
// 			Timestamp:       timestamp,
// 			TransactionHash: hashData,
// 			TransactionSign: []byte(nil),
// 		}

// 		tx := &transaction.Transaction{
// 			SmartContract: []byte(operation),
// 			Header:        txHead,
// 		}
// 		Tx = append(Tx, tx)
// 	}
// 	var wg sync.WaitGroup
// 	runtime.GOMAXPROCS(runtime.NumCPU())
// 	fmt.Println("cpu =", runtime.NumCPU())

// 	for i := 0; i < Times; i++ {
// 		wg.Add(1)
// 		go func(index int) {
// 			defer wg.Add(-1)

// 			t2 := time.Now()
// 			res, err := client.NewTransaction(context.Background(), Tx[index])
// 			if err != nil {
// 				//logger.Fatal(err)
// 				logger.Error(err)
// 			}
// 			time2 := time.Since(t2)
// 			historgam2.Update(int64(time2))

// 			//交易成功
// 			//	logger.Info("transcation res:", res)

// 			//GroupCache.Add()
// 			if res != nil {
// 				if res.Ok == true {
// 					cSuccess.Inc(1)
// 					cfail.Dec(1)
// 				}
// 			}

// 		}(i)
// 	}

// 	//time.Sleep(time.Second * time.Duration(Times/100))
// 	wg.Wait()

// 	//time.Sleep(time.Second * 1)
// 	s := fmt.Sprintf("交易新建次数:%d;交易成功次数:%d;交易失败次数:%d", Times, cSuccess.Count(), cfail.Count())
// 	logger.Info(s)
// 	// 时间取平均值
// 	// data := &TestData{
// 	// 	Function:               "addTx",
// 	// 	RutineTimes:            Times,
// 	// 	CreateConn:             historgam1.Mean(),
// 	// 	RequestAndResponseTime: historgam2.Mean(),
// 	// 	Result:                 s,
// 	// }

// 	// x := xlsxInit()
// 	// defer x.Save()
// 	// x.addFunction(data)

// 	//var wgs sync.WaitGroup
// 	fmt.Println("key len:", len(Key))

// 	for i := 0; i < len(Key); i++ {
// 		//	wgs.Add(1)
// 		//	go func(index int) {
// 		//	wgs.Add(-1)
// 		_, err := client.Search(context.Background(), &peer.BlockchainHash{[]byte(Key[i])})
// 		if err != nil {
// 			//fmt.Println(err)
// 		} else {
// 			cSearch.Inc(1)
// 		}

// 		//}(i)

// 	}
// 	//wgs.Wait()
// 	//time.Sleep(time.Second * 2)
// 	fmt.Println("search success times:", cSearch.Count())
// }

// func DBTest() {
// 	client, err := connection()
// 	if err != nil {
// 		logger.Error("connection err!")
// 		return
// 	}

// 	//
// 	for {
// 		timestamp := uint64(time.Now().Unix())
// 		hashData, currentKey, operation := genHash(timestamp)

// 		txHead := &transaction.TransactionHeader{
// 			Version:         0,
// 			Timestamp:       timestamp,
// 			TransactionHash: hashData,
// 			TransactionSign: []byte(nil),
// 		}

// 		tx := &transaction.Transaction{
// 			SmartContract: []byte(operation),
// 			Header:        txHead,
// 		}

// 		res, err := client.NewTransaction(context.Background(), tx)
// 		if err != nil {
// 			//logger.Fatal(err)
// 			logger.Error(err)
// 		} else {
// 			fmt.Println("addtx res = ", res)
// 		}
// 		time.Sleep(time.Millisecond * 100)
// 		ress, err := client.Search(context.Background(), &peer.BlockchainHash{[]byte(currentKey)})
// 		if err != nil {
// 			fmt.Println(err)
// 		} else {
// 			fmt.Println("Search Key #### ", currentKey)
// 			fmt.Println("Search res = = = ", ress)
// 		}
// 	}

// }

// // func SearchTx() {
// // 	client, err := connection()
// // 	if err != nil {
// // 		logger.Error("connection err---", err)
// // 		return
// // 	}
// // 	res, err := client.Search(context.Background(), &peer.BlockchainHash{[]byte("a")})
// // 	if err != nil {
// // 		logger.Fatal("SearchTx err---", err)
// // 	}

// // 	logger.Info("search success:", res)
// // }

// // //获取区块高度
// // func BlockchainGetHeight() {
// // 	client, err := connection()
// // 	if err != nil {
// // 		logger.Error("connection err---", err)
// // 		return
// // 	}

// // 	if Height > 0 {
// // 		res, err := client.BlockchainGetBlockByHeight(context.Background(), &peer.BlockchainNumber{uint64(Height)})
// // 		if err != nil {
// // 			logger.Fatal(err)
// // 		}
// // 		logger.Info("block height:%d", Height)
// // 		logger.Info("block data:%+v", res)
// // 		return
// // 	}

// // 	//BlockChainBool == true
// // 	res, err := client.BlockchainGetHeight(context.Background(), &peer.BlockchainBool{true})
// // 	if err != nil {
// // 		logger.Fatal(err)
// // 	}

// // 	logger.Info("get height success", res)
// // }

// // func testBlockchainGetBlockByHash(hash []byte) {

// // 	client, err := connection()
// // 	if err != nil {
// // 		logger.Error("connection err---", err)
// // 		return
// // 	}

// // 	res, err := client.BlockchainGetBlockByHash(context.Background(), &peer.BlockchainHash{hash})
// // 	if err != nil {
// // 		logger.Fatal("get blockchain by hash err---", err)
// // 	}

// // 	logger.Info("get blockchain by hash success:", res)
// // }

// // func GetBlockByHeight(inNum uint64) {

// // 	client, err := connection()
// // 	if err != nil {
// // 		logger.Error("connection err---", err)
// // 		return
// // 	}

// // 	res, err := client.BlockchainGetBlockByHeight(context.Background(), &peer.BlockchainNumber{inNum})
// // 	if err != nil {
// // 		logger.Fatal("Get blockchain by height err---", err)
// // 	}

// // 	logger.Info("get blockchain by height:", res)
// // }

// // func testBlockchainGetTransaction(blockHash []byte) {

// // 	client, err := connection()
// // 	if err != nil {
// // 		logger.Error("connection err---", err)
// // 		return
// // 	}

// // 	res, err := client.BlockchainGetTransaction(context.Background(), &peer.BlockchainHash{blockHash})
// // 	if err != nil {
// // 		logger.Fatal("blockchain get transcation err---", err)

// // 	}
// // 	logger.Fatal("blockchain get transcation success:", res)
// // }

// // func testBlockchainGetTransactionIndex(blockHash []byte) {
// // 	client, err := connection()
// // 	if err != nil {
// // 		logger.Error("connection err---", err)
// // 		return
// // 	}

// // 	res, err := client.BlockchainGetTransactionIndex(context.Background(), &peer.BlockchainHash{blockHash})
// // 	if err != nil {
// // 		logger.Fatal("blockchain get tx index err---", err)
// // 	}

// // 	logger.Info("Blockchain get transcation index success:", res)
// // }

// // func testBlockchainGetTransactionBlock(blockHash []byte) {

// // 	t1 := time.Now()
// // 	client, err := connection()
// // 	if err != nil {
// // 		logger.Error("connection err---", err)
// // 		return
// // 	}
// // 	e := time.Since(t1)
// // 	logger.Info("create connection time:", e)

// // 	t2 := time.Now()
// // 	res, err := client.BlockchainGetTransactionBlock(context.Background(), &peer.BlockchainHash{blockHash})
// // 	if err != nil {
// // 		logger.Fatal("blockchain get transcation block err---", err)
// // 	}
// // 	logger.Info("blockchain get tx block success:", res)
// // 	e = time.Since(t2)
// // 	logger.Info("handle time:", e)
// // }

// // func SaveTestData(data *TestData) {
// // 	x := xlsxInit()
// // 	defer x.Save()
// // 	x.addFunction(data)
// // }

// // func BlockQuery() {
// // 	client, err := connection()
// // 	if err != nil {
// // 		logger.Error("connection err---", err)
// // 		return
// // 	}

// // 	timestamp := uint64(time.Now().Unix())
// // 	hashData, currentKey, operation := genHash(timestamp)
// // 	fmt.Println("currentKey = ", currentKey)
// // 	txHead := &transaction.TransactionHeader{
// // 		Version:         0,
// // 		Timestamp:       timestamp,
// // 		TransactionHash: hashData,
// // 		TransactionSign: []byte(nil),
// // 	}

// // 	tx := &transaction.Transaction{
// // 		SmartContract: []byte(operation),
// // 		Header:        txHead,
// // 	}

// // 	res, err := client.NewTransaction(context.Background(), tx)
// // 	if err != nil {
// // 		//logger.Fatal(err)
// // 		logger.Error(err)
// // 	} else {
// // 		fmt.Println("addtx res = ", res)
// // 	}

// // 	//query height

// // 	num, err := client.BlockchainGetTransactionIndex(context.Background(), &peer.BlockchainHash{hashData})
// // 	if err != nil {
// // 		logger.Fatal(err)
// // 	} else {
// // 		fmt.Println("TX index = ", num)
// // 	}

// // 	//query block by height

// // 	block, err := client.BlockchainGetBlockByHeight(context.Background(), num)

// // 	if err != nil {
// // 		logger.Error(err)
// // 	} else {
// // 		fmt.Println("block data  = ", block)
// // 	}

// // 	//query block by hash

// // 	blockSelf, err := client.BlockchainGetBlockByHash(context.Background(), &peer.BlockchainHash{block.Header.BlockHash})
// // 	if err != nil {
// // 		logger.Error(err)
// // 	} else {
// // 		if block == blockSelf {
// // 			fmt.Println("block == blockSelf")
// // 		} else {
// // 			fmt.Println("block = ", block)
// // 			fmt.Println("blockSelf = ", blockSelf)
// // 		}

// // 	}

// // 	//pervious hash query block

// // 	perviousBlock, err := client.BlockchainGetBlockByHash(context.Background(), &peer.BlockchainHash{block.Header.PreviousHash})
// // 	if err != nil {
// // 		logger.Error(err)
// // 	} else {
// // 		fmt.Println("pervious block = ", perviousBlock)
// // 	}

// // 	fmt.Println("-----------END---------------")
// // }
