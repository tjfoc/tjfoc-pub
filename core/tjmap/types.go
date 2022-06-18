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
package tjmap

import (
	"sync"

	"github.com/tjfoc/tjfoc/core/rbtree"
	"github.com/tjfoc/tjfoc/core/typeclass"
)

type MapKey interface {
	typeclass.Eq
	typeclass.Ord
}

type MapNode struct {
	key   MapKey
	value interface{}
}

type Map struct {
	lock sync.Mutex
	tree *rbtree.RBtree
}

//a 和 b MapNode是否相等
func (a MapNode) Eq(b typeclass.Eq) bool {
	v, ok := b.(MapNode)
	if !ok {
		return false
	}
	return a.key.Eq(v.key)
}

// a和 b是否不等
func (a MapNode) NotEq(b typeclass.Eq) bool {
	return !a.Eq(b)
}

//a 是否小于 b
func (a MapNode) LessThan(b typeclass.Ord) bool {
	v, ok := b.(MapNode)
	if !ok {
		return false
	}
	return a.key.LessThan(v.key)
}

//a是否大于b
func (a MapNode) MoreThan(b typeclass.Ord) bool {
	return !a.LessThan(b)
}
