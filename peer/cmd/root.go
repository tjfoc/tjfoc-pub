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
package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	cfg "github.com/tjfoc/tjfoc/core/config"
)

const (
	Version = "1.0.1beta"
)

var cfgFile string
var Config cfg.PeerConfig

var RootCmd = &cobra.Command{
	Use:   "peer",
	Short: "peer",
	Long:  "Copyright 2018 Suzhou Tongji Fintech Research Institute",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

func init() {
	RootCmd.Flags().StringVarP(&cfgFile, "config", "c", "./conf/base.yaml", "config file path")
	initConfig()
	RootCmd.Flags().IntVarP(&Config.Rpc.Port, "port", "p", Config.Rpc.Port, "rpc port")
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	if err := viper.Unmarshal(&Config); err != nil {
		log.Fatal(err)
	}
}

func logInit() {
	flogging.LogTxt("logs/" + time.Now().Format("2006_01_02_15_04") + ".log")
	//日志默认级别
	flogging.SetModuleLevel("", Config.Log.Level)

	//其他模块日志级别需要单独显示的在 conf/base.yaml 配置
	flogging.SetModuleLevel("raft", viper.GetString("Log.RaftLevel"))
}
