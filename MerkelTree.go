package main

import (
	"crypto/sha256"
	"fmt"
)

// my merkle node consists of hashes
type MerkleNode struct {
	left        *MerkleNode
	right       *MerkleNode
	Transaction string
	Hash        []byte // 32 byte
}

type MerkleTree struct {
	root *MerkleNode
}

func NewMerkleNode(left *MerkleNode, right *MerkleNode, data []byte, transaction string) *MerkleNode {
	newNode := &MerkleNode{nil, nil, transaction, data}
	if left == nil && nil == right {
		hash := sha256.Sum256([]byte(data))
		newNode.Hash = hash[:]
	} else {
		prevHashes := append(left.Hash, right.Hash...)
		hash := sha256.Sum256(prevHashes)
		newNode.Hash = hash[:]

	}
	newNode.left = left
	newNode.right = right
	return newNode
}
func tranHash(transactions []string) [][]byte {
	var hashes [][]byte
	for i := 0; i < len(transactions); i++ {
		hash := sha256.Sum256([]byte(transactions[i]))
		hashes = append(hashes, hash[:])
	}
	return hashes
}
func transaction_to_Nodes(data [][]byte, transactions []string) []*MerkleNode {
	var nodes []*MerkleNode
	for i := 0; i < len(data); i++ {
		node := NewMerkleNode(nil, nil, data[i], transactions[i])
		nodes = append(nodes, node)
	}
	return nodes
}

// We COnstruct the tree on the basis of given slice not the typical  insertion processes
// check if the length of transactions is odd make it even
func CreateMerkelTree(transactions []string) *MerkleTree {

	txHashes := tranHash(transactions)
	myTree := &MerkleTree{nil}
	Nodes := transaction_to_Nodes(txHashes, transactions)

	// make sure the no of transactions are even
	if len(transactions)%2 != 0 {
		Nodes = append(Nodes, Nodes[len(Nodes)-1])
		transactions = append(transactions, transactions[len(transactions)-1])
	}

	// form bottom to top  doing a level order transversal to create each level
	for len(Nodes) > 1 {
		var upperLevel []*MerkleNode

		for i := 0; i < len(Nodes); i += 2 {
			// In Merkle Tree,transactions are only part of leaf nodes
			node := NewMerkleNode(Nodes[i], Nodes[i+1], nil, "")
			upperLevel = append(upperLevel, node)
		}

		Nodes = upperLevel
	}
	myTree.root = Nodes[0]
	return myTree
}
func DisplayMerkleTree(root *MerkleNode, prefix string, isLeft bool) {
	if root != nil {
		fmt.Printf("%s", prefix)
		if isLeft {
			fmt.Printf("├── ")
		} else {
			fmt.Printf("└── ")
		}
		if root.Transaction != "" { // Check if it's a leaf node
			fmt.Printf(root.Transaction)
		} else {
			fmt.Printf("Hash: %x", root.Hash)
		}
		fmt.Println()

		// Recursive calls for child nodes
		DisplayMerkleTree(root.left, prefix+"│   ", true)
		DisplayMerkleTree(root.right, prefix+"    ", false)
	}
}
