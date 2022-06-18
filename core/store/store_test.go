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
package store

import (
	"fmt"
	"testing"

	"github.com/tjfoc/tjfoc/core/queue"
)

func TestStore(t *testing.T) {
	var s Store

	s, _ = NewMap()
	//s, _ = NewDb("test")
	q := queue.New()
	for i := 0; i < 10; i++ {
		q.PushBack(NewFactor([]byte(fmt.Sprintf("id_%d", i)), []byte(fmt.Sprintf("data_%d", i))))
	}
	s.BatchWrite(q)
	/*
		for i := 0; i < 10; i++ {
			s.Set([]byte(fmt.Sprintf("id_%d", i)), []byte(fmt.Sprintf("data_%d", i)))
		}
	*/
	for i := 0; i < 10; i++ {
		v, _ := s.Get([]byte(fmt.Sprintf("id_%d", i)))
		fmt.Printf("%s: %s\n", fmt.Sprintf("id_%d", i), string(v))
	}
}
