package main

import (
  "fmt"
  "strconv"
  "runtime"
  "flag"
  "os"
  "github.com/LidoKing/learnBlockchain/blockchain"
)

type CommandLine struct {
  blockchain *blockchain.BlockChain
}

func (cli *CommandLine) printUsage() {
  fmt.Println("Usage:")
  fmt.Println(" add -block BLOCK_DATA -> Add a block to the chain")
  fmt.Println(" print -> Prints the blocks in the chain")
}

func (cli *CommandLine) validateArgs() {
  // Check command line arguments in the form of array (e.g. >> a b c d, len(os.Args) = 4)
  if len(os.Args) < 2 {
    cli.printUsage()

    // Exit application by shutting down GO routine
    runtime.Goexit()
  }
}

func (cli *CommandLine) addBlock(data string) {
  cli.blockchain.AddBlock(data)
  fmt.Println("Added Block!")
}

func (cli *CommandLine) printChain() {
  iter := cli.blockchain.Iterator()

  for {
    block := iter.Next()

    // Block info
    fmt.Printf("Previous hash: %x\n", block.PrevHash)
    fmt.Printf("data: %s\n", block.Data)
    fmt.Printf("hash: %x\n", block.Hash)
    fmt.Printf("nonce: %d\n", block.Nonce)

    // PoW validation
    pow := blockchain.NewProofOfWork(block)
    fmt.Printf("Pow: %s\n", strconv.FormatBool(pow.Validate()))
    fmt.Println()

    // Genesis block has no previous hash and therefore loop breaks
    if len(block.PrevHash) == 0 {
      break
    }
  }
}

func (cli *CommandLine) run() {
  cli.validateArgs()

  addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
  printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
  addBlockData := addBlockCmd.String("block", "", "Block data")

  switch os.Args[1] {
  case "add":
    err := addBlockCmd.Parse(os.Args[2:])
    blockchain.Handle(err)

  case "print":
    err := printChainCmd.Parse(os.Args[2:])
    blockchain.Handle(err)

  default:
    cli.printUsage()
    runtime.Goexit()
  }

  if addBlockCmd.Parsed() {
    if *addBlockData == "" {
      addBlockCmd.Usage()
      runtime.Goexit()
    }
    cli.addBlock(*addBlockData)
  }

  if printChainCmd.Parsed() {
    cli.printChain()
  }
}

func main() {
  defer os.Exit(0)
  chain := blockchain.InitBlockChain()
  defer chain.Database.Close()

  cli := CommandLine{chain}
  cli.run()
}
