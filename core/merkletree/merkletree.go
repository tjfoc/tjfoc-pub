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
	"crypto/sha256"
	"errors"
	"fmt"
)

type Merkletree_Data struct {
	Hash_str string
	Parent   *Merkletree_Data
	Left     *Merkletree_Data
	Right    *Merkletree_Data
}

type Merkletree struct {
	root          *Merkletree_Data
	treetall      uint64
	level_1       map[string]*Merkletree_Data
	list_1        []string
	level_2       map[string]*Merkletree_Data
	list_2        []string
	hash_queue    []string
	current_level uint8
	init_flag     bool
}

func (tree *Merkletree) Init() {
	fmt.Println("init")
	tree.root = nil
	tree.treetall = uint64(0)
	tree.level_1 = make(map[string]*Merkletree_Data, 0)
	tree.list_1 = make([]string, 0)
	tree.level_2 = make(map[string]*Merkletree_Data, 0)
	tree.list_2 = make([]string, 0)
	tree.hash_queue = make([]string, 0)
	tree.current_level = 0
	tree.init_flag = true
}
func (tree *Merkletree) GetRootHash() string {
	fmt.Println("getroothash")
	return tree.root.Hash_str
}
func (tree *Merkletree) GetTreeTall() uint64 {
	fmt.Println("gettreetall")
	return tree.treetall
}
func (tree *Merkletree) Push(new_hash string) error {
	fmt.Println("push")
	if !tree.init_flag {
		return errors.New("Merkletree was not initialized!")
	}
	tree.hash_queue = append(tree.hash_queue, new_hash)
	switch tree.current_level {
	case 0:
		new_tree_data := &Merkletree_Data{
			Hash_str: new_hash,
			Left:     nil,
			Right:    nil,
			Parent:   nil,
		}
		tree.root = new_tree_data
		tree.current_level = 1
		tree.level_1[new_hash] = new_tree_data
		tree.list_1 = append(tree.list_1, new_hash)
		tree.treetall++
	case 1:
		if tree.treetall == 1 {
			temp := tree.list_1[0]
			temp_tree_data := tree.level_1[temp]
			new_tree_data := &Merkletree_Data{
				Hash_str: new_hash,
				Left:     nil,
				Right:    nil,
				Parent:   nil,
			}
			new_root := &Merkletree_Data{
				Hash_str: "",
				Left:     nil,
				Right:    nil,
				Parent:   nil,
			}
			new_root.Left = new_tree_data
			new_root.Right = tree.root
			new_tree_data.Parent = new_root
			tree.root.Parent = new_root
			tree.root = new_root
			h := sha256.New()
			h.Write([]byte(tree.root.Left.Hash_str + tree.root.Right.Hash_str))
			tree.root.Hash_str = string(h.Sum(nil))
			tree.list_2 = append(tree.list_2, temp, new_hash)
			tree.level_2[temp] = temp_tree_data
			tree.level_2[new_hash] = new_tree_data
			delete(tree.level_1, temp)
			tree.list_1 = tree.list_1[:0]
			tree.current_level = 2
			tree.treetall++
		} else {
			new_left_data := tree.level_1[tree.list_1[0]]
			new_right_data := &Merkletree_Data{
				Hash_str: new_hash,
				Left:     nil,
				Right:    nil,
				Parent:   nil,
			}
			new_parent_data := &Merkletree_Data{
				Hash_str: "",
				Left:     nil,
				Right:    nil,
				Parent:   nil,
			}
			if new_left_data.Parent.Left == new_left_data {
				new_left_data.Parent.Left = new_parent_data
			} else {
				new_left_data.Parent.Right = new_parent_data
			}
			new_parent_data.Parent = new_left_data.Parent
			new_parent_data.Left = new_left_data
			new_parent_data.Right = new_right_data
			new_left_data.Parent = new_parent_data
			new_right_data.Parent = new_parent_data
			tree.list_2 = append(tree.list_2, tree.list_1[0], new_hash)
			tree.level_2[tree.list_1[0]] = new_left_data
			tree.level_2[new_hash] = new_right_data
			delete(tree.level_1, tree.list_1[0])
			tree.list_1 = tree.list_1[1:]
			if len(tree.list_1) == 0 {
				tree.treetall++
				tree.current_level = 2
			}
			for {
				h := sha256.New()
				h.Write([]byte(new_parent_data.Left.Hash_str + new_parent_data.Right.Hash_str))
				new_parent_data.Hash_str = string(h.Sum(nil))
				new_parent_data = new_parent_data.Parent
				if new_parent_data == nil {
					break
				}
			}
		}
	case 2:
		new_left_data := tree.level_2[tree.list_2[0]]
		new_right_data := &Merkletree_Data{
			Hash_str: new_hash,
			Left:     nil,
			Right:    nil,
			Parent:   nil,
		}
		new_parent_data := &Merkletree_Data{
			Hash_str: "",
			Left:     nil,
			Right:    nil,
			Parent:   nil,
		}
		if new_left_data.Parent.Left == new_left_data {
			new_left_data.Parent.Left = new_parent_data
		} else {
			new_left_data.Parent.Right = new_parent_data
		}
		new_parent_data.Parent = new_left_data.Parent
		new_parent_data.Left = new_left_data
		new_parent_data.Right = new_right_data
		new_left_data.Parent = new_parent_data
		new_right_data.Parent = new_parent_data
		tree.list_1 = append(tree.list_1, tree.list_2[0], new_hash)
		tree.level_1[tree.list_2[0]] = new_left_data
		tree.level_1[new_hash] = new_right_data
		delete(tree.level_2, tree.list_2[0])
		tree.list_2 = tree.list_2[1:]
		if len(tree.list_2) == 0 {
			tree.treetall++
			tree.current_level = 1
		}
		for {
			h := sha256.New()
			h.Write([]byte(new_parent_data.Left.Hash_str + new_parent_data.Right.Hash_str))
			new_parent_data.Hash_str = string(h.Sum(nil))
			new_parent_data = new_parent_data.Parent
			if new_parent_data == nil {
				break
			}
		}
	}
	return nil
}
func (tree *Merkletree) Pushn(new_hash []string) {
	fmt.Println("pushn")
	for _, v := range new_hash {
		tree.Push(v)
	}
}
func (tree *Merkletree) Rebuild(new_hash []string) {
	fmt.Println("rebuild")
	tree.Init()
	for _, v := range new_hash {
		tree.Push(v)
	}
}
func (tree *Merkletree) Search(search_hash string) bool {
	fmt.Println("search")
	_, ok := tree.level_1[search_hash]
	if !ok {
		_, ok = tree.level_2[search_hash]
		if !ok {
			return false
		}
	}
	return true
}
func (tree *Merkletree) Searchn(search_hash []string) []bool {
	fmt.Println("searchn")
	result := make([]bool, 0)
	for _, v := range search_hash {
		_, ok := tree.list_1[v]
		if !ok {
			_, ok = tree.level_2(v)
			if !ok {
				result = append(result, false)
			} else {
				result = append(result, true)
			}
		} else {
			result = append(result, true)
		}
	}
	return result
}
func (tree *Merkletree) Exoprt() []string {
	fmt.Println("export")
	return tree.hash_queue
}
