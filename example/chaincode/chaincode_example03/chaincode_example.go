package main

import (
	"fmt"
	"github.com/tjfoc/tjfoc/core/chaincode/shim"
	"github.com/tjfoc/tjfoc/protos/chaincode"
	"strconv"
)

type SimpleChaincode struct {
}

func (s *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) chaincode.Response {
	fmt.Println("in Init")
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 2 {
		fmt.Println("args too short")
		return shim.Error("args too short")
	}
	key := args[0]
	value := args[1]
	fmt.Printf("init Key:%s,Value:%s\n", key, value)
	err := stub.PutState(key, []byte(value))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}
func (s *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) chaincode.Response {
	fmt.Println("in Invoke")
	_, args := stub.GetFunctionAndParameters()
	if len(args) < 3 {
		fmt.Println("args too short")
		return shim.Error("args too short")
	}
	keya := args[0]
	keyb := args[1]
	value, e := strconv.ParseInt(string(args[2]), 10, 64)
	if e != nil {
		fmt.Println("third arg strconv to int64 error!")
		return shim.Error("third arg strconv to int64 error!")
	}
	valuea, e := stub.GetState(keya)
	if e != nil {
		fmt.Println("get first value failed!")
		return shim.Error("get first value failed!")
	}
	fmt.Println("before value a:", valuea)
	valueb, e := stub.GetState(keyb)
	if e != nil {
		fmt.Println("get second value failed!")
		return shim.Error("get second value failed!")
	}
	fmt.Println("before value b:", valueb)
	if r, e := stub.Cmp(keya, valuea, value); e != nil || r < 0 {
		fmt.Println("no enough money!")
		return shim.Error("no enough money!")
	}
	valuea, e = stub.Sub(keya, valuea, value)
	if e != nil {
		fmt.Println("sub error:", e)
		return shim.Error("sub error")
	}
	fmt.Println("after sub value a:", valuea)
	valueb, e = stub.Add(keyb, valueb, value)
	if e != nil {
		fmt.Println("add error:", e)
		return shim.Error("add error")
	}
	fmt.Println("after add value b:", valueb)
	e = stub.PutState(keya, valuea)
	if e != nil {
		fmt.Println("putstate first value error!")
		return shim.Error("putstate first value error!")
	}
	e = stub.PutState(keyb, valueb)
	if e != nil {
		fmt.Println("putstate second value error!")
		return shim.Error("putstate second value error!")
	}
	return shim.Success(nil)
}
func main() {
	shim.Start(new(SimpleChaincode))
}
