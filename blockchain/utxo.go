package blockchain

import (
  "bytes"
  "log"
  "encoding/hex"
  "github.com/dgraph-io/badger"
)

var (
  utxoPrefix = []byte("utxo-")
  prefixLength = len(utxoPrefix)
)

type UTXOSet struct {
  // For connecting to blockchain database
  Blockchain *BlockChain
}

// Replaced iterating from scratch with just checking if UTXOs can be unlocked


// Find all UTXOs owned by specific address (pubKeyHash)
// Used for checking balance of an address only
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TxOutput {
  var UTXOs []TxOutput
  var outs TxOutputs
  db := u.Blockchain.Database

  err := db.View(func(txn *badger.Txn) error {
    it := txn.NewIterator(badger.DefaultIteratorOptions)
    defer it.Close()

    for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
      item := it.Item()
      err := item.Value(func(val []byte) error {
        outs = DeserializeOutputs(val)
        return nil
      })
      Handle(err)

      for _, out := range outs.Outputs {
        if out.IsLockedWithKey(pubKeyHash) {
          UTXOs = append(UTXOs, out)
        }
      }
    }
    return nil
  })
  Handle(err)

  return UTXOs
}

// Used for transaction
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, sendAmount int) (int, map[string][]int) {
  var outs TxOutputs
  unspentOuts := make(map[string][]int)
  accumulated := 0
  db := u.Blockchain.Database

  err := db.View(func(txn *badger.Txn) error {
    it := txn.NewIterator(badger.DefaultIteratorOptions)
    defer it.Close()

    for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
      item := it.Item()
      k := item.Key()
      k = bytes.TrimPrefix(k, utxoPrefix)
      err := item.Value(func(val []byte) error {
        outs = DeserializeOutputs(val)
        return nil
      })
      Handle(err)
      txID := hex.EncodeToString(k)

      for outIdx, out := range outs.Outputs {
        if out.IsLockedWithKey(pubKeyHash) && accumulated < sendAmount {
          accumulated += out.Value
          unspentOuts[txID] = append(unspentOuts[txID], outIdx)
        }
      }
    }
    return nil
  })
  Handle(err)

  return accumulated, unspentOuts
}

// Count txs with unspent outputs
func (u UTXOSet) CountTransactions() int {
  db := u.Blockchain.Database
  counter := 0

  err := db.View(func(txn *badger.Txn) error {
    it := txn.NewIterator(badger.DefaultIteratorOptions)
    defer it.Close()

    for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
      counter++
    }

    return nil
  })
  Handle(err)

  return counter
}

func (u *UTXOSet) DeleteByPrefix(prefix []byte) {
  deleteKeys := func(keysForDelete [][]byte) error {
    // Run Update() func first and return err if there is one
    if err := u.Blockchain.Database.Update(func(txn *badger.Txn) error {
      for _, key := range keysForDelete {
        // Run Delete() func first and return err if there is one
        if err := txn.Delete(key); err != nil {
          return err
        }
      }
      return nil
    }); err != nil {
      return err
    }
    return nil
  }

  collectSize := 100000
  u.Blockchain.Database.View(func(txn *badger.Txn) error {
    opts := badger.DefaultIteratorOptions
    // Enable key-only iteration for faster iteration
    opts.PrefetchValues = false
    it := txn.NewIterator(opts)
    defer it.Close()

    keysForDelete := make([][]byte, 0, collectSize)
    keysCollected := 0

    for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
      item := it.Item()
      key := item.KeyCopy(nil)
      keysForDelete = append(keysForDelete, key)
      keysCollected++
      // Delete batch first as number of keys has reached optimum
      if keysCollected == collectSize {
        if err := deleteKeys(keysForDelete); err != nil {
          log.Panic(err)
        }
        keysForDelete = make([][]byte, 0, collectSize)
        keysCollected = 0
      }
    }
    // Delete keys if number of keys is still less than optimum having completed the loop
    if keysCollected > 0 {
       if err := deleteKeys(keysForDelete); err != nil {
         log.Panic(err)
       }
    }
    return nil
  })
}

// Match with blocchain updates
// Prevent conflict when comparing indices in Update()
func (u *UTXOSet) Reindex() {
  db := u.Blockchain.Database

  // Clear out database 'utxo-' prefix
  // Rebuild set in the following
  u.DeleteByPrefix(utxoPrefix)

  UTXO := u.Blockchain.FindUTXO()

  // Get key of map and add prefix
  err := db.Update(func(txn *badger.Txn) error {
    for txId, outs := range UTXO {
      key, err := hex.DecodeString(txId)
      if err != nil {
        return err
      }
      key = append(utxoPrefix, key...)

      // Use 'prefixed key' as key in database
      // Key: txID with prefix, Value: serialized UTXOs in the tx
      err = txn.Set(key, outs.Serialize())
      Handle(err)
    }
    return nil
  })
  Handle(err)
}

// Remove outsputs that are originally unspent but are now referenced and used
// Add newest unspent outputs if there are any
func (u *UTXOSet) Update(block *Block) {
  var outs TxOutputs
  db := u.Blockchain.Database

  err := db.Update(func(txn *badger.Txn) error {
    for _, tx := range block.Transactions {
      if tx.IsCoinbase() == false {
        for _, in := range tx.Inputs {
          updatedOuts := TxOutputs{}
          inID := append(utxoPrefix, in.ID...) // i.e. prefixed ID of transaction referenced by input
          item, err := txn.Get(inID)
          Handle(err)
          // v: serialized TxOutputs struct which contains UTXOs of tx referenced by the input
          err = item.Value(func(val []byte) error {
            outs = DeserializeOutputs(val)
            return nil
          })
          Handle(err)

          for outIdx, out := range outs.Outputs {
            // Potential conflict solved by reindexing
            if outIdx != in.Out {
              // Add ouput to updatedOuts if it remains unspent after new transaction
              updatedOuts.Outputs = append (updatedOuts.Outputs, out)
            }
          }

          // Delete key if original UTXOs are spent with no more UTXOs in the tx
          // (i.e value of the item has nothing)
          if len(updatedOuts.Outputs) == 0 {
            if err := txn.Delete(inID); err != nil {
              log.Panic(err)
            }
          } else {
            // Update the item otherwise
            if err := txn.Set(inID, updatedOuts.Serialize()); err != nil {
              log.Panic(err)
            }
          }
        }
      }
      // Logic for coinbase tx
      newOutputs:= TxOutputs{}
      // Output must be unspent for coinbase tx
      // No checking is needed
      for _, out := range tx.Outputs {
        newOutputs.Outputs = append(newOutputs.Outputs, out)
      }

      txID := append(utxoPrefix, tx.ID...)
      if err := txn.Set(txID, newOutputs.Serialize()); err != nil {
        log.Panic(err)
      }
    }
    return nil
  })
  Handle(err)
}
