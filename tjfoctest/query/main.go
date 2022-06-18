package main

import (
	"fmt"
	"os"

	"github.com/tjfoc/tjfoc/tjfoctest/common"
)

func main() {

	address := os.Args[1]

	endHeight := uint64(common.GetBlockHeight(address))

	sum := common.GetTxNumbersInBlockChain(address, 0, endHeight)

	fmt.Printf("Address: %s, Block height: %d\n", address, sum)
	fmt.Printf("Total numbers of transactions added: %d\n", sum)
}
