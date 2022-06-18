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

package chaincode

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	"github.com/tjfoc/tjfoc/core/common/ccprovider"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/container/util"
	"github.com/tjfoc/tjfoc/core/crypt"
	pb "github.com/tjfoc/tjfoc/protos/chaincode"
)

var chaincodeStartLogger = flogging.MustGetLogger("chaincode_start")

type singleValue struct {
	Value    []byte
	IsSecret bool
}
type singleResult struct {
	value   singleValue
	duiChen bool  //该值是否需要进行对称密钥加密
	tongTai bool  //如果需要对称密钥加密，是否需要先进行同态解密
	action  int32 //导致该结果改变的动作
}

type result struct {
	c                 crypt.Crypto
	peerID            string
	monitorAddr       string
	tempResult        map[string]*singleResult //所有交易结果
	singleTxResult    map[string]*singleResult //当前正在执行交易的结果
	resLocker         sync.Mutex
	currentTxIsSecret bool
	currentTxId       string
}

func (r *result) SetAttribute(id string, cc crypt.Crypto) {
	r.peerID = id
	r.c = cc
}

var res *result

func startDocker() {
	if !viper.GetBool("Docker.Enable") {
		return
	}
	NewChaincodeSupport()
	//开启与docker的GRPC服务
	lis, err := net.Listen("tcp", theChaincodeSupport.ip+":"+theChaincodeSupport.port)
	if err != nil {
		panic(fmt.Errorf("Listen err. %s", err))
	}
	grpcServer := grpc.NewServer()
	pb.RegisterChaincodeSupportServer(grpcServer, theChaincodeSupport)
	go grpcServer.Serve(lis)
	//启动未启动的docker
	listImagesAndStartDocker()
}

//listImagesAndStartDocker 判断当前docker是否存在,存在则启动
func listImagesAndStartDocker() {
	client, err := util.NewDockerClient()
	if err != nil {
		panic(fmt.Errorf("fatal newClient err %s", err))
	}
	imgs, err := client.ListImages(docker.ListImagesOptions{All: false})
	if err != nil {
		panic(err)
	}
	for _, img := range imgs {
		s := img.RepoTags[0]
		imageName := strings.Split(s, ":")[0]
		chaincodeName, chaincodeVersion, ip, port := getContainerNameFromImageName(imageName)
		//跳过非合约容器
		if chaincodeName == "" && chaincodeVersion == "" {
			continue
		}
		//跳过ip端口不对的合约容器，此处不对表明修改过配置文件
		if ip != theChaincodeSupport.ip || port != theChaincodeSupport.port {
			continue
		}
		cccid := ccprovider.NewCCContext("tjfoc", chaincodeName, chaincodeVersion, "", theChaincodeSupport.ip, theChaincodeSupport.port, false)
		theChaincodeSupport.Launch(context.Background(), cccid, []byte(""))
	}
}

func hasChainCode(name string) bool {
	client, err := util.NewDockerClient()
	if err != nil {
		panic(fmt.Errorf("fatal newClient err %s", err))
	}
	imgs, err := client.ListImages(docker.ListImagesOptions{All: false})
	if err != nil {
		panic(err)
	}
	for _, img := range imgs {
		s := img.RepoTags[0]
		imageName := strings.Split(s, ":")[0]
		chaincodeName, chaincodeVersion, _, _ := getContainerNameFromImageName(imageName)
		if chaincodeName == "" && chaincodeVersion == "" {
			continue
		}
		if chaincodeName == name {
			return true
		}
	}
	return false
}

// ChaincodeNormalTx 对普通交易在docker中进行运行、验证
func chaincodeNormalTx(name string, version string, args [][]byte, txID string) string {
	cccid := ccprovider.NewCCContext("tjfoc", name, version, txID, theChaincodeSupport.ip, theChaincodeSupport.port, false)
	canName := cccid.GetCanonicalName()
	if _, ok := theChaincodeSupport.runningChaincodes.chaincodeMap[canName]; !ok {
		chaincodeStartLogger.Errorf("chaincode [%s] doesn't exist!", canName)
		temp := &ccResult{Status: CCError, Message: fmt.Sprintf("Transaction excute failed!chaincode [%s] doesn't exist!", cccid.Name+"_"+cccid.Version), ChangedKv: nil, Response: nil}
		resB, _ := json.Marshal(temp)
		return string(resB)
	}
	cMsg := &pb.ChaincodeInput{Args: args}
	payload, err := proto.Marshal(cMsg)
	if err != nil {
		chaincodeStartLogger.Error(err)
	}
	var ccMsg *pb.ChaincodeMessage
	if len(args) != 0 && (string(args[0]) == "init" || string(args[0]) == "Init") {
		ccMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_INIT, Payload: payload, Txid: txID}
	} else {
		ccMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: txID}
	}
	theChaincodeSupport.currentTxContent = ccMsg
	resp, err := theChaincodeSupport.Execute(context.Background(), cccid, ccMsg, time.Duration(30)*time.Second)
	theChaincodeSupport.currentTxContent = nil
	if resp == nil {
		//超时
		chaincodeStartLogger.Error("normal tx Timeout!")
		temp := &ccResult{Status: CCTimeout, Message: "Transaction excute timeout!", ChangedKv: nil, Response: nil}
		resB, _ := json.Marshal(temp)
		return string(resB)
	} else if err != nil {
		//系统出错
		chaincodeStartLogger.Errorf("normal tx System Error:%s", err)
		temp := &ccResult{Status: CCError, Message: fmt.Sprintf("Transaction excute failed!System Error!"), ChangedKv: nil, Response: nil}
		resB, _ := json.Marshal(temp)
		return string(resB)
	} else {
		//正常
		chaincodeStartLogger.Info("Normal Transaction excute success!")
		temp := &ccResult{Status: CCSuccess, Message: "Transaction excute success!", ChangedKv: nil, Response: resp.Payload}
		resB, _ := json.Marshal(temp)
		return string(resB)
	}
}

func chaincodeSecretTx(name string, version string, args [][]byte, txID string) {
	cccid := ccprovider.NewCCContext("tjfoc", name, version, txID, theChaincodeSupport.ip, theChaincodeSupport.port, false)
	canName := cccid.GetCanonicalName()
	if _, ok := theChaincodeSupport.runningChaincodes.chaincodeMap[canName]; !ok {
		chaincodeStartLogger.Errorf("chaincode [%s] doesn't exist!", canName)
		return
	}
	cMsg := &pb.ChaincodeInput{Args: args}
	payload, err := proto.Marshal(cMsg)
	if err != nil {
		chaincodeStartLogger.Error(err)
	}
	var ccMsg *pb.ChaincodeMessage
	if len(args) != 0 && (string(args[0]) == "init" || string(args[0]) == "Init") {
		ccMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_INIT, Payload: payload, Txid: txID}
	} else {
		ccMsg = &pb.ChaincodeMessage{Type: pb.ChaincodeMessage_TRANSACTION, Payload: payload, Txid: txID}
	}
	theChaincodeSupport.currentTxContent = ccMsg
	resp, err := theChaincodeSupport.Execute(context.Background(), cccid, ccMsg, time.Duration(30)*time.Second)
	theChaincodeSupport.currentTxContent = nil
	if resp == nil {
		//超时
		chaincodeStartLogger.Error("secret tx Timeout!")
	} else if err != nil {
		//系统出错
		chaincodeStartLogger.Errorf("secret tx System Error:%s", err)
	} else {
		//正常
		chaincodeStartLogger.Info("Secret Transaction excute success!")
	}
}

// chaincodeSpecialTxCreate 处理安装chaincode容器的交易
func chaincodeSpecialTxCreate(chaincodeName string, version string, txID string, content []byte, sign string) string {
	chaincodeStartLogger.Debug("enter ChaincodeSpecialTxCreate,create sign: ", sign)
	cccid := ccprovider.NewCCContext("tjfoc", chaincodeName, version, txID, theChaincodeSupport.ip, theChaincodeSupport.port, false)
	canName := cccid.GetCanonicalName()
	if _, ok := theChaincodeSupport.runningChaincodes.chaincodeMap[canName]; ok {
		chaincodeStartLogger.Error("chaincode already exist!")
		temp := &ccResult{Status: CCError, Message: fmt.Sprintf("Install chaincode [%s] failed!chaincode already exist!", cccid.Name+"_"+cccid.Version), ChangedKv: nil, Response: nil}
		resB, _ := json.Marshal(temp)
		return string(resB)
	}
	_, _, err := theChaincodeSupport.Launch(context.Background(), cccid, content)
	if err != nil {
		chaincodeStartLogger.Errorf("Install chaincode:[%s] failed,err:%+v\n", canName, err)
		temp := &ccResult{Status: CCError, Message: fmt.Sprintf("Install chaincode [%s] failed!System Error!", cccid.Name+"_"+cccid.Version), ChangedKv: nil, Response: nil}
		resB, _ := json.Marshal(temp)
		return string(resB)
	}
	chaincodeStartLogger.Infof("Install chaincode [%s] success!\n", canName)
	//res.resLocker.Lock()
	//name := "chaincodename_" + chaincodeName
	//res.tempResult[name] = sign
	//res.singleTxResult[name] = sign
	//res.resLocker.Unlock()
	temp := &ccResult{Status: CCSuccess, Message: fmt.Sprintf("Install chaincode [%s] success!", cccid.Name+"_"+cccid.Version), ChangedKv: nil, Response: nil}
	resB, _ := json.Marshal(temp)
	return string(resB)
}

// chaincodeSpecialTxDelete 处理删除chaincode容器的交易
func chaincodeSpecialTxDelete(chaincodeName string, version string, txID string, deleteDocker bool, sign string) string {
	chaincodeStartLogger.Debug("enter ChaincodeSpecialTxDelete,delete sign: ", sign)
	cccid := ccprovider.NewCCContext("tjfoc", chaincodeName, version, txID, theChaincodeSupport.ip, theChaincodeSupport.port, false)
	canName := cccid.GetCanonicalName()
	if _, ok := theChaincodeSupport.runningChaincodes.chaincodeMap[canName]; !ok {
		chaincodeStartLogger.Error("chaincode doesn't exist!")
		temp := &ccResult{Status: CCError, Message: fmt.Sprintf("Delete chaincode [%s] failed!chaincode doesn't exist!", cccid.Name+"_"+cccid.Version), ChangedKv: nil, Response: nil}
		resB, _ := json.Marshal(temp)
		return string(resB)
	}
	err := theChaincodeSupport.Stop(context.Background(), cccid, deleteDocker)
	if err != nil {
		chaincodeStartLogger.Errorf("Delete container:[%s] failed err:%s\n", cccid.GetCanonicalName(), err)
		temp := &ccResult{Status: CCError, Message: fmt.Sprintf("Delete chaincode [%s] failed!", cccid.Name+"_"+cccid.Version), ChangedKv: nil, Response: nil}
		resB, _ := json.Marshal(temp)
		return string(resB)
	}

	//theChaincodeSupport.rebootingChaincodes.Lock()
	//delete(theChaincodeSupport.rebootingChaincodes.chaincodeMap, canName)
	//theChaincodeSupport.rebootingChaincodes.Unlock()

	chaincodeStartLogger.Infof("Delete chaincode [%s] success!\n", canName)

	//if hasChainCode(chaincodeName) {
	//还存在别的version的该chaincode
	temp := &ccResult{Status: CCSuccess, Message: fmt.Sprintf("Delete chaincode [%s] success!", cccid.Name+"_"+cccid.Version), ChangedKv: nil, Response: nil}
	resB, _ := json.Marshal(temp)
	return string(resB)
	//} else {
	//不存在别的version的该chaincode
	//res.resLocker.Lock()
	//name := "chaincodename_" + chaincodeName
	//res.tempResult[name] = ""
	//res.singleTxResult[name] = ""
	//res.resLocker.Unlock()
	//temp := &ccResult{Status: CCSuccess, Message: fmt.Sprintf("Delete chaincode [%s] success!", cccid.Name+"_"+cccid.Version), ChangedKv: res.singleTxResult, Response: nil}
	//resB, _ := json.Marshal(temp)
	//return string(resB)
	//}
}
func getContainerNameFromImageName(imageName string) (string, string, string, string) {
	index := strings.LastIndex(imageName, "-")
	if index == -1 {
		return "", "", "", ""
	}
	//去掉hash字符串
	containerName := imageName[:index]
	iname := getImageNameFromContainerName(containerName)
	if iname == imageName {
		//去掉port
		index = strings.LastIndex(containerName, "_")
		port := containerName[index+1:]
		containerName = containerName[:index]
		//去掉ip
		index = strings.LastIndex(containerName, "_")
		ip := containerName[index+1:]
		containerName = containerName[:index]
		//获得合约名字和版本
		index = strings.LastIndex(containerName, "_")
		chaincodeVersion := containerName[index+1:]
		chaincodeName := containerName[:index]
		return chaincodeName, chaincodeVersion, ip, port
	}
	return "", "", "", ""
}
func getImageNameFromContainerName(containerName string) string {
	imageName := fmt.Sprintf("%s-%s", containerName, hex.EncodeToString([]byte(containerName)))
	return imageName
}
