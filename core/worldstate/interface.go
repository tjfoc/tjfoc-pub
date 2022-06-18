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
package worldstate

import "github.com/tjfoc/tjfoc/proto"
import "fmt"

type SearchResult struct {
	Key   string
	Value string
	Err   error
}
type WorldStateData map[string]string

type Worldstate_interface interface {
	Search(key string) SearchResult                                      //搜索单个key
	Searchn(key []string) []SearchResult                                 //搜索多个key
	SearchPrefix(prefix string) (map[string]string, error)               //前缀查询
	SearchRange(start string, stop string) (map[string]string, error)    //范围查询
	GetRootHash() string                                                 //获取根hash
	Check(bIndex uint64, compareRootHash string, wd WorldStateData) bool //检查是否能插入，b_index是block的index用来表示新旧状态，wd是在worldstate中type的一个类型
	Push() bool                                                          //将缓存中数据插入数据库
	Update(peerid string) int                                            //向指定peer同步数据，0-同步失败，1-同步成功，2-两个节点间的数据一样，3-超时（10秒），4-同步成功，写入数据库失败（数据库error）
}

var wd *WorldState

func New(buckettreeTall uint64, dbPath string, sp proto.SP) {
	wd = new(WorldState)
	wd.initwd(buckettreeTall, dbPath, sp)
}

func GetWorldState() Worldstate_interface {
	if wd != nil {
		return wd
	} else {
		fmt.Println("worldstate is nil!Please init wroldstate first!")
		return nil
	}
}
