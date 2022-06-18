package crypt

import (
	"crypto"
	"crypto/ecdsa"
	"hash"
	"io"

	"github.com/tjfoc/gmsm/sm2"
)

type Crypt struct {
	h    hash.Hash
	s    crypto.Signer
	opts crypto.SignerOpts

	pki Pki
	cip Cipher
	v   Verifyer
}

// 通用加密接口
type Crypto interface {
	hash.Hash
	crypto.Signer
	crypto.SignerOpts

	Pki
	Cipher
	Verifyer
}

type Sm2Ecc struct {
	Pub  *sm2.PublicKey
	Priv *sm2.PrivateKey
}

type Sm2Verify struct {
	Pub *sm2.PublicKey
}

type EccVerify struct {
	Pub *ecdsa.PublicKey
}

type Sm4Cipher struct {
}

type Cipher interface {
	// DecryptBlock decrypts ciphertext with key.
	DecryptBlock(k, ciphertext []byte) (plaintext []byte, err error)

	// EncryptBlock decrypts plaintext with key.
	EncryptBlock(k, plaintext []byte) (ciphertext []byte, err error)
}

type Pki interface {
	// Decrypt decrypts ciphertext with priv.
	Decrypt(ciphertext []byte) (plaintext []byte, err error)

	// Encrypt encrypts plaintext with pub.
	Encrypt(plaintext []byte) (ciphertext []byte, err error)
}

type Verifyer interface {
	Verify([]byte, []byte) bool
}

func (c *Crypt) Sum(b []byte) []byte {
	return c.h.Sum(b)
}

func (c *Crypt) Reset() {
	c.h.Reset()
}

func (c *Crypt) Size() int {
	return c.h.Size()
}

func (c *Crypt) Write(b []byte) (int, error) {
	return c.h.Write(b)
}

func (c *Crypt) BlockSize() int {
	return c.h.BlockSize()
}

func (c *Crypt) Public() crypto.PublicKey {
	return c.s.Public()
}

func (c *Crypt) HashFunc() crypto.Hash {
	return c.opts.HashFunc()
}

func (c *Crypt) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	return c.s.Sign(rand, digest, opts)
}

func (c *Crypt) Verify(h []byte, s []byte) bool {
	return c.v.Verify(h, s)
}

func (c *Crypt) Decrypt(plaintext []byte) ([]byte, error) {
	return c.pki.Decrypt(plaintext)
}

func (c *Crypt) Encrypt(ciphertext []byte) ([]byte, error) {
	return c.pki.Encrypt(ciphertext)
}

func (c *Crypt) DecryptBlock(k, plaintext []byte) ([]byte, error) {
	return c.cip.DecryptBlock(k, plaintext)
}

func (c *Crypt) EncryptBlock(k, ciphertext []byte) ([]byte, error) {
	return c.cip.EncryptBlock(k, ciphertext)
}
