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

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/container/util"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/spf13/viper"
)

var mainLogger = flogging.MustGetLogger("createDocker")

func init() {
	viper.SetEnvPrefix("base")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetConfigName("base")
	viper.AddConfigPath("./conf")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../conf")
	viper.AddConfigPath("../../peer/conf")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		mainLogger.Errorf("Fatal error when reading %s config file: %s\n", "base", err)
	}
}

// existBaseImage 判断tjfoc-ccenv镜像是否存在
func existBaseImage() bool {
	client, err := util.NewDockerClient()
	if err != nil {
		panic(fmt.Errorf("fatal newClient err %s", err))
	}
	imgs, err := client.ListImages(docker.ListImagesOptions{All: false})
	if err != nil {
		panic(err)
	}
	var isExist = false
	for _, img := range imgs {
		for _, v := range img.RepoTags {
			if v == "tjfoc/tjfoc-ccenv:"+viper.GetString("Docker.BaseImageVersion") {
				isExist = true
				break
			}
		}
	}
	return isExist
}

func main() {
	newImage := "tjfoc/tjfoc-ccenv"
	version := viper.GetString("Docker.BaseImageVersion")
	if existBaseImage() {
		mainLogger.Warningf("docker images: %s:%s is already exist, please check if it is the latest version\n", newImage, version)
		return
	}
	//基础镜像
	baseImage := "golang:1.9"
	//编译tjcc的docker镜像名
	tjccDockerName := newImage + ":" + version
	//tjcc的源码，tar包
	tjfocSrcTarFile := getCurrentDirectory() + "/tjfoc.tar.gz"
	err := createImage(baseImage, tjccDockerName, tjfocSrcTarFile)
	if err != nil {
		mainLogger.Errorf("create docker error:%s\n", err)
		return
	}
}

//getCurrentDirectory 获取当前工作目录
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1) //将\替换成/
}

//createImage 创建一个docker镜像
func createImage(baseImage, newImageName string, tarfile string) error {

	client, err := util.NewDockerClient()
	if err != nil {
		panic(fmt.Errorf("fatal newClient err %s", err))
	}

	reader, err := generateDockerBuild(baseImage, tarfile)
	outputbuf := bytes.NewBuffer(nil)
	opts := docker.BuildImageOptions{
		Name:         newImageName, //指定新镜像的名字
		Pull:         false,
		InputStream:  reader,
		OutputStream: outputbuf,
	}
	if err = client.BuildImage(opts); err != nil {
		mainLogger.Errorf("Error building images: %s", err)
		mainLogger.Errorf("Image Output:\n********************\n%s\n********************", outputbuf.String())
		return err
	}
	mainLogger.Infof("created docker images  '%s' ok.\n", newImageName)
	return nil
}

//InputFiles 数组-存储dockerfile
type InputFiles map[string][]byte

//generateDockerBuild 创建dockerfile,并创建镜像
func generateDockerBuild(baseImage, tarfile string) (io.Reader, error) {
	inputFiles := make(InputFiles)

	var buf []string
	buf = append(buf, "FROM "+baseImage)
	buf = append(buf, "ADD binpackage.tar /go/src/github.com/")
	contents := strings.Join(buf, "\n")
	mainLogger.Infof("\n%s", contents)

	dockerFile := []byte(contents)
	inputFiles["Dockerfile"] = dockerFile

	input, output := io.Pipe()
	go func() {
		gw := gzip.NewWriter(output)
		tw := tar.NewWriter(gw)

		err := createDockerFile(inputFiles, tw, tarfile)
		if err != nil {
			mainLogger.Error(err)
		}
		tw.Close()
		gw.Close()
		output.CloseWithError(err)
	}()
	return input, nil
}

//创建 binpackage.tar 文件
func createDockerFile(inputFiles InputFiles, tw *tar.Writer, tarfile string) error {
	var err error
	for name, data := range inputFiles {
		err = util.WriteBytesToPackage(name, data, tw)
		mainLogger.Infof("finished WriteBytesToPackage name:%s", name)
		if err != nil {
			return fmt.Errorf("Failed to inject \"%s\": %s", name, err)
		}
	}

	//shim层相关代码
	bytes, err := ioutil.ReadFile(tarfile)
	if err != nil {
		return err
	}
	mainLogger.Infof("read test file len:%d\n", len(bytes))
	err = util.WriteBytesToPackage("binpackage.tar", bytes, tw)
	if err != nil {
		return fmt.Errorf("Failed to WriteBytesToPackage: %s", err)
	}
	return nil
}
