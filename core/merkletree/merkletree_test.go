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
package merkletree

import (
	"fmt"
	"testing"
)

func TestMerkletree(t *testing.T) {
	tree := Merkletree{}
	tree.Init()
	new_hash1 := "123"
	new_hash2 := "456"
	new_hash3 := "789"
	new_hash4 := "abc"
	tree.Push(new_hash1)
	tree.Push(new_hash2)
	tree.Push(new_hash3)
	tree.Push(new_hash4)
	fmt.Println("tall: ", tree.GetTreeTall())
	fmt.Printf("root: %x\n", tree.GetRootHash())
	fmt.Println("export: ", tree.Exoprt())
	a := tree.Search(new_hash1)
	if a {
		fmt.Println("exist!")
	} else {
		fmt.Println("not exist!")
	}
	b := make([]string, 0)
	b = append(b, new_hash1)
	b = append(b, new_hash2)
	b = append(b, new_hash3)
	b = append(b, new_hash4)
	tree.Rebuild(b)
	fmt.Println("export: ", tree.Exoprt())
}
