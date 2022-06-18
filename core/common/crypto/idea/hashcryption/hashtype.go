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
package hashcryption

type DataHash struct {
	Type string
}

//	=====================================================================
//	function name: DataHashByType
//	function type: public
//	function receiver: na
//  switch hashtype(sha256/ripemd160/md5/sha1/doublesha256/ShakeSum256)
//	=====================================================================
func (dataHash *DataHash) DataHashByType(data string) (dataHashResult []byte) {
	hashtype := dataHash.Type
	switch hashtype {
	case "sha256":
		dataHashResult = Sha256Hash([]byte(data))
	case "ripemd160":
		dataHashResult = Ripemd160Hash([]byte(data))
	case "md5":
		dataHashResult = []byte(Md5Hash(data))
	case "sha1":
		dataHashResult = Sha1Hash([]byte(data))
	case "doublesha256":
		dataHashResult = DoubleSha256([]byte(data))
	case "ShakeSum256":
		dataHashResult = ShakeSum256([]byte(data))
	default:
		dataHashResult = []byte("unknown Hash")
		panic("not found hashtype for " + hashtype)
	}
	return
}
