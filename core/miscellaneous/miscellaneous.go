package miscellaneous

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
)

func Sha1Hash(data []byte) []byte {
	h := sha1.New()
	h.Write(data)
	return h.Sum(nil)
}

func Sha256Hash(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

func Sha512Hash(data []byte) []byte {
	h := sha512.New()
	h.Write(data)
	return h.Sum(nil)
}

//将[]byte以16进制打印
func Byte2String(data []byte) string {
	fmt.Sprintf("%x", data)
}
