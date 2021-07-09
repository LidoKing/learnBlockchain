package blockchain

import "github.com/dgraph-io/badger"

// Get chain from database for printing info
type BlockChainIterator struct {
  CurrentHash []byte
  Database    *badger.DB
}

// Turn blockchain struct to iterator struct
func (chain *BlockChain) Iterator() *BlockChainIterator {
  iter := &BlockChainIterator{chain.LastHash, chain.Database}

  return iter
}

// Iterate chain from newest block to oldest
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
