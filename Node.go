package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

func (n *Node) VerifyAndUpdateLedger(NewTx string) bool {
	// Verify the transaction
	// New Transaction Format "TransactionType:Ipfs_Cid"
	lastHash := n.blockchain.latestHash
	parts := strings.Split(NewTx, ":")
	TransactionType := parts[0]

	if TransactionType == "RegisterVehicle" {
		NewTx := RegisterVehicle{parts[1], parts[2]}
		NewCid, _ := n.storage.StoreData(&NewTx)

		for lastHash != "" {
			block := n.blockchain.Blocks[lastHash]

			for _, tx := range block.Transactions {
				CurrentCid := strings.Split(tx, ":")[1]
				if strings.Split(tx, ":")[0] == "RegisterVehicle" {
					var vehicle RegisterVehicle
					_ = n.storage.RetrieveData(CurrentCid, &vehicle)
					if CurrentCid == NewCid || (vehicle.VIN == NewTx.VIN && vehicle.Owner != NewTx.Owner) {
						return false
					}
				} else if strings.Split(tx, ":")[0] == "ChangeOwnership" {
					var ownership ChangeOwnership
					_ = n.storage.RetrieveData(CurrentCid, &ownership)
					if ownership.VIN == NewTx.VIN && ownership.To == NewTx.Owner {
						return false
					} else if CurrentCid == NewCid {
						return false
					}
				}
			}

			lastHash = hex.EncodeToString(block.PreviousHash)
		}
	} else if TransactionType == "ChangeOwnership" {
		NewTx := ChangeOwnership{parts[1], parts[2], parts[3]}
		NewCid, _ := n.storage.StoreData(&NewTx)

		for lastHash != "" {
			block := n.blockchain.Blocks[lastHash]

			for _, tx := range block.Transactions {
				CurrentCid := strings.Split(tx, ":")[1]
				if strings.Split(tx, ":")[0] == "RegisterVehicle" {
					var vehicle RegisterVehicle
					_ = n.storage.RetrieveData(CurrentCid, &vehicle)
					if vehicle.VIN == NewTx.VIN && vehicle.Owner != NewTx.From {
						return false
					}
				} else if strings.Split(tx, ":")[0] == "ChangeOwnership" {
					var ownership ChangeOwnership
					_ = n.storage.RetrieveData(CurrentCid, &ownership)
					if ownership.VIN == NewTx.VIN && ownership.To == NewTx.From {
						return false
					} else if CurrentCid == NewCid {
						return false
					}
				}
			}

			lastHash = hex.EncodeToString(block.PreviousHash)
		}
	}

	return true
}

func NewNode(ID string, listenAddr string) *Node {
	return &Node{

		listenAddr: listenAddr,
		status:     true,
	}
}

type Node struct {
	blockchain *Blockchain
	storage    *Storage
	listenAddr string
	ln         net.Listener

	ID           string
	status       bool
	transactions []string // keep track of all recieved transaction (like  a local MemPool)
	PeerList     []NodeInfo
	// Ledger       []string
}
type NodeInfo struct {
	Address string
	ID      string
}

func (n *Node) StartServer() {
	ln, err := net.Listen("tcp", n.listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	n.ln = ln
	fmt.Println(n.ID, "listening for connections on", n.listenAddr)
	for {
		conn, err := n.ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go n.handleConnection(conn)

	}
}
func (n *Node) ForwardTransaction(wg *sync.WaitGroup, msg *Message) {

	for _, peer := range n.PeerList {
		if peer.Address != msg.CameFrom {
			wg.Add(1)
			go func(peerAddr string) {
				defer wg.Done() // Decrement the wait group when the goroutine completes

				msg.CameFrom = n.listenAddr
				msg.HopCount++
				time.Sleep(1 * time.Second)
				conn, err := net.Dial("tcp", peerAddr)
				if err != nil {
					log.Println("Error connecting to server:", err)
					return
				}
				defer conn.Close()
				data, err := EncodeMessage(msg)
				if err != nil {
					log.Println("Error encoding transaction:", err)
					return
				}

				// Send the encoded transaction to the server
				_, err = conn.Write(data)
				if err != nil {
					log.Println("Error sending transaction to:", peerAddr, err)
					return
				}
				fmt.Println("Transaction  " + " sent to " + peerAddr)
			}(peer.Address)
		}
	}
}
func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read data from the client
	buffer := make([]byte, 4096)
	_, err := conn.Read(buffer)
	if err != nil {
		if err != io.EOF {
			log.Println("Error reading:", err)
		}
		return
	}

	// Decode the received data into a Message object
	msg, err := DecodeMessage(buffer)
	if err != nil {
		log.Println("Error decoding Message:", err)
		return
	}

	switch msg.Command {
	case "New Node":
		var peerAddr string
		DecodeReq(msg.Req, &peerAddr)
		// New node wants to establish connection
		fmt.Println("Recieved HandShake Request From" + string(peerAddr))
		if len(n.PeerList) == 2 {
			n.PeerList = n.PeerList[1:]
		}
		n.PeerList = append(n.PeerList, NodeInfo{Address: string(peerAddr), ID: ""})
	case "RegisterVehicle":

		var req Transaction
		err := DecodeReq(msg.Req, &req)
		if err != nil {
			fmt.Println("Error Decoding Message")
			return

		}
		if msg.HopCount > 5 { // ***change the HopAccount w.r.t NetworkSize***
			return
		}
		if n.UniqueTransaction(req.Tx) {
			if VerifyTransaction(&req) && n.VerifyAndUpdateLedger(req.Tx) {
				// sotre the data into
				parts := strings.Split(req.Tx, ":")
				NewTx := RegisterVehicle{parts[1], parts[2]}
				cid, _ := n.storage.StoreData(NewTx)

				n.transactions = append(n.transactions, "RegisterVehicle:"+cid)

			} else {
				fmt.Println("Invalid Transaction")
				return
			}

		}

		var wg sync.WaitGroup

		n.ForwardTransaction(&wg, msg)

		fmt.Print(n.transactions)

		// Wait for all transactions to be propagated to peers
		wg.Wait()
	case "ChangeOwnership":

		var req Transaction
		err := DecodeReq(msg.Req, &req)
		if err != nil {
			fmt.Println("Error Decoding Message")
			return

		}
		if msg.HopCount > 5 {
			return
		}

		if n.UniqueTransaction(req.Tx) {
			// check if the owner mentioned is correct
			if VerifyTransaction(&req) && n.VerifyAndUpdateLedger(req.Tx) {
				// sotre the data into
				NewOwnerPubKey := strings.Split(req.Tx, ":")[2]
				Vin := strings.Split(req.Tx, ":")[3]                                         // split the add vehicle transaction
				NewTx := ChangeOwnership{strings.Split(req.Tx, ":")[1], NewOwnerPubKey, Vin} // contruct
				cid, _ := n.storage.StoreData(&NewTx)
				n.transactions = append(n.transactions, "ChangeOwnership"+":"+cid)
				//n.Ledger = append(n.Ledger, req.Tx)
			} else {
				fmt.Println("Invalid Transaction")
				return
			}

		}

		var wg sync.WaitGroup

		n.ForwardTransaction(&wg, msg)

		fmt.Print(n.transactions)

		// Wait for all transactions to be propagated to peers
		wg.Wait()

	default:
		// Echo back the current time to the client
		_, err := conn.Write([]byte(time.Now().Format("2006-01-02 15:04:05")))
		if err != nil {
			log.Println("Error writing:", err)
			return
		}

	}
}
func (n *Node) UniqueTransaction(tx string) bool {
	for i := 0; i < len(n.transactions); i++ {
		if n.transactions[i] == tx {
			return false
		}
	}
	return true
}
func (n *Node) JoinNetwork(BootStrapAddress string) {
	// Connect to the bootstrap node
	conn, err := net.Dial("tcp", BootStrapAddress)
	if err != nil {
		log.Println("Error connecting:", err)
		return
	}
	defer conn.Close()
	// Read the response from the bootstrap node into a bytes.Buffer
	fmt.Println(n.ID + " Connected to BootStrapNode")
	_, err = conn.Write([]byte(n.listenAddr))
	if err != nil {
		log.Println("Error writing to BootStrapServer:", err)
		return
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, conn); err != nil {
		log.Println("Error reading from BootStrapServer:", err)
		return
	}

	// Decode the received data into the peer list
	var peerList []NodeInfo
	dec := gob.NewDecoder(&buf)
	if err := dec.Decode(&peerList); err != nil {
		if err != io.EOF {
			log.Println("Error decoding:", err)
		} else {
			log.Println("Decoder reached EOF before decoding complete value")
		}
		return
	}
	n.PeerList = peerList

}

func (n *Node) StartMinning() {
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
func (n *Node) HandShake() {
	for i := 0; i < len(n.PeerList); i++ {

		conn, err := net.Dial("tcp", n.PeerList[i].Address)
		if err != nil {
			log.Println("Error connecting to server:", err)
			continue
		}

		msg := &Message{Command: "New Node", Req: EncodeReq(n.listenAddr)}
		rq, err := EncodeMessage(msg)
		if err != nil {
			log.Println("Error Encoding:", err)
			continue
		}

		// Send the message

		_, err = conn.Write(rq)
		if err != nil {
			log.Println("Error writing to server:", err)
			continue
		}

		conn.Close()

	}
}
