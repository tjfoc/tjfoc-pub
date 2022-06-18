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
	"github.com/tjfoc/tjfoc/core/functor"
	"github.com/tjfoc/tjfoc/core/rbtree"
)

func New() *Map {
	return &Map{
		tree: rbtree.New(),
	}
}

//移除
func (m *Map) Remove(key MapKey) bool {
	return m.tree.Remove(MapNode{
		key: key,
	})
}

func (m *Map) Insert(key MapKey, value interface{}) bool {
	return m.tree.Insert(MapNode{
		key:   key,
		value: value,
	})
}

func (m *Map) Update(key MapKey, value interface{}) bool {
	return m.tree.Update(MapNode{key: key}, MapNode{
		key:   key,
		value: value,
	})
}

func (m *Map) Elem(key MapKey) interface{} {
	n := m.tree.Elem(MapNode{
		key: key,
	})
	if n == nil {
		return n
	}
	v, ok := n.(MapNode)
	if !ok {
		return nil
	}
	return v.value
}

func (m *Map) Map(f functor.MapFunc) functor.Functor {
	return m.tree.Map(func(a interface{}) interface{} {
		v, _ := a.(MapNode)
		v.value = f(v.value)
		return v
	})
}

func (m *Map) Foldl(f functor.FoldFunc, a interface{}) interface{} {
	return m.tree.Foldl(func(x interface{}, y interface{}) interface{} {
		v, _ := y.(MapNode)
		return f(x, v.value)
	}, a)
}

func (m *Map) Foldr(f functor.FoldFunc, a interface{}) interface{} {
	return m.tree.Foldr(func(x interface{}, y interface{}) interface{} {
		v, _ := y.(MapNode)
		return f(x, v.value)
	}, a)
}
