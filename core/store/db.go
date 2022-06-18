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
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tjfoc/tjfoc/core/functor"
)

type storeDb struct {
	db *leveldb.DB
}

func batchDbWrite(a, b interface{}) interface{} {
	if bc, ok := a.(*leveldb.Batch); ok {
		if v, ok := b.(StoreFactor); ok {
			bc.Put(v.GetKey(), v.GetValue())
		}
	}
	return a
}

func NewDb(path string) (*storeDb, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &storeDb{db}, nil
}

func (a *storeDb) Close() error {
	if a == nil || a.db == nil {
		return errors.New("store db Close: null pointer")
	}
	err := a.db.Close()
	defer func() { a.db = nil }()
	return err
}

func (a *storeDb) Del(k []byte) error {
	if a == nil || a.db == nil {
		return errors.New("store db Del: null pointer")
	}
	return a.db.Delete(k, nil)
}

func (a *storeDb) Set(k, v []byte) error {
	if a == nil || a.db == nil {
		return errors.New("store db Set: null pointer")
	}
	return a.db.Put(k, v, nil)
}

func (a *storeDb) Get(k []byte) ([]byte, error) {
	if a == nil || a.db == nil {
		return nil, errors.New("store db Get: null pointer")
	}
	return a.db.Get(k, nil)
}

func (a *storeDb) BatchWrite(b functor.Functor) error {
	if a == nil || a.db == nil {
		return errors.New("store db BatchWrite: null pointer")
	}
	bc := new(leveldb.Batch)
	b.Foldl(batchDbWrite, bc)
	return a.db.Write(bc, nil)
}
