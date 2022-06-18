package verify

import (
	"fmt"
	"testing"
)

func TestGetMostResult(t *testing.T) {

	r := [][]byte{{1}, {2}, {3}, {1}, {1}}

	b, n := getTheMost(r)

	if b == nil {
		fmt.Println("error!")
	} else {
		fmt.Println("result:", b)
		fmt.Println("len:", n)
	}
}
