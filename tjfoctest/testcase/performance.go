package testcase

import (
	"os"
	"strconv"
	"strings"

	"github.com/tjfoc/tjfoc/tjfoctest/common"
)

func CalculateTPSForTxs() {
	var address string
	var txnNumbers int

	//add for multi address for transaction query
	addr := strings.Split(os.Args[1], "-")
	address = addr[0]

	txnNumbers, _ = strconv.Atoi(os.Args[2])

	common.CalculateAverageTPSForTxs(address, txnNumbers)

}
