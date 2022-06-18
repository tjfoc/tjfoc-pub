package main

import (
	"fmt"
	"strconv"

	shim "github.com/tjfoc/tjfoc/core/chaincode/shim"

	pb "github.com/tjfoc/tjfoc/protos/chaincode"
)

//一份有效的智能合约必须实现Init和Invoke方法
type MyChaincode struct {
}

func (cc MyChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fun, args := stub.GetFunctionAndParameters()
	fmt.Println("function:", fun)
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// 向账本写入数据,为账号A初始化金额
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}
	// 初始化账号B的金额
	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (cc MyChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// Make payment of X units from A to B
		return cc.invoke(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke
		return cc.query(stub, args)
	} else if function == "queryn" {
		return cc.queryn(stub, args)
	} else if function == "delete" {
		return cc.delete(stub, args)
	} else if function == "deleten" {
		return cc.deleten(stub, args)
	} else if function == "querysame" {
		return cc.querysame(stub, args)
	}
	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (cc MyChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Println(jsonResp)
	return shim.Success(Avalbytes)
}

// query callback representing the query of a chaincode
func (cc MyChaincode) queryn(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	// Get the state from the ledger
	Avalbytes, err := stub.GetStaten(args)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	fmt.Println(Avalbytes)
	return shim.Success(nil)
}
func (cc MyChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	err := stub.DelState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	fmt.Println("delete  删除完成")
	return shim.Success(nil)
}

func (cc MyChaincode) deleten(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	// Get the state from the ledger
	err := stub.DelStaten(args)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state " + "\"}"
		return shim.Error(jsonResp)
	}

	fmt.Println("deleten  删除完成")
	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (cc MyChaincode) querysame(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	var A string // Entities

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetStateByPrefix(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	fmt.Println(Avalbytes)
	return shim.Success(nil)
}

// Transaction makes payment of X units from A to B
func (cc MyChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))
	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func main() {
	// 注册自定义合约
	fmt.Println("enter main")
	err := shim.Start(new(MyChaincode))
	if err != nil {
		fmt.Printf("Error starting my chaincode : %s", err)
	}
}
