package blockchain

import (
  "fmt"
  "github.com/dgraph-io/badger"
  "os"
  "encoding/hex"
  "runtime"
  "bytes"
  "errors"
  "crypto/ecdsa"
)

const (
  dbPath = "./tmp/blocks"
  // Check if a blockchain exists or not
  dbFile = "./tmp/blocks/MANIFEST"
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

func (chain *BlockChain) AddBlock(transactions []*Transaction) *Block {
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
  newBlock := CreateBlock(transactions, lastHash)

  // Add new block to database and update lastHash
  err = chain.Database.Update(func(txn *badger.Txn) error { // error 3
    err := txn.Set(newBlock.Hash, newBlock.Serialize()) // error 4
    Handle(err) // Handle error 4

    err = txn.Set([]byte("lh"), newBlock.Hash) // error 3

    chain.LastHash = newBlock.Hash

    return err // error 3
  })
  Handle(err) // Handle error 3

  return newBlock
}

// Only called for starting completely new chain
func InitBlockChain(address string) *BlockChain { // miner's wallet pubKeyHash
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
    cbtx := CoinbaseTx(address, genesisData)

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

  chain := BlockChain{lastHash, db}
  return &chain
}

// Only called for continuing with existing chain
func ContinueBlockChain(address string) *BlockChain { // miner's wallet pubKeyHash
  var lastHash []byte

  if DBexists() == false {
    fmt.Println("No blockchain found, call 'InitBlockChain' to create one.")
    runtime.Goexit()
  }

  // Open database
  db, err := badger.Open(badger.DefaultOptions(dbPath)) // error 1
  Handle(err) // Handle error 1

  err = db.Update(func(txn *badger.Txn) error { // error 2
    item, err := txn.Get([]byte("lh")) // error 3
    Handle(err) // Handle error 3

    // Get LastHash instance
    err = item.Value(func(val []byte) error { // error 4
      lastHash = val
      return nil
    })
    Handle(err) // Handle error 4

    return err // error 2
  })
  Handle(err) // Handle error 2

  // Set LastHash instance
  chain := BlockChain{lastHash, db}
  return &chain
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

// Get UTXOs from scratch by iterating through whole chain
func (chain *BlockChain) FindUTXO() map[string]TxOutputs {
  // Value (i.e. TxOutputs struct) contains all UTXOs in the corresponding transaction
  UTXO := make(map[string]TxOutputs)
  spentTXOs := make(map[string][]int)

  iter := chain.Iterator()

  for {
    block := iter.Next()

    for _, tx := range block.Transactions {
      txID := hex.EncodeToString(tx.ID)

      if tx.IsCoinbase() == false {
        for _, in := range tx.Inputs {
          inTxID := hex.EncodeToString(in.ID)
          spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
        }
      }

      Outputs:
      for outIdx, out := range tx.Outputs {
        if spentTXOs[txID] != nil {
          for _, spentOut := range spentTXOs[txID] {
            if spentOut == outIdx {
              continue Outputs
            }
          }
        }
        // Get/Initialize map
        outs := UTXO[txID]
        // Modify map
        outs.Outputs = append(outs.Outputs, out)
        // Set map
        UTXO[txID] = outs
      }
    }

    if len(block.PrevHash) == 0 {
      break
    }
  }
  return UTXO
}

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
  iter := bc.Iterator()

  for {
    block := iter.Next()

    for _,tx := range block.Transactions {
      if bytes.Compare(tx.ID, ID) == 0 {
        return *tx, nil
      }
    }

    if len(block.PrevHash) == 0 {
      break
    }
  }
  return Transaction{}, errors.New("Transaction does not exist")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) {
  prevTXs := make(map[string]Transaction)

  for _, in := range tx.Inputs {
    prevTX, err := bc.FindTransaction(in.ID)
    Handle(err)
    prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
  }

  tx.Sign(privateKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
  prevTXs := make(map[string]Transaction)

  for _, in := range tx.Inputs {
    prevTX, err := bc.FindTransaction(in.ID)
    Handle(err)
    prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
  }

  return tx.Verify(prevTXs)
}
