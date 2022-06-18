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
	"fmt"
	"os"

	"github.com/tjfoc/tjfoc/core/common/copydir"
	"github.com/tjfoc/tjfoc/core/common/zip"
)

// Gopath 获取当前环境的GOPATH
var Gopath string

func packageShim() {
	copydir.CopyDir(Gopath+"/src/github.com/tjfoc/tjfoc/core/chaincode", Gopath+"/src/github.com/tjfoc/tjfoc/tools/createdocker/tjfoc/tjfoc/core/chaincode")
	copydir.CopyDir(Gopath+"/src/github.com/tjfoc/tjfoc/core/common/ccprovider", Gopath+"/src/github.com/tjfoc/tjfoc/tools/createdocker/tjfoc/tjfoc/core/common/ccprovider")
	copydir.CopyDir(Gopath+"/src/github.com/tjfoc/tjfoc/core/common/flogging", Gopath+"/src/github.com/tjfoc/tjfoc/tools/createdocker/tjfoc/tjfoc/core/common/flogging")
	copydir.CopyDir(Gopath+"/src/github.com/tjfoc/tjfoc/core/container", Gopath+"/src/github.com/tjfoc/tjfoc/tools/createdocker/tjfoc/tjfoc/core/container")
	copydir.CopyDir(Gopath+"/src/github.com/tjfoc/tjfoc/vendor", Gopath+"/src/github.com/tjfoc/tjfoc/tools/createdocker/tjfoc/tjfoc/vendor")
	copydir.CopyDir(Gopath+"/src/github.com/tjfoc/tjfoc/protos", Gopath+"/src/github.com/tjfoc/tjfoc/tools/createdocker/tjfoc/tjfoc/protos")
	copydir.CopyDir(Gopath+"/src/github.com/tjfoc/gmsm", Gopath+"/src/github.com/tjfoc/tjfoc/tools/createdocker/tjfoc/gmsm")
	os.RemoveAll(Gopath + "/src/github.com/tjfoc/tjfoc/tools/createdocker/tjfoc/tjfoc/protos/raft")
	zip.Tar(Gopath+"/src/github.com/tjfoc/tjfoc/tools/createdocker/tjfoc", "../createdocker/tjfoc.tar.gz", false)
}

func main() {
	Gopath = os.Getenv("GOPATH")
	if Gopath == "" {
		fmt.Errorf("GOPATH is not set, please checkout $GOPATH!")
		return
	}
	packageShim()

	err := os.RemoveAll("../createdocker/tjfoc")
	if err != nil {
		fmt.Errorf("delete tjfoc dir:%s\n", err)
	}
}
