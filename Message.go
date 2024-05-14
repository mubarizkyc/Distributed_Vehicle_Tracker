package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"sync"
)

func EncodePublicKeyToBytes(pubKey *ecdsa.PublicKey) ([]byte, error) {
	// Marshal the public key to DER format
	derBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, err
	}
	return derBytes, nil
}

// DecodePublicKeyFromBytes decodes a byte slice to a public key
func DecodePublicKeyFromBytes(derBytes []byte) (*ecdsa.PublicKey, error) {
	// Parse the DER-encoded public key
	pubKeyInterface, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		return nil, err
	}

	// Assert that the parsed public key is of type *ecdsa.PublicKey
	pubKey, ok := pubKeyInterface.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to parse public key: not an ECDSA public key")
	}

	return pubKey, nil
}

var privateKeyMap = make(map[string]*ecdsa.PrivateKey)

func StorePrivateKey(keyString string, privateKey *ecdsa.PrivateKey) {
	privateKeyMap[keyString] = privateKey
}

func GetPrivateKey(keyString string) (*ecdsa.PrivateKey, bool) {
	privateKey, ok := privateKeyMap[keyString]
	return privateKey, ok
}
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2) // Since 1 byte is 2 hex characters
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

type Message struct {
	Command  string
	HopCount int
	Req      []byte
	CameFrom string
}

func DecodeReq(data []byte, target interface{}) error {
	buff := bytes.NewBuffer(data)

	dec := gob.NewDecoder(buff)
	err := dec.Decode(target)
	if err != nil {
		log.Panic(err)
	}

	return nil
}

type Transaction struct {
	Tx string

	Signature []byte
}

func EncodeReq(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()

}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 4096))
	},
}

// EncodeMessage encodes a Message object into a byte slice.

func EncodeMessage(msg *Message) ([]byte, error) {
	// Get a buffer from the pool

	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
	}()

	// Encode the Message into the buffer
	enc := gob.NewEncoder(buf)
	err := enc.Encode(msg)
	if err != nil {
		return nil, err
	}

	// Return the byte slice
	return buf.Bytes(), nil

}

// DecodeMessage decodes a byte slice into a Message object.

func DecodeMessage(data []byte) (*Message, error) {
	// Get a buffer from the pool
	buf := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
	}()

	// Write data to the buffer
	_, err := buf.Write(data)
	if err != nil {
		return nil, err
	}

	// Decode the buffer into a Message
	msg := &Message{}
	dec := gob.NewDecoder(buf)
	err = dec.Decode(msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
