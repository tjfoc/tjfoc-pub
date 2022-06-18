package common

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/config"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/tjfoc/tjfoc/protos/peer"
	"github.com/tjfoc/tjfoc/protos/transaction"
)

//QueryMultiTransaction add transaction and check transaction by hash
/*************************************************************************************/
//Data:20180410
//func QueryMultiTransaction(addressAdd string, addressQuery string, times int)  --> to query multi existed transaction
//addressAdd:the address of the block chain node which is added a transaction
//addressQuery:the address of the block chain node which query the transaction
//times:the added transaction times
/*************************************************************************************/
func QueryMultiTransaction(addressAdd string, addressQuerylist []string, times int, sleepTime int) {

	var s [][]byte
	//add transactions and save all transaction hash
	s = AddTxnsConSaveHash(addressAdd, times)
	//fmt.Println(len(s))
	//need wait for a moment for the block will be packeted and added to the block chain
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	//time.Sleep(3000 * time.Millisecond)
	fmt.Printf("Add transaction complete and Delay %d ms\n", sleepTime)
	//query block height
	WriteInfoToLog("Address: " + addressQuerylist[0] + "Block Height:" + strconv.Itoa(GetBlockHeight(addressQuerylist[0])))
	//use all transactions query function

	for _, addressQuery := range addressQuerylist {
		start := time.Now()
		var successqueryno = QueryTrsnByHashDataList(addressQuery, s)
		end := time.Since(start) / 1000000.0
		fmt.Printf("Query Transaction No.: %d in %.2f ms\n", successqueryno, float64(end))
	}
}

//QueryTrsnByHashDataList check transaction by hash list
/*************************************************************************************/
//Data:20180410
//func QueryTrsnByHashDataList(address string, hashDatalist [][]byte) int64   --> to query mutli existed transaction by hashdata
//address:the address of the block chain node which query the transactions
//hashDatalist:the query hash list of the transactions
//this will be a return value of success query result
/*************************************************************************************/
func QueryTrsnByHashDataList(address string, hashDatalist [][]byte) int64 {
	client, conn, err := connect(address)
	if err != nil {
		logger.Error("gRPC query connection error!")
	}
	//生成metric对象
	cQureyOK := metrics.NewCounter()

	runtime.GOMAXPROCS(runtime.NumCPU())
	var wg sync.WaitGroup
	//fmt.Printf("hash data list length: %d", len(hashDatalist))
	for i := 0; i < len(hashDatalist); i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			res1, err1 := client.BlockchainGetTransaction(context.Background(), &peer.BlockchainHash{hashDatalist[index]})
			if err1 != nil {
				fmt.Printf("address:%s Query Transaction:%x Error\n", address, res1.Header.TransactionHash)
				WriteInfoToLog(address + "query transaction error")
				fmt.Println(err1)
				//return
			} else {
				fmt.Printf("address:%s Query transaction:%x complete!\n", address, res1.Header.TransactionHash)
				WriteInfoToLog(address + "query transaction complete")
				cQureyOK.Inc(1)
			}

		}(i)
	}

	wg.Wait()

	conn.Close()
	return cQureyOK.Count()
}

//AddTxnsConSaveHash add transaction and save transaction hash data
/*************************************************************************************/
//Data:20180410
//func AddTxnsConSaveHash(address string, times int) [][]byte   --> add transactions and save transaction hash data
//address:the address of the block chain node which query the transactions
//times:the transaction count
//return the column of hash data
/*************************************************************************************/
func AddTxnsConSaveHash(address string, times int) [][]byte {
	//client := createConn(address)
	var s1 [][]byte
	client, conn, err := connect(address)
	if err != nil {
		logger.Error("gRPC connection error!")
	}

	//生成metric对象 to calculator the success newtransaction
	cSuccess := metrics.NewCounter()

	runtime.GOMAXPROCS(runtime.NumCPU())

	var txn []*transaction.Transaction
	//generate hashdata by timestamp for every transaction
	txn = genTxn(times)

	start := time.Now()

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
					s1 = append(s1, txn[index].GetHeader().GetTransactionHash())
					fmt.Printf("the added hash data :%x\n", txn[index].GetHeader().GetTransactionHash())
				}
			}

		}(i)
	}

	wg.Wait()

	conn.Close()

	end := time.Since(start)

	tps := (float64(times) / float64(end)) * 1000000000.0

	fmt.Printf("Total: %d times, Success: %d times, TPS: %.2f\n", times, cSuccess.Count(), tps)
	fmt.Printf("hash data length: %d \n", len(s1))
	return s1

}

//addressValidChk to check the address format :10.1.3.160:9000
/*************************************************************************************/
//Data:20180413
//address:the node address
/*************************************************************************************/
func AddressValidChk(address string) bool {
	var bValid bool
	if address != "" {
		arr := strings.Split(address, ":")
		if arr[1] != "" {
			bValid = true
		} else {
			fmt.Printf("Please check port is null\n")
			bValid = false
		}
	}
	return bValid
}

//SendTxToMutliPeers send transactions to multi peers
/*************************************************************************************/
//Data:20180418
//func SendTxToMutliPeers(addrList []string ,times int)   --> send transaction to multi peers
//addrList:the addresses of the block chain node
//times:the transaction count
/*************************************************************************************/
func SendTxToMutliPeers(addrList []string, times int) {

	initblockHeight := GetBlockHeight(addrList[0])
	var wg sync.WaitGroup

	for i := 0; i < len(addrList); i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			fmt.Println(addrList[index])
			AddTxForSingleTesting(addrList[index], times)

		}(i)
	}

	wg.Wait()

	//time.Sleep(time.Duration(sleepTime) * time.Microsecond)
	txNo := GetTxNumbersInBlockChain(addrList[0], uint64(initblockHeight), uint64(GetBlockHeight(addrList[0])))

	for {
		if txNo == len(addrList)*times {
			break
		} else {
			//fmt.Printf("Sleep for %d ms,query transaction number:%d\n", sleepTime, txNo)
			//time.Sleep(time.Duration(sleepTime) * time.Microsecond)
			k := 1
			for k <= 800000000 {
				k = k + 1
			}
			txNo = GetTxNumbersInBlockChain(addrList[0], uint64(initblockHeight), uint64(GetBlockHeight(addrList[0])))
		}
	}
}

//CheckSameTxResult check the number of transaction while send same transactions to peer
/*************************************************************************************/
//Data:20180419
//addrList:the addresses of the block chain node
//times:the transaction count
/*************************************************************************************/
func CheckSameTxResult(addrList []string, times int) bool {

	initblockHeight := GetBlockHeight(addrList[0])
	bPass := SendSameTxToMultiPeer(addrList, times)

	txNo := GetTxNumbersInBlockChain(addrList[0], uint64(initblockHeight), uint64(GetBlockHeight(addrList[0])))
	for {
		if txNo >= times {
			break
		} else {
			k := 1
			for k <= 800000000 {
				k = k + 1
			}
			txNo = GetTxNumbersInBlockChain(addrList[0], uint64(initblockHeight), uint64(GetBlockHeight(addrList[0])))
		}
	}
	return bPass
}

//SendSameTxToMultiPeer send same transactions to multi peers
/*************************************************************************************/
//Data:20180419
//addrList:the addresses of the block chain node
//times:the transaction count
/*************************************************************************************/
func SendSameTxToMultiPeer(addrList []string, times int) bool {

	var bPass bool

	var txn []*transaction.Transaction
	txn = genTxn(times)
	//var errArr []error
	for _, value := range addrList {
		client, conn, err := connect(value)
		if err != nil {
			logger.Error("gRPC connection error!")
		}
		SendTxToPeer(client, txn)

		conn.Close()
	}

	fmt.Printf("Complete\n")
	bPass = true
	return bPass
}

//SendTxToPeer with the same transactions
/*************************************************************************************/
//Data:20180419
//client:PeerClient which has already been connected
//txn:the transaction which be added to peers
/*************************************************************************************/
func SendTxToPeer(client peer.PeerClient, txn []*transaction.Transaction) {

	//生成metric对象
	cSuccess := metrics.NewCounter()

	runtime.GOMAXPROCS(runtime.NumCPU())

	var wg sync.WaitGroup

	for i := 0; i < len(txn); i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			res, err := client.NewTransaction(context.Background(), txn[index])
			if err != nil {
				WriteInfoToLog(err.Error())
				logger.Error(err)
			} else {
				if res.Ok {
					cSuccess.Inc(1)
				}
			}

		}(i)

	}

	wg.Wait()

	//conn.Close()

	fmt.Printf("Total: %d times, Success: %d times\n", len(txn), cSuccess.Count())
}

//SendTxToOnePeerMutiTime send one transaction to the same node by multi times
/*************************************************************************************/
//Data:20180419
//client:PeerClient which has already been connected
//times:the transaction which be added to peers
/*************************************************************************************/
func SendTxToOnePeerMutiTime(address string, txnumb int, times int) {

	initblockHeight := GetBlockHeight(address)
	//WriteInfoToLog(strconv.Itoa(initblockHeight))
	client, conn, err := connect(address)
	if err != nil {
		WriteInfoToLog("error")
		logger.Error("gRPC connection error!")
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	var txn []*transaction.Transaction

	txn = genTxn(txnumb)

	var wg sync.WaitGroup

	for i := 0; i < times; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			SendTxToPeer(client, txn)
		}(i)

	}

	wg.Wait()
	fmt.Println("111")

	conn.Close()
	txNo := GetTxNumbersInBlockChain(address, uint64(initblockHeight), uint64(GetBlockHeight(address)))
	WriteInfoToLog("Init Block Height: " + strconv.Itoa(initblockHeight))

	for {
		if txNo >= txnumb {
			break
		} else {
			k := 1
			for k <= 800000000 {
				k = k + 1
			}
			txNo = GetTxNumbersInBlockChain(address, uint64(initblockHeight), uint64(GetBlockHeight(address)))
		}
	}
	WriteInfoToLog("Query tx no changes: " + strconv.Itoa(txNo))
	fmt.Printf("Success send TX No.: %d by %d times\n", txNo, times)
}

//WriteInfoToLog write information to log file
/*************************************************************************************/
//Data:20180419
//info:the test information will be written into logs
/*************************************************************************************/
func WriteInfoToLog(info string) {

	var filenameOnly string
	filenameOnly = GetCurFilename()
	//fmt.Println("filenameOnly=", filenameOnly)

	var logFilename string = filenameOnly + ".log"
	//fmt.Println("logFilename=", logFilename)
	logFile, err := os.OpenFile(logFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		fmt.Printf("open file error=%s\r\n", err.Error())
		os.Exit(-1)
	}

	defer logFile.Close()
	//logger:=log.New(logFile,"\r\n", log.Ldate | log.Ltime | log.Llongfile)
	logger := log.New(logFile, "\r\n", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println(info)

}

// GetCurFilename  Get current file name, without suffix
/*************************************************************************************/
//Data:20180419
/*************************************************************************************/
func GetCurFilename() string {

	// fulleFilename will be :/root/go/src/github.com/tjfoc/tjfoc/tjfoctest/common/function.go
	_, fulleFilename, _, _ := runtime.Caller(0)

	var filenameWithSuffix string
	//filenameWithSuffix will be:function.go
	filenameWithSuffix = path.Base(fulleFilename)

	var fileSuffix string
	//fileSuffix will be:.go
	fileSuffix = path.Ext(filenameWithSuffix)

	//filenameOnly will be:function
	var filenameOnly string
	filenameOnly = strings.TrimSuffix(filenameWithSuffix, fileSuffix)

	return filenameOnly
}

//GetParamFromIni  get config parameters
func GetParamFromIni(inifile string, section string, option string) string {
	c, err := config.ReadDefault(inifile)
	if err != nil {
		fmt.Println("Read error:", err)

		return ""
	}

	name, err := c.String(section, option)
	if err != nil {
		fmt.Println("Get name failed:", err)
		return ""
	}
	return name
}

func GetTxNoWithoutPrint(address string, startHeight, endHeight uint64) int {

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
	}

	conn.Close()
	//fmt.Printf("sum: %d \n", sum)
	return sum

}
