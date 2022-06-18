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

func New() *Queue {
	return &Queue{
		queue: nil,
	}
}

//清空
func (a *Queue) Reset() {
	a.queue = nil
}

//返回队列长度
func (a *Queue) Length() int {
	return a.queue.length()
}

//返回队列尾节点
func (a *Queue) Last() QueueNode {
	node := a.queue.last()
	if node != nil {
		return node.node
	}
	return nil
}

//返回队列头节点
func (a *Queue) Head() QueueNode {
	node := a.queue.head()
	if node != nil {
		return node.node
	}
	return nil
}

//返回第n个节点
func (a *Queue) Index(n int) QueueNode {
	node := a.queue.index(n)
	if node != nil {
		return node.node
	}
	return nil
}

//寻找b节点
func (a *Queue) Elem(b QueueNode) QueueNode {
	node := a.queue.elem(b)
	if node != nil {
		return node.node
	}
	return nil
}

// 去除队列的前n个元素
func (a *Queue) Drop(n int) {
	a.queue = a.queue.drop(n)
}

// 取队列的前n个元素
func (a *Queue) Take(n int) {
	if n == 0 {
		a.Reset()
	}
	a.queue = a.queue.take(n - 1)
}

//删除b节点
func (a *Queue) Remove(b QueueNode) {
	a.queue = a.queue.remove(b)
}

//从尾部插入b节点（对外）
func (a *Queue) PushBack(b QueueNode) {
	a.queue = a.queue.pushBack(b)
}

//从头部插入b节点（对外）
func (a *Queue) PushFront(b QueueNode) {
	a.queue = a.queue.pushFront(b)
}

//判断队列是否为空，空为true
func (a *queueNode) isEmpty() bool {
	return a == nil
}

//后节点
func (a *queueNode) Next() *queueNode {
	return a.next
}

//前节点
func (a *queueNode) Prev() *queueNode {
	return a.prev
}

//返回队列长度
func (a *queueNode) length() int {
	if a.isEmpty() {
		return 0
	}
	return 1 + a.Next().length()
}

//返回队列头节点
func (a *queueNode) head() *queueNode {
	if a.isEmpty() || a.Prev().isEmpty() {
		return a
	}
	return a.Prev().head()
}

//返回队列尾节点
func (a *queueNode) last() *queueNode {
	if a.isEmpty() || a.Next().isEmpty() {
		return a
	}
	return a.next.last()
}

//返回队列第N个节点
func (a *queueNode) index(n int) *queueNode {
	if a.isEmpty() || n <= 0 {
		return a
	}
	return a.Next().index(n - 1)
}

//寻找b节点
func (a *queueNode) elem(b QueueNode) *queueNode {
	if a.isEmpty() {
		return a
	}
	next := a.Next()
	if a.node.Eq(b) {
		return a
	}
	return next.elem(b)
}

//删除b节点
func (a *queueNode) remove(b QueueNode) *queueNode {
	if a.isEmpty() {
		return nil
	}
	if a.node.Eq(b) {
		a.Next().prev = nil
		return a.Next()
	}
	node := a.elem(b)
	if node == nil {
		return a
	}
	if !node.Prev().isEmpty() {
		node.Prev().next = node.Next()
	}
	if !node.Next().isEmpty() {
		node.Next().prev = node.Prev()
	}
	node.prev = nil
	node.next = nil
	return a
}

//从尾部插入b节点
func (a *queueNode) pushBack(b QueueNode) *queueNode {
	return link(a, &queueNode{
		prev: nil,
		next: nil,
		node: b,
	}, nil).head()
}

//从头部插入b节点
func (a *queueNode) pushFront(b QueueNode) *queueNode {
	return link(nil, &queueNode{
		prev: nil,
		next: nil,
		node: b,
	}, a).head()
}

//
func (a *queueNode) drop(n int) *queueNode {
	if a.isEmpty() {
		return a
	}
	if n <= 0 {
		a.prev = nil
		return a
	}
	return a.Next().drop(n - 1)
}

//
func (a *queueNode) take(n int) *queueNode {
	if a.isEmpty() {
		return a
	}
	if n <= 0 {
		a.next = nil
		return a
	}
	next := a.Next()
	return link(nil, a, next.take(n-1)).head()
}

//
func link(prev, node, next *queueNode) *queueNode {
	node.next = next
	if next != nil {
		next.prev = node
	}
	if !prev.isEmpty() {
		for !prev.Next().isEmpty() {
			prev = prev.Next()
		}
		prev.next = node
	}
	node.prev = prev
	return node
}
