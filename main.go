package main

import (
	"crypto/ecdsa" // Import the crypto/rand package
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var transactionDone sync.WaitGroup

func main() {

	for {
		var input rune
		fmt.Print("Start BootStrapNode : 1 \nAdd Node to Network : 2 \nCreateAccount : 3 \nRegister Vehicle : 4 \nExit : 5 \nChangeOwnership-1\n")
		fmt.Scan(&input)
		switch input {
		case -1:
			fmt.Println("Exiting...")
			return
		case 1:
			startBootstrapServer("6000") // Default Port
		case 2:
			startNodeInteractive()
		case 3:
			CreateAccount()
		case 4:
			AddVehicleInteractive()
		case 5:
			ChangeOwnerShip()
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
func ChangeOwnerShip() {
	var OwnerKey string
	var NewOwnerKey string
	fmt.Print("Enter Owner Private Key: ")
	_, _ = fmt.Scan(&OwnerKey)
	fmt.Print("Enter NewOwner Public Key: ")
	_, _ = fmt.Scan(&NewOwnerKey)

	privateKey, flag := GetPrivateKey(OwnerKey)
	if !flag {
		fmt.Println("Error: Wrong Private Key")
		return
	}
	var NodeId string
	fmt.Println("Enter NodeID:")
	_, _ = fmt.Scan(&NodeId)
	var Vin string
	fmt.Println("Enter VIN:")
	_, _ = fmt.Scan(&Vin)

	publicKey, _ := EncodePublicKeyToBytes(&privateKey.PublicKey)

	tx := Transaction{Tx: "ChangeOwnership" + ":" + hex.EncodeToString(publicKey) + ":" + NewOwnerKey + ":" + Vin}
	fmt.Println("Transaction Change Ownership:", Vin)
	tx.Signature, _ = SignTransaction(tx, privateKey)
	fmt.Println("Transaction Signed Successfully")

	msg := Message{Command: "ChangeOwnership", HopCount: 0, Req: EncodeReq(tx)}

	SendTransaction("localhost:"+NodeId, msg)
}
func startNodeInteractive() {
	var port string
	fmt.Print("Enter Port: ")
	_, _ = fmt.Scan(&port)
	n := startNode(port)
	n.status = true

	// Set up signal handling for interrupt signals
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		// Wait for an interrupt signal
		<-interruptChan

		fmt.Println("Received interrupt signal. Turning off the node.")
		n.status = false // Turn off the node's activities
		os.Exit(0)       // Exit the program
	}()

	go func() {
		for {
			n.StartMinning()
			ManageConnections(n) // remove the inactive Ones
			n.DisplayChain()
			time.Sleep(2 * time.Second)
		}
	}()
}

func startNode(port string) *Node {
	node := NewNode("node_"+fmt.Sprint(port), "localhost:"+port)
	node.storage = NewStorage()
	node.CreateChain()
	go node.StartServer()
	node.JoinNetwork("localhost:6000")
	go node.HandShake()

	return node
}

func AddVehicleInteractive() {
	var privateKeyInput string
	fmt.Print("Enter Private Key: ")
	_, _ = fmt.Scan(&privateKeyInput)

	key, flag := GetPrivateKey(privateKeyInput)
	if !flag {
		fmt.Println("Error: Wrong Private Key")
		return
	}
	AddVehicle(key)
}

func AddVehicle(privateKey *ecdsa.PrivateKey) {
	var NodeId string
	fmt.Println("Enter NodeID:")
	_, _ = fmt.Scan(&NodeId)
	var Vin string
	fmt.Println("Enter VIN:")
	_, _ = fmt.Scan(&Vin)

	publicKey, _ := EncodePublicKeyToBytes(&privateKey.PublicKey)

	tx := Transaction{Tx: "RegisterVehicle" + ":" + hex.EncodeToString(publicKey) + ":" + Vin}
	fmt.Println("Transaction AddVehicle:", Vin)
	tx.Signature, _ = SignTransaction(tx, privateKey)
	fmt.Println("Transaction Signed Successfully")

	msg := Message{Command: "RegisterVehicle", HopCount: 0, Req: EncodeReq(tx)}

	SendTransaction("localhost:"+NodeId, msg)

}
func CreateAccount() {
	acc := MakeAccount()
	//	AddUser(acc)

	publicKeyString := EncodePublicKeyToBase64(acc.PublicKey)
	fmt.Println("PubKey:", publicKeyString)
	key, _ := generateRandomString(8)
	StorePrivateKey(key, acc.PrivateKey)

	fmt.Println("PrivKey:", key)
}

func startBootstrapServer(port string) {
	bt := NewBootStrapNode("localhost:" + port)
	go BootStrapServer(bt) // host the incoming ones

	go func() {
		for {
			ManageConnections(bt) // remove the inactive Ones
			DipslayPeer(bt)
			time.Sleep(2 * time.Second)
		}
	}()
}

/*
func (n *Node) SendData(address string, msg string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("Error connecting to server:", err)
	}
	defer conn.Close()
	// Send message to server
	_, err = conn.Write([]byte(msg))
	if err != nil {
		log.Println("Error writing to server:", err)
		return
	}
	// Read response from server
	buffer := make([]byte, 4096)
	len, err := conn.Read(buffer)
	if err != nil {
		log.Println("Error reading from server:", err)
		return
	}
	fmt.Printf(n.ID+" received: %s\n", buffer[:len])
}
*/
