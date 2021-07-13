package network

import (
  "bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"syscall"
	"runtime"
	"os"
  "github.com/vrecan/death/v3"
  "github.com/LidoKing/learnBlockchain/blockchain"
)

const (
  protocol = "tcp"
  version = 1
  commandLength = 12
)

var (
  nodeAddress string
  minerAddress string
  KnownNodes = []string{"localhost:3000"}
  // Blocks being sent from one client to the next
  blocksInTransit = [][]byte{}
  // Store txs until there are enough for mining
  memPool = make(map[string]blockchain.Transaction)
)

// Following structures help identify type of data to be sent back and forth

type Addr struct {
  AddrList []string
}

type Block struct {
  AddrFrom string
  Block []byte
}

type GetBlocks struct {
  AddrFrom string
}

type GetData struct {
  AddrFrom string
  Type string
  ID []byte
}

type Inv struct {
  AddrFrom string
  Type string
  Items [][]byte
}

type Tx struct {
  AddrFrom string
  Transaction []byte
}

// Identify when a chain needs to be updated
type Version struct {
  Version int
  BestHeight int
  AddrFrom string
}

func CmdToBytes(cmd string) []byte {
  var bytes [commandLength]byte

  for i, c := range cmd {
    // Convert each charater into byte respectively
    bytes[i] = bytes(c)
  }

  return bytes[:]
}

func BytesToCmd(bytes []byte) string {
  var cmd []byte

  for _, b := range bytes {
    if b != 0x0 {
      cmd = append(cmd, b)
    }
  }

  return fmt.Sprintf("%s", cmd)
}

func GobEncode(data interface{}) []byte {
  var buff bytes.Buffer

  enc := gob.NewEncoder(&buff)
  err := enc.Encode(data)
  if err != nil {
    log.Panic(err)
  }

  return buff.Bytes()
}

func CloseDB(chain *blockchain.Blockchain) {
  d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

  d.WaitForDeathWithFunc(func() {
    defer os.Exit(1)
    defer runtime.Goexit()
    chain.Database.Close()
  })
}

func HandleConnection(conn net.Conn, chain *blockchain.BlockChain) {
  req, err := ioutil.ReadAll(conn)
  defer conn.Close()

  if err != nil {
    log.Panic(err)
  }
  command := BytesToCmd(req[:commandLength])
  fmt.Printf("Received %s command\n", command)

  switch command {
  case "addr":
		HandleAddr(req)
	case "block":
		HandleBlock(req, chain)
	case "inv":
		HandleInv(req, chain)
	case "getblocks":
		HandleGetBlocks(req, chain)
	case "getdata":
		HandleGetData(req, chain)
	case "tx":
		HandleTx(req, chain)
	case "version":
		HandleVersion(req, chain)
	default:
		fmt.Println("Unknown command")
  }
}

// Send data from one node to another
func SendData(addr string, data []byte) {
  conn, err := net.Dial(protocol, addr)

  if err != nil {
    fmt.Printf("%s is not available\n", addr)

    var updatedNodes []string

    // Remove unavailable node from KnownNodes array
    for _, node := range KnownNodes {
      if node != addr {
        updatedNodes = append(updatedNodes, node)
      }
    }

    KnownNodes = updatedNodes

    return
  }

  defer conn.Close()

  // Copy() params: destination, source
  _, err = io.Copy(conn, bytes.NewReader(data))

  if err != nil {
    log.Panic(err)
  }
}

// Send funcs pattern:
// 1. Create struct
// 2. Encode struct to bytes as payload
// 3. Concatenate command and payload to form request
// 4. Call SendData()

func SendAddr(address string) {
  nodes := Addr{KnownNodes}
  nodes.AddrList = append(nodes.AddrList, nodeAddress)
  payload := GobEncode(nodes)
  request := append(CmdToBytes("addr"), payload...)

  SendData(address, request)
}

func SendBlock(address string, b *blockchain.Block) {
  payload := GobEncode(Block{nodeAddress, b.Serialize()})
  request := append(CmdToBytes("block"), payload...)

  SendData(address, request )
}

func SendInv(address, kind string, items [][]byte) {
  payload := GobEncode(Inv{nodeAddress, kind, items})
  request := append(CmdToBytes("inv"), payload...)

  SendData(address, request)
}

func SendTx(address string, tx *blockchain.Transaction) {
  payload := GobEncode(Tx{nodeAddress, tx.Serialize()})
  request := append(CmdToBytes("tx"), payload...)

  SendData(address, request)
}

func SendVersion(address string, chain *blockchain.BlockChain) {
  bestHeight := chain.GetBestHeight()
  payload := GobEncode(Version{version, bestHeight, nodeAddress})
  request := append(CmdToBytes("version"), payload...)

  SendData(address, request)
}

func SendGetBlocks(address string) {
  payload := GobEncode(GetBlocks{nodeAddress})
  request := append(CmdToBytes("getblocks"), payload...)

  SendData(address, request)
}

func SendGetData(address, kind string, id []byte) {
  payload := GobEncode(GetData{nodeAddress, kind, id})
  request := append(CmdToBytes("getdata"), payload...)

  SendData(address, request)
}

// Make sure all blockchains are synced
func RequestBlocks() {
  for _, node := range KnownNodes {
    SendGetBlocks(node)
  }
}

func HandleAddr(request []byte) {
  var buff bytes.Buffer
  var payload Addr

  buff.Write(request[commandLength:]) // Extract payload
  dec := gob.NewDecoder(&buff) // Read from buff
  if err := dec.Decode(&payload); err != nil { // Write decoded buff to payload
    log.Panic(err)
  }

  KnownNodes = append(KnownNodes, payload.AddrList...)
  fmt.Printf("There are %d known nodes\n", len(KnownNodes))
  RequestBlocks()
}

func HandleBlock(request []byte, chain *blockchain.BlockChain) {
  var buff bytes.Buffer
  var payload Block

  buff.Write(request[commandLength:])
  dec := gob.NewDecoder(&buff)
  if err := dec.Decode(&payload); err != nil {
    log.Panic(err)
  }

  blockData := payload.Block
  block := blockchain.Deserialize(blokData)

  fmt.Println("Received a block!")
  chain.AddBlock(block)

  fmt.Printf("Added block %x\n", block.Hash)

  if len(blocksInTransit) > 0 {
    blockHash := blocksInTransit[0]
    SendGetData(payload.AddrFrom, "block", blockHash)

    blocksInTransit = blocksInTransit[1:]
  } else {
    UTXOSet := blockchain.UTXOSet{chain}
    UTXOSet.Reindex()
  }
}

func HandleGetBlocks(request []byte, chain *blockchain.BlockChain) {
  var buff bytes.Buffer
  var payload GetBlocks

  buff.Write(request[commandLength:])
  dec := gob.NewDecoder(&buff)
  if err := dec.Decode(&payload); err != nil {
    log.Panic(err)
  }

  blocks := chain.GetBlockHashes()
  SendInv(payload.AddrFrom, "block", blocks)
}

func HandleGetData(request []byte, chain *blockchain.BlockChain) {
  var buff bytes.Buffer
  var payload GetData

  buff.Write(request[commandLength:])
  dec := gob.NewDecoder(&buff)
  if err := dec.Decode(&payload); err != nil {
    log.Panic(err)
  }

  if payload.Type == "block" {
    block, err := chain.GetBlock([]byte(payload.ID))
    if err != nil {
      return
    }
    // Send block to peers for downloading
    SendBlock(payload.AddrFrom, &block)

  }

  if payload.Type == "tx" {
    txID := hex.EncodeToString(payload.ID)
    tx := memPool[txID]

    SendTx(payload.AddrFrom, &tx)
  }
}

func NodeIsKnown(addr string) bool {
  for _, node := range KnownNodes {
    if node == addr {
      return true
    } else {
      return false
    }
  }
}

func HandleVersion(request []byte, chain *blockchain.BlockChain) {
  var buff bytes.Buffer
  var payload Version

  buff.Write(request[commandLength:])
  dec := gob.NewDecoder(&buff)
  if err := dec.Decode(&payload); err != nil {
    log.Panic(err)
  }

  // Height of local chain
  bestHeight := chain.GetBestHeight()
  // Height of a peer's chain
  othersHeight := payload.BestHeight

  if bestHeight > othersHeight {
    SendVersion(payload.AddrFrom, chain)
  } else if othersHeight > bestHeight {
    SendGetBlocks(payload.AddrFrom)
  }

  if !NodeIsKnown() {
    KnownNodes = append(KnownNodes, payload.AddrFrom)
  }
}

func HandleTx(request []byte, chain *blockchain.BlockChain) {
  var buff bytes.Buffer
  var payload Tx

  buff.Write(request[commandLength:])
  dec := gob.NewDecoder(&buff)
  if err := dec.Decode(&payload); err != nil {
    log.Panic(err)
  }

  tx := blockchain.Deserialize(payload.Transaction)
  memPool[hex.EncodeToString(tx.ID)] = tx

  if nodeAddress == KnownNodes[0] { // central node
    for _, node := range KnownNodes {
      if node != nodeAddress && node != payload.AddrFrom {
        SendInv(node, "tx", [][]byte{tx.ID})
      }
    }
  } else { // miner node
    if len(memPool) >= 2 && len(minerAddress) > 0 {
      MineTx(chain)
    }
  }
}

func MineTx(chain *blockchain.BlockChain) {
  var txs []*blockchain.Transaction

  /* for id := range memPool {
    fmt.Printf("tx: %s\n", memPool[id].ID)
    tx := memPool[id]
    if chain.VerifyTransaction(&tx) {
      txs = append(txs, &tx)
    }
  } */

  for _, tx := range memPool {
    fmt.Printf("tx: %s\n", tx.ID)
    if chain.VerifyTransaction(tx) {
      txs = append(txs, tx)
    }
  }


  if len(txs) == 0 {
    fmt.Println("All txs are invalid")
    return
  }

  cbtx := blockchain.CoinbaseTx(minerAddress, "")
  txs = append(txs, cbtx)

  newBlock := chain.MineBlock(txs)
  UTXOSet := blockchain.UTXOSet{chain}
  UTXOSet.Reindex()

  fmt.Println("New block mined")

  for _, tx := range txs {
    txID := hex.EncodeToString(tx.ID)
    delete(memPool, txID)
  }

  for _, node := range KnownNodes {
    if node != nodeAddress {
      SendInv(node, "block", [][]byte{newBlock.Hash})
    }
  }

  if len(memPool) > 0 {
    MineTx(chain)
  }
}

func HandleInv(request []byte, chain *blockchain.BlockChain) {
  var buff bytes.Buffer
  var payload Inv

  buff.Write(request[commandLength:])
  dec := gob.NewDecoder(&buff)
  if err := dec.Decode(&payload); err != nil {
    log.Panic(err)
  }

  fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

  if payload.Type == "block" {
    blocksInTransit = payload.Items
    blockHash := payload.Items[0]
    SendGetData(payload.AddrFrom, "block", blockHash)

    newInTransit := [][]byte{}
    for _, b := range blocksInTransit {
      if bytes.Compare(b, blockHash) != 0 {
        newInTransit = append(newInTransit, b)
      }
    }
    blocksInTransit = newInTransit
  }

  if payload.Type == "tx" {
    txID := payload.Items[0]

    if memPool[hex.EncodeToString(txID)].ID == nil {
      SendGetData(payload.AddrFrom, "tx", txID)
    }
  }
}

func StartServer(nodeID, minerAddress string) {
  nodeAddress = fmt.Sprintf("localhost: %s", nodeID)
  minerAddress = minerAddress

  ln, err := net.Listen(protocol, nodeAddress)
  if err != nil {
    log.Panic(err)
  }

  defer ln.Close()

  chain := blockchain.ContinueBlockChain(nodeID)
  defer chain.Databae.Close()
  go CloseDB(chain)

  if nodeAddress != KnownNodes[0] {
    SendVersion(KnownNodes[0], chain)
  }

  for {
    conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go HandleConnection(conn, chain)
  }
}
