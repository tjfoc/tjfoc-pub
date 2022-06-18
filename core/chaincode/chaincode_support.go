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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/tjfoc/tjfoc/core/common/ccprovider"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/container"
	"github.com/tjfoc/tjfoc/core/container/api"
	"github.com/tjfoc/tjfoc/core/container/util"
	pb "github.com/tjfoc/tjfoc/protos/chaincode"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/golang/protobuf/proto"
	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
)

// 一些全局默认的配置,可通过配置文件读取
const (
	DevModeUserRunsChaincode       string = "dev"
	chaincodeStartupTimeoutDefault int    = 5000
	peerIpDefault                  string = "0.0.0.0"
	peerPortDefault                string = "7051"
)

//ChaincodeSupport 基本结构
type ChaincodeSupport struct {
	runningChaincodes   *runningChaincodes
	rebootingChaincodes *rebootingChaincodes
	ip                  string
	port                string
	ccStartupTimeout    time.Duration
	peerTLSCertFile     string
	peerTLSKeyFile      string
	peerTLSSvrHostOrd   string
	keepalive           time.Duration
	chaincodeLogLevel   string
	shimLogLevel        string
	logFormat           string
	executetimeout      time.Duration
	userRunsCC          bool
	peerTLS             bool
}

type rebootingChaincodes struct {
	sync.Mutex
	chaincodeMap map[string]*rebootInfo
}

type rebootInfo struct {
	needReboot bool
	nextstate  *nextStateInfo
	txid       string
	txctx      *transactionContext
}

type runningChaincodes struct {
	sync.RWMutex
	chaincodeMap  map[string]*chaincodeRTEnv
	launchStarted map[string]bool
}

type chaincodeRTEnv struct {
	handler *Handler
}

var chaincodeLogger = flogging.MustGetLogger("chaincode_support")

var theChaincodeSupport *ChaincodeSupport

//NewChaincodeSupport ,用来实现chaincode_shim的grpc服务注册
func NewChaincodeSupport() *ChaincodeSupport {
	theChaincodeSupport = &ChaincodeSupport{
		runningChaincodes:   &runningChaincodes{chaincodeMap: make(map[string]*chaincodeRTEnv), launchStarted: make(map[string]bool)},
		rebootingChaincodes: &rebootingChaincodes{chaincodeMap: make(map[string]*rebootInfo)},
	}

	theChaincodeSupport.ip = viper.GetString("Docker.Ip")
	theChaincodeSupport.port = viper.GetString("Docker.Port")
	if theChaincodeSupport.ip == "" {
		theChaincodeSupport.ip = peerIpDefault
	}
	if theChaincodeSupport.port == "" {
		theChaincodeSupport.port = peerPortDefault
	}

	theChaincodeSupport.keepalive = time.Duration(0) * time.Second

	execto := time.Duration(30) * time.Second
	if eto := viper.GetDuration("Chaincode.Timeout"); eto <= time.Duration(1)*time.Second {
		chaincodeLogger.Warningf("Invalid execute timeout value %s (should be at least 1s); defaulting to %s", eto, execto)
	} else {
		chaincodeLogger.Infof("Setting execute timeout value to %s", eto)
		execto = eto
	}
	ccstartuptimeout := viper.GetDuration("Chaincode.Timeout")
	theChaincodeSupport.executetimeout = execto
	theChaincodeSupport.ccStartupTimeout = ccstartuptimeout

	return theChaincodeSupport
}

func getLogLevelFromViper(module string) string {
	levelString := viper.GetString("chaincode.logging." + module)
	_, err := logging.LogLevel(levelString)

	if err == nil {
		chaincodeLogger.Debugf("CORE_CHAINCODE_%s set to level %s", strings.ToUpper(module), levelString)
	} else {
		chaincodeLogger.Warningf("CORE_CHAINCODE_%s has invalid log level %s. defaulting to %s", strings.ToUpper(module), levelString, flogging.DefaultLevel())
		levelString = flogging.DefaultLevel()
	}
	return levelString
}

//Register chaincode_shim的grpc实现方法
func (chaincodeSupport *ChaincodeSupport) Register(stream pb.ChaincodeSupport_RegisterServer) error {
	chaincodeLogger.Debug("entry Register")
	return HandleChaincodeStream(chaincodeSupport, stream.Context(), stream)
}

//registerHandler shim注册时会调用，用来管理运行的chaincode
func (chaincodeSupport *ChaincodeSupport) registerHandler(chaincodehandler *Handler) error {
	chaincodeLogger.Debug("entry registerHandler")

	key := chaincodehandler.ChaincodeID.Name

	chaincodeSupport.runningChaincodes.Lock()
	chaincodehandler.readyNotify = theChaincodeSupport.runningChaincodes.chaincodeMap[key].handler.readyNotify
	theChaincodeSupport.runningChaincodes.chaincodeMap[key].handler = chaincodehandler
	chaincodehandler.registered = true
	chaincodehandler.txCtxs = make(map[string]*transactionContext)
	chaincodehandler.txidMap = make(map[string]bool)
	chaincodeSupport.runningChaincodes.Unlock()

	chaincodeSupport.rebootingChaincodes.Lock()
	if _, ok := chaincodeSupport.rebootingChaincodes.chaincodeMap[key]; !ok {
		chaincodeSupport.rebootingChaincodes.chaincodeMap[key] = &rebootInfo{
			needReboot: false,
		}
	}
	chaincodeSupport.rebootingChaincodes.Unlock()

	chaincodeLogger.Infof("registered handler complete for chaincode [%s]!", key)
	return nil
}

//Execute 执行chaincode 并返回执行结果
func (chaincodeSupport *ChaincodeSupport) Execute(ctxt context.Context, cccid *ccprovider.CCContext, msg *pb.ChaincodeMessage, timeout time.Duration) (*pb.ChaincodeMessage, error) {
	chaincodeLogger.Debug("entry Execute")
	canName := cccid.GetCanonicalName()
	var notfy chan *pb.ChaincodeMessage
	var err error
	if notfy, err = theChaincodeSupport.runningChaincodes.chaincodeMap[canName].handler.sendExecuteMessage(ctxt, cccid.ChainID, msg); err != nil {
		return nil, fmt.Errorf("Error sending %s: %s", msg.Type.String(), err)
	}
	var ccresp *pb.ChaincodeMessage
	select {
	case ccresp = <-notfy:
	case <-time.After(timeout):
	}
	theChaincodeSupport.runningChaincodes.chaincodeMap[canName].handler.deleteTxContext(msg.Txid)
	return ccresp, err
}

//Stop 停止正在运行的chaincode
func (chaincodeSupport *ChaincodeSupport) Stop(context context.Context, cccid *ccprovider.CCContext, deleteDocker bool) error {
	canName := cccid.GetCanonicalName()
	if canName == "" {
		return fmt.Errorf("chaincode name not set")
	}
	sir := container.StopImageReq{
		ContainerName:  canName,
		Timeout:        10,
		DonDeleteImage: deleteDocker,
	}
	_, err := container.VMCProcess(context, sir)
	if err != nil {
		err = fmt.Errorf("Error stopping container: %s", err)
	}
	chaincodeSupport.runningChaincodes.Lock()
	delete(chaincodeSupport.runningChaincodes.chaincodeMap, canName)
	chaincodeSupport.runningChaincodes.Unlock()

	theChaincodeSupport.rebootingChaincodes.Lock()
	delete(theChaincodeSupport.rebootingChaincodes.chaincodeMap, canName)
	theChaincodeSupport.rebootingChaincodes.Unlock()
	return err
}

//chaincode 容器起来以后，会向此 Peer发送注册操作
//此处创建了一个通道，用来监控等待 chaincode 发送注册指令
func (chaincodeSupport *ChaincodeSupport) preLaunchSetup(chaincode string) chan bool {
	chaincodeSupport.runningChaincodes.Lock()
	defer chaincodeSupport.runningChaincodes.Unlock()
	//register placeholder Handler. This will be transferred in registerHandler
	//NOTE: from this point, existence of handler for this chaincode means the chaincode
	//is in the process of getting started (or has been started)
	notfy := make(chan bool, 1)
	chaincodeSupport.runningChaincodes.chaincodeMap[chaincode] = &chaincodeRTEnv{handler: &Handler{readyNotify: notfy}}
	return notfy
}

//Launch 启动Chaincode容器，并完成chaincode跟peer的grpc链接
//如果已经启动，则仅仅返回客户端参数pb.Input
func (chaincodeSupport *ChaincodeSupport) Launch(context context.Context, cccid *ccprovider.CCContext, srcContent []byte) (*pb.ChaincodeID, *pb.ChaincodeInput, error) {
	var cID *pb.ChaincodeID
	var cMsg *pb.ChaincodeInput
	builder := func() (io.Reader, error) { return GenerateDockerBuild(srcContent, cccid.Name, cccid.Version) }
	err := chaincodeSupport.launchAndWaitForRegister(context, cccid, builder)
	if err == nil {
		err = chaincodeSupport.sendReady(context, cccid, chaincodeSupport.ccStartupTimeout)
		if err != nil {
			chaincodeLogger.Errorf("sending init failed(%s)", err)
			err = fmt.Errorf("Failed to init chaincode(%s)", err)
		}
		chaincodeLogger.Debug("sending init completed")
	} else {
		return cID, cMsg, err
	}
	chaincodeLogger.Debug("LaunchChaincode complete")
	//注册完成之后恢复之前触发重启的交易
	theChaincodeSupport.rebootingChaincodes.Lock()
	if theChaincodeSupport.rebootingChaincodes.chaincodeMap[cccid.GetCanonicalName()].needReboot {
		chaincodeLogger.Info("this is a reboot container,register end retry tx!")
		temptxid := theChaincodeSupport.rebootingChaincodes.chaincodeMap[cccid.GetCanonicalName()].txid
		temptxctx := theChaincodeSupport.rebootingChaincodes.chaincodeMap[cccid.GetCanonicalName()].txctx
		theChaincodeSupport.runningChaincodes.chaincodeMap[cccid.GetCanonicalName()].handler.txCtxs[temptxid] = temptxctx
		theChaincodeSupport.runningChaincodes.chaincodeMap[cccid.GetCanonicalName()].handler.nextState <- theChaincodeSupport.rebootingChaincodes.chaincodeMap[cccid.GetCanonicalName()].nextstate
	} else {
		chaincodeLogger.Info("this is not a reboot container!")
	}
	theChaincodeSupport.rebootingChaincodes.Unlock()
	return cID, cMsg, err
}

//launchAndWaitForRegister 启动一个运行chaincode的docker容器，并等待docker容器[chaincode] 发送注册指令
func (chaincodeSupport *ChaincodeSupport) launchAndWaitForRegister(ctxt context.Context, cccid *ccprovider.CCContext, builder api.BuildSpecFactory) error {
	chaincodeLogger.Debug("create container, start container , build docker ")
	//chaincodeName:version
	canName := cccid.GetCanonicalName()

	//此通道等待chaincode 向次peer端发送注册指令
	var notfy chan bool
	preLaunchFunc := func() error {
		notfy = chaincodeSupport.preLaunchSetup(canName)
		return nil
	}
	shimLogging := viper.GetString("Docker.Loglevel")
	//此参数为测试参数，正常是启动某个服务（会一直阻塞），这里做测试写个死循环，保证容器不退出
	//args := []string{"bash", "-c", "cd /tmp;  while true; do sleep 20170504; done"}
	args := []string{"chaincode", theChaincodeSupport.ip + ":" + theChaincodeSupport.port, canName, cccid.Name, shimLogging}

	//环境变量
	envs := []string{"CORE_CHAINCODE_LOGGING_SHIM=debug"}
	envs = append(envs, "CORE_PEER_TLS_ENABLED=true")

	//创建启动 docker 容器结构体
	sir := container.StartImageReq{
		ContainerName: canName,
		Builder:       builder,
		Args:          args,
		Env:           envs,
		PrelaunchFunc: preLaunchFunc,
	}
	ipcCtxt := context.WithValue(ctxt, "CCHANDLER", chaincodeSupport)

	//通过该函数调用container中的的do函数，执行dockercontroler中的start函数
	resp, err := container.VMCProcess(ipcCtxt, sir)
	if err != nil || (resp != nil && resp.(container.VMCResp).Err != nil) {
		if err == nil {
			err = resp.(container.VMCResp).Err
		}
		err = fmt.Errorf("Error starting container: %s", err)

		//启动失败
		chaincodeSupport.runningChaincodes.Lock()
		delete(chaincodeSupport.runningChaincodes.chaincodeMap, canName)
		chaincodeSupport.runningChaincodes.Unlock()
		return err
	}

	//启动成功，等待chaincode容器 发送注册指令
	chaincodeLogger.Debugf("waiting for chaincode register...")
	select {
	case ok := <-notfy:
		chaincodeLogger.Infof("xxx in case . ok:%v", ok)
		if !ok {
			err = fmt.Errorf("registration failed for %s(tx:%s)", canName, cccid.TxID)
		}
	case <-time.After(chaincodeSupport.ccStartupTimeout):
		err = fmt.Errorf("Timeout expired while starting chaincode %s(tx:%s)", canName, cccid.TxID)
	}
	if err != nil {
		chaincodeLogger.Debugf("stopping due to error while launching %s", err)
		// errIgnore := chaincodeSupport.Stop(ctxt, cccid, cds)
		// if errIgnore != nil {
		// 	chaincodeLogger.Debugf("error on stop %s(%s)", errIgnore, err)
		// }
		chaincodeLogger.Info("等待超时，需要关闭docker。。")
	}
	return err
}

// sendReady 发送ready消息
func (chaincodeSupport *ChaincodeSupport) sendReady(context context.Context, cccid *ccprovider.CCContext, timeout time.Duration) error {
	canName := cccid.GetCanonicalName()
	chaincodeSupport.runningChaincodes.Lock()
	var chrte *chaincodeRTEnv
	var ok bool
	if chrte, ok = chaincodeSupport.runningChaincodes.chaincodeMap[canName]; !ok {
		chaincodeSupport.runningChaincodes.Unlock()
		chaincodeLogger.Debugf("handler not found for chaincode %s", canName)
		return fmt.Errorf("handler not found for chaincode %s", canName)
	}
	chaincodeSupport.runningChaincodes.Unlock()

	var notfy chan *pb.ChaincodeMessage
	var err error
	if notfy, err = chrte.handler.ready(context, cccid.ChainID, cccid.TxID); err != nil {
		return fmt.Errorf("Error sending %s: %s", pb.ChaincodeMessage_READY, err)
	}
	if notfy != nil {
		select {
		case ccMsg := <-notfy:
			chaincodeLogger.Debugf("case notfy. %s", ccMsg.Type.String())
			if ccMsg.Type == pb.ChaincodeMessage_ERROR {
				err = fmt.Errorf("Error initializing container %s: %s", canName, string(ccMsg.Payload))
			}
			if ccMsg.Type == pb.ChaincodeMessage_COMPLETED {
				// res := &pb.Response{}
				// _ = proto.Unmarshal(ccMsg.Payload, res)
				// if res.Status != shim.OK {
				// 	err = fmt.Errorf("Error initializing container %s: %s", canName, string(res.Message))
				// }
				// TODO
				// return res so that endorser can anylyze it.
			}
		case <-time.After(timeout):
			chaincodeLogger.Error("Timeout")
			err = fmt.Errorf("Timeout expired while executing send init message")
		}
	}

	chrte.handler.deleteTxContext(cccid.TxID)

	return err
}

//创建一个新的 ChaincodeMessage，用来跟CC通信
func createCCMessage(typ pb.ChaincodeMessage_Type, txid string, cMsg *pb.ChaincodeInput) (*pb.ChaincodeMessage, error) {
	payload, err := proto.Marshal(cMsg)
	if err != nil {
		return nil, err
	}
	return &pb.ChaincodeMessage{Type: typ, Payload: payload, Txid: txid}, nil
}

// InputFiles 暂存dockerfile
type InputFiles map[string][]byte

var dockerBuildLogger = flogging.MustGetLogger("chaincode_support_build")

//GenerateDockerBuild 根据基础镜像生成需要的镜像
func GenerateDockerBuild(srcContent []byte, chaincodeName, chaincodeVersion string) (io.Reader, error) {

	inputFiles := make(InputFiles)
	var BaseDockerLabel = "tjfoc.tjcc"
	var buf []string
	buf = append(buf, "FROM "+"tatsushid/tinycore:latest")
	buf = append(buf, "ADD binpackage.tar /usr/local/bin")
	buf = append(buf, fmt.Sprintf("LABEL %s.chaincode.id.name=\"%s\" \\", BaseDockerLabel, chaincodeName))
	buf = append(buf, fmt.Sprintf("      %s.chaincode.id.version=\"%s\" \\", BaseDockerLabel, chaincodeVersion))
	buf = append(buf, fmt.Sprintf("      %s.chaincode.type=\"%s\" \\", BaseDockerLabel, "GOLANG"))
	buf = append(buf, fmt.Sprintf("      %s.version=\"%s\" \\", BaseDockerLabel, "1.0.0"))
	buf = append(buf, fmt.Sprintf("      %s.base.version=\"%s\"", BaseDockerLabel, "0.3.1"))
	buf = append(buf, fmt.Sprintf("ENV CORE_CHAINCODE_BUILDLEVEL=%s", "1.0.0"))
	contents := strings.Join(buf, "\n")
	dockerFile := []byte(contents)
	inputFiles["Dockerfile"] = dockerFile
	input, output := io.Pipe()
	go func() {
		gw := gzip.NewWriter(output)
		tw := tar.NewWriter(gw)
		//上述的 dockerFile需要一个 binpackage.tar 文件
		err := generateDockerBuild(inputFiles, tw, srcContent, chaincodeName)
		tw.Close()
		gw.Close()
		output.CloseWithError(err)
	}()
	return input, nil
}

func generateDockerBuild(inputFiles InputFiles, tw *tar.Writer, srcContent []byte, chaincodeName string) error {
	var err error
	for name, data := range inputFiles {
		err = util.WriteBytesToPackage(name, data, tw)
		dockerBuildLogger.Infof("finished WriteBytesToPackage name:%s", name)
		if err != nil {
			return fmt.Errorf("Failed to inject \"%s\": %s", name, err)
		}
	}
	if err = GenerateDockerBuildGolang(tw, srcContent, chaincodeName); err != nil {
		return err
	}
	return nil
}

//GenerateDockerBuildGolang 根据dockerfile生成镜像
func GenerateDockerBuildGolang(tw *tar.Writer, srcContent []byte, chaincodeName string) error {
	dockerBuildLogger.Infof("in GenerateDockerBuildGolang\n")

	if bytes := DockerBuild(srcContent, chaincodeName); bytes == nil {
		return errors.New("compile golang code failed in chaincode environment docker!")
	} else {
		dockerBuildLogger.Infof("read test file len:%d\n", len(bytes))
		util.WriteBytesToPackage("binpackage.tar", bytes, tw)
	}
	return nil
}

//DockerBuild 创建镜像
func DockerBuild(srcContent []byte, chaincodeName string) []byte {
	envImage := viper.GetString("Docker.BaseImage") + ":" + viper.GetString("Docker.Version")
	//chaincode在docker容器中的路径
	pkgname := "/home/" + chaincodeName + ".go"
	//编译chaincode的参数
	const ldflags = "-linkmode external -extldflags '-static'"

	envs := []string{}
	cmd := fmt.Sprintf("GOPATH=/home:$GOPATH go build -ldflags \"%s\" -o /boot/chaincode %s", ldflags, pkgname)

	client, err := util.NewDockerClient()
	if err != nil {
		panic(fmt.Errorf("fatal newClient err %s\n", err))
	}

	_, err = client.InspectImage(envImage)
	if err != nil {
		panic(fmt.Errorf("fatal InspectImage err %s\n", err))
	}

	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        envImage,
			Env:          envs,
			Cmd:          []string{"/bin/sh", "-c", cmd},
			AttachStdout: true,
			AttachStderr: true,
		},
	})
	if err != nil {
		dockerBuildLogger.Errorf("Error creating container: %s\n", err)
	}

	codepackage := bytes.NewReader(srcContent)
	err = client.UploadToContainer(container.ID, docker.UploadToContainerOptions{
		Path:        "/home",
		InputStream: codepackage,
	})
	if err != nil {
		dockerBuildLogger.Errorf("Error uploading input to container: %s\n", err)
	}

	stdout := bytes.NewBuffer(nil)
	_, err = client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    container.ID,
		OutputStream: stdout,
		ErrorStream:  stdout,
		Logs:         true,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
	})
	if err != nil {
		dockerBuildLogger.Errorf("Error attaching to container: %s\n", err)
	}

	err = client.StartContainer(container.ID, nil)
	if err != nil {
		dockerBuildLogger.Errorf("Error StartContainer : %s \"%s\"", err, stdout.String())
	}

	retval, err := client.WaitContainer(container.ID)
	if err != nil {
		dockerBuildLogger.Errorf("Error waiting for container to complete: %s", err)
	}
	if retval > 0 {
		dockerBuildLogger.Errorf("Error returned from build: %d \"%s\"", retval, stdout.String())
		return nil
	}

	binpackage := bytes.NewBuffer(nil)
	err = client.DownloadFromContainer(container.ID, docker.DownloadFromContainerOptions{
		Path:         "/boot/.",
		OutputStream: binpackage,
	})

	if err != nil {
		dockerBuildLogger.Errorf("Error downloading output: %s", err)
	}
	dockerBuildLogger.Infof("downloaded file len %d\n", binpackage.Len)

	return binpackage.Bytes()
}
