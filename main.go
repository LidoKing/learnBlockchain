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

// Print all possible actions an instructions
func (cli *CommandLine) printUsage() {
  fmt.Println()
  fmt.Println("Usage:")
  fmt.Println(" add -block BLOCK_DATA -> Add a block to the chain")
  fmt.Println(" print -> Prints the blocks in the chain")
}

// Ensure valid input is given
func (cli *CommandLine) validateArgs() {
  // Check command line arguments in the form of a string array
  // with program name included (e.g. >> main.go print, len(os.Args) = 2)
  if len(os.Args) < 2 {
    cli.printUsage()

    // Exit application by shutting down GO routine
    // to prevent data corruption
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
    fmt.Println()
    fmt.Printf("Previous hash: %x\n", block.PrevHash)
    fmt.Printf("data: %s\n", block.Data)
    fmt.Printf("hash: %x\n", block.Hash)
    fmt.Printf("nonce: %d\n", block.Nonce)

    // PoW validation
    pow := blockchain.NewProofOfWork(block)
    fmt.Printf("Pow: %s\n", strconv.FormatBool(pow.Validate()))
    fmt.Println()

    // Length of slice of byte is 0 = no data
    if len(block.PrevHash) == 0 {
      // Genesis block has no previous hash and therefore loop breaks
      break
    }
  }
}

func (cli *CommandLine) run() {
  cli.validateArgs()

  addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
  printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)

  // String() params: name, value, usage
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

  // Parsed() will return true if the object it was used on has been called
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
  // defer prevents corruption of bytes that are going into database
  // by making sure all actions are finished before closing database
  defer os.Exit(0)
  chain := blockchain.InitBlockChain()
  defer chain.Database.Close()

  cli := CommandLine{chain}
  cli.run()
}
