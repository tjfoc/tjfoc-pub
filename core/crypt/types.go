package crypt

import (
	"crypto"
	"hash"
	"io"
)

type Crypt struct {
	h    hash.Hash
	s    crypto.Signer
	opts crypto.SignerOpts
}

// 通用加密接口
type Crypto interface {
	hash.Hash
	crypto.Signer
	crypto.SignerOpts
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
