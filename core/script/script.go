package script

import (
	"bytes"

	"github.com/tjfoc/tjfoc/core/stack"
)

// 关键词列表
var reservedRegistry = []dsReserved{
	// 数学运算
	{MIN, []byte("MIN")},
	{MAX, []byte("MAX")},
	{ADD, []byte("ADD")},
	{SUB, []byte("SUB")},
	{MUL, []byte("MUL")},
	{DIV, []byte("DIV")},
	{MOD, []byte("MOD")},

	// 栈操作符
	{POP, []byte("POP")},
	{DUP, []byte("DUP")},
	{NOOP, []byte("NOOP")},
	{PUSH, []byte("PUSH")},
	{SWAP, []byte("SWAP")},
	{EQUAL, []byte("EQUAL")},
	{EQUALQ, []byte("EQUALQ")},
	// 位操作
	{OR, []byte("OR")},
	{XOR, []byte("XOR")},
	{AND, []byte("AND")},
	{LSHIFT, []byte("LSHIFT")},
	{RSHIFT, []byte("RSHIFT")},
	// 密码学操作
	{SM3, []byte("SM3")},
	{HASH1, []byte("HASH1")},
	{HASH256, []byte("HASH256")},
	{HASH512, []byte("HASH512")},
	{RIPEMD160, []byte("RIPEMD160")},
	{CHECKSIGSM2, []byte("CHECKSIGSM2")},
	{CHECKSIGRSA1024, []byte("CHECKSIGRSA1024")},
	{CHECKSIGECDSA256, []byte("CHECKSIGECDSA256")},
	// 类型
	{BOOL, []byte("BOOL")},
	{INT8, []byte("INT8")},
	{INT16, []byte("INT16")},
	{INT32, []byte("INT32")},
	{INT64, []byte("INT64")},
	{STRING, []byte("STRING")},
	//IF ELSE
	{IF, []byte("IF")},
	{IF_END, []byte(":")},
	{ELSE, []byte("ELSE")},
	{ENDIF, []byte("ENDIF")},
}

var opCodeRegistry = []dsOpCodeFunc{
	{OR, or},
	{XOR, xor},
	{AND, and},
	{MIN, min},
	{MAX, max},
	{ADD, add},
	{SUB, sub},
	{MUL, mul},
	{DIV, div},
	{MOD, mod},
	{POP, pop},
	{DUP, dup},
	{NOOP, noop},
	{PUSH, push},
	{SWAP, swap},
	{EQUAL, equal},
	{EQUALQ, equalQ},
	{SM3, hashSm3},
	{HASH1, hash1},
	{LSHIFT, lshift},
	{RSHIFT, rshift},
	{HASH256, hash256},
	{HASH512, hash512},
	{RIPEMD160, hashRipemd160},
	{CHECKSIGECDSA256, checkSigEcdsa256},
}
var opCodeRegistryIFELSE = []dsOpCodeFunc{
	{IF, s_if},
	{ELSE, s_else},
}

// 将脚本参数放入栈中
func newStack(a []*DsValue) *stack.Stack {
	stack := stack.New()
	for _, v := range a {
		stack.Push(v)
	}
	return stack
}

// 判断单词类型
func wordType(w []byte) int {
	for _, v := range reservedRegistry {
		if bytes.Compare(w, v.name) == 0 {
			return v.typ
		}
	}
	return DATA
}

// 简单的词法解析
func Parse(script []byte) *DS {
	d := &DS{}
	w := []byte{}
	for _, v := range script {
		switch v {
		case ' ', '\t', '\n':
			if len(w) > 0 {
				d.words = append(d.words, &dsWord{
					data: w,
					typ:  wordType(w),
				})
				w = []byte{}
			}
		default:
			w = append(w, v)
		}
	}
	if len(w) > 0 {
		d.words = append(d.words, &dsWord{
			data: w,
			typ:  wordType(w),
		})
	}
	return d
}

func (d *DS) Run(args []*DsValue) *DsValue {
	var rtv *DsValue
	var f opCodeFunc

	d.stack = newStack(args)
	for {
		if len(d.words) == 0 {
			break
		}
		f = nil
		for _, v := range opCodeRegistry {
			if v.typ == d.words[0].typ {
				f = v.f
				break
			}
		}
		if f == nil {
			for _, v := range opCodeRegistryIFELSE {
				if v.typ == d.words[0].typ {
					f = v.f
					break
				}
			}
		}
		if f == nil {
			return nil
		}
		if rtv = f(d); rtv == nil {
			return nil
		}
	}
	if d.stack.Count() != 0 {
		return nil
	}
	return rtv
}
