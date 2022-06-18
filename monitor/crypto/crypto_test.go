package crypto

import (
	"crypto/rand"
	"fmt"
	"io"
	"testing"
)

var data1 = []byte{1, 2, 3, 4, 5}
var data2 = []byte{1, 2, 3, 4, 5, 6}
var data3 = []byte{1, 2, 3, 4, 5, 6, 7}

func TestSm4(t *testing.T) {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		fmt.Println(err)
	}

	buf, err := Sm4Encryption(b, data1)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("data1:", data1)
	fmt.Println("encryption data:", buf)
	buff, err := Sm4Decryption(b, buf)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("decryption data:", buff)

}
