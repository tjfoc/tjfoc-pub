package tools

import (
	"fmt"
	"testing"
)

func TestPrint(t *testing.T) {

	for i := 0; i < 100; i++ {

		fmt.Println(GenNum())
	}
}
