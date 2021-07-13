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
  "github.com/LidoKing/learnBlockchain/network"
)

type CommandLine struct {
}

// Print all possible actions an instructions
func (cli *CommandLine) printUsage() {
  fmt.Println()
  fmt.Println("Usage:")
  // Get balance of ADDRESS
  fmt.Println(" 1. balance -a ADDRESSS")
  // Creates a blockchain and rewards the mining fee
  fmt.Println(" 2. createchain -a ADDRESS")
  // Send coins from one address to another, -mine allows sender to mine own block
  fmt.Println(" 3. send -f FROM -t TO -amount AMOUNT -mine")
  // Prints the blocks in the chain
  fmt.Println(" 4. print")
  // Creates new wallets
  fmt.Println(" 5. createwallet -n NUMBER OF WALLETS")
  // Lists all existing addresses
  fmt.Println(" 6. listaddresses")
  // Rebuilds the UTXO set
  fmt.Println(" 7. reindexutxo")
  // Start node with ID specified in NODE_ID env. var., -miner indicates that the node is a miner node
  fmt.Println(" 8. startnode -miner ADDRESS")
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

func (cli *CommandLine) createBlockChain(address, nodeID string) {
  if !wallet.ValidateAddress(address) {
    log.Panic("Address is not valid.")
  }

  newChain := blockchain.InitBlockChain(address, nodeID)
  defer newChain.Database.Close()

  UTXOSet := blockchain.UTXOSet{newChain}
  UTXOSet.Reindex()

  fmt.Println("Finished creating chain")
}

func (cli *CommandLine) getBalance(address, nodeID string) {
  if !wallet.ValidateAddress(address) {
    log.Panic("Address is not valid.")
  }

  chain := blockchain.ContinueBlockChain(nodeID)
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

func (cli *CommandLine) send(from, to string, amount int, nodeID string, mineNow bool) {
  if !wallet.ValidateAddress(from) {
    log.Panic("Address is not valid.")
  }

  if !wallet.ValidateAddress(to) {
    log.Panic("Address is not valid.")
  }

  chain := blockchain.ContinueBlockChain(nodeID)
  UTXOSet := blockchain.UTXOSet{chain}
  defer chain.Database.Close()

  // Retrieve wallets 'managed' by the node
  wallets, err := wallet.CreateWallets(nodeID)
  if err != nil {
    log.Panic(err)
  }
  fromAddress := wallets.GetWallet(from)

  tx := blockchain.NewTransaction(&fromAddress, to, amount, &UTXOSet)

  if mineNow {
    // Tx for rewarding miner
    cbtx := blockchain.CoinbaseTx(from, "")
    txs := []*blockchain.Transaction{cbtx, tx}
    block := chain.MineBlock(txs)
    UTXOSet.Update(block)
  } else {
    network.SendTx(network.KnownNodes[0], tx)
    fmt.Println("Tx sent")
  }

  fmt.Println()
  fmt.Println("Success. Details:")
  fmt.Printf("  From: %s\n", from)
  fmt.Printf("  To: %s\n", to)
  fmt.Printf("  Amount: %d\n", amount)
  fmt.Println()
}

func (cli *CommandLine) printChain(nodeID string) {
  chain := blockchain.ContinueBlockChain(nodeID)
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

func (cli *CommandLine) listAddresses(nodeID string) {
  wallets, _ := wallet.CreateWallets(nodeID)
  addresses := wallets.GetAllAddresses()

  fmt.Println()
  for index, address := range addresses {
    _index := index + 1
    fmt.Printf("%d: %s\n", _index, address)
  }
  fmt.Println()
}

func (cli *CommandLine) createWallet(nodeID string, num int) {
  wallets, _ := wallet.CreateWallets(nodeID)

  for i := 0; i < num; i++ {
    address := wallets.AddWallet()
    wallets.SaveFile(nodeID)

    fmt.Println()
    fmt.Printf("New address created: %s\n", address)
    fmt.Println()
  }
}

func (cli *CommandLine) reindexUTXO(nodeID string) {
  chain := blockchain.ContinueBlockChain(nodeID)
  defer chain.Database.Close()
  UTXOSet := blockchain.UTXOSet{chain}
  UTXOSet.Reindex()

  count := UTXOSet.CountTransactions()

  fmt.Println()
  fmt.Printf("Done! There are %d transactions in the UTXO set. \n", count)
  fmt.Println()
}

func (cli *CommandLine) startNode(nodeID, minerAddress string) {
  fmt.Printf("Starting Node %s\n", nodeID)

  if len(minerAddress) != 0 {
    if wallet.ValidateAddress(minerAddress) {
      fmt.Printf("Mining is on. Address to receive rewards: %s\n", minerAddress)
    } else {
      log.Panic("Wrong miner address!")
    }
  }

  network.StartServer(nodeID, minerAddress)
}

func (cli *CommandLine) Run() {
  cli.validateArgs()

  nodeID := os.Getenv("NODE_ID")
  if nodeID == "" {
    fmt.Println("NODE_ID env is not set")
    runtime.Goexit()
  }

  getBalanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
  createBlockchainCmd := flag.NewFlagSet("createchain", flag.ExitOnError)
  sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
  printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
  createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
  listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
  reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
  startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

  // String() params: name, value, usage
  getBalanceAddress := getBalanceCmd.String("a", "", "The address to get balance for")
  createBlockchainAddress := createBlockchainCmd.String("a", "", "The address to send genesis block reward to")
  sendFrom := sendCmd.String("f", "", "Sender wallet address")
  sendTo := sendCmd.String("t", "", "Receiver wallet address")
  sendAmount := sendCmd.Int("amount", 0, "Amount to send")
  sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
  numOfWallets := createWalletCmd.Int("n", 1, "Number of wallets to be created")
  startNodeMiner := startNodeCmd.String("miner", "", "Enable mining node and send reward to ADDRESS")

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

  case "startnode":
    err := startNodeCmd.Parse(os.Args[2:])
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
    cli.getBalance(*getBalanceAddress, nodeID)
  }

  if createBlockchainCmd.Parsed() {
    if *createBlockchainAddress == "" {
      createBlockchainCmd.Usage()
      runtime.Goexit()
    }
  cli.createBlockChain(*createBlockchainAddress, nodeID)
  }

  if printChainCmd.Parsed() {
    cli.printChain(nodeID)
  }

  if sendCmd.Parsed() {
    if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
      sendCmd.Usage()
      runtime.Goexit()
    }
    cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
  }

  if createWalletCmd.Parsed() {
    cli.createWallet(nodeID, *numOfWallets)
  }

  if listAddressesCmd.Parsed() {
    cli.listAddresses(nodeID)
  }

  if reindexUTXOCmd.Parsed() {
    cli.reindexUTXO(nodeID)
  }

  if startNodeCmd.Parsed() {
    nodeID := os.Getenv("NODE_ID")
    if nodeID == "" {
      fmt.Println("NODE_ID env is not set")
      runtime.Goexit()
    }
    cli.startNode(nodeID, *startNodeMiner)
  }
}
