package crypto

import (
	"github.com/spf13/viper"
	"github.com/tjfoc/tjfoc/core/common/flogging"
)

type Cipher interface {
	//	=====================================================================
	//	function name: GetAymCryptionByFile
	//	function type: public
	//	function receiver: Idea
	//	instantiate via a pair of public and private key files and
	//	using ecdsa or rsa algorithm, return AymCryption
	//	=====================================================================
	GetAymCryptionByFile(privateKeyPath, publicKeyPath string) AymCryption
	//	=====================================================================
	//	function name: GetAymCryption
	//	function type: public
	//	function receiver: Idea
	//	instantiate via a pair of public and private key byte array and
	//	using ecdsa or rsa algorithm, return AymCryption
	//	=====================================================================
	GetAymCryption(privateKey, publicKey []byte) AymCryption
	//	=====================================================================
	//	function name: GetAymCryption
	//	function type: public
	//	function receiver: Idea
	//	instantiate using aes or des algorithmalgorithm, return SymCryption
	//	=====================================================================
	GetSymCryption() SymCryption
	//	=====================================================================
	//	function name: Hash
	//	function type: public
	//	function receiver: Idea
	//	select hash algorithm via hash type and source byte array,
	//	return the hash result(byte array)
	//	=====================================================================
	Hash(hashtype string, src []byte) []byte
}

var logger = flogging.MustGetLogger("crypto")

//	=====================================================================
//	function name: GetCipherHelper
//	function type: public
//	function receiver: na
//	load the cryptographic algorithm(Idea or SM) via viper config file,
//	route the cryptographic algorithm implement
//	=====================================================================
func GetCipherHelper() Cipher {
	Cryptotype := viper.GetString("crypto.type")
	switch Cryptotype {
	//	case "sm":
	//		return NewSMS()
	case "idea":
		return NewIdea()
	default:
		panic("not found Cryptotype for " + Cryptotype)
	}

}

//	=====================================================================
//	function name: PayloadHash
//	function type: public
//	function receiver: na
//	load the payload hash algorithm(Idea or SM) via viper config file,
//	return byte array
//	=====================================================================
func PayloadHash(payload []byte) []byte {
	hashtype := viper.GetString("crypto.hash.payload")

	return GetCipherHelper().Hash(hashtype, payload)
}

//	=====================================================================
//	function name: Hash
//	function type: public
//	function receiver: na
//	load the hash algorithm(Idea or SM) via viper config file,
//	return byte array
//	=====================================================================
func Hash(src []byte) []byte {
	hashtype := viper.GetString("crypto.hash")

	return GetCipherHelper().Hash(hashtype, src)
}
