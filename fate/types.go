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
package fate

import "bytes"

const (
	T_DIG = 0x01          // 数字
	T_UPH = 0x02          // 大写字谜
	T_LPH = 0x04          // 小写字谜
	T_PHA = T_UPH | T_LPH // 字谜
	T_NUM = T_PHA | T_DIG // 字母或者数字
	T_SPC = 0x08          // 空白符
	T_EOL = 0x10
	T_LIM = 0xFF
)

const (
	_DIG = T_DIG
	_TAB = T_SPC // tab
	_SPC = T_SPC // 空格
	_UPH = T_UPH
	_LPH = T_LPH
	_PHA = T_PHA
	_EOL = T_EOL // '\n'
)

type Fate interface {
	Run() map[string]string
}

type fateReserved struct {
	typ  int
	name []byte
}

var reservedRegistry = []fateReserved{
	{SET, []byte("SET")},
	{GET, []byte("GET")},
}

var charList = [...]int{
	0, 0, 0, 0, 0, 0, 0, 0,

	0, _TAB, _EOL, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	_SPC, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	_DIG, _DIG, _DIG, _DIG, _DIG, _DIG, _DIG, _DIG,

	_DIG, _DIG, 0, 0, 0, 0, 0, 0,

	0, _UPH, _UPH, _UPH, _UPH, _UPH, _UPH, _UPH,

	_UPH, _UPH, _UPH, _UPH, _UPH, _UPH, _UPH, _UPH,

	_UPH, _UPH, _UPH, _UPH, _UPH, _UPH, _UPH, _UPH,

	_UPH, _UPH, _UPH, 0, 0, 0, 0, _PHA,

	0, _LPH, _LPH, _LPH, _LPH, _LPH, _LPH, _LPH,

	_LPH, _LPH, _LPH, _LPH, _LPH, _LPH, _LPH, _LPH,

	_LPH, _LPH, _LPH, _LPH, _LPH, _LPH, _LPH, _LPH,

	_LPH, _LPH, _LPH, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,

	0, 0, 0, 0, 0, 0, 0, 0,
}

func wordType(w []byte) int {
	for _, v := range reservedRegistry {
		if bytes.Compare(w, v.name) == 0 {
			return v.typ
		}
	}
	return DATA
}

// 空格或者制表符
func isSpace(c int) bool {
	return (charList[c&T_LIM] & T_SPC) != 0
}

func isDigit(c int) bool {
	return (charList[c&T_LIM] & T_DIG) != 0
}

func isUpper(c int) bool {
	return (charList[c&T_LIM] & T_UPH) != 0
}

func isLower(c int) bool {
	return (charList[c&T_LIM] & T_LPH) != 0
}

// 包含特殊字符_
func isAlpha(c int) bool {
	return (charList[c&T_LIM] & T_PHA) != 0
}

func isAlnum(c int) bool {
	return (charList[c&T_LIM] & T_NUM) != 0
}
