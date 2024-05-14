package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"net"
	"strings"
)

type RegisterVehicle struct {
	Owner string
	VIN   string
}
type ChangeOwnership struct {
	From string
	To   string
	VIN  string
}

func SendTransaction(addr string, msg Message) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Encode the transaction object using gob
	data, err := EncodeMessage(&msg)
	if err != nil {
		log.Println("Error encoding transaction:", err)
		return
	}

	// Send the encoded transaction to the server
	_, err = conn.Write(data)
	if err != nil {
		log.Println("Error sending transaction to:", addr, err)
		return
	}

	// Notify completion of transaction propagation
	transactionDone.Add(1)
}

func SignTransaction(tx Transaction, privateKey *ecdsa.PrivateKey) ([]byte, error) {

	// Calculate the hash of the transaction data
	hash := sha256.Sum256([]byte(fmt.Sprintf(tx.Tx)))

	// Sign the hash using the private key
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return []byte{}, err
	}

	// Encode the signature (r, s) into bytes
	signature := append(r.Bytes(), s.Bytes()...)

	return signature, nil
}
func (tx Transaction) GetPublicKey() *ecdsa.PublicKey {
	parts := strings.Split(tx.Tx, ":")
	xxx, _ := hex.DecodeString(parts[1])
	pubKey, _ := DecodePublicKeyFromBytes(xxx)
	return pubKey

}
func VerifyTransaction(tx *Transaction) bool {

	// Decode the public key from Base64
	fmt.Println(tx.GetPublicKey())
	hash := sha256.Sum256([]byte(fmt.Sprintf(tx.Tx)))

	// Extract the signature (r, s) from the transaction
	rBytes := tx.Signature[:len(tx.Signature)/2]
	sBytes := tx.Signature[len(tx.Signature)/2:]

	// Parse the signature components (r, s)
	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)

	fmt.Println("Verifying Signature...")

	if tx.GetPublicKey() == nil {
		// Handle nil public key error gracefully
		fmt.Println("Error: Public key is nil")
		return false
	}
	// Validate other input parameters (hash, r, s) as necessary

	// Perform ECDSA signature verification
	verified := ecdsa.Verify(tx.GetPublicKey(), hash[:], r, s)

	fmt.Println("Signature Verification Result:", verified)

	return verified

}
