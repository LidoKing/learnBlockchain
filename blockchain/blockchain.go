package blockchain

import (
  "fmt"
  "github.com/dgraph-io/badger"
)

const dbPath = "./tmp/blocks"

type BlockChain struct{
  LastHash []*Block
  Database *badger.DB
}

func (chain *BlockChain) AddBlock(data string) {
  prevBlock := chain.Blocks[len(chain.Blocks)-1] // 'len(chain.blocks)-1 gets the index of current latest block (i.e. previous block)'
  // ^ Get old block to get hash to create new block

  new := CreateBlock(data, prevBlock.Hash)
  // ^ Create new block

  chain.Blocks = append(chain.Blocks, new)
  // ^ Append new block to chain
}

func InitBlockChain() *BlockChain {
  var lastHash []byte

  // Open database, create if one doesn't exist
  db, err := badger.Open(badger.DefaultOptions(dbPath)) // error 1
  Handle(err) // Handle error 1

  err = db.Update(func(txn *badger.Txn) error { //error 4

    // Create chain with Genesis()
    if _, err :=txn.Get([]byte("lh")); err == badger.ErrKeyNotFound { // "lh" name of key, no key also means no chain
      fmt.Println("No existing blockchain found")
      genesis := Genesis()
      fmt.Println("Genesis proved")
      err = txn.Set(genesis.Hash, genesis.Serialize()) // genesis.Hash as key, serialized as value // error 2
      Handle(err) // Handle error 2
      err = txn.Set([]byte("lh"), genesis.Hash) // error 4.1

      lastHash = genesis.Hash

      return err // error 4.1

    } else  {
      // Get item with key "lh" from database if a chain already exists
      item, err := txn.Get([]byte("lh")) // error 3
      Handle(err) //Handle error 3

      lastHash, err = item.Value() // error 4.2

      return err // error 4.2
    }
  })

  Handle(err) // Handle error 4

  blockchain := Blockchain{lastHash, db}
  return &blockchain
}
