package network

import (
  "fmt"
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
