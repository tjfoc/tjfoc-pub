package main

import "github.com/tjfoc/tjfoc/tjfoctest/testcase"

func main() {

	//testcase.SingleTesting()
	//testcase.StressTesting()
	//testcase.CalculateTPSForTxs()

	//parameter should be ./MTP 10.1.3.160:9000-10.1.3.161:9000 1 1000
	testcase.QueryTxByHash("TestCase_1")
	//parameter should be ./MTP 10.1.3.160:9000-10.1.3.161:9000-10.1.3.162:9000 100
	testcase.SendTxToMutliPeersChk("TestCase_2")
	//parameter should be ./MTP 10.1.3.160:9000-10.1.3.161:9000-10.1.3.162:9000 100
	testcase.CheckSameTxToMultiPeers("TestCase_3")
	//parameter should be ./MTP 10.1.3.160:9000 2 100
	testcase.CheckMultiSameTxToOnePeer("TestCase_4")
}
