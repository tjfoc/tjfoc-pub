package common

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/tjfoc/gmsm/sm3"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/protos/peer"
	"github.com/tjfoc/tjfoc/protos/transaction"
	"google.golang.org/grpc"
)

var logger = flogging.MustGetLogger("process")

type EnterKey []byte

var Key = []EnterKey{}

var TestDB = true

var ca = "./ca.pem"

func connect(address string) (peer.PeerClient, *grpc.ClientConn, error) { //connect to peer TLS

	// admin := "./cert6.pem"
	// const adminkey = "./priv6.pem"
	// cert, err := gmtls.LoadX509KeyPair(admin, adminkey)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// certPool := sm2.NewCertPool()
	// cacert, err := ioutil.ReadFile(ca)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// certPool.AppendCertsFromPEM(cacert)
	// creds := gmcredentials.NewTLS(&gmtls.Config{
	// 	ServerName:   "test.example.com",
	// 	Certificates: []gmtls.Certificate{cert},
	// 	RootCAs:      certPool,
	// })

	// conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))

	conn, err := grpc.Dial(address, grpc.WithInsecure()) // Insecure

	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}

	return peer.NewPeerClient(conn), conn, nil

}

// func createConn(address string) peer.PeerClient {

// 	client, err := conn(address)
// 	if err != nil {
// 		logger.Error("gRPC connection error!")
// 		return nil
// 	}

// 	return client
// }

func QueryBlockHeight(address string) {
	client, conn, err := connect(address)
	if err != nil {
		logger.Error("gRPC connection error!")
	}

	res, err := client.BlockchainGetHeight(context.Background(), &peer.BlockchainBool{true})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("address : %s, Block height: %d \n", address, res.Number)
	conn.Close()

}

func GetBlockHeight(address string) int {
	client, conn, err := connect(address)
	if err != nil {
		logger.Error("gRPC connection error!")
	}

	res, err := client.BlockchainGetHeight(context.Background(), &peer.BlockchainBool{true})
	if err != nil {
		fmt.Println(err)
		return -1
	}

	fmt.Printf("Address : %s, Block height: %d \n", address, res.Number)
	conn.Close()

	return int(res.Number)

}

func GetTxNumbersInBlockChain(address string, startHeight, endHeight uint64) int {

	var sum = 0

	client, conn, err := connect(address)
	if err != nil {
		logger.Error("gRPC connection error!")
	}

	for i := startHeight; i < endHeight; i++ {
		res, err := client.BlockchainGetBlockByHeight(context.Background(), &peer.BlockchainNumber{i})

		if err != nil {
			fmt.Println(err)
			return -1
		}

		//fmt.Printf("address : %s, Block txns number: %d \n", address, res.Header.Height)

		sum = sum + len(res.Txs)

		// fmt.Printf("Block txs: %d \n", len(res.Txs))
		// fmt.Printf("Block txn numbers: %d \n", res.Header.Height)
		// // fmt.Printf("Block hash: %d \n", res.Header.BlockHash)
		// fmt.Println("-----------------------------------")
	}

	conn.Close()
	fmt.Printf("sum: %d \n", sum)
	return sum

}

func AddTxForTPS(address string, times int) float64 {

	client, conn, err := connect(address)
	if err != nil {
		logger.Error("gRPC connection error!")
	}

	//生成metric对象
	cSuccess := metrics.NewCounter()

	runtime.GOMAXPROCS(runtime.NumCPU())

	var txn []*transaction.Transaction

	txn = genTxn(times)

	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < times; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			res, err := client.NewTransaction(context.Background(), txn[index])
			if err != nil {
				logger.Error(err)
			} else {
				if res.Ok {
					cSuccess.Inc(1)
				}
			}

		}(i)

	}

	wg.Wait()

	conn.Close()

	end := time.Since(start)

	tps := (float64(times) / float64(end)) * 1000000000.0

	if int64(times) != cSuccess.Count() {
		fmt.Printf("Total: %d times, Success: %d times\n", times, cSuccess.Count())
	}

	return tps

}

func AddTxForStressTesting(address string, times int) {

	client, conn, err := connect(address)
	if err != nil {
		logger.Error("gRPC connection error!")
	}

	//生成metric对象
	cSuccess := metrics.NewCounter()

	runtime.GOMAXPROCS(runtime.NumCPU())

	var txn []*transaction.Transaction

	txn = genTxn(times)

	var wg sync.WaitGroup

	for i := 0; i < times; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			res, err := client.NewTransaction(context.Background(), txn[index])
			if err != nil {
				logger.Error(err)
			} else {
				if res.Ok {
					cSuccess.Inc(1)
				}
			}

		}(i)

	}

	wg.Wait()
	conn.Close()

	fmt.Printf("Total: %d times, Success: %d times\n", times, cSuccess.Count())

}

func AddTxnsForAllPeers(address []string, times int) {

	for _, value := range address {
		client, conn, err := connect(value)
		if err != nil {
			logger.Error("gRPC connection error!")
		}

		//生成metric对象
		cSuccess := metrics.NewCounter()

		runtime.GOMAXPROCS(runtime.NumCPU())

		var txn []*transaction.Transaction

		txn = genTxn(times)

		var wg sync.WaitGroup

		for i := 0; i < times; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				res, err := client.NewTransaction(context.Background(), txn[index])
				if err != nil {
					logger.Error(err)
				} else {
					if res.Ok {
						cSuccess.Inc(1)
					}
				}

			}(i)

		}

		wg.Wait()

		conn.Close()

		fmt.Printf("Total: %d times, Success: %d times, Address: %v\n", times, cSuccess.Count(), value)

	}

}

func CalculateAverageTPSForTxs(address string, numbers int) {

	var avg, sum, max, min, tps float64

	sum = 0.0

	for i := 0; i < 10; i++ {
		tps = AddTxForTPS(address, numbers)

		sum = sum + tps

		if i == 0 {
			max, min = tps, tps
		} else {
			if tps > max {
				max = tps
			}
			if tps < min {
				min = tps
			}
		}
	}

	avg = (sum - max - min) / 8.0

	fmt.Printf("Average TPS for %d times is %.2f(counting exclued max and min)\n", numbers, avg)
	fmt.Printf("Max TPS for %d times is %.2f\n", numbers, max)
	fmt.Printf("Min TPS for %d times is %.2f\n", numbers, min)

}

func AddTxForSingleTesting(address string, times int) {

	client, conn, err := connect(address)
	if err != nil {
		logger.Error("gRPC connection error!")
	}

	//生成metric对象
	cSuccess := metrics.NewCounter()

	runtime.GOMAXPROCS(runtime.NumCPU())

	var txn []*transaction.Transaction

	txn = genTxn(times)

	var wg sync.WaitGroup

	for i := 0; i < times; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			res, err := client.NewTransaction(context.Background(), txn[index])
			if err != nil {
				logger.Error(err)
			} else {
				if res.Ok {
					cSuccess.Inc(1)
				}
			}

		}(i)

	}

	wg.Wait()

	conn.Close()

	fmt.Printf("Total: %d times, Success: %d times\n", times, cSuccess.Count())

}

func genTxn(times int) []*transaction.Transaction {

	var Txn []*transaction.Transaction

	for i := 0; i < times; i++ {

		timestamp := uint64(time.Now().Unix())
		hashData, _, operation := genHash(timestamp)
		txnHead := &transaction.TransactionHeader{
			Version:         0,
			Timestamp:       timestamp,
			TransactionHash: hashData,
			TransactionSign: []byte(nil),
		}

		txn := &transaction.Transaction{
			SmartContract: []byte(operation),
			Header:        txnHead,
		}
		Txn = append(Txn, txn)
	}

	return Txn
}

func genHash(timestamp uint64) ([]byte, string, string) {

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	currentKey, operation := genOperation(rnd.Intn(50), rnd.Intn(50), rnd.Intn(50))
	buf := []byte{}
	//version =0
	buf = append(buf, miscellaneous.E32func(0)...)
	//取时间戳
	//timestamp := uint64(time.Now().Unix())
	//buf = append(buf, []byte(timestamp)...)
	buf = append(buf, miscellaneous.E64func(timestamp)...)
	buf = append(buf, miscellaneous.E32func(uint32(len([]byte(operation))))...)
	buf = append(buf, []byte(operation)...)
	buf = append(buf, miscellaneous.E32func(0)...)
	hashData, _ := miscellaneous.GenHash(sm3.New(), buf)

	return hashData, currentKey, operation
}

func genOperation(i int, index int, in int) (string, string) {
	sA := randUpString(i)
	sB := randUpString(index)
	sC := randUpString(in)
	nD := (i + index + in)

	currentKey := fmt.Sprintf("%s%s%s%d", sA, sB, sC, nD)

	if TestDB == false {
		Key = append(Key, []byte(currentKey))
	}
	//fmt.Println(fmt.Sprintf("SET %s 12345%d", currentKey))
	return currentKey, fmt.Sprintf("SET %s 12345%d", currentKey, index)
}

func randUpString(l int) string {
	var result bytes.Buffer
	var temp byte
	for i := 0; i < l; {
		if randInt(65, 91) != temp {
			temp = randInt(65, 91)
			result.WriteByte(temp)
			i++
		}
	}
	return result.String()
}

func randInt(min int, max int) byte {
	rand.Seed(time.Now().UnixNano())
	return byte(min + rand.Intn(max-min))
}
