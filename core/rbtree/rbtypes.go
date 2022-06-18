package rbtree

import (
	"core/functor"
	"core/typeclass"
	"errors"
)

const (
	R = iota
	B
	BB
)

type Color int8

type TreeNode interface {
	typeclass.Eq
	typeclass.Ord
}

type RBtreeNode struct {
	bbtag bool
	color Color
	node  TreeNode
	left  *RBtreeNode
	right *RBtreeNode
}

type RBtree struct {
	tree *RBtreeNode
}

func (a Color) Eq(b typeclass.Eq) bool {
	v, ok := b.(Color)
	if !ok {
		return false
	}
	return a == v
}

func (a Color) NotEq(b typeclass.Eq) bool {
	return !a.Eq(b)
}

func (a Color) Show() string {
	switch a {
	case R:
		return "RED"
	case B:
		return "BLACK"
	case BB:
		return "DOUBLE BLACK"
	}
	return "UNKOWN"
}

func (a *Color) Read(s string) error {
	switch s {
	case "RED":
		*a = R
	case "BLACK":
		*a = B
	case "DOUBLE BLACK":
		*a = BB
	default:
		return errors.New("unknown color")
	}
	return nil
}

func (a *RBtree) Map(f functor.MapFunc) functor.Functor {
	return a.tree.Map(f)
}

func (b *RBtree) Foldl(f functor.FoldFunc, a interface{}) interface{} {
	return b.tree.Foldl(f, a)
}

func (b *RBtree) Foldr(f functor.FoldFunc, a interface{}) interface{} {
	return b.tree.Foldr(f, a)
}

func (a *RBtreeNode) Map(f functor.MapFunc) functor.Functor {
	if a.isEmpty() {
		return a
	}
	left := a.left
	right := a.right
	left.Map(f)
	f(a.node)
	right.Map(f)
	return a
}

func (b *RBtreeNode) Foldl(f functor.FoldFunc, a interface{}) interface{} {
	if b.isEmpty() {
		return a
	}
	left := b.left
	right := b.right
	a = left.Foldl(f, a)
	a = f(a, b.node)
	a = right.Foldl(f, a)
	return a
}

func (b *RBtreeNode) Foldr(f functor.FoldFunc, a interface{}) interface{} {
	if b.isEmpty() {
		return a
	}
	left := b.left
	right := b.right
	a = right.Foldl(f, a)
	a = f(a, b.node)
	a = left.Foldl(f, a)
	return a
}
