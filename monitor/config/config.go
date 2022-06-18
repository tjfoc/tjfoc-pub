package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

var configName = "./config.yaml"

func InitConfig() {
	viper.SetConfigFile("config.yaml")
	//viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("read config err:%s", err))
	}

	log.Println("init config success!")
}
