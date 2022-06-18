package crypto

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"github.com/tjfoc/gmsm/sm4"
)

var commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

//sm4对称密钥加密
func Sm4Encryption(key, plainText []byte) ([]byte, error) {
	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData := PKCS5Padding(plainText, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, commonIV)
	cryted := make([]byte, len(origData))
	blockMode.CryptBlocks(cryted, origData)
	return cryted, nil
}

//sm4对称密钥解密
func Sm4Decryption(key, cipherText []byte) ([]byte, error) {
	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, commonIV)
	origData := make([]byte, len(cipherText))
	blockMode.CryptBlocks(origData, cipherText)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

func PKCS5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func PKCS5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

func Encrypt(b []byte) ([]byte, []byte, error) {
	if b == nil {
		return nil, nil, errors.New("empty data")
	}
	key := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, nil, err
	}
	buf, err := Sm4Encryption(key, b)
	return key, buf, err
}
