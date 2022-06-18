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
package proto

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestProto(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().Unix()))
	for i := 0; i < 1000000; i++ {
		a := &spProto{
			size:    rand.Uint32(),
			class:   rand.Uint32(),
			version: rand.Uint32(),
		}
		b, _ := spProtoToBytes(a)
		c, _ := bytesToSpProto(b)
		if a.size != c.size || a.class != c.class || a.version != c.version {
			fmt.Printf("%d: fail\n", i)
		}
	}
}
