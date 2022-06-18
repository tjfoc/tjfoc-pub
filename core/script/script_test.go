package script

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"

	base58 "github.com/jbenet/go-base58"
	"github.com/tjfoc/tjfoc/core/miscellaneous"
)

func TestScript(t *testing.T) {
	//	ds := Parse([]byte("PUSH INT8 3 PUSH INT8 2 ADD NOOP NOOP POP"))

	m := []byte("test")
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	r, s, _ := ecdsa.Sign(rand.Reader, priv, m)
	pubKey := priv.PublicKey
	sig := make([]byte, 64)
	key := make([]byte, 64)
	miscellaneous.Memmove(sig[:32], r.Bytes())
	miscellaneous.Memmove(sig[32:], s.Bytes())
	miscellaneous.Memmove(key[:32], pubKey.X.Bytes())
	miscellaneous.Memmove(key[32:], pubKey.Y.Bytes())

	/*
		ds := Parse([]byte(fmt.Sprintf("PUSH STRING %s PUSH STRING %s EQUAL"+
			" IF DUP POP : "+
			" POP IF PUSH BOOL FALSE : PUSH INT8 2 PUSH INT8 2 ADD POP ELSE PUSH INT8 1 PUSH INT8 1 ADD POP ENDIF "+
			" ELSE IF PUSH BOOL FALSE : "+
			" PUSH INT8 3 PUSH INT8 3 ADD "+
			" ELSE PUSH BOOL TRUE POP ENDIF POP",
			base58.Encode(m), base58.Encode(m))))
	*/
	ds := Parse([]byte(fmt.Sprintf("PUSH STRING %s PUSH STRING %s EQUAL"+
		" IF DUP POP : "+
		" POP IF PUSH BOOL FALSE : PUSH INT8 2 PUSH INT8 2 ADD POP ELSE "+
		"NOOP IF PUSH INT8 1 PUSH INT8 1 EQUAL : POP ENDIF POP PUSH INT8 1 PUSH INT8 1 ADD POP ENDIF "+
		"ENDIF",
		base58.Encode(m), base58.Encode(m))))
	//	ds := Parse([]byte(fmt.Sprintf("PUSH STRING %s PUSH STRING %s PUSH STRING %s CHECKSIGECDSA256"+
	//		"IF : PUSH INT8 2 PUSH INT8 2 ADD POP ELSE IF PUSH BOOL FALSE : PUSH INT8 3 PUSH INT8 3 ADD ELSE PUSH BOOL TRUE ENDIF",
	//		base58.Encode(m), base58.Encode(sig), base58.Encode(key))))

	/*
		m := []byte("test")
		n := []byte("test1")

		a := base58.Encode(m)
		b := base58.Encode(n)
		ds := Parse([]byte(fmt.Sprintf("PUSH STRING %s PUSH STRING %s EQUAL POP PUSH INT8 3 PUSH INT8 2 ADD POP", a, b)))
	*/

	//ds := Parse([]byte("PUSH INT64 66666 IF PUSH BOOL FALSE : POP PUSH INT32 6 ELSE IF PUSH BOOL FALSE : POP POP PUSH INT16 8 ELSE PUSH INT32 100 ENDIF POP POP POP POP"))
	v := ds.Run(nil)
	if v == nil {
		fmt.Printf("script: run error\n")
		return
	}
	fmt.Printf("%v: %v\n", v.Typeof(), v.Valueof())
}
