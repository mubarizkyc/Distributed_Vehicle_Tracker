// Create Block
// Mine Block
// Display Blocks
package main

//on creation every node will create its own chain and will get rest of blocks from peers
import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const target string = "00ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" // you can change the difficulty

func (n *Node) DisplayChain() {
	fmt.Println("Current Chain")
	// Get the hash of the latest block
	lastHash := n.blockchain.latestHash

	for lastHash != "" { // unless lastHash!=prevHash og gensis Block
		block := n.blockchain.Blocks[lastHash]

		// Display the block
		block.DisplayBlock(n.storage)

		// Update lastHash to previous block's hash
		lastHash = hex.EncodeToString(block.PreviousHash)
		time.Sleep(2 * time.Second)
	}
}

func (n *Node) CreateChain() {
	a, _ := hex.DecodeString("")
	GenesisBlock := Block{PreviousHash: a, Transactions: []string{"Genesis:Genesis:Genesis"}}
	hash := sha256.Sum256([]byte{})

	GenesisBlock.Hash = hash[:]
	chain := &Blockchain{
		latestHash: hex.EncodeToString(GenesisBlock.Hash),
		Blocks:     make(map[string]*Block),
	}
	chain.Blocks[chain.latestHash] = &GenesisBlock
	n.blockchain = chain
}

type Block struct {
	Hash           []byte
	PreviousHash   []byte
	Transactions   []string
	MerkleRootHash [32]byte
	Nonce          int
}

type Blockchain struct {
	latestHash string // hexadecimal string
	Blocks     map[string]*Block
}

func (n *Node) AddtoChain(b *Block) {

	n.blockchain.Blocks[hex.EncodeToString(b.Hash)] = b
	n.blockchain.latestHash = hex.EncodeToString(b.Hash)

}

func (b *Block) CalculateHash(nonce int) []byte {
	header := fmt.Sprintf("%x", b.PreviousHash) // conversion to string for hashng

	for _, d := range b.Transactions {
		header += d
	}

	header += fmt.Sprintf("%d", nonce)

	hash := sha256.Sum256([]byte(header))
	b.Hash = hash[:]
	return b.Hash
}
func (b *Block) TransactionsHash() {
	mt := CreateMerkelTree(b.Transactions)

	b.MerkleRootHash = [32]byte(mt.root.Hash)
}
func CreateBlock(prevHash []byte, transactions []string) (*Block, error) {
	newBlock := &Block{
		PreviousHash: prevHash,
		Transactions: transactions,
		Nonce:        0,
	}
	nonce, hash, err := MineBlock(newBlock)
	if err != nil {
		return &Block{}, err
	}
	newBlock.Nonce = nonce
	newBlock.Hash = hash[:]
	if len(transactions) != 0 { // for genisis block
		newBlock.TransactionsHash()
	}
	newBlock.CalculateHash(newBlock.Nonce)
	return newBlock, nil
}
func (n *Node) Mine() {
	// check for tx validity and remove the invalid form then list

	if len(n.transactions) == 0 {
		return
	}
	GenisisHash, _ := hex.DecodeString(n.blockchain.latestHash)
	block, err := CreateBlock(GenisisHash, n.transactions)
	n.transactions = []string{} // remove from pool as u will validate all tx at Once
	if err != nil {
		fmt.Println("Error Minning Block", err)
		return
	}

	n.AddtoChain(block)

}
func MineBlock(block *Block) (int, []byte, error) {
	timeout := 10 * time.Second
	targetBytes, err := hex.DecodeString(target)
	if err != nil {
		return 0, []byte{}, err
	}

	startTime := time.Now() // Get the current time

	nonce := 0

	for {
		hash := block.CalculateHash(nonce)

		// Check if the hash meets the target criteria
		if compareHashWithTarget(hash[:], targetBytes) {
			return nonce, hash, nil // Return the nonce and hash if found
		}

		// Check if the time elapsed exceeds the timeout duration
		if time.Since(startTime) >= timeout {
			return 0, []byte{}, fmt.Errorf("timeout occurred while mining") // Return an error if timeout
		}

		nonce++ // Increment the nonce for the next iteration
	}

}
func (b Block) DisplayBlock(storage *Storage) {
	//fmt.Printf("PreviousHash %x\n:", b.PreviousHash)

	//fmt.Printf("Hash %x\n:", b.Hash)
	fmt.Println("###   Block   ###")
	for _, tx := range b.Transactions {

		parts := strings.Split(tx, ":")

		if parts[0] == "RegisterVehicle" {
			var vehicle RegisterVehicle
			_ = storage.RetrieveData(parts[1], &vehicle)

			if len(vehicle.Owner) > 0 {
				fmt.Println("Register Vehicle")
				fmt.Println("   Vehicle :", vehicle.VIN)
				fmt.Println("     Owner :", vehicle.Owner[:6]+"...")
			}
		} else if parts[0] == "ChangeOwnership" {
			var vehicle ChangeOwnership
			_ = storage.RetrieveData(parts[1], &vehicle)
			if len(vehicle.From) > 0 && len(vehicle.To) > 0 {
				fmt.Println("Transaction Change Ownership")
				fmt.Println("   Vehicle :", vehicle.VIN)
				fmt.Println("     From :", vehicle.From[:6]+"...")
				fmt.Println("     To :", vehicle.To[:6]+"...")
			}

		}

	}
}

func compareHashWithTarget(hash, target []byte) bool {
	// the idea as ,as we reach the first byte differenct the hash byte should be the smaller One
	for i := 0; i < len(hash); i++ {

		if hash[i] < target[i] {

			return true
		} else if hash[i] > target[i] {
			return false
		}
	}
	return true
}
