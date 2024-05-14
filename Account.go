package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"log"
	"math/big"
)

func EncodePublicKeyToBase64(publicKey *ecdsa.PublicKey) string {
	// Serialize the public key to bytes
	publicKeyBytes := elliptic.MarshalCompressed(publicKey.Curve, publicKey.X, publicKey.Y)
	// Encode the bytes to Base64
	base64String := base64.StdEncoding.EncodeToString(publicKeyBytes)
	return base64String
}

// DecodePublicKeyFromBase64 decodes a Base64 string to an ECDSA public key
func DecodePublicKeyFromBase64(base64String string) (*ecdsa.PublicKey, error) {
	// Decode the Base64 string to bytes
	publicKeyBytes, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return nil, err
	}
	// Unmarshal the bytes to an ECDSA public key
	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(publicKeyBytes[:32]), // X coordinate
		Y:     new(big.Int).SetBytes(publicKeyBytes[32:]), // Y coordinate
	}
	return publicKey, nil
}

// EncodePrivateKeyToBase64 encodes an ECDSA private key to Base64.
func EncodePrivateKeyToBase64(privateKey *ecdsa.PrivateKey) (string, error) {
	// Marshal the private key to bytes
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	// Encode the bytes to Base64
	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKeyBytes)
	return privateKeyBase64, nil
}

// DecodePrivateKeyFromBase64 decodes a Base64-encoded ECDSA private key.
func DecodePrivateKeyFromBase64(privateKeyBase64 string) (*ecdsa.PrivateKey, error) {
	// Decode the Base64 string to bytes
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, err
	}

	// Parse the bytes to an ECDSA private key
	privateKey, err := x509.ParseECPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
func NewKeyPair() *ecdsa.PrivateKey {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	// Return both private and public keys
	return private
}

type Account struct {
	PublicKey  *ecdsa.PublicKey
	PrivateKey *ecdsa.PrivateKey
}

func MakeAccount() *Account {
	private := NewKeyPair()
	public := &private.PublicKey
	acc := Account{public, private}

	return &acc
}
