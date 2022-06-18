package he

import (
	"crypto/rand"
	"fmt"
	"github.com/tjfoc/gmsm/sm2"
	"math/big"
	"testing"
)

func TestHe(t *testing.T) {
	p, _ := sm2.GenerateKey()
	k, _ := RandFieldElement(p.Curve, rand.Reader)
	a := int64(1)
	b := int64(11)
	c := int64(111)
	ba := new(big.Int).SetInt64(a)
	bb := new(big.Int).SetInt64(b)
	bc := new(big.Int).SetInt64(c)
	sa, _ := Encrypt(p, ba, k)
	sb, _ := Encrypt(p, bb, k)
	sc, _ := Encrypt(p, bc, k)
	r1, _ := Sub(sb, sa)
	d1, _ := Decrypt(p, r1)
	r2, _ := Add(sc, r1)
	d2, _ := Decrypt(p, r2)
	fmt.Println("result1:", d1)
	fmt.Println("result2:", d2)
	da, _ := Decrypt(p, sa)
	db, _ := Decrypt(p, sb)
	dc, _ := Decrypt(p, sc)
	fmt.Println("a:", da)
	fmt.Println("b:", db)
	fmt.Println("c:", dc)

}
