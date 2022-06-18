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
	"fmt"
	"testing"

	"github.com/tjfoc/tjfoc/core/typeclass"
)

type TestInt int

func (a TestInt) Eq(b typeclass.Eq) bool {
	v, ok := b.(TestInt)
	if !ok {
		return false
	}
	return a == v
}

func (a TestInt) NotEq(b typeclass.Eq) bool {
	return !a.Eq(b)
}

func intPrint(a interface{}, b interface{}) interface{} {
	fmt.Printf("%v\n", b)
	return a
}

func Test(t *testing.T) {
	queue := New()
	for i := 0; i < 100; i++ {
		queue.PushBack(TestInt(i))
	}
	fmt.Printf("length: %d\n", queue.Length())
	for i := 0; i < 100; i = i + 2 {
		queue.Remove(TestInt(i))
	}
	fmt.Printf("length: %d\n", queue.Length())
	queue.Drop(20)
	fmt.Printf("length: %d\n", queue.Length())
	queue.Take(20)
	fmt.Printf("length: %d\n", queue.Length())
	queue.Foldl(intPrint, TestInt(1))
}
