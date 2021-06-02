package main

import (
  "fmt"
  "strconv"
  "github.com/LidoKing/learnBlockchain/blockchain"
)

func main() {
  chain := blockchain.InitBlockChain()

  chain.AddBlock("first block after genesis")
  chain.AddBlock("second block after genesis")
  chain.AddBlock("third block after genesis")

  for _, block := range chain.Blocks {
    pow := blockchain.NewProofOfWork(block)
    fmt.Printf("Pow: %s\n", strconv.FormatBool(pow.Validate()))
    fmt.Println()

    fmt.Printf("Previous hash: %x\n", block.PrevHash)
    fmt.Printf("data: %s\n", block.Data)
    fmt.Printf("hash: %x\n", block.Hash)
    fmt.Printf("nonce: %d\n", block.Nonce)
  }
}
