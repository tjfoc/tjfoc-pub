package rbtree

import (
	"github.com/tjfoc/tjfoc/core/functor"
	"github.com/tjfoc/tjfoc/core/typeclass"
)

const (
	R = iota
	B
	BB
)

type Color int8

// tree node must implement the eq and ord interface
type TreeNode interface {
	typeclass.Eq
	typeclass.Ord
}

// typical rbtree structure
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

// functor map
func (a *RBtree) Map(f functor.MapFunc) functor.Functor {
	return a.tree.Map(f)
}

// functor foldl
func (b *RBtree) Foldl(f functor.FoldFunc, a interface{}) interface{} {
	return b.tree.Foldl(f, a)
}

// functor foldr
func (b *RBtree) Foldr(f functor.FoldFunc, a interface{}) interface{} {
	return b.tree.Foldr(f, a)
}

// functor map
func (a *RBtreeNode) Map(f functor.MapFunc) functor.Functor {
	if a.isEmpty() {
		return a
	}
	left := a.left
	right := a.right
	left.Map(f)
	v, _ := f(a.node).(TreeNode)
	a.node = v
	right.Map(f)
	return a
}

// functor foldl
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

// functor foldr
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
