package hashcryption

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
)

//	=====================================================================
//	function name: Sha256Hash
//	function type: public
//	function receiver: na
//	GO SDK package（crypto） hash function（Sha256Hash）
//	Sha256Hash calculates the SHA256 hash algorithms
//	and returns the resulting bytes.
//	=====================================================================
func Sha256Hash(b []byte) []byte {
	h := sha256.New()
	h.Write(b)
	bytes := h.Sum(nil)

	return bytes
}

//	=====================================================================
//	function name: Ripemd160Hash
//	function type: public
//	function receiver: na
//	GO SDK package（crypto） hash function（Ripemd160Hash）
//	Ripemd160Hash calculates the RIPEMD-160 hash algorithm
//	and returns the resulting bytes.
//	=====================================================================
func Ripemd160Hash(b []byte) []byte {
	h := ripemd160.New()
	h.Write(b)
	bytes := h.Sum(nil)

	return bytes
}

//	=====================================================================
//	function name: Sha1Hash
//	function type: public
//	function receiver: na
//	GO SDK package（crypto） hash function（Sha1Hash）
//	Sha1Hash calculates the SHA1 hash algorithm
//	and returns the resulting bytes.
//	=====================================================================
func Sha1Hash(b []byte) []byte {
	h := sha1.New()
	h.Write(b)
	bytes := h.Sum(nil)

	return bytes
}

//	=====================================================================
//	function name: DoubleSha256
//	function type: public
//	function receiver: na
//	GO SDK package（crypto） hash function（DoubleSha256）
//	DoubleSha256 calculates the double SHA256 hash algorithms
//	and returns the resulting bytes.
//	=====================================================================
func DoubleSha256(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])

	return second[:]
}

//	=====================================================================
//	function name: ShakeSum256
//	function type: public
//	function receiver: na
//	GO SDK package（crypto） hash function（ShakeSum256）
//	ShakeSum256 writes an arbitrary-length digest of data into hash
//	and returns the resulting bytes.
//	=====================================================================
func ShakeSum256(b []byte) []byte {
	hash := make([]byte, 64)
	sha3.ShakeSum256(hash, b)

	return hash
}

//	=====================================================================
//	function name: ShakeSum256
//	function type: public
//	function receiver: na
//	GO SDK package（crypto） hash function（Md5Hash）
//	Md5Hash calculates the MD5 hash algorithm
//	and returns the resulting strings.
//	=====================================================================
func Md5Hash(s string) string {
	h := md5.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}
