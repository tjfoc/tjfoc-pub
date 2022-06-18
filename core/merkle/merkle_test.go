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
*/package merkle

import (
	"fmt"
	"testing"

	"github.com/tjfoc/gmsm/sm3"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
)

func TestMerkle(t *testing.T) {
	tree := MkMerkleTree(sm3.New(), [][]byte{[]byte("0"), []byte("1"), []byte("2"), []byte("3"), []byte("4")})
	hash := tree.GetMtHash()
	fmt.Printf("hash = %s\n", miscellaneous.Bytes2HexString(hash))
	fmt.Printf("VerifyNode = %v\n", tree.VerifyNode([]byte("4")))
}
