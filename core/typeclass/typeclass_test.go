package typeclass

import (
	"fmt"
	"log"
	"strconv"
	"testing"
)

type TestInt int64

func (a TestInt) Eq(b Eq) bool {
	v, ok := b.(TestInt)
	if !ok {
		return false
	}
	return a == v
}

func (a TestInt) NotEq(b Eq) bool {
	return !a.Eq(b)
}

func (a TestInt) LessThan(b Ord) bool {
	v, ok := b.(TestInt)
	if !ok {
		return false
	}
	return a < v
}

func (a TestInt) MoreThan(b Ord) bool {
	return !a.LessThan(b)
}

func (a TestInt) Show() string {
	return strconv.FormatInt(int64(a), 10)
}

func (a *TestInt) Read(s string) error {
	b, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*a = TestInt(b)
	return nil
}

func TestTypeClass(t *testing.T) {
	var a TestInt
	var b TestInt
	a = 1
	b = 1

	fmt.Printf("a = %s, b = %s\n", a.Show(), b.Show())
	fmt.Printf("== %v, != %v\n", a.Eq(b), a.NotEq(b))
	fmt.Printf("< %v, > %v\n", a.LessThan(b), a.MoreThan(b))
	err := a.Read("10")
	if err != nil {
		log.Fatalf("failed to read: %v", err)
	}
	fmt.Printf("a update %s\n", a.Show())
}
