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
	"errors"
	"sync"

	"github.com/tjfoc/tjfoc/core/functor"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/tjmap"
	"github.com/tjfoc/tjfoc/core/typeclass"
)

type storeMapKey struct {
	key []byte
}

type storeMap struct {
	m *tjmap.Map
	l sync.RWMutex
}

func batchMapWrite(a, b interface{}) interface{} {
	if bs, ok := a.(*storeMap); ok {
		if v, ok := b.(StoreFactor); ok {
			bs.m.Update(&storeMapKey{v.GetKey()}, v.GetValue())
		}
	}
	return a
}

func NewMap() (*storeMap, error) {
	return &storeMap{
		m: tjmap.New(),
	}, nil
}

func (a *storeMap) Close() error {
	return nil
}

func (a *storeMap) Del(k []byte) error {
	if a == nil || a.m == nil {
		return errors.New("store map Del: null pointer")
	}
	a.l.Lock()
	defer a.l.Unlock()
	if ok := a.m.Remove(&storeMapKey{k}); ok {
		return nil
	}
	return errors.New("store map Del: key not exist")
}

func (a *storeMap) Set(k, v []byte) error {
	if a == nil || a.m == nil {
		return errors.New("store map Set: null pointer")
	}
	a.l.Lock()
	defer a.l.Unlock()
	if ok := a.m.Update(&storeMapKey{miscellaneous.Dup(k)},
		miscellaneous.Dup(v)); ok {
		return nil
	}
	return errors.New("store map Set: failed")
}

func (a *storeMap) Get(k []byte) ([]byte, error) {
	if a == nil || a.m == nil {
		return nil, errors.New("store map Get: null pointer")
	}
	a.l.RLock()
	defer a.l.RUnlock()
	b := a.m.Elem(&storeMapKey{k})
	if v, ok := b.([]byte); ok {
		return v, nil
	}
	return nil, errors.New("store map Get: key not exist")
}

func (a *storeMap) BatchWrite(b functor.Functor) error {
	if a == nil || a.m == nil {
		return errors.New("store map BatchWrite: null pointer")
	}
	a.l.Lock()
	defer a.l.Unlock()
	b.Foldl(batchMapWrite, a)
	return nil
}

func (a *storeMapKey) Eq(b typeclass.Eq) bool {
	v, ok := b.(*storeMapKey)
	if !ok {
		return false
	}
	return bytes.Compare(a.key, v.key) == 0
}

func (a *storeMapKey) NotEq(b typeclass.Eq) bool {
	return !a.Eq(b)
}

func (a *storeMapKey) LessThan(b typeclass.Ord) bool {
	v, ok := b.(*storeMapKey)
	if !ok {
		return false
	}
	return bytes.Compare(a.key, v.key) < 0
}

func (a *storeMapKey) MoreThan(b typeclass.Ord) bool {
	return !a.LessThan(b)
}
