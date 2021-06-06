package blockchain

import (
  "fmt"
  "github.com/dgraph-io/badger"
  "os"
)

const (
  dbPath = "./tmp/blocks"

  // Check if a blockchain exists or not
  dbFile = "./tmpm/blocks/MANIFEST"

  // Arbitrary data for filling up empty genesis
  genesisData = "First transaction from genesis"
)

type BlockChain struct{
  LastHash []byte
  Database *badger.DB
}

// Get chain from database for printing info
type BlockChainIterator struct {
  CurrentHash []byte
  Database    *badger.DB
}

/*-------------------------------utils-------------------------------*/

func DBexists() bool {
  if _, err := os.Stat(dbFile); os.IsNotExist(err) {
    return false
  }
  return true
}

/*-------------------------------main-------------------------------*/

func (chain *BlockChain) AddBlock(data string) {
  var lastHash []byte

  // Get lastHash from database
  err := chain.Database.View(func(txn *badger.Txn) error { // error 1
    item, err := txn.Get([]byte("lh")) // error 2
    Handle(err) // Handle error 2

    err = item.Value(func(val []byte) error { // error 1
      lastHash = val
      return nil
    })
    Handle(err) // Handle error 1

    return err // error 1
  })
  Handle(err) // Handle error 1

  // Create new block with hash retrieved from database
  newBlock := CreateBlock(data, lastHash)

  // Add new block to database and update lastHash
  err = chain.Database.Update(func(txn *badger.Txn) error { // error 3
    err := txn.Set(newBlock.Hash, newBlock.Serialize()) // error 4
    Handle(err) // Handle error 4

    err = txn.Set([]byte("lh"), newBlock.Hash) // error 3

    chain.LastHash = newBlock.Hash

    return err // error 3
  })
  Handle(err) // Handle error 3
}

// Only called for starting completely new chain
func InitBlockChain(address string) *BlockChain { // miner's wallet address
  var lastHash []byte

  if DBexists() {
    fmt.Println("Blockchain already exists, call 'ContinueBlockChain' instead.")
    runtime.Goexit()
  }

  // Open database
  db, err := badger.Open(badger.DefaultOptions(dbPath)) // error 1
  Handle(err) // Handle error 1

  err = db.Update(func(txn *badger.Txn) error { // error 2
    // Create coinbase transaction
    cbtx := CoinbaseTx(Address, genesisData)

    // Create genesis block with coinbase transaction
    genesis := Genesis(cbtx)
    fmt.Println("Genesis created.")

    // Add block to database
    err = txn.Set(genesis.Hash, genesis.Serialize()) // error 3
    Handle(err) // Handle error 3

    // Set genesis block hash as last hash
    err = txn.Set([]byte("lh"), genesis.Hash) // error 2
    lastHash = genesis.Hash

    return err // error 2
  })
  Handle(err) // error 2
}

// Only called for continuing with existing chain
func ContinueBlockChain(address string) *BlockChain { // miner's wallet address
  var lastHash []byte

  if DBexists(dbFile) == false {
    fmt.Println("No blockchain found, call 'InitBlockChain' to create one.")
    runtime.Gexit()
  }

  // Open database
  db, err := badger.Open(badger.DefaultOptions(dbPath)) // error 1
  Handle(err) // Handle error 1

  err = db.Update(func(txn *badger.Txn) error { // error 2
    item, err := txn.Get([]byte("lh")) // error 3
    Handle(err) // Handle error 3

    err = item.Value(func(val []byte) error { // error 4
      lastHash = val
      return nil
    })
    Handle(err) // Handle error 4

    return err // error 2
  })
  Handle(err) // Handle error 2

  chain := BlockChain{lastHash, db}
  return &chain
}

// Turn blockchain struct to iterator struct
func (chain *BlockChain) Iterator() *BlockChainIterator {
  iter := &BlockChainIterator{chain.LastHash, chain.Database}

  return iter
}

func (iter *BlockChainIterator) Next() *Block {
  var block *Block

  err := iter.Database.View(func(txn *badger.Txn) error { // error 1
    item, err := txn.Get(iter.CurrentHash) // error 2
    Handle(err) // Handle error 2

    // Get most recent block of chain with CurrentHash
    // and convert it from byte to block struct
    err = item.Value(func(val []byte) error { // error 1
      block = Deserialize(val)
      return nil
    })
    Handle(err) // Handle error 1

    return err // error 1
  })
  Handle(err) // Handle error 1

  // Set current hash of iterator to hash of previous block
  // so next iteration will get the previous block (i.e. second most recent)
  iter.CurrentHash = block.PrevHash

  return block
}
