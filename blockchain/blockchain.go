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
  "path/filepath"
  "strings"
  "log"
)

const (
  dbPath = "./tmp/blocks_%s"
  genesisData = "First Transaction from Genesis"
)

type BlockChain struct{
  LastHash []byte
  Database *badger.DB
}

/*-------------------------------utils-------------------------------*/

func DBexists(path string) bool {
  if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
    return false
  }
  return true
}

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	db, err := badger.Open(retryOpts)
	return db, err
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				log.Println("database unlocked, value log truncated")
				return db, nil
			}
			log.Println("could not unlock database:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}

/*-------------------------------main-------------------------------*/

func (chain *BlockChain) AddBlock(block *Block) {
  var lastHash []byte
  var lastBlockData [] byte

  err := chain.Database.Update(func(txn *badger.Txn) error {
    if _, err := txn.Get(block.Hash); err == nil {
      return nil
    }

    // Add new block to database
    blockData := block.Serialize()
    err := txn.Set(block.Hash, blockData)
    Handle(err)

    // Get last hash
    item, err := txn.Get([]byte("lh"))
    Handle(err)
    err = item.Value(func(val []byte) error {
      lastHash = val
      return nil
    })
    Handle(err)

    // Get last block with last hash
    item, err = txn.Get(lastHash)
    Handle(err)

    err = item.Value(func(val []byte) error {
      lastBlockData = val
      return nil
    })
    Handle(err)

    lastBlock := Deserialize(lastBlockData)

    // Set new block hash as last hash if height of new block > last block
    if block.Height > lastBlock.Height {
      err = txn.Set([]byte("lh"), block.Hash)
      Handle(err)
      chain.LastHash = block.Hash
    }

    return nil
  })
  Handle(err)
}

func ( chain *BlockChain) GetBlock(blockHash []byte) (Block,error) {
  var block Block

  err := chain.Database.View(func(txn *badger.Txn) error {
    if item, err := txn.Get(blockHash); err != nil {
      return errors.New("Block is not found")
    } else {
      err = item.Value(func(val []byte) error {
        block = *Deserialize(val)
        return nil
      })
      Handle(err)
     }

    return nil
  })

  if err != nil {
    return block, err
  }

  return block, nil
}

func (chain *BlockChain) GetBlockHashes() [][]byte {
  var blocks [][]byte

  iter := chain.Iterator()

  for {
    block := iter.Next()

    blocks = append(blocks, block.Hash)

    if len(block.PrevHash) == 0 {
      break
    }
  }

  return blocks
}

func (chain *BlockChain) GetBestHeight() int {
  var lastHash []byte
  var lastBlock Block

  err := chain.Database.View(func(txn *badger.Txn) error {
    item, err := txn.Get([]byte("lh"))
    Handle(err)

    err = item.Value(func(val []byte) error {
      lastHash = val
      return nil
    })
    Handle(err)

    item, err = txn.Get(lastHash)
    Handle(err)

    // Value is serialized block
    err = item.Value(func(val []byte) error {
      lastBlock = *Deserialize(val)
      return nil
    })

    return err
  })
  Handle(err)

  return lastBlock.Height
}

func (chain *BlockChain) MineBlock(transactions []*Transaction) *Block {
  var lastHash []byte
  var lastHeight int

  for _, tx := range transactions {
		if chain.VerifyTransaction(tx) != true {
			log.Panic("Invalid Transaction")
		}
	}

  // Get lastHash from database
  err := chain.Database.View(func(txn *badger.Txn) error { // error 1
    item, err := txn.Get([]byte("lh")) // error 2
    Handle(err) // Handle error 2

    err = item.Value(func(val []byte) error { // error 3
      lastHash = val
      return nil
    })
    Handle(err) // Handle error 3

    // Use last hash to get last height
    item, err = txn.Get(lastHash) // error 4
    Handle(err) // Handle error 4

    // Value is serialized block
    err = item.Value(func(val []byte) error { // error 1
      lastBlock := Deserialize(val)
      lastHeight = lastBlock.Height
      return nil
    })

    return err // error 1
  })
  Handle(err) // Handle error 1

  // Create new block with hash retrieved from database
  newBlock := CreateBlock(transactions, lastHash, lastHeight+1)

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
func InitBlockChain(address, nodeID string) *BlockChain { // miner's wallet pubKeyHash
  path := fmt.Sprintf(dbPath, nodeID)
  if DBexists(path) {
    fmt.Println("Blockchain already exists, call 'ContinueBlockChain' instead.")
    runtime.Goexit()
  }

  var lastHash []byte

  // Open database
  opts := badger.DefaultOptions(path)

  db, err := openDB(path, opts) // error 1
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
func ContinueBlockChain(nodeID string) *BlockChain { // miner's wallet pubKeyHash
  path := fmt.Sprintf(dbPath, nodeID)
  if DBexists(path) == false {
    fmt.Println("No blockchain found, call 'InitBlockChain' to create one.")
    runtime.Goexit()
  }

  var lastHash []byte

  // Open database
  opts := badger.DefaultOptions(path)

  db, err := openDB(path, opts) // error 1
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
  if tx.IsCoinbase() {
    return true
  }

  prevTXs := make(map[string]Transaction)

  for _, in := range tx.Inputs {
    prevTX, err := bc.FindTransaction(in.ID)
    Handle(err)
    prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
  }

  return tx.Verify(prevTXs)
}
