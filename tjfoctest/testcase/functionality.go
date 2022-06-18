package testcase

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tjfoc/tjfoc/tjfoctest/common"
)

func SingleTesting() {

	var address string
	var txnNumbers int

	//add for multi address for transaction query
	addr := strings.Split(os.Args[1], "-")
	address = addr[0]

	txnNumbers, _ = strconv.Atoi(os.Args[2])

	// //add for delay check transactions
	// var sleepTime int

	// if len(os.Args) == 4 {
	// 	sleepTime, _ = strconv.Atoi(os.Args[3])
	// } else {
	// 	sleepTime = 500
	// }

	common.AddTxForSingleTesting(address, txnNumbers)

	time.Sleep(time.Second * 3)

	for _, value := range addr {
		common.QueryBlockHeight(value)
	}

}

//QueryTxByHash check the added transactions by hash data
func QueryTxByHash(section string) {
	//parse parameters for testcase
	var addressAdd string
	var sleepTime, txnNumbers int

	addrListstr := common.GetParamFromIni("./conf/conf.ini", "Peer_Node", "Address")
	addrList := strings.Split(addrListstr, ";")
	port := common.GetParamFromIni("./conf/conf.ini", "Peer_Node", "gRPC_Port")
	if port == "" {
		fmt.Println("the gRPC is null,please check [Peer_Node] gRPC_Port settings")
		return
	}
	for i := 0; i < len(addrList); i++ {
		addrList[i] = addrList[i] + ":" + port
	}

	txnNumbersstr := common.GetParamFromIni("./conf/conf.ini", section, "TX_NO")
	txnNumbers, _ = strconv.Atoi(txnNumbersstr)
	addressAdd = addrList[0]

	if common.GetParamFromIni("./conf/conf.ini", section, "SleepTime") == "" {
		sleepTime = 500
	} else {
		sleepTime, _ = strconv.Atoi(common.GetParamFromIni("./conf/conf.ini", section, "SleepTime"))
	}

	//record execution start local time
	start := time.Now()
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	common.WriteInfoToLog(section + ":Query transaction Start........" +
		"\nPeers:\n" + strings.Replace(addrListstr, ";", "\n", -1) +
		"\nTransaction Number: " + txnNumbersstr)

	common.QueryMultiTransaction(addressAdd, addrList, txnNumbers, sleepTime)

	//record execution duration
	end := time.Since(start)
	fmt.Printf("Execution period: %.2f ms\n", float64(end)/1000000)

	info := strconv.FormatFloat(float64(end)/1000000, 'g', -1, 64)
	common.WriteInfoToLog(section + " execution period: " + info + " ms")
	common.WriteInfoToLog("#############################end##################################")
}

//SendTxToMutliPeersChk send transactions to multi peers
func SendTxToMutliPeersChk(section string) {
	//parse parameters for testcase
	var txnNumbers int

	addrListstr := common.GetParamFromIni("./conf/conf.ini", "Peer_Node", "Address")
	addrList := strings.Split(addrListstr, ";")
	port := common.GetParamFromIni("./conf/conf.ini", "Peer_Node", "gRPC_Port")
	if port == "" {
		fmt.Println("the gRPC is null,please check [Peer_Node] gRPC_Port settings")
		return
	}
	for i := 0; i < len(addrList); i++ {
		addrList[i] = addrList[i] + ":" + port
	}
	txnNumbersstr := common.GetParamFromIni("./conf/conf.ini", section, "TX_NO")
	txnNumbers, _ = strconv.Atoi(txnNumbersstr)

	//record execution start local time
	start := time.Now()
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	common.WriteInfoToLog(section + ":Send Diff Multi Tx to Multi Peers Start........" +
		"\nPeers:\n" + strings.Replace(addrListstr, ";", "\n", -1) +
		"\nTransaction Number: " + txnNumbersstr)

	//execute test case function
	common.SendTxToMutliPeers(addrList, txnNumbers)

	//record execution duration
	end := time.Since(start)
	fmt.Printf("Execution period: %.2f ms\n", float64(end)/1000000)

	info := strconv.FormatFloat(float64(end)/1000000, 'g', -1, 64)
	common.WriteInfoToLog(section + " execution period: " + info + " ms")
	common.WriteInfoToLog("#############################end##################################")
}

//CheckSameTxToMultiPeers check transaction number when send same transacitons to multi peers
func CheckSameTxToMultiPeers(section string) {
	//parse parameters for testcase

	var txnNumbers int
	addrListstr := common.GetParamFromIni("./conf/conf.ini", "Peer_Node", "Address")
	addrList := strings.Split(addrListstr, ";")
	port := common.GetParamFromIni("./conf/conf.ini", "Peer_Node", "gRPC_Port")
	if port == "" {
		fmt.Println("the gRPC is null,please check [Peer_Node] gRPC_Port settings")
		return
	}
	for i := 0; i < len(addrList); i++ {
		addrList[i] = addrList[i] + ":" + port
	}
	txnNumbersstr := common.GetParamFromIni("./conf/conf.ini", section, "TX_NO")
	txnNumbers, _ = strconv.Atoi(txnNumbersstr)

	//record execution start local time
	start := time.Now()
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	common.WriteInfoToLog(section + ":Send Same Tx to Multi Peers Start........" +
		"\nPeers:\n" + strings.Replace(addrListstr, ";", "\n", -1) +
		"\nTransaction Number: " + txnNumbersstr)

	//execute test case function
	if common.CheckSameTxResult(addrList, txnNumbers) {
		common.WriteInfoToLog("CheckSameTxToMultiPeers Pass")
	} else {
		fmt.Println("CheckSameTxToMultiPeers Fail******")
		common.WriteInfoToLog("CheckSameTxToMultiPeers Fail******")
	}

	//record execution duration
	end := time.Since(start)
	fmt.Printf("Execution period: %.2f ms\n", float64(end)/1000000)

	info := strconv.FormatFloat(float64(end)/1000000, 'g', -1, 64)
	common.WriteInfoToLog(section + " execution period: " + info + " ms")
	common.WriteInfoToLog("#############################end##################################")
}

//CheckMultiSameTxToOnePeer check transaction number when send same transacitons to one peer by multi times
func CheckMultiSameTxToOnePeer(section string) {
	//parse parameters for testcase
	var txnNumbers, repeatTimes int
	var address string
	addrListstr := common.GetParamFromIni("./conf/conf.ini", "Peer_Node", "Address")
	addrList := strings.Split(addrListstr, ";")
	port := common.GetParamFromIni("./conf/conf.ini", "Peer_Node", "gRPC_Port")
	if port == "" {
		fmt.Println("the gRPC is null,please check [Peer_Node] gRPC_Port settings")
		return
	}
	for i := 0; i < len(addrList); i++ {
		addrList[i] = addrList[i] + ":" + port
	}
	txnNumbersstr := common.GetParamFromIni("./conf/conf.ini", section, "TX_NO")
	txnNumbers, _ = strconv.Atoi(txnNumbersstr)
	repeatTimesstr := common.GetParamFromIni("./conf/conf.ini", section, "RepeatTime")
	repeatTimes, _ = strconv.Atoi(repeatTimesstr)

	address = addrList[0]

	//record execution start local time
	start := time.Now()
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	common.WriteInfoToLog(section + ":Send Multi Same Tx to the same Peer Start........" +
		"\nPeer:\n" + address +
		"\nTransaction Number: " + txnNumbersstr + "\nRepeat Times: " + repeatTimesstr)

	//execute test case function
	common.SendTxToOnePeerMutiTime(address, txnNumbers, repeatTimes)

	//record execution duration
	end := time.Since(start)
	fmt.Printf("Execution period: %.2f ms\n", float64(end)/1000000)

	info := strconv.FormatFloat(float64(end)/1000000, 'g', -1, 64)
	common.WriteInfoToLog(section + " execution period: " + info + " ms")
	common.WriteInfoToLog("#############################end##################################")
}
