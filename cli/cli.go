package cli

import (
  "fmt"
  "strconv"
  "runtime"
  "flag"
  "os"
  "log"
  "github.com/LidoKing/learnBlockchain/blockchain"
  "github.com/LidoKing/learnBlockchain/blockchain/wallet"
)

type CommandLine struct {
}

// Print all possible actions an instructions
func (cli *CommandLine) printUsage() {
  fmt.Println()
  fmt.Println("Usage:")
  fmt.Println(" balance -a ADDRESSS -> Get balance of ADDRESS")
  fmt.Println(" createchain -a ADDRESS -> Creates a blockchain and rewards the mining fee")
  fmt.Println(" send -f FROM -t TO -amount AMOUNT -> Send coins from one address to another")
  fmt.Println(" print -> Prints the blocks in the chain")
  fmt.Println(" createwallet -n NUMBER OF WALLETS -> Creates new wallets")
  fmt.Println(" listaddresses -> Lists all existing addresses")
  fmt.Println(" reindexutxo -> Rebuilds the UTXO set")
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

/*
FUNCTION ABOLISHED
func (cli *CommandLine) addBlock(data string) {
  cli.blockchain.AddBlock(data)
  fmt.Println("Added Block!")
}
*/

func (cli *CommandLine) createBlockChain(address string) {
  if !wallet.ValidateAddress(address) {
    log.Panic("Address is not valid.")
  }

  newChain := blockchain.InitBlockChain(address)
  defer newChain.Database.Close()

  UTXOSet := blockchain.UTXOSet{newChain}
  UTXOSet.Reindex()

  fmt.Println("Finished creating chain")
}

func (cli *CommandLine) getBalance(address string) {
  if !wallet.ValidateAddress(address) {
    log.Panic("Address is not valid.")
  }

  chain := blockchain.ContinueBlockChain(address)
  UTXOSet := blockchain.UTXOSet{chain}
  defer chain.Database.Close()

  balance := 0
  fullHash := wallet.Base58Decode([]byte(address))
  pubKeyHash := fullHash[1:len(fullHash) - 4]
  UTXOs := UTXOSet.FindUTXO(pubKeyHash)

  for _, out := range UTXOs {
    balance += out.Value
  }

  fmt.Println()
  fmt.Printf("Balance of %s: %d\n", address, balance)
  fmt.Println()
}

func (cli *CommandLine) send(from, to string, amount int) {
  if !wallet.ValidateAddress(from) {
    log.Panic("Address is not valid.")
  }

  if !wallet.ValidateAddress(to) {
    log.Panic("Address is not valid.")
  }

  chain := blockchain.ContinueBlockChain(from)
  UTXOSet := blockchain.UTXOSet{chain}
  defer chain.Database.Close()

  tx := blockchain.NewTransaction(from, to, amount, &UTXOSet)
  // Tx for rewarding miner, which is the sender
  coinbaseTx := blockchain.CoinbaseTx(from, "")

  block := chain.AddBlock([]*blockchain.Transaction{coinbaseTx, tx})
  UTXOSet.Update(block)

  fmt.Println()
  fmt.Println("Transaction complete. Details:")
  fmt.Printf("  From: %s\n", from)
  fmt.Printf("  To: %s\n", to)
  fmt.Printf("  Amount: %d\n", amount)
  fmt.Println()
}

func (cli *CommandLine) printChain() {
  chain := blockchain.ContinueBlockChain("")
  defer chain.Database.Close()
  iter := chain.Iterator()

  for {
    block := iter.Next()

    // Block info
    fmt.Println()
    fmt.Printf("Previous hash: %x\n", block.PrevHash)
    fmt.Printf("hash: %x\n", block.Hash)
    fmt.Printf("nonce: %d\n", block.Nonce)

    // PoW validation
    pow := blockchain.NewProofOfWork(block)
    fmt.Printf("Pow: %s\n", strconv.FormatBool(pow.Validate()))

    // Transactoins
    for _, tx := range block.Transactions {
      fmt.Println(tx)
      fmt.Println()
    }

    // Length of slice of byte is 0 = no data
    if len(block.PrevHash) == 0 {
      // Genesis block has no previous hash and therefore loop breaks
      break
    }
  }
}

func (cli *CommandLine) listAddresses() {
  wallets, _ := wallet.CreateWallets()
  addresses := wallets.GetAllAddresses()

  fmt.Println()
  for index, address := range addresses {
    fmt.Printf("%d: %s\n", index, address)
  }
  fmt.Println()
}

func (cli *CommandLine) createWallet(num int) {
  wallets, _ := wallet.CreateWallets()

  for i := 0; i < num; i++ {
    address := wallets.AddWallet()
    wallets.SaveFile()

    fmt.Println()
    fmt.Printf("New address created: %s\n", address)
    fmt.Println()
  }
}

func (cli *CommandLine) reindexUTXO() {
  chain := blockchain.ContinueBlockChain("")
  defer chain.Database.Close()
  UTXOSet := blockchain.UTXOSet{chain}
  UTXOSet.Reindex()

  count := UTXOSet.CountTransactions()

  fmt.Println()
  fmt.Printf("Done! There are %d transactions in the UTXO set. \n", count)
  fmt.Println()
}

func (cli *CommandLine) Run() {
  cli.validateArgs()

  getBalanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
  createBlockchainCmd := flag.NewFlagSet("createchain", flag.ExitOnError)
  sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
  printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
  createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
  listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
  reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)

  // String() params: name, value, usage
  getBalanceAddress := getBalanceCmd.String("a", "", "The address to get balance for")
  createBlockchainAddress := createBlockchainCmd.String("a", "", "The address to send genesis block reward to")
  sendFrom := sendCmd.String("f", "", "Sender wallet address")
  sendTo := sendCmd.String("t", "", "Receiver wallet address")
  sendAmount := sendCmd.Int("amount", 0, "Amount to send")
  numOfWallets := createWalletCmd.Int("n", 1, "Number of wallets to be created")

  // Parse arguments for checking afterwards
  switch os.Args[1] {
  case "balance":
    err := getBalanceCmd.Parse(os.Args[2:])
    blockchain.Handle(err)

  case "createchain":
    err := createBlockchainCmd.Parse(os.Args[2:])
    blockchain.Handle(err)

  case "print":
    err := printChainCmd.Parse(os.Args[2:])
    blockchain.Handle(err)

  case "send":
    err := sendCmd.Parse(os.Args[2:])
    blockchain.Handle(err)

  case "createwallet":
    err := createWalletCmd.Parse(os.Args[2:])
    blockchain.Handle(err)

  case "listaddresses":
    err := listAddressesCmd.Parse(os.Args[2:])
    blockchain.Handle(err)

  case "reindexutxo":
    err := reindexUTXOCmd.Parse(os.Args[2:])
    blockchain.Handle(err)

  default:
    cli.printUsage()
    runtime.Goexit()
  }

  // Parsed() will return true if the object it was used on has been called
  if getBalanceCmd.Parsed() {
    if *getBalanceAddress == "" {
      getBalanceCmd.Usage()
      runtime.Goexit()
    }
    cli.getBalance(*getBalanceAddress)
  }

  if createBlockchainCmd.Parsed() {
    if *createBlockchainAddress == "" {
      createBlockchainCmd.Usage()
      runtime.Goexit()
    }
  cli.createBlockChain(*createBlockchainAddress)
  }

  if printChainCmd.Parsed() {
    cli.printChain()
  }

  if sendCmd.Parsed() {
    if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
      sendCmd.Usage()
      runtime.Goexit()
    }
    cli.send(*sendFrom, *sendTo, *sendAmount)
  }

  if createWalletCmd.Parsed() {
    cli.createWallet(*numOfWallets)
  }

  if listAddressesCmd.Parsed() {
    cli.listAddresses()
  }

  if reindexUTXOCmd.Parsed() {
    cli.reindexUTXO()
  }
}
