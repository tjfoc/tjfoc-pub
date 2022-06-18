package rbtree

import (
	"fmt"
	"testing"

	"github.com/tjfoc/tjfoc/core/typeclass"
)

type TestInt int

func intPrint(a interface{}, b interface{}) interface{} {
	fmt.Printf("%v\n", b)
	return a
}

func Test(t *testing.T) {
	tree := New()
	for i := 0; i < 100; i++ {
		tree.Insert(TestInt(i))
	}
	for i := 1; i < 100; i += 2 {
		tree.Remove(TestInt(i))
	}
	fmt.Printf("Update = %v\n", tree.Update(TestInt(4), TestInt(5)))
	tree.Foldl(intPrint, TestInt(1))
}

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

func (a TestInt) LessThan(b typeclass.Ord) bool {
	v, ok := b.(TestInt)
	if !ok {
		return false
	}
	return a < v
}

func (a TestInt) MoreThan(b typeclass.Ord) bool {
	return !a.LessThan(b)
}
