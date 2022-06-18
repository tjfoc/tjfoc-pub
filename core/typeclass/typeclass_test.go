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
package typeclass

import (
	"fmt"
	"log"
	"testing"

	"github.com/tjfoc/tjfoc/core/miscellaneous"
)

type TestInt struct {
	Val int64
}

func (a *TestInt) Eq(b Eq) bool {
	v, ok := b.(*TestInt)
	if !ok {
		return false
	}
	return a.Val == v.Val
}

func (a *TestInt) NotEq(b Eq) bool {
	return !a.Eq(b)
}

func (a *TestInt) LessThan(b Ord) bool {
	v, ok := b.(*TestInt)
	if !ok {
		return false
	}
	return a.Val < v.Val
}

func (a *TestInt) MoreThan(b Ord) bool {
	return !a.LessThan(b)
}

func (a *TestInt) Show() ([]byte, error) {
	return miscellaneous.Marshal(*a)
}

func (a *TestInt) Read(data []byte) ([]byte, error) {
	return []byte{}, miscellaneous.Unmarshal(data, a)
}

func TestTypeClass(t *testing.T) {
	a := &TestInt{1}
	b := &TestInt{2}
	as, _ := a.Show()
	bs, _ := b.Show()
	fmt.Printf("a = %v, b = %v\n", as, bs)
	fmt.Printf("== %v, != %v\n", a.Eq(b), a.NotEq(b))
	fmt.Printf("< %v, > %v\n", a.LessThan(b), a.MoreThan(b))
	c := &TestInt{10}
	cs, _ := c.Show()
	_, err := a.Read(cs)
	if err != nil {
		log.Fatalf("failed to read: %v", err)
	}
	as, _ = a.Show()
	fmt.Printf("a update %v, %v\n", *a, as)
}
