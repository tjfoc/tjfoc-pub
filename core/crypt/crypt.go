package crypt

import (
	"crypto"
	"crypto/ecdsa"
	"encoding/asn1"
	"hash"
	"math/big"

	"github.com/tjfoc/gmsm/sm2"
)

type cryptSignature struct {
	R, S *big.Int
}

func New(h hash.Hash, s crypto.Signer, opts crypto.SignerOpts) *Crypt {
	return &Crypt{
		h:    h,
		s:    s,
		opts: opts,
	}
}

func Verify(k crypto.PublicKey, h []byte, s []byte, opts crypto.SignerOpts) bool {
	switch k.(type) {
	case *sm2.PublicKey:
		pub, _ := k.(*sm2.PublicKey)
		return pub.Verify(h, s)
	case *ecdsa.PublicKey:
		var sig cryptSignature
		pub, _ := k.(*ecdsa.PublicKey)
		if _, err := asn1.Unmarshal(s, &sig); err != nil {
			return false
		}
		return ecdsa.Verify(pub, h, sig.R, sig.S)
	}
	return false
}
