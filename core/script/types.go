/*
	DS是一种强类型的栈式脚本，代码比较写的比较粗漏
*/

package script

import (
	"reflect"
	"strconv"

	base58 "github.com/jbenet/go-base58"
	"github.com/tjfoc/tjfoc/core/stack"
	"github.com/tjfoc/tjfoc/core/typeclass"
)

const ( // 关键词列表, 0x00 ~ 0x1F为类型关键词
	// 数学运算
	ADD = 0x30
	SUB = 0x31
	MUL = 0x32
	DIV = 0x33
	MOD = 0x34
	MIN = 0x36
	MAX = 0x37
	// 栈操作
	POP    = 0x40
	DUP    = 0x41
	PUSH   = 0x42
	SWAP   = 0x43
	NOOP   = 0x44
	EQUAL  = 0x45 // 相等比较，适应于任何类型
	EQUALQ = 0x46 // 相等比较，如果不等直接脚本出错，Q是quit的意思
	// 位操作
	OR     = 0x20
	XOR    = 0x21
	AND    = 0x22
	LSHIFT = 0x23
	RSHIFT = 0x24
	// 密码学操作
	SM3              = 0x50
	HASH1            = 0x51
	HASH256          = 0x52
	HASH512          = 0x53
	RIPEMD160        = 0x54
	CHECKSIGSM2      = 0x55
	CHECKSIGRSA1024  = 0x56
	CHECKSIGECDSA256 = 0x57
	// 类型
	DATA   = 0xFF
	BOOL   = int(reflect.Bool)
	INT8   = int(reflect.Int8)
	INT16  = int(reflect.Int16)
	INT32  = int(reflect.Int32)
	INT64  = int(reflect.Int64)
	STRING = int(reflect.String) // 脚本中的字符串需要进行base58编码
	//IF ELSE
	IF     = 0x60
	IF_END = 0x61
	ELSE   = 0x62
	ENDIF  = 0x63
)

type opCodeFunc (func(*DS) *DsValue)

type dsOpCodeFunc struct {
	typ int
	f   opCodeFunc
}

// 保留关键词注册信息
type dsReserved struct {
	typ  int
	name []byte
}

type dsWord struct {
	typ  int
	data []byte
}

type DsValue struct {
	typ   int
	value interface{}
}

type DS struct {
	words []*dsWord
	stack *stack.Stack
}

func newDsValue(typ int, value []byte) *DsValue {
	switch typ {
	case BOOL:
		if v, err := strconv.ParseBool(string(value)); err == nil {
			return NewDsValue(typ, v)
		}
	case INT8:
		if v, err := strconv.ParseInt(string(value), 0, 8); err == nil {
			return NewDsValue(typ, int8(v))
		}
	case INT16:
		if v, err := strconv.ParseInt(string(value), 0, 16); err == nil {
			return NewDsValue(typ, int16(v))
		}
	case INT32:
		if v, err := strconv.ParseInt(string(value), 0, 32); err == nil {
			return NewDsValue(typ, int32(v))
		}
	case INT64:
		if v, err := strconv.ParseInt(string(value), 0, 64); err == nil {
			return NewDsValue(typ, int64(v))
		}
	case STRING:
		return NewDsValue(typ, string(base58.Decode(string(value))))
	}
	return nil
}

func NewDsValue(typ int, value interface{}) *DsValue {
	if t := int(reflect.TypeOf(value).Kind()); t == typ {
		switch t {
		case BOOL, INT8, INT16, INT32, INT64, STRING:
			return &DsValue{
				typ:   typ,
				value: value,
			}
		}
	}
	return nil
}

func (a *DsValue) Eq(b typeclass.Eq) bool {
	v, ok := b.(*DsValue)
	if !ok {
		return false
	}
	return reflect.DeepEqual(a.value, v.value)
}

func (a *DsValue) NotEq(b typeclass.Eq) bool {
	return !a.Eq(b)
}

func (a *DsValue) Typeof() int {
	return a.typ
}

func (a *DsValue) Valueof() interface{} {
	return a.value
}
