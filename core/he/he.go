package he

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"io"
	"math/big"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm3"
)

var one = new(big.Int).SetInt64(1)

// 32byte
func zeroByteSlice() []byte {
	return []byte{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
	}
}

func RandFieldElement(c elliptic.Curve, rand io.Reader) (k *big.Int, err error) {
	params := c.Params()
	b := make([]byte, params.BitSize/8+8)
	_, err = io.ReadFull(rand, b)
	if err != nil {
		return
	}

	k = new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(params.N, one)
	k.Mod(k, n)
	k.Add(k, one)
	return
}

func Add(c1, c2 []byte) ([]byte, error) {
	if bytes.Compare(c1[:64], c2[:64]) != 0 {
		return []byte{}, errors.New("Sub: Failed to Add")
	}
	x, _ := new(big.Int).SetString(string(c1[96:]), 10)
	y, _ := new(big.Int).SetString(string(c2[96:]), 10)
	//return append(c1[:96], []byte(new(big.Int).Add(x, y).String())...), nil
	buf := []byte{}
	buf = append(buf, c1[:96]...)
	return append(buf, []byte(new(big.Int).Add(x, y).String())...), nil
}

func Sub(c1, c2 []byte) ([]byte, error) {
	if bytes.Compare(c1[:64], c2[:64]) != 0 {
		return []byte{}, errors.New("Sub: Failed to Sub")
	}

	x, _ := new(big.Int).SetString(string(c1[96:]), 10)
	y, _ := new(big.Int).SetString(string(c2[96:]), 10)

	buf := []byte{}
	buf = append(buf, c1[:96]...)
	return append(buf, []byte(new(big.Int).Sub(x, y).String())...), nil

}

func Encrypt(priv interface{}, m, r *big.Int) ([]byte, error) {
	var x1, x2, y1, y2 *big.Int
	switch priv.(type) {
	case *sm2.PrivateKey:
		v, _ := priv.(*sm2.PrivateKey)
		x1, y1 = v.Curve.ScalarBaseMult(r.Bytes())
		//x1, y1 = priv.Curve.ScalarBaseMult(r.Bytes())
		x2, y2 = v.Curve.ScalarMult(v.X, v.Y, r.Bytes())

	case *ecdsa.PrivateKey:
		v, _ := priv.(*ecdsa.PrivateKey)
		x1, y1 = v.Curve.ScalarBaseMult(r.Bytes())
		x2, y2 = v.Curve.ScalarMult(v.X, v.Y, r.Bytes())

	default:
		return nil, errors.New("error key type")
	}

	x1Buf := x1.Bytes()
	y1Buf := y1.Bytes()
	x2Buf := x2.Bytes()
	y2Buf := y2.Bytes()

	if n := len(x1Buf); n < 32 {
		x1Buf = append(zeroByteSlice()[:32-n], x1Buf...)
	}
	if n := len(y1Buf); n < 32 {
		y1Buf = append(zeroByteSlice()[:32-n], y1Buf...)
	}
	if n := len(x2Buf); n < 32 {
		x2Buf = append(zeroByteSlice()[:32-n], x2Buf...)
	}
	if n := len(y2Buf); n < 32 {
		y2Buf = append(zeroByteSlice()[:32-n], y2Buf...)
	}

	c := []byte{}
	c = append(c, x1Buf...) // R
	c = append(c, y1Buf...) // R
	cc := new(big.Int).Mul(m, new(big.Int).SetBytes(append(x2Buf, y2Buf...)))
	tm := []byte{}
	tm = append(tm, x2Buf...)
	tm = append(tm, y2Buf...)
	c = append(c, sm3.Sm3Sum(tm)...)
	c = append(c, []byte(cc.String())...)

	return c, nil

}

func Decrypt(private interface{}, data []byte) (*big.Int, error) {

	if len(data) < 64 {
		return nil, errors.New("error input data")
	}

	x := new(big.Int).SetBytes(data[:32])
	y := new(big.Int).SetBytes(data[32:64])

	var x2, y2 *big.Int
	switch private.(type) {
	case *sm2.PrivateKey:
		priv, _ := private.(*sm2.PrivateKey)
		x2, y2 = priv.Curve.ScalarMult(x, y, priv.D.Bytes())

	case *ecdsa.PrivateKey:

		priv, _ := private.(*ecdsa.PrivateKey)
		x2, y2 = priv.Curve.ScalarMult(x, y, priv.D.Bytes())

	default:
		return nil, errors.New("error key type")
	}

	//x2, y2 := priv.Curve.ScalarMult(x, y, priv.D.Bytes())
	x2Buf := x2.Bytes()
	y2Buf := y2.Bytes()
	if n := len(x2Buf); n < 32 {
		x2Buf = append(zeroByteSlice()[:32-n], x2Buf...)
	}
	if n := len(y2Buf); n < 32 {
		y2Buf = append(zeroByteSlice()[:32-n], y2Buf...)
	}
	cc, _ := new(big.Int).SetString(string(data[96:]), 10)
	tm := []byte{}
	tm = append(tm, x2Buf...)
	tm = append(tm, y2Buf...)
	if bytes.Compare(sm3.Sm3Sum(tm), data[64:96]) != 0 {
		return nil, errors.New("Decrypt: Failed to Decrypt")
	}
	return new(big.Int).Div(cc, new(big.Int).SetBytes(append(x2Buf, y2Buf...))), nil
}
