/*
Copyright Suzhou Tongji Fintech Research Institute 2018 All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package merkle

import (
	"hash"

	"github.com/tjfoc/tjfoc/core/miscellaneous"
)

// a new hash struct
func mkRoot(hash []byte) *MerkleRoot {
	return &MerkleRoot{
		root: hash,
	}
}

// create hash of leaf node
func mkLeafRootHash(h hash.Hash, data []byte) *MerkleRoot {
	hashData, _ := miscellaneous.GenHash(h, data)
	return mkRoot(hashData)
}

// create hash of branch node
func mkRootHash(h hash.Hash, a MKRoot, b MKRoot) *MerkleRoot {
	hashData, _ := miscellaneous.GenHash(h, append(a.getMerkleRoot(), b.getMerkleRoot()...))
	return mkRoot(hashData)
}

func powerOfTwo0(n int) int {
	if n&(n-1) == 0 {
		return n
	} else {
		return powerOfTwo0(n & (n - 1))
	}
}

// calculate the minimum value of 2 ^ n for this party
func powerOfTwo(n int) int {
	if n&(n-1) == 0 {
		return n >> 1
	}
	return powerOfTwo0(n)
}

// create a new merkle tree
func MkMerkleTree(hash hash.Hash, data [][]byte) *MerkleTree {
	if data == nil {
		return nil
	}
	len := len(data)
	if len == 0 {
		return nil
	}
	r := &MerkleTree{
		hash: hash,
	}
	r.tree = r.mkMerkleTreeRoot(len, data)
	return r
}

// creat proof path
func constructPath(proof MerkleProof, leaf MKNode, node MKNode) MerkleProof {
	if leaf.isLeafNode() {
		if leaf.getMKRoot().Eq(node.getMKRoot()) {
			return proof
		}
		return MerkleProof{}
	}
	ln := leaf.leftNode()
	rn := leaf.rightNode()
	lProofElem := ProofElem{
		isLeft:      true,
		nodeRoot:    ln.getMKRoot(),
		siblingRoot: rn.getMKRoot(),
	}
	rProofElem := ProofElem{
		isLeft:      false,
		nodeRoot:    rn.getMKRoot(),
		siblingRoot: ln.getMKRoot(),
	}
	lpath := constructPath(append(proof, lProofElem), ln, node)
	rpath := constructPath(append(proof, rProofElem), rn, node)
	return append(lpath, rpath...)
}

// creat proof path
func (t *MerkleTree) merkleProof(node MKNode) MerkleProof {
	return constructPath(MerkleProof{}, t.GetMtRoot(), node)
}

// check whether the data is in a given proof path
func validate(hash hash.Hash, proof MerkleProof, root MKNode, leaf MKNode) bool {
	len := len(proof)
	if len == 0 {
		return leaf.getMKRoot().Eq(root.getMKRoot())
	}
	if !proof[len-1].nodeRoot.Eq(leaf.getMKRoot()) {
		return false
	}
	var node *MerkleNode
	if proof[len-1].isLeft {
		node = &MerkleNode{
			root: mkRootHash(hash, proof[len-1].nodeRoot, proof[len-1].siblingRoot),
		}
	} else {
		node = &MerkleNode{
			root: mkRootHash(hash, proof[len-1].siblingRoot, proof[len-1].nodeRoot),
		}
	}
	return validate(hash, proof[:len-1], root, node)

}

// check whether the data is in a given proof path
func (t *MerkleTree) validateMerkleProof(proof MerkleProof, node MKNode) bool {
	return validate(t.hash, proof, t.GetMtRoot(), node)
}

// detects whether the data exists in the tree
func (t *MerkleTree) VerifyNode(data []byte) bool {
	node := t.mkLeaf(data)
	proof := t.merkleProof(node)
	return t.validateMerkleProof(proof, node)
}
