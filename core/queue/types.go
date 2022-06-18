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
package queue

import (
	"github.com/tjfoc/tjfoc/core/functor"
	"github.com/tjfoc/tjfoc/core/typeclass"
)

type QueueNode interface {
	typeclass.Eq
}

type queueNode struct {
	node QueueNode
	prev *queueNode
	next *queueNode
}

type Queue struct {
	queue *queueNode
}

func (a *Queue) Map(f functor.MapFunc) functor.Functor {
	return a.queue.Map(f)
}

func (b *Queue) Foldl(f functor.FoldFunc, a interface{}) interface{} {
	return b.queue.Foldl(f, a)
}

func (b *Queue) Foldr(f functor.FoldFunc, a interface{}) interface{} {
	return b.queue.Foldr(f, a)
}

func (a *queueNode) Map(f functor.MapFunc) functor.Functor {
	if a.isEmpty() {
		return a
	}
	next := a.Next()
	v, _ := f(a.node).(QueueNode)
	a.node = v
	next.Map(f)
	return a
}

func (b *queueNode) Foldl(f functor.FoldFunc, a interface{}) interface{} {
	if b.isEmpty() {
		return a
	}
	next := b.Next()
	return next.Foldl(f, f(a, b.node))
}

func (b *queueNode) Foldr(f functor.FoldFunc, a interface{}) interface{} {
	if b.isEmpty() {
		return a
	}
	next := b.Next()
	return f(next.Foldr(f, a), b.node)
}
