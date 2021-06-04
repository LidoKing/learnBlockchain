package blockchain

import (
  "fmt"
  "github.com/dgraph-io/badger"
)

const dbPath = "./tmp/blocks"

type BlockChain struct{
  LastHash []byte
  Database *badger.DB
}

// Get chain from database for printing info
type BlockChainIterator struct {
  CurrentHash []byte
  Database    *badger.DB
}

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

func InitBlockChain() *BlockChain {
  var lastHash []byte

  // Open database, create if one doesn't exist
  db, err := badger.Open(badger.DefaultOptions(dbPath)) // error 1
  Handle(err) // Handle error 1

  err = db.Update(func(txn *badger.Txn) error { //error 4

    // Create chain with Genesis()
    // "lh" is name of key, no last hash also means no blocks created yet
    if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
      fmt.Println("No existing blockchain found")
      genesis := Genesis()
      fmt.Println("Genesis proved")

      // Add block to database
      err = txn.Set(genesis.Hash, genesis.Serialize()) // error 2
      Handle(err) // Handle error 2

      // Update last hash in database
      err = txn.Set([]byte("lh"), genesis.Hash) // error 4.1

      lastHash = genesis.Hash

      return err // error 4.1

    } else  {
      // Get item with key "lh" from database if a chain already exists
      item, err := txn.Get([]byte("lh")) // error 3
      Handle(err) //Handle error 3

      err = item.Value(func(val []byte) error { // error 4.2
        lastHash = val
        return nil
      })
      Handle(err) // Handle error 4.2

      return err // error 4.2
    }
  })

  Handle(err) // Handle error 4

  blockchain := BlockChain{lastHash, db}
  return &blockchain
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
  iter := &BlockChainIterator{chain.LastHash, chain.Database}

  return iter
}

func (iter *BlockChainIterator) Next() *Block {
  var block *Block

  err := iter.Database.View(func(txn *badger.Txn) error { // error 1
    item, err := txn.Get(iter.CurrentHash) // error 2
    Handle(err) // Handle error 2

    // Get current block with current hash and convert it from byte to block struct
    err = item.Value(func(val []byte) error { // error 1
      block = Deserialize(val)
      return nil
    })
    Handle(err) // Handle error 1

    return err // error 1
  })
  Handle(err) // Handle error 1

  // Set current hash of iterator to hash of previous block for next iteration
  iter.CurrentHash = block.PrevHash

  return block
}
