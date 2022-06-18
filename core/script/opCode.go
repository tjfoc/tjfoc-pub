package script

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"math/big"

	"golang.org/x/crypto/ripemd160"

	"github.com/tjfoc/gmsm/sm3"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
)

func pop(d *DS) *DsValue {
	d.words = d.words[1:]
	if d.stack.Count() < 1 {
		return nil
	}
	if v, ok := d.stack.Pop().(*DsValue); ok {
		return v
	}
	return nil
}

func dup(d *DS) *DsValue {
	d.words = d.words[1:]
	if d.stack.Count() < 1 {
		return nil
	}
	if v, ok := d.stack.Top().(*DsValue); ok {
		d.stack.Push(v)
		return v
	}
	return nil
}

func noop(d *DS) *DsValue {
	d.words = d.words[1:]
	return NewDsValue(BOOL, true)
}

func push(d *DS) *DsValue {
	d.words = d.words[1:]
	if len(d.words) < 2 { // 数据必须包含类型声明
		return nil
	}
	if v := newDsValue(d.words[0].typ, d.words[1].data); v != nil && d.words[1].typ == DATA {
		d.stack.Push(v)
		d.words = d.words[2:]
		return v
	}
	return nil
}

// 交换栈顶元素，并返回原先的栈顶元素
func swap(d *DS) *DsValue {
	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	a, _ := d.stack.Pop().(*DsValue)
	b, _ := d.stack.Pop().(*DsValue)
	d.stack.Push(a)
	d.stack.Push(b)
	return a
}

// 比较栈顶元素，如果相等则弹出元素并返回栈顶元素，失败返回nil
func equal(d *DS) *DsValue {
	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	a, _ := d.stack.Pop().(*DsValue)
	b, _ := d.stack.Pop().(*DsValue)
	v := NewDsValue(BOOL, a.Eq(b))
	d.stack.Push(v)
	return v
}

// 比较栈顶元素，如果相等则弹出元素并返回栈顶元素，失败返回nil
func equalQ(d *DS) *DsValue {
	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	a, _ := d.stack.Pop().(*DsValue)
	b, _ := d.stack.Pop().(*DsValue)
	if a.Eq(b) {
		return NewDsValue(BOOL, true)
	}
	return nil
}

// (...b, a) -> (... a | b)
func or(d *DS) *DsValue {
	var v *DsValue

	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok {
		if b, ok := d.stack.Pop().(*DsValue); ok && a.typ == b.typ {
			switch a.typ {
			case INT8:
				v1, _ := a.value.(int8)
				v2, _ := b.value.(int8)
				v = NewDsValue(a.typ, v1|v2)
				goto out
			case INT16:
				v1, _ := a.value.(int16)
				v2, _ := b.value.(int16)
				v = NewDsValue(a.typ, v1|v2)
				goto out
			case INT32:
				v1, _ := a.value.(int32)
				v2, _ := b.value.(int32)
				v = NewDsValue(a.typ, v1|v2)
				goto out
			case INT64:
				v1, _ := a.value.(int64)
				v2, _ := b.value.(int64)
				v = NewDsValue(a.typ, v1|v2)
				goto out
			}
		}
	}
	return nil
out:
	d.stack.Push(v)
	return v
}

// (...b, a) -> (... a ^ b)
func xor(d *DS) *DsValue {
	var v *DsValue

	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok {
		if b, ok := d.stack.Pop().(*DsValue); ok && a.typ == b.typ {
			switch a.typ {
			case INT8:
				v1, _ := a.value.(int8)
				v2, _ := b.value.(int8)
				v = NewDsValue(a.typ, v1^v2)
				goto out
			case INT16:
				v1, _ := a.value.(int16)
				v2, _ := b.value.(int16)
				v = NewDsValue(a.typ, v1^v2)
				goto out
			case INT32:
				v1, _ := a.value.(int32)
				v2, _ := b.value.(int32)
				v = NewDsValue(a.typ, v1^v2)
				goto out
			case INT64:
				v1, _ := a.value.(int64)
				v2, _ := b.value.(int64)
				v = NewDsValue(a.typ, v1^v2)
				goto out
			}
		}
	}
	return nil
out:
	d.stack.Push(v)
	return v
}

// (...b, a) -> (... a & b)
func and(d *DS) *DsValue {
	var v *DsValue

	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok {
		if b, ok := d.stack.Pop().(*DsValue); ok && a.typ == b.typ {
			switch a.typ {
			case INT8:
				v1, _ := a.value.(int8)
				v2, _ := b.value.(int8)
				v = NewDsValue(a.typ, v1&v2)
				goto out
			case INT16:
				v1, _ := a.value.(int16)
				v2, _ := b.value.(int16)
				v = NewDsValue(a.typ, v1&v2)
				goto out
			case INT32:
				v1, _ := a.value.(int32)
				v2, _ := b.value.(int32)
				v = NewDsValue(a.typ, v1&v2)
				goto out
			case INT64:
				v1, _ := a.value.(int64)
				v2, _ := b.value.(int64)
				v = NewDsValue(a.typ, v1&v2)
				goto out
			}
		}
	}
	return nil
out:
	d.stack.Push(v)
	return v
}

// (...b, a) -> (... b << a)
func lshift(d *DS) *DsValue {
	var v *DsValue

	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok {
		if b, ok := d.stack.Pop().(*DsValue); ok && a.typ == b.typ {
			switch a.typ {
			case INT8:
				v1, _ := a.value.(int8)
				v2, _ := b.value.(int8)
				v = NewDsValue(a.typ, int8(uint8(v2)<<uint8(v1)))
				goto out
			case INT16:
				v1, _ := a.value.(int16)
				v2, _ := b.value.(int16)
				v = NewDsValue(a.typ, int16(uint16(v2)<<uint16(v1)))
				goto out
			case INT32:
				v1, _ := a.value.(int32)
				v2, _ := b.value.(int32)
				v = NewDsValue(a.typ, int32(uint32(v2)<<uint32(v1)))
				goto out
			case INT64:
				v1, _ := a.value.(int64)
				v2, _ := b.value.(int64)
				v = NewDsValue(a.typ, int64(uint64(v2)<<uint64(v1)))
				goto out
			}
		}
	}
	return nil
out:
	d.stack.Push(v)
	return v
}

// (...b, a) -> (... b >> a)
func rshift(d *DS) *DsValue {
	var v *DsValue

	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok {
		if b, ok := d.stack.Pop().(*DsValue); ok && a.typ == b.typ {
			switch a.typ {
			case INT8:
				v1, _ := a.value.(int8)
				v2, _ := b.value.(int8)
				v = NewDsValue(a.typ, int8(uint8(v2)>>uint8(v1)))
				goto out
			case INT16:
				v1, _ := a.value.(int16)
				v2, _ := b.value.(int16)
				v = NewDsValue(a.typ, int16(uint16(v2)>>uint16(v1)))
				goto out
			case INT32:
				v1, _ := a.value.(int32)
				v2, _ := b.value.(int32)
				v = NewDsValue(a.typ, int32(uint32(v2)>>uint32(v1)))
				goto out
			case INT64:
				v1, _ := a.value.(int64)
				v2, _ := b.value.(int64)
				v = NewDsValue(a.typ, int64(uint64(v2)>>uint64(v1)))
				goto out
			}
		}
	}
	return nil
out:
	d.stack.Push(v)
	return v
}

// (...b, a) -> (... b + a)
func add(d *DS) *DsValue {
	var v *DsValue

	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok {
		if b, ok := d.stack.Pop().(*DsValue); ok && a.typ == b.typ {
			switch a.typ {
			case INT8:
				v1, _ := a.value.(int8)
				v2, _ := b.value.(int8)
				v = NewDsValue(a.typ, v1+v2)
				goto out
			case INT16:
				v1, _ := a.value.(int16)
				v2, _ := b.value.(int16)
				v = NewDsValue(a.typ, v1+v2)
				goto out
			case INT32:
				v1, _ := a.value.(int32)
				v2, _ := b.value.(int32)
				v = NewDsValue(a.typ, v1+v2)
				goto out
			case INT64:
				v1, _ := a.value.(int64)
				v2, _ := b.value.(int64)
				v = NewDsValue(a.typ, v1+v2)
				goto out
			}
		}
	}
	return nil
out:
	d.stack.Push(v)
	return v
}

// (...b, a) -> (... b - a)
func sub(d *DS) *DsValue {
	var v *DsValue

	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok {
		if b, ok := d.stack.Pop().(*DsValue); ok && a.typ == b.typ {
			switch a.typ {
			case INT8:
				v1, _ := a.value.(int8)
				v2, _ := b.value.(int8)
				v = NewDsValue(a.typ, v2-v1)
				goto out
			case INT16:
				v1, _ := a.value.(int16)
				v2, _ := b.value.(int16)
				v = NewDsValue(a.typ, v2-v1)
				goto out
			case INT32:
				v1, _ := a.value.(int32)
				v2, _ := b.value.(int32)
				v = NewDsValue(a.typ, v2-v1)
				goto out
			case INT64:
				v1, _ := a.value.(int64)
				v2, _ := b.value.(int64)
				v = NewDsValue(a.typ, v2-v1)
				goto out
			}
		}
	}
	return nil
out:
	d.stack.Push(v)
	return v
}

// (...b, a) -> (... b * a)
func mul(d *DS) *DsValue {
	var v *DsValue

	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok {
		if b, ok := d.stack.Pop().(*DsValue); ok && a.typ == b.typ {
			switch a.typ {
			case INT8:
				v1, _ := a.value.(int8)
				v2, _ := b.value.(int8)
				v = NewDsValue(a.typ, v2*v1)
				goto out
			case INT16:
				v1, _ := a.value.(int16)
				v2, _ := b.value.(int16)
				v = NewDsValue(a.typ, v2*v1)
				goto out
			case INT32:
				v1, _ := a.value.(int32)
				v2, _ := b.value.(int32)
				v = NewDsValue(a.typ, v2*v1)
				goto out
			case INT64:
				v1, _ := a.value.(int64)
				v2, _ := b.value.(int64)
				v = NewDsValue(a.typ, v2*v1)
				goto out
			}
		}
	}
	return nil
out:
	d.stack.Push(v)
	return v
}

// (...b, a) -> (... b / a)
func div(d *DS) *DsValue {
	var v *DsValue

	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok {
		if b, ok := d.stack.Pop().(*DsValue); ok && a.typ == b.typ {
			switch a.typ {
			case INT8:
				v1, _ := a.value.(int8)
				v2, _ := b.value.(int8)
				if v1 == 0 {
					return nil
				}
				v = NewDsValue(a.typ, v2/v1)
				goto out
			case INT16:
				v1, _ := a.value.(int16)
				v2, _ := b.value.(int16)
				if v1 == 0 {
					return nil
				}
				v = NewDsValue(a.typ, v2/v1)
				goto out
			case INT32:
				v1, _ := a.value.(int32)
				v2, _ := b.value.(int32)
				if v1 == 0 {
					return nil
				}
				v = NewDsValue(a.typ, v2/v1)
				goto out
			case INT64:
				v1, _ := a.value.(int64)
				v2, _ := b.value.(int64)
				if v1 == 0 {
					return nil
				}
				v = NewDsValue(a.typ, v2/v1)
				goto out
			}
		}
	}
	return nil
out:
	d.stack.Push(v)
	return v
}

// (...b, a) -> (... b % a)
func mod(d *DS) *DsValue {
	var v *DsValue

	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok {
		if b, ok := d.stack.Pop().(*DsValue); ok && a.typ == b.typ {
			switch a.typ {
			case INT8:
				v1, _ := a.value.(int8)
				v2, _ := b.value.(int8)
				if v1 == 0 {
					return nil
				}
				v = NewDsValue(a.typ, v2%v1)
				goto out
			case INT16:
				v1, _ := a.value.(int16)
				v2, _ := b.value.(int16)
				if v1 == 0 {
					return nil
				}
				v = NewDsValue(a.typ, v2%v1)
				goto out
			case INT32:
				v1, _ := a.value.(int32)
				v2, _ := b.value.(int32)
				if v1 == 0 {
					return nil
				}
				v = NewDsValue(a.typ, v2%v1)
				goto out
			case INT64:
				v1, _ := a.value.(int64)
				v2, _ := b.value.(int64)
				if v1 == 0 {
					return nil
				}
				v = NewDsValue(a.typ, v2%v1)
				goto out
			}
		}
	}
	return nil
out:
	d.stack.Push(v)
	return v
}

// (... b, a) -> (... b < a ? b : a)
func min(d *DS) *DsValue {
	var v *DsValue

	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok {
		if b, ok := d.stack.Pop().(*DsValue); ok && a.typ == b.typ {
			switch a.typ {
			case INT8:
				v1, _ := a.value.(int8)
				if v2, _ := b.value.(int8); v1 < v2 {
					v = NewDsValue(a.typ, v1)
				} else {
					v = NewDsValue(a.typ, v2)
				}
				goto out
			case INT16:
				v1, _ := a.value.(int16)
				if v2, _ := b.value.(int16); v1 < v2 {
					v = NewDsValue(a.typ, v1)
				} else {
					v = NewDsValue(a.typ, v2)
				}
				goto out
			case INT32:
				v1, _ := a.value.(int32)
				if v2, _ := b.value.(int32); v1 < v2 {
					v = NewDsValue(a.typ, v1)
				} else {
					v = NewDsValue(a.typ, v2)
				}
				goto out
			case INT64:
				v1, _ := a.value.(int64)
				if v2, _ := b.value.(int64); v1 < v2 {
					v = NewDsValue(a.typ, v1)
				} else {
					v = NewDsValue(a.typ, v2)
				}
				goto out
			}
		}
	}
	return nil
out:
	d.stack.Push(v)
	return v
}

// (... b, a) -> (... b < a ? a : b)
func max(d *DS) *DsValue {
	var v *DsValue

	d.words = d.words[1:]
	if d.stack.Count() < 2 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok {
		if b, ok := d.stack.Pop().(*DsValue); ok && a.typ == b.typ {
			switch a.typ {
			case INT8:
				v1, _ := a.value.(int8)
				if v2, _ := b.value.(int8); v1 < v2 {
					v = NewDsValue(a.typ, v2)
				} else {
					v = NewDsValue(a.typ, v1)
				}
				goto out
			case INT16:
				v1, _ := a.value.(int16)
				if v2, _ := b.value.(int16); v1 < v2 {
					v = NewDsValue(a.typ, v2)
				} else {
					v = NewDsValue(a.typ, v1)
				}
				goto out
			case INT32:
				v1, _ := a.value.(int32)
				if v2, _ := b.value.(int32); v1 < v2 {
					v = NewDsValue(a.typ, v2)
				} else {
					v = NewDsValue(a.typ, v1)
				}
				goto out
			case INT64:
				v1, _ := a.value.(int64)
				if v2, _ := b.value.(int64); v1 < v2 {
					v = NewDsValue(a.typ, v2)
				} else {
					v = NewDsValue(a.typ, v1)
				}
				goto out
			}
		}
	}
	return nil
out:
	d.stack.Push(v)
	return v
}

// (... a) -> (... h(a))
func hashSm3(d *DS) *DsValue {
	d.words = d.words[1:]
	if d.stack.Count() < 1 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok && a.typ == STRING {
		m, _ := a.value.(string)
		if c, err := miscellaneous.GenHash(sm3.New(), []byte(m)); err == nil {
			v := NewDsValue(STRING, string(c))
			d.stack.Push(v)
			return v
		}
	}
	return nil
}

// (... a) -> (... h(a))
func hash1(d *DS) *DsValue {
	d.words = d.words[1:]
	if d.stack.Count() < 1 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok && a.typ == STRING {
		m, _ := a.value.(string)
		if c, err := miscellaneous.GenHash(sha1.New(), []byte(m)); err == nil {
			v := NewDsValue(STRING, string(c))
			d.stack.Push(v)
			return v
		}
	}
	return nil
}

// (... a) -> (... h(a))
func hash256(d *DS) *DsValue {
	d.words = d.words[1:]
	if d.stack.Count() < 1 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok && a.typ == STRING {
		m, _ := a.value.(string)
		if c, err := miscellaneous.GenHash(sha256.New(), []byte(m)); err == nil {
			v := NewDsValue(STRING, string(c))
			d.stack.Push(v)
			return v
		}
	}
	return nil
}

// (... a) -> (... h(a))
func hash512(d *DS) *DsValue {
	d.words = d.words[1:]
	if d.stack.Count() < 1 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok && a.typ == STRING {
		m, _ := a.value.(string)
		if c, err := miscellaneous.GenHash(sha512.New(), []byte(m)); err == nil {
			v := NewDsValue(STRING, string(c))
			d.stack.Push(v)
			return v
		}
	}
	return nil
}

// (... a) -> (... h(a))
func hashRipemd160(d *DS) *DsValue {
	d.words = d.words[1:]
	if d.stack.Count() < 1 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok && a.typ == STRING {
		m, _ := a.value.(string)
		if c, err := miscellaneous.GenHash(ripemd160.New(), []byte(m)); err == nil {
			v := NewDsValue(STRING, string(c))
			d.stack.Push(v)
			return v
		}
	}
	return nil
}

// (... msg, sig, pubkey) -> (... bool)
func checkSigEcdsa256(d *DS) *DsValue {
	d.words = d.words[1:]
	if d.stack.Count() < 3 {
		return nil
	}
	if a, ok := d.stack.Pop().(*DsValue); ok { // pubkey
		if b, ok := d.stack.Pop().(*DsValue); ok && a.typ == b.typ { // sig
			if c, ok := d.stack.Pop().(*DsValue); ok && b.typ == c.typ { // msg
				switch a.typ {
				case STRING:
					pubStr, _ := a.value.(string)
					sigStr, _ := b.value.(string)
					msgStr, _ := c.value.(string)
					sigBytes := []byte(sigStr)
					pubBytes := []byte(pubStr)
					x := new(big.Int).SetBytes(pubBytes[:len(pubBytes)/2])
					y := new(big.Int).SetBytes(pubBytes[len(pubBytes)/2:])
					r := new(big.Int).SetBytes(sigBytes[:len(sigBytes)/2])
					s := new(big.Int).SetBytes(sigBytes[len(sigBytes)/2:])
					return NewDsValue(BOOL, ecdsa.Verify(&ecdsa.PublicKey{X: x, Y: y, Curve: elliptic.P256()}, []byte(msgStr), r, s))
				}
			}
		}
	}
	return nil
}
func s_if(d *DS) *DsValue {
	d.words = d.words[1:]
	if_end_flag := false
	for _, v := range d.words {
		if v.typ == ELSE || v.typ == ENDIF {
			break
		} else if v.typ == IF_END {
			if_end_flag = true
		}
	}
	if !if_end_flag {
		return nil
	}
	for {
		if d.words[0].typ == IF_END {
			break
		} else {
			var rtv *DsValue
			var f opCodeFunc
			for _, v := range opCodeRegistry {
				if v.typ == d.words[0].typ {
					f = v.f
					break
				}
			}
			if rtv = f(d); rtv == nil {
				return nil
			}
		}
	}
	d.words = d.words[1:]
	if d.stack.Count() < 1 {
		return nil
	}
	var if_flag bool
	if v, ok := d.stack.Top().(*DsValue); ok && v.typ == BOOL {
		if_flag = v.value.(bool)
	} else {
		return nil
	}
	if if_flag {
		end_flag := false
		else_flag := false
		else_index := 0
		endif_index := 0
		end_last_flag := false
		if_count := 0
		last_type := 0
		for i, v := range d.words {
			if v.typ == ENDIF {
				if_count--
				if if_count == -1 {
					end_flag = true
					if i+1 == len(d.words) {
						end_last_flag = true
					} else {
						endif_index = i + 1
					}
					break
				}
			} else if v.typ == ELSE && if_count == 0 {
				if !else_flag {
					else_flag = true
					else_index = i
				}
				last_type = ELSE
			} else if last_type != ELSE && v.typ == IF {
				if_count++
			} else if v.typ == ELSE {
				last_type = ELSE
			} else {
				last_type = 0
			}
		}
		if else_flag {
			if end_last_flag {
				d.words = d.words[:else_index]
			} else {
				d.words = append(d.words[:else_index], d.words[endif_index:]...)
			}
		} else if end_flag {
			if end_last_flag {
				d.words = d.words[:len(d.words)-1]
			} else {
				d.words = append(d.words[:endif_index-1], d.words[endif_index:]...)
			}
		} else {
			return nil
		}
		return &DsValue{
			typ:   BOOL,
			value: true,
		}
	} else {
		else_flag := false
		end_flag := false
		index := 0
		if_count := 0
		end_last_flag := false
		last_type := 0
		for i, v := range d.words {
			if v.typ == ELSE && if_count == 0 {
				else_flag = true
				index = i
				break
			} else if v.typ == ENDIF {
				if_count--
				if if_count == -1 {
					end_flag = true
					if i+1 == len(d.words) {
						end_last_flag = true
					} else {
						index = i + 1
					}
					break
				}
			} else if last_type != ELSE && v.typ == IF {
				if_count++
			} else if v.typ == ELSE {
				last_type = ELSE
			} else {
				last_type = 0
			}
		}
		if else_flag {
			d.words = d.words[index:]
		} else if end_flag {
			if end_last_flag {
				d.words = d.words[:len(d.words)-1]
			} else {
				d.words = d.words[index:]
			}
		} else {
			return nil
		}
		return &DsValue{
			typ:   BOOL,
			value: false,
		}
	}
}
func s_else(d *DS) *DsValue {
	d.words = d.words[1:]
	if d.words[0].typ != IF {
		end_flag := false
		index := 0
		if_count := 0
		end_last_flag := false
		last_type := 0
		for i, v := range d.words {
			if v.typ == ENDIF {
				if_count--
				if if_count == -1 {
					end_flag = true
					if i+1 == len(d.words) {
						end_last_flag = true
					} else {
						index = i
					}
				}
			} else if last_type != ELSE && v.typ == IF {
				if_count++
			} else if v.typ == ELSE {
				last_type = ELSE
			} else {
				last_type = 0
			}
		}
		if !end_flag {
			return nil
		}
		if end_last_flag {
			d.words = d.words[:len(d.words)-1]
		} else {
			d.words = append(d.words[:index], d.words[index+1:]...)
		}
	}
	return &DsValue{
		typ:   BOOL,
		value: true,
	}
}
