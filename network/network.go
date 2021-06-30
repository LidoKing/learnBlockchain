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
  "gopkg.in/vrecan/death.v3"
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
  Tpe string
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
  data := Block {nodeAddress, b.Serialize()}
  payload := GobEncode(data)
  request := append(CmdToBytes("block"), payload...)

  SendData(address, request )
}

func SendInv(address, kind string, items [][]byte) {
  inventory := Inv{nodeAddress, kind, items}
  payload := GobEncode(inventory)
  request := append(CmdToBytes("inv"), payload...)

  SendData(address, request)
}

func SendTx(address string, tx *blockchain.Transaction) {
  data := Tx{nodeAddress, tx.Serialize()}
  payload := GobEncode(data)
  request := append(CmdToBytes("tx"), payload...)

  SendData(address, request)
}

func SendVersion(address string, chain *blockchain.BlockChain) {
  bestHeight := chain.GetBestHeight()
  version := Version{version, bestHeight, nodeAddress}
  payload := GobEncode(version)
  request := append(CmdToBytes("version"), payload...)

  SendData(address, request)
}
