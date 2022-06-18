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
package buckettree

import (
	"crypto/sha256"
	"hash"
)

type Bucket struct {
	HashStr string
	Index   uint64
	Parent  *Bucket
	Left    *Bucket
	Right   *Bucket
}
type BucketTree struct {
	buckets  map[uint64]*Bucket
	treetall uint64
	root     *Bucket
	encoder  hash.Hash
}

func (b *BucketTree) Init(treeTall uint64) {
	b.buckets = make(map[uint64]*Bucket, 0)
	b.treetall = treeTall
	b.encoder = sha256.New()

	b.encoder.Write([]byte(""))
	tempStr := string(b.encoder.Sum(nil))
	b.encoder.Reset()
	for i := uint64(0); i < uint64(1<<treeTall); i++ {
		b.buckets[i] = &Bucket{
			HashStr: tempStr,
			Index:   i,
			Parent:  nil,
			Left:    nil,
			Right:   nil,
		}
	}

	if treeTall > 0 {
		temp := make([]*Bucket, 0)
		v, _ := b.buckets[0]
		b.encoder.Write([]byte(v.HashStr + v.HashStr))
		tempStr = string(b.encoder.Sum(nil))
		b.encoder.Reset()
		for i := uint64(0); i < uint64(1<<treeTall); i += 2 {
			v1, _ := b.buckets[i]
			v2, _ := b.buckets[i+1]
			tempBucket := &Bucket{
				HashStr: tempStr,
				Index:   0,
				Parent:  nil,
				Left:    nil,
				Right:   nil,
			}
			tempBucket.Left = v1
			v1.Parent = tempBucket
			tempBucket.Right = v2
			v2.Parent = tempBucket
			temp = append(temp, tempBucket)
		}
		for i := (treeTall - 1); i > 0; i-- {
			v = temp[0]
			b.encoder.Write([]byte(v.HashStr + v.HashStr))
			tempStr = string(b.encoder.Sum(nil))
			b.encoder.Reset()
			for j := 0; j < (1 << i); j += 2 {
				v1 := temp[j]
				v2 := temp[j+1]
				tempBucket := &Bucket{
					HashStr: tempStr,
					Index:   0,
					Parent:  nil,
					Left:    nil,
					Right:   nil,
				}
				tempBucket.Left = v1
				v1.Parent = tempBucket
				tempBucket.Right = v2
				v2.Parent = tempBucket
				temp = append(temp, tempBucket)
			}
			temp = temp[(1 << i):]
		}
		b.root = temp[0]
	} else {
		b.root = b.buckets[0]
	}
}
func (b *BucketTree) UpdateSingle(index uint64, newHash string) {
	if index < uint64(1<<b.treetall) && index >= 0 {
		v, _ := b.buckets[index]
		v.HashStr = newHash
		tempParent := v.Parent
		for tempParent != nil {
			b.encoder.Write([]byte(tempParent.Left.HashStr + tempParent.Right.HashStr))
			tempParent.HashStr = string(b.encoder.Sum(nil))
			b.encoder.Reset()
			tempParent = tempParent.Parent
		}
	}
}
func (b *BucketTree) Rebuild(temp []string) {
	for i := uint64(0); i < uint64(1<<b.treetall); i++ {
		b.buckets[i].HashStr = temp[i]
	}
	if b.treetall > 0 {
		b.UpdateAll()
	}
}
func (b *BucketTree) UpdateAll() {
	temp := make([]*Bucket, 0)
	for i := uint64(0); i < uint64(1<<b.treetall); i += 2 {
		v1, _ := b.buckets[i]
		v2, _ := b.buckets[i+1]
		b.encoder.Write([]byte(v1.HashStr + v2.HashStr))
		tempStr := string(b.encoder.Sum(nil))
		b.encoder.Reset()
		tempBucket := v1.Parent
		tempBucket.HashStr = tempStr
		temp = append(temp, tempBucket)
	}
	for i := (b.treetall - 1); i > 0; i-- {
		for j := 0; j < (1 << i); j += 2 {
			v1 := temp[j]
			v2 := temp[j+1]
			b.encoder.Write([]byte(v1.HashStr + v2.HashStr))
			tempStr := string(b.encoder.Sum(nil))
			b.encoder.Reset()
			tempBucket := v1.Parent
			tempBucket.HashStr = tempStr
			temp = append(temp, tempBucket)
		}
		temp = temp[(1 << i):]
	}
}
func (b *BucketTree) GetRootHash() string {
	return b.root.HashStr
}
func (b *BucketTree) GetIndexHash(i uint64) string {
	return b.buckets[i].HashStr
}
func (b *BucketTree) GetTreeTall() uint64 {
	return b.treetall
}
func (b *BucketTree) CompareRoot(roothash string) bool {
	return roothash == b.root.HashStr
}
func (b *BucketTree) SearchDifferent(temp []string) []uint64 {
	var temptree BucketTree
	temptree.Init(b.treetall)
	temptree.Rebuild(temp)
	if b.root.HashStr == temptree.root.HashStr {
		return nil
	}
	index := make([]uint64, 0)
	compare(b.root, temptree.root, &index)
	return index
}
func compare(local *Bucket, other *Bucket, index *[]uint64) {
	localLeft := local.Left
	localRight := local.Right
	otherLeft := other.Left
	otherRight := other.Right
	if localLeft == nil && localRight == nil && otherLeft == nil && otherRight == nil {
		*index = append(*index, local.Index)
	} else {
		if localLeft.HashStr != otherLeft.HashStr {
			compare(localLeft, otherLeft, index)
		}
		if localRight.HashStr != otherRight.HashStr {
			compare(localRight, otherRight, index)
		}
	}
}
func (b *BucketTree) Export() (temp []string) {
	for i := uint64(0); i < uint64(1<<b.treetall); i++ {
		v, _ := b.buckets[i]
		temp = append(temp, v.HashStr)
	}
	return
}
