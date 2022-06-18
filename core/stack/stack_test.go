package stack

import (
	"fmt"
	"testing"

	"github.com/tjfoc/tjfoc/core/typeclass"
)

type TestInt int

func (a TestInt) Eq(b typeclass.Eq) bool {
	v, ok := b.(TestInt)
	if !ok {
		return false
	}
	return a == v
}

func (a TestInt) NotEq(b typeclass.Eq) bool {
	return !a.Eq(b)
}

func Test(t *testing.T) {
	s := New()
	for i := 0; i < 100; i++ {
		s.Push(TestInt(i))
	}
	fmt.Printf("count: %d\n", s.Count())
	fmt.Printf("top: %v\n", s.Top())
	for i := 0; i < 100; i++ {
		fmt.Printf("%d pop: %v\n", i, s.Pop())
	}
	fmt.Printf("count: %d\n", s.Count())
}
