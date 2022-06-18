package ecdsa

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	cr "crypto/rand"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"io/ioutil"
	"math/big"

	"github.com/tjfoc/tjfoc/core/common/errors"
	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/common/util/file"
)

type Idea struct {
	PrivateKey     *ecdsa.PrivateKey
	PublickKey     *ecdsa.PublicKey
	PriKeyFilePath string
	PubKeyFilePath string
}

// ============================================================
// Signature is a type representing an ecdsa signature.
type Signature struct {
	R *big.Int
	S *big.Int
}

const (
	PrFilePath = "Ecdsaprivatekey.pem"
	PuFilePath = "Ecdsapublickey.pem"
)

var logger = flogging.MustGetLogger("crypto.ecdsa")

//	=====================================================================
//	function name: GenerateKeys
//	function type: public
//	function receiver: Idea
//  generate a pair of public and private key, return error
//	=====================================================================
func (idea *Idea) GenerateKeys() errors.CallStackError {
	idea.PriKeyFilePath = PrFilePath
	idea.PubKeyFilePath = PuFilePath

	pri, err := ecdsa.GenerateKey(elliptic.P256(), cr.Reader)
	if err != nil {
		e := errors.ErrorWithCallstack("0008", "000010", "create keys fail")
		logger.Error(e.Message())
		return e
	}

	idea.PrivateKey = pri
	idea.PublickKey = &pri.PublicKey

	return nil
}

//	=====================================================================
//	function name: Sign
//	function type: public
//	function receiver: Idea
//	generate signature according to the private key,
//	return byte array or error
//	=====================================================================
func (idea *Idea) Sign(src []byte) ([]byte, errors.CallStackError) {
	sign, err := idea.PrivateKey.Sign(cr.Reader, src, crypto.SHA256)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000002", "Marshal sign error")
		logger.Error(e.Message())
		return nil, e
	}

	return sign, nil
}

//	=====================================================================
//	function name: Verify
//	function type: public
//	function receiver: Idea
//	verify signature, return bool
//	=====================================================================
func (idea *Idea) Verify(src, sign, pubKey []byte) bool {
	var signS Signature
	_, err := asn1.Unmarshal(sign, &signS)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000003", sign)
		logger.Error(e.Message())
		return false
	}

	err = idea.GetPubKeyByByte(pubKey)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000003", pubKey)
		logger.Error(e.Message())
		return false
	}

	flg := ecdsa.Verify(idea.PublickKey, src, signS.R, signS.S)

	return flg
}

//	=====================================================================
//	function name: WriteKeyFile
//	function type: public
//	function receiver: Idea
//	write a pair of public and private key into two files, return error
//	=====================================================================
func (idea *Idea) WriteKeyFile() errors.CallStackError {
	//write the private key into the file
	prk, err := x509.MarshalECPrivateKey(idea.PrivateKey)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000002", prk)
		logger.Error(e.Message())
		return e
	}

	e := file.WriteFile(idea.PriKeyFilePath, "PRIVATE KEY", prk)
	if e != nil {
		logger.Error(e.Message())
		return e
	}

	//write the public key into the file
	puk, err := x509.MarshalPKIXPublicKey(idea.PublickKey)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000002", puk)
		logger.Error(e.Message())
		return e
	}

	e = file.WriteFile(idea.PubKeyFilePath, "PUBLIC KEY", puk)
	if e != nil {
		logger.Error(e.Message())
		return e
	}

	return nil
}

//	=====================================================================
//	function name: GetPriKey
//	function type: public
//	function receiver: Idea
//	get private key from underlying file, return error
//	=====================================================================
func (idea *Idea) GetPriKey(pivKeyFilePath string) errors.CallStackError {
	buf, err := ioutil.ReadFile(pivKeyFilePath)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000013", pivKeyFilePath)
		logger.Error(e.Message())
		return e
	}

	p, rest := pem.Decode(buf)
	var out *ecdsa.PrivateKey
	if p == nil || len(p.Bytes) == 0 {
		out, err = x509.ParseECPrivateKey(rest)
		if err != nil {
			e := errors.ErrorWithCallstack("0006", "000003", rest)
			logger.Error(e.Message())
			return e
		}
	} else {
		out, err = x509.ParseECPrivateKey(p.Bytes)
		if err != nil {
			e := errors.ErrorWithCallstack("0006", "000003", p.Bytes)
			logger.Error(e.Message())
			return e
		}
	}
	idea.PrivateKey = out

	return nil
}

//	=====================================================================
//	function name: GetPriKeyByByte
//	function type: public
//	function receiver: Idea
//	byte array decode into the private key struct, return error
//	=====================================================================
func (idea *Idea) GetPriKeyByByte(pivKey []byte) errors.CallStackError {
	var out *ecdsa.PrivateKey
	p, rest := pem.Decode(pivKey)
	var err error
	if p != nil || len(p.Bytes) != 0 {
		out, err = x509.ParseECPrivateKey(rest)
	} else {
		out, err = x509.ParseECPrivateKey(p.Bytes)
	}
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000003", pivKey)
		logger.Error(e.Message())
		return e
	}
	idea.PrivateKey = out

	return nil
}

//	=====================================================================
//	function name: GetPubKey
//	function type: public
//	function receiver: Idea
//	get public key from underlying file, return error
//	=====================================================================
func (idea *Idea) GetPubKey(PubKeyFilePath string) errors.CallStackError {
	buf, err := ioutil.ReadFile(PubKeyFilePath)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000013", PubKeyFilePath)
		logger.Error(e.Message())
		return e
	}

	p, rest := pem.Decode(buf)
	if p == nil || len(p.Bytes) == 0 {
		pubk, err := x509.ParsePKIXPublicKey(rest)
		if err != nil {
			e := errors.ErrorWithCallstack("0006", "000003", rest)
			logger.Error(e.Message())
			return e
		}

		switch pub := pubk.(type) {
		case *ecdsa.PublicKey:
			idea.PublickKey = pub
		}
	} else {
		pubk, err := x509.ParsePKIXPublicKey(p.Bytes)
		if err != nil {
			e := errors.ErrorWithCallstack("0006", "000003", p.Bytes)
			logger.Error(e.Message())
			return e
		}

		switch pub := pubk.(type) {
		case *ecdsa.PublicKey:
			idea.PublickKey = pub
		}
	}

	return nil
}

//	=====================================================================
//	function name: GetPubKeyByByte
//	function type: public
//	function receiver: Idea
//	byte array decode into the public key struct, return error
//	=====================================================================
func (idea *Idea) GetPubKeyByByte(PubKey []byte) errors.CallStackError {
	p, rest := pem.Decode(PubKey)
	var err errors.CallStackError
	if p == nil || len(p.Bytes) == 0 {
		err = idea.ParsePub(rest)
	} else {
		err = idea.ParsePub(p.Bytes)
	}
	if err != nil {
		logger.Error(err.Message())
		return err
	}

	return nil
}

//	=====================================================================
//	function name: NewEcdsa
//	function type: public
//	instantiate via a pair of public and private key file,
//	return Idea struct or error
//	=====================================================================
func NewEcdsa(priKeyFilePath, publickKeyPath string) (*Idea, errors.CallStackError) {
	idea := Idea{}
	if len(priKeyFilePath) != 0 {
		idea.PriKeyFilePath = priKeyFilePath
		if e := idea.GetPriKey(idea.PriKeyFilePath); e != nil {
			return nil, e
		}
		idea.PublickKey = &idea.PrivateKey.PublicKey
	}
	if len(publickKeyPath) != 0 {
		idea.PubKeyFilePath = publickKeyPath
		if e := idea.GetPubKey(idea.PubKeyFilePath); e != nil {
			return nil, e
		}
	}

	return &idea, nil
}

//	=====================================================================
//	function name: New
//	function type: public
//	instantiate via a pair of public and private key byte array,
//	return Idea struct or error
//	=====================================================================
func New(priKey, publickKey []byte) (*Idea, errors.CallStackError) {
	idea := Idea{}
	if len(priKey) != 0 {
		out, err := x509.ParseECPrivateKey(priKey)
		if err != nil {
			e := errors.ErrorWithCallstack("0006", "000003", priKey)
			logger.Error(e.Message())
			return nil, e
		}
		idea.PrivateKey = out
		idea.PublickKey = &idea.PrivateKey.PublicKey
	}
	if len(publickKey) != 0 {
		err := idea.ParsePub(publickKey)
		if err != nil {
			logger.Error(err.Message())
			return nil, err
		}
	}

	return &idea, nil
}

//	=====================================================================
//	function name: GetPublickKey
//	function type: public
//	function receiver: Idea
//	marshal public key, return byte array or error
//	=====================================================================
func (idea *Idea) GetPublickKey() ([]byte, errors.CallStackError) {
	pu := &idea.PrivateKey.PublicKey
	puk, err := x509.MarshalPKIXPublicKey(pu)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000002", puk)
		logger.Error(e.Message())
		return nil, e
	}

	return puk, nil
}

//	=====================================================================
//	function name: ParsePub
//	function type: public
//	function receiver: Idea
// 	unmarshal public key, return error
//	=====================================================================
func (idea *Idea) ParsePub(pub []byte) errors.CallStackError {
	pubk, err := x509.ParsePKIXPublicKey(pub)
	if err != nil {
		e := errors.ErrorWithCallstack("0006", "000003", pub)
		logger.Error(e.Message())
		return e
	}

	switch pub := pubk.(type) {
	case *ecdsa.PublicKey:
		idea.PublickKey = pub
	}
	return nil
}

//	=====================================================================
//	function name: MessageEncrypt
//	function type: public
//	function receiver: Idea
//  encrypt message, return byte array or error
//	=====================================================================
func (idea *Idea) MessageEncrypt(mes_hash []byte) ([]byte, errors.CallStackError) {
	//TODO
	return mes_hash, nil
}

//	=====================================================================
//	function name: MessageDecrypt
//	function type: public
//	function receiver: Idea
//	decrypt message, return byte array or error
//	=====================================================================
func (idea *Idea) MessageDecrypt(mes_hash []byte) ([]byte, errors.CallStackError) {
	//TODO
	return mes_hash, nil
}
