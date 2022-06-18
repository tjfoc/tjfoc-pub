package crypto

import (
	"github.com/spf13/viper"
	//	ecdsa "github.com/tjfoc/tjfoc/core/crypto/idea/aymcryption/ecdsa"
	"github.com/tjfoc/tjfoc/core/common/crypto/idea/aymcryption/ecdsa"
	"github.com/tjfoc/tjfoc/core/common/crypto/idea/aymcryption/rsa"
	"github.com/tjfoc/tjfoc/core/common/crypto/idea/hashcryption"
	"github.com/tjfoc/tjfoc/core/common/crypto/idea/symcryption"
	//rsa "github.com/tjfoc/tjfoc/core/crypto/idea/aymcryption/rsa"
	//hash "github.com/tjfoc/tjfoc/core/crypto/idea/hashcryption"
)

type Idea struct {
}

//	=====================================================================
//	function name: GetAymCryptionByFile
//	function type: public
//	function receiver: Idea
//	instantiate via a pair of public and private key files and
//	using ecdsa or rsa algorithm, return AymCryption
//	=====================================================================
func (idea *Idea) GetAymCryptionByFile(privateKeyPath, publicKeyPath string) AymCryption {
	algotype := viper.GetString("crypto.algo")
	switch algotype {
	case "ecdsa":
		ecdsa, e := ecdsa.NewEcdsa(privateKeyPath, publicKeyPath)
		if e != nil {
			logger.Error(e.Message())
		}
		return ecdsa

	case "rsa":
		rsa, e := rsa.NewRSA(privateKeyPath, publicKeyPath)
		if e != nil {
			logger.Error(e.Message())
		}
		return rsa

	default:
		panic("not found algotype for " + algotype)
	}
}

//	=====================================================================
//	function name: GetAymCryption
//	function type: public
//	function receiver: Idea
//	instantiate via a pair of public and private key byte array and
//	using ecdsa or rsa algorithm, return AymCryption
//	=====================================================================
func (idea *Idea) GetAymCryption(privateKey, publicKey []byte) AymCryption {
	algotype := viper.GetString("crypto.algo")
	switch algotype {
	case "ecdsa":
		ecdsa, e := ecdsa.New(privateKey, publicKey)
		if e != nil {
			logger.Error(e.Message())
		}
		return ecdsa

	case "rsa":
		rsa, e := rsa.New(privateKey, publicKey)
		if e != nil {
			logger.Error(e.Message())
		}
		return rsa

	default:
		panic("not found algotype for " + algotype)
	}
}

//	=====================================================================
//	function name: GetAymCryption
//	function type: public
//	function receiver: Idea
//	instantiate using aes or des algorithmalgorithm, return SymCryption
//	=====================================================================
func (idea *Idea) GetSymCryption() SymCryption {
	Symtype := viper.GetString("crypto.sym")
	switch Symtype {
	case "aes":
		aes := symcryption.New()
		return aes

	case "des":
		return nil

	default:
		panic("not found Symtype for " + Symtype)
	}
}

//	=====================================================================
//	function name: Hash
//	function type: public
//	function receiver: Idea
//	select hash algorithm via hash type and source byte array,
//	return the hash result(byte array)
//	=====================================================================
func (idea *Idea) Hash(hashtype string, src []byte) []byte {
	DataHash := hashcryption.DataHash{
		Type: hashtype,
	}
	return DataHash.DataHashByType(string(src))
}

//	=====================================================================
//	function name: Hash
//	function type: public
//	function receiver: Idea
//	instantiate Idea
//	=====================================================================
func NewIdea() *Idea {
	return &Idea{}
}
