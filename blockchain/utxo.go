package blockchain

var (
  utxoPrefix = []byte("utxo-")
  prefixLength = len(utxoPrefix)
)

type UTXOSet struct {
  // For connecting to blockchain database
  Blockchain *BlockChain
}

func (u *UTXOSet) Reindex() {
  db := u.Blockchain.Database

  u.DeleteByPrefix(utxoPrefix)

  UTXO := u.Blockchain.FindUTXO()

  err := db.Update(func(txn *badger.Txn) error {
    for txId, outs := range UTXO {
      key, err := hex.DecodeString(txId)
      if err != nil {
        return err
      }
      key = append(utxoPrefix, key...)

      err = txn.Set(key, outs.Serialize())
    }
  })
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
  u.Blockchain.Database.View(func(txn *Bader.Txn) error {
    opts := badger.DefaultIteratorOptions
    // Enable key-only iteration for faster iteration
    opts.PrefetchValues := false
    it := txn.NewIterator(opts)
    defer it.Close()

    keysForDelete := make([][]byte, 0, collectSize)
    keysCollected := 0

    for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
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
