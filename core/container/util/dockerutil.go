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

package util

import (
	"runtime"
	"strings"

	"github.com/fsouza/go-dockerclient"
	"github.com/spf13/viper"
)

//根据配置文件创建一个 dockerClient
func NewDockerClient() (client *docker.Client, err error) {
	endpoint := viper.GetString("Docker.Endpoint")
	client, err = docker.NewClient(endpoint)
	return
}

//系统平台架构
var archRemap = map[string]string{
	"amd64": "x86_64",
}

func getArch() string {
	if remap, ok := archRemap[runtime.GOARCH]; ok {
		return remap
	} else {
		return runtime.GOARCH
	}
}

func ParseDockerfileTemplate(template string) string {
	r := strings.NewReplacer()

	return r.Replace(template)
}

func GetDockerfileFromConfig(path string) string {
	return ParseDockerfileTemplate(viper.GetString(path))
}
