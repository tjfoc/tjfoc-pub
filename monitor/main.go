package main

import (
	"strings"
	"time"

	logging "github.com/tjfoc/tjfoc/core/common/flogging"

	"github.com/spf13/viper"
	"github.com/tjfoc/tjfoc/monitor/config"
	"github.com/tjfoc/tjfoc/monitor/serv"
)

func init() {
	config.InitConfig()
}

var logger = logging.MustGetLogger("main")

func main() {
	logInit()
	s, err := serv.NewService()
	if err != nil {
		return
	}

	go s.GrpcServerRun()
	//每隔3秒处理一下交易结果
	time.Sleep(time.Second * 2)
	logger.Info("====开始处理交易结果====")

	for {
		s.HandleTxResults()
		time.Sleep(time.Millisecond * 500)
	}
}

func logInit() {
	logging.SetModuleLevel("", strings.ToUpper(viper.GetString("Log.LogLevel")))
}
