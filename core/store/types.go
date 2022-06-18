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
	"bytes"

	"github.com/tjfoc/tjfoc/core/functor"
	"github.com/tjfoc/tjfoc/core/typeclass"
)

type storeFactor struct {
	key   []byte
	value []byte
}

type StoreFactor interface {
	typeclass.Eq
	GetKey() []byte
	GetValue() []byte
}

type Store interface {
	Close() error
	Del([]byte) error
	Set([]byte, []byte) error
	Get([]byte) ([]byte, error)
	BatchWrite(functor.Functor) error // 批量写
}

func (a *storeFactor) GetKey() []byte {
	return a.key
}

func (a *storeFactor) GetValue() []byte {
	return a.value
}

func (a *storeFactor) Eq(b typeclass.Eq) bool {
	v, ok := b.(*storeFactor)
	if !ok {
		return false
	}
	return bytes.Compare(a.key, v.key) == 0
}

func (a *storeFactor) NotEq(b typeclass.Eq) bool {
	return !a.Eq(b)
}
