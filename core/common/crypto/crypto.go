package crypto

import (
	"github.com/tjfoc/tjfoc/core/common/errors"
)

//	=====================================================================
//	interface name:Datahash
//	interface type: public
//	interface function: DataHashByType
//	=====================================================================
type Datahash interface {
	//	=====================================================================
	//	function name: DataHashByType
	//	function type: public
	//	function receiver: na
	//  switch hashtype(sha256/ripemd160/md5/sha1/doublesha256/ShakeSum256)
	//	return byte array
	//	=====================================================================
	DataHashByType(data string) []byte
}

//	=====================================================================
//	interface name:SymCryption
//	interface type: public
//	interface function: Genkeys/Encrypt/Decrypt
//	=====================================================================
type SymCryption interface {
	//	=====================================================================
	//	function name: Genkeys
	//	function type: public
	//	function receiver: AesEncrypt
	//  generate a key, return byte array
	//	====================================================================
	Genkeys() []byte

	//	=====================================================================
	//	function name: Encrypt
	//	function type: public
	//	function receiver: AesEncrypt
	//  encrypt message, return byte array or error
	//	=====================================================================
	Encrypt(mes_hash, key []byte) ([]byte, errors.CallStackError)

	//	=====================================================================
	//	function name: Decrypt
	//	function type: public
	//	function receiver: AesEncrypt
	//  decrypt message, return byte array or error
	//	=====================================================================
	Decrypt(mes_hash, key []byte) ([]byte, errors.CallStackError)
}

//	=====================================================================
//	interface name:AymCryption
//	interface type: public
//	interface function: GenerateKeys/WriteKeyFile/GetPublickKey/GetPriKey
//						GetPubKey/Sign/Verify/MessageEncrypt/MessageDecrypt
//	=====================================================================
type AymCryption interface {
	//	=====================================================================
	//	function name: GenerateKeys
	//	function type: public
	//	function receiver: Idea/PKCS1v15/SMS2
	//  generate a pair of public and private key, return error
	//	=====================================================================
	GenerateKeys() errors.CallStackError

	//	=====================================================================
	//	function name: WriteKeyFile
	//	function type: public
	//	function receiver: Idea/PKCS1v15/SMS2
	//	write a pair of public and private key into two files, return error
	//	=====================================================================
	WriteKeyFile() errors.CallStackError

	//	=====================================================================
	//	function name: GetPublickKey
	//	function type: public
	//	function receiver: Idea/PKCS1v15/SMS2
	//	marshal public key, return byte array or error
	//	=====================================================================
	GetPublickKey() ([]byte, errors.CallStackError)

	//	=====================================================================
	//	function name: GetPriKey
	//	function type: public
	//	function receiver: Idea/PKCS1v15/SMS2
	//	get private key from underlying file, return error
	//	=====================================================================
	GetPriKey(pivKeyFilePath string) errors.CallStackError

	//	=====================================================================
	//	function name: GetPubKey
	//	function type: public
	//	function receiver: Idea/PKCS1v15/SMS2
	//	get public key from underlying file, return error
	//	=====================================================================
	GetPubKey(pubKeyFilePath string) errors.CallStackError

	//	=====================================================================
	//	function name: Sign
	//	function type: public
	//	function receiver: Idea/PKCS1v15/SMS2
	//	generate signature according to the private key,
	//	return byte array or error
	//	=====================================================================
	Sign(hash []byte) ([]byte, errors.CallStackError)

	//	=====================================================================
	//	function name: Verify
	//	function type: public
	//	function receiver: Idea/PKCS1v15/SMS2
	//	verify signature, return bool
	//	=====================================================================
	Verify(src, sign, pubKey []byte) bool

	//	=====================================================================
	//	function name: MessageEncrypt
	//	function type: public
	//	function receiver: Idea/PKCS1v15/SMS2
	//  encrypt message, return byte array or error
	//	=====================================================================
	MessageEncrypt(mes_hash []byte) ([]byte, errors.CallStackError)

	//	=====================================================================
	//	function name: MessageDecrypt
	//	function type: public
	//	function receiver: Idea/PKCS1v15/SMS2
	//	decrypt message, return byte array or error
	//	=====================================================================
	MessageDecrypt(mes_hash []byte) ([]byte, errors.CallStackError)
}
