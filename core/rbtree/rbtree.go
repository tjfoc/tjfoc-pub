package rbtree

func T(color Color, left, node, right *RBtreeNode) *RBtreeNode {
	node.left = left
	node.right = right
	node.color = color
	return node
}

func RBtreeNew() *RBtree {
	return &RBtree{nil}
}

func (a *RBtree) Elem(node TreeNode) TreeNode {
	treeNode := a.tree.elem(node)
	if treeNode == nil {
		return nil
	}
	return treeNode.node
}

func (a *RBtree) Update(old, new TreeNode) bool {
	treeNode := a.tree.elem(old)
	if treeNode == nil {
		return false
	}
	treeNode.node = new
	return true
}

func (a *RBtree) Insert(node TreeNode) bool {
	return a.InsertP(&node)
}

func (a *RBtree) Remove(node TreeNode) bool {
	return a.RemoveP(&node)
}

func (a *RBtree) InsertP(node *TreeNode) bool {
	a.tree = a.tree.insert(&RBtreeNode{
		bbtag: false,
		color: B,
		node:  *node,
		left:  nil,
		right: nil,
	})
	return true
}

func (a *RBtree) RemoveP(node *TreeNode) bool {
	a.tree = a.tree.remove(&RBtreeNode{
		bbtag: false,
		color: B,
		node:  *node,
		left:  nil,
		right: nil,
	})
	return true
}

func (a *RBtreeNode) isEmpty() bool {
	return a == nil
}

func (a *RBtreeNode) isBBEmpty() bool {
	return a != nil && a.bbtag
}

func (a *RBtree) isEmpty() bool {
	return a == nil || a.tree.isEmpty()
}

func (a *RBtreeNode) min() *RBtreeNode {
	if a.isEmpty() || a.left.isEmpty() {
		return a
	}
	return a.left.min()
}

func (a *RBtreeNode) elem(node TreeNode) *RBtreeNode {
	if a.isEmpty() {
		return nil
	}
	if a.node.Eq(node) {
		return a
	}
	if node.LessThan(a.node) {
		return a.left.elem(node)
	}
	return a.right.elem(node)
}

func (a *RBtreeNode) makeblack() *RBtreeNode {
	if a.isEmpty() || a.isBBEmpty() {
		return nil
	}
	a.color = B
	return a
}

func (a *RBtreeNode) makeblack1() *RBtreeNode {
	if a.isEmpty() {
		return &RBtreeNode{
			bbtag: true,
		}
	}
	switch a.color {
	case B:
		a.color = BB
	default:
		a.color = B
	}
	return a
}

func (a *RBtreeNode) insert(node *RBtreeNode) *RBtreeNode {
	return ins(a, node).makeblack()
}

func (a *RBtreeNode) remove(node *RBtreeNode) *RBtreeNode {
	return del(a, node).makeblack()
}

func ins(a, node *RBtreeNode) *RBtreeNode {
	if a.isEmpty() {
		return T(R, nil, node, nil)
	}
	if a.node.Eq(node.node) {
		return a
	}
	if node.node.LessThan(a.node) {
		return balance(a.color, ins(a.left, node), a, a.right)
	} else {
		return balance(a.color, a.left, a, ins(a.right, node))
	}
}

func del(a, node *RBtreeNode) *RBtreeNode {
	if a.isEmpty() {
		return a
	}
	if a.node.Eq(node.node) {
		if a.left.isEmpty() {
			if a.color == B {
				return a.right.makeblack1()
			}
			return a.right
		}
		if a.right.isEmpty() {
			if a.color == B {
				return a.left.makeblack1()
			}
			return a.left
		}
		min := a.right.min()
		return fixdb(a.color, a.left, min, del(a.right, min))
	}
	if node.node.LessThan(a.node) {
		return fixdb(a.color, del(a.left, node), a, a.right)
	} else {
		return fixdb(a.color, a.left, a, del(a.right, node))
	}
}

func (a *RBtreeNode) set(b *RBtreeNode) bool {
	if b == nil {
		return false
	}
	*a = *b
	return true
}

func balance(color Color, l, a, r *RBtreeNode) *RBtreeNode {
	var t RBtreeNode

	if color == B {
		if l != nil && l.color == R && t.set(l.left) && t.color == R {
			return T(R, l.left.makeblack(), l, T(B, l.right, a, r))
		}
		if l != nil && l.color == R && t.set(l.right) && t.color == R {
			return T(R, T(B, l.left, l, t.left), &t, T(B, t.right, a, r))
		}
		if r != nil && r.color == R && t.set(r.right) && t.color == R {
			return T(R, T(B, l, a, r.left), r, r.right.makeblack())
		}
		if r != nil && r.color == R && t.set(r.left) && t.color == R {
			return T(R, T(B, l, a, t.left), &t, T(B, t.right, r, r.right))
		}
	}
	return T(color, l, a, r)
}

func fixdb(color Color, l, a, r *RBtreeNode) *RBtreeNode {
	var t, t1 RBtreeNode

	if l.isBBEmpty() && r.isEmpty() {
		return T(BB, nil, a, nil)
	}
	if l.isEmpty() && r.isBBEmpty() {
		return T(BB, nil, a, nil)
	}
	if l.isBBEmpty() {
		return T(color, nil, a, r)
	}
	if r.isBBEmpty() {
		return T(color, l, a, nil)
	}
	if l != nil && l.color == BB && r != nil && r.color == B && t.set(r.left) && t.color == R {
		return T(color, T(B, l.makeblack1(), a, t.left), &t, T(B, t.right, r, r.right))
	}
	if l != nil && l.color == BB && r != nil && r.color == B && t.set(r.right) && t.color == R {
		return T(color, T(B, l.makeblack1(), a, r.left), r, T(B, t.left, &t, t.right))
	}
	if r != nil && r.color == BB && l != nil && l.color == B && t.set(l.right) && t.color == R {
		return T(color, T(B, l.left, l, t.left), &t, T(B, t.right, a, r.makeblack1()))
	}
	if r != nil && r.color == BB && l != nil && l.color == B && t.set(l.left) && t.color == R {
		return T(color, T(B, t.left, &t, t.right), l, T(B, l.right, a, r.makeblack1()))
	}
	if l != nil && l.color == BB && r != nil && r.color == B && t.set(r.left) && t.color == B &&
		t1.set(r.right) && t.color == B {
		return T(color, l.makeblack1(), a, T(R, &t, r, &t1)).makeblack1()
	}
	if r != nil && r.color == BB && l != nil && l.color == B && t.set(l.left) && t.color == B &&
		t1.set(l.right) && t.color == B {
		return T(color, T(R, &t, l, &t1), a, r.makeblack1()).makeblack1()
	}
	if color == B && l != nil && l.color == BB && r != nil && r.color == R {
		return fixdb(B, fixdb(R, l, a, r.left), r, r.right)
	}
	if color == B && r != nil && r.color == BB && l != nil && l.color == R {
		return fixdb(B, l.left, l, fixdb(R, l.right, a, r))
	}
	return T(color, l, a, r)
}
