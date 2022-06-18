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

func (a TestInt) LessThan(b typeclass.Ord) bool {
	v, ok := b.(TestInt)
	if !ok {
		return false
	}
	return a < v
}

func (a TestInt) MoreThan(b typeclass.Ord) bool {
	return !a.LessThan(b)
}

func intPrint(a interface{}, b interface{}) interface{} {
	fmt.Printf("%v\n", b)
	return a
}

func intAdd1(a interface{}) interface{} {
	x, _ := a.(TestInt)
	return x + 1
}

func intAdd(a interface{}, b interface{}) interface{} {
	x, _ := a.(TestInt)
	y, _ := b.(TestInt)
	return x + y
}

func Test(t *testing.T) {
	m := New()
	for i := 0; i < 5; i++ {
		m.Insert(TestInt(i), TestInt(i))
	}
	m.Foldl(intPrint, TestInt(1))
	m.Map(intAdd1)
	fmt.Printf("--------------------\n")
	m.Foldl(intPrint, TestInt(1))
	fmt.Printf("--------------------\n")
	x := m.Foldl(intAdd, TestInt(0))
	fmt.Printf("x = %v\n", x)
}
