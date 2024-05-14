package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

func NewBootStrapNode(address string) *BootStrapNode {
	return &BootStrapNode{
		address: address,
	}
}

type BootStrapNode struct {
	address  string
	PeerList []NodeInfo
}

func shuffle(slice []NodeInfo) {
	rand.Seed(time.Now().UnixNano())
	for i := len(slice) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}
func SubsetPeers(list []NodeInfo) []NodeInfo { // 2
	var Result []NodeInfo
	shuffle(list)
	for i := 0; i < len(list); i++ {
		if len(Result) == 2 {
			return Result
		}
		Result = append(Result, list[i])

	}
	return Result
}

func provideInfo(bt *BootStrapNode, conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		if err != io.EOF {
			log.Println("Error reading:", err)
		}
		return
	}
	client_address := string(buffer[:n])

	// Create a buffer to hold the encoded data
	var buf bytes.Buffer

	// Encode the peer list into the buffer

	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(SubsetPeers(bt.PeerList)); err != nil {
		log.Println("Error encoding:", err)
		return
	}

	// Write the encoded data to the client
	if _, err := conn.Write(buf.Bytes()); err != nil {
		log.Println("Error writing:", err)
		return
	}
	str := strings.Split(client_address, ":")
	bt.PeerList = append(bt.PeerList, NodeInfo{client_address, str[1]})
}
func BootStrapServer(bt *BootStrapNode) {
	ln, err := net.Listen("tcp", bt.address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("BootStrap Server Started On" + bt.address)
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go provideInfo(bt, conn)
	}
}

type PeerManager interface {
	GetPeerList() []NodeInfo
	RemovePeer(address string)
}

func (n *Node) GetPeerList() []NodeInfo {
	return n.PeerList
}

func (bt *BootStrapNode) GetPeerList() []NodeInfo {
	return bt.PeerList
}
func (bt *BootStrapNode) RemovePeer(address string) {
	for i := range bt.PeerList {
		if bt.PeerList[i].Address == address {
			bt.PeerList = append(bt.PeerList[:i], bt.PeerList[i+1:]...)
			break // Found and removed the peer, exit the loop
		}
	}
}
func (n *Node) RemovePeer(address string) {
	for i := range n.PeerList {
		if n.PeerList[i].Address == address {
			n.PeerList = append(n.PeerList[:i], n.PeerList[i+1:]...)
			break // Found and removed the peer, exit the loop
		}
	}
}

// ManageConnections checks the connection status of peers in a PeerManager and removes unreachable peers.
func ManageConnections(pm PeerManager) {
	for i := range pm.GetPeerList() {
		if !isPortOpen(pm.GetPeerList()[i].Address) {
			// Handle closed or unreachable port
			fmt.Printf("Port is closed or unreachable for peer %s\n", pm.GetPeerList()[i].Address)
			// Remove the peer from the peer list
			pm.RemovePeer(pm.GetPeerList()[i].Address)
		}
	}
}
func isPortOpen(address string) bool {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		// Check if error indicates closed or unreachable port
		if strings.Contains(err.Error(), "refused") || strings.Contains(err.Error(), "unreachable") {
			// Port is closed or unreachable
			return false
		}
		// Other error occurred
		fmt.Printf("Error while dialing %s: %s\n", address, err.Error())
		return false
	}
	defer conn.Close()
	// Port is open
	return true
}
func DipslayPeer(bt *BootStrapNode) {
	fmt.Println("***     Network     ***")
	for i := range bt.PeerList {
		fmt.Println(bt.PeerList[i].ID)

	}
}
