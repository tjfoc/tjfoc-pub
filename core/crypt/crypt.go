package crypt

import (
	"bytes"
	"crypto"
	"crypto/cipher"
	"crypto/ecdsa"
	"encoding/asn1"
	"hash"
	"math/big"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm4"
)

var commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

type cryptSignature struct {
	R, S *big.Int
}

func New(h hash.Hash, s crypto.Signer, opts crypto.SignerOpts, p Pki, cip Cipher, v Verifyer) *Crypt {
	return &Crypt{
		h:    h,
		s:    s,
		opts: opts,
		pki:  p,
		cip:  cip,
		v:    v,
	}
}

// func (pub *ecdsa.PublicKey) Verify(h, s []byte, opts crypto.SignerOpts) bool {
// 	// return ecdsa.Verify(pub, h, r, s)
// }

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

func (v *Sm2Verify) Verify(h []byte, s []byte) bool {
	return v.Pub.Verify(h, s)
}

func (v *EccVerify) Verify(h []byte, s []byte) bool {
	var sig cryptSignature
	if _, err := asn1.Unmarshal(s, &sig); err != nil {
		return false
	}
	return ecdsa.Verify(v.Pub, h, sig.R, sig.S)
}

func (e *Sm2Ecc) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	return e.Priv.Decrypt(ciphertext)
}

func (e *Sm2Ecc) Encrypt(plaintext []byte) (ciphertext []byte, err error) {
	return e.Pub.Encrypt(plaintext)
}

func (s *Sm4Cipher) DecryptBlock(k, ciphertext []byte) (plaintext []byte, err error) {
	block, err := sm4.NewCipher(k)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, commonIV)
	origData := make([]byte, len(ciphertext))
	blockMode.CryptBlocks(origData, ciphertext)
	origData = pkcs5UnPadding(origData)
	return origData, nil
}

func (s *Sm4Cipher) EncryptBlock(k, plaintext []byte) (ciphertext []byte, err error) {
	block, err := sm4.NewCipher(k)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData := pkcs5Padding(plaintext, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, commonIV)
	cryted := make([]byte, len(origData))
	blockMode.CryptBlocks(cryted, origData)
	return cryted, nil
}

func pkcs5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...) 
}

func pkcs5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}
